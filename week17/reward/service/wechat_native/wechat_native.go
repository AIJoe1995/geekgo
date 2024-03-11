package wechat_native

import (
	"context"
	"errors"
	"fmt"
	accDomain "geekgo/week17/account/domain"
	accSvc "geekgo/week17/account/service"
	domain2 "geekgo/week17/payment/domain"
	"geekgo/week17/payment/service/wechat"
	"geekgo/week17/reward/domain"
	"geekgo/week17/reward/repository"
	"strconv"
	"strings"
)

// WechatNativeRewardService 实现RewardService接口
type WechatNativeRewardService struct {
	// 调用Payment模块 引入Payment的客户端
	wechatPaySvc wechat.NativePaymentService

	// 操作Reward数据库表 引入repo
	repo repository.RewardRepository

	// 打赏模块在完成支付后 调用记账模块Account(引入Account模块的客户端) 同时在打赏模块内部完成用户和平台分账
	acli accSvc.AccountService
}

func (w *WechatNativeRewardService) PreReward(ctx context.Context, r domain.Reward) (domain.CodeURL, error) {
	// 根据domain.Reward 调用wechat Prepay 构造打赏二维码 返回
	// 先查询缓存

	rid, err := w.repo.CreateReward(ctx, r)
	// 如果reward id已经存在了 怎么办 会有gorm DuplicatedKeyError 这时
	if err != nil {
		return domain.CodeURL{}, err
	}
	url, err := w.wechatPaySvc.Prepay(ctx, domain2.Payment{
		Amt: domain2.Amount{
			Currency: "CNY",
			Total:    r.Amt,
		},
		BizTradeNO:  fmt.Sprintf("reward-%d", rid),          // 业务生成 交易序号
		Description: fmt.Sprintf("打赏-%s", r.Target.BizName), // 订单的描述
	})
	if err != nil {
		return domain.CodeURL{}, err
	}
	cu := domain.CodeURL{
		Rid: rid,
		URL: url,
	}
	return cu, err

}

// GetReward 根据rid查询reward 如果reward状态是还未完成 可以去查询payment状态来更新reward状态
func (w *WechatNativeRewardService) GetReward(ctx context.Context, rid, uid int64) (domain.Reward, error) {
	r, err := w.repo.GetReward(ctx, rid)
	if err != nil {
		return domain.Reward{}, err
	}
	if r.Uid != uid {
		return domain.Reward{}, errors.New("非本人打赏的记录")
	}
	if r.Completed() {
		return r, nil
	}
	// reward记录的状态还没完结 去Payment的表里查询payment来更新reward
	pmt, err := w.wechatPaySvc.GetPayment(ctx, w.bizTradeNO(r.Id))
	if err != nil {
		return domain.Reward{}, err
	}
	switch pmt.Status {
	case domain2.PaymentStatusFailed:
		r.Status = domain.RewardStatusFailed
	case domain2.PaymentStatusInit:
		r.Status = domain.RewardStatusInit
	case domain2.PaymentStatusRefund:
		r.Status = domain.RewardStatusFailed
	case domain2.PaymentStatusSuccess:
		r.Status = domain.RewardStatusPayed
	case domain2.PaymentStatusUnknown:
		r.Status = domain.RewardStatusUnknown
	}
	err = w.repo.UpdateStatus(ctx, rid, r.Status)
	if err != nil {
		//s.l.Error("更新本地打赏状态失败",
		//	logger.Int64("rid", r.Id), logger.Error(err))
		return r, nil
	}
	return r, nil
}

func (w *WechatNativeRewardService) bizTradeNO(rid int64) string {
	return fmt.Sprintf("reward-%d", rid)
}

func (w *WechatNativeRewardService) toRid(tradeNO string) int64 {
	ridStr := strings.Split(tradeNO, "-")
	val, _ := strconv.ParseInt(ridStr[1], 10, 64)
	return val
}

// UpdateReward 可能是根据PaymentEvent 消费逻辑 来updatereward 如果update之后reward状态是Payed表示支付完成 需要分账记账
func (w *WechatNativeRewardService) UpdateReward(ctx context.Context, bizTradeNO string, status domain.RewardStatus) error {
	rid := w.toRid(bizTradeNO)
	err := w.repo.UpdateStatus(ctx, rid, status)
	if err != nil {
		return err
	}

	// 完成了支付，准备入账
	if status == domain.RewardStatusPayed {
		r, err := w.repo.GetReward(ctx, rid)
		if err != nil {
			return err
		}
		weAmt := int64(float64(r.Amt) * 0.1)

		// 同一订单之分账一次

		err = w.acli.Credit(ctx, accDomain.Credit{
			Biz:   "reward",
			BizId: rid,
			Items: []accDomain.CreditItem{
				{
					AccountType: accDomain.AccountTypeSystem,
					// 虽然可能为 0，但是也要记录出来
					Amt:      weAmt,
					Currency: "CNY",
				},
				{
					Account:     r.Uid,
					Uid:         r.Uid,
					AccountType: accDomain.AccountTypeReward,
					Amt:         r.Amt - weAmt,
					Currency:    "CNY",
				},
			},
		})

		if err != nil {
			//s.l.Error("入账失败了，快来修数据啊！！！",
			//	logger.String("biz_trade_no", bizTradeNO),
			//	logger.Error(err))
			// 做好监控和告警，这里
			return err
		}

	}
	return nil
}
