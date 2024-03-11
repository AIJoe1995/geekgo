package wechat

import (
	"context"
	"errors"
	"fmt"
	"geekgo/week17/payment/domain"
	"geekgo/week17/payment/events"
	"geekgo/week17/payment/repository"
	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments/native"
	"time"

	"github.com/wechatpay-apiv3/wechatpay-go/services/payments"
)

var _ PaymentService = (*NativePaymentService)(nil)

var errUnknownTransactionState = errors.New("未知的微信事务状态")

// 生成二维码 扫码支付 的接口, 分为Prepay生成CodeUrl二维码(数据库建支付订单), 和处理支付回调结果(支付订单状态更新)
// 处理回调结果，微信会调用注册好的路由，在web模块里提供了callback路由
type NativePaymentService struct {
	svc       *native.NativeApiService // 调用wechat第三方接口
	appID     string
	mchID     string
	notifyURL string // 回调路由
	repo      repository.PaymentRepository
	producer  events.Producer

	// 在微信 native 里面，分别是
	// SUCCESS：支付成功
	// REFUND：转入退款
	// NOTPAY：未支付
	// CLOSED：已关闭
	// REVOKED：已撤销（付款码支付）
	// USERPAYING：用户支付中（付款码支付）
	// PAYERROR：支付失败(其他原因，如银行返回失败)
	nativeCBTypeToStatus map[string]domain.PaymentStatus
}

func NewNativePaymentService(svc *native.NativeApiService,
	repo repository.PaymentRepository,
	producer events.Producer,
	appid, mchid string) *NativePaymentService {
	return &NativePaymentService{
		repo:  repo,
		svc:   svc,
		appID: appid,
		mchID: mchid,
		// 一般来说，这个都是固定的，基本不会变的
		notifyURL: "http://wechat.meoying.com/pay/callback",
		nativeCBTypeToStatus: map[string]domain.PaymentStatus{
			"SUCCESS":  domain.PaymentStatusSuccess,
			"PAYERROR": domain.PaymentStatusFailed,
			"NOTPAY":   domain.PaymentStatusInit,
			"CLOSED":   domain.PaymentStatusFailed,
			"REVOKED":  domain.PaymentStatusFailed,
			"REFUND":   domain.PaymentStatusRefund,
			// 其它状态你都可以加
		},
		producer: producer,
	}
}

func (w *NativePaymentService) Prepay(ctx context.Context, pmt domain.Payment) (string, error) {
	err := w.repo.AddPayment(ctx, pmt)
	if err != nil {
		return "", err
	}
	resp, _, err := w.svc.Prepay(ctx, native.PrepayRequest{
		Appid:       core.String(w.appID),
		Mchid:       core.String(w.mchID),
		Description: core.String(pmt.Description),
		OutTradeNo:  core.String(pmt.BizTradeNO),
		TimeExpire:  core.Time(time.Now().Add(time.Minute * 30)),
		NotifyUrl:   core.String(w.notifyURL),
		Amount: &native.Amount{
			Currency: core.String(pmt.Amt.Currency),
			Total:    core.Int64(pmt.Amt.Total),
		},
	})
	if err != nil {
		return "", err
	}
	return *resp.CodeUrl, nil

}

// HandleCallback 根据微信的回调信息 更新交易结果
func (w *NativePaymentService) HandleCallback(ctx context.Context, txn *payments.Transaction) error {
	//
	return w.updateByTxn(ctx, txn)
}

func (w *NativePaymentService) updateByTxn(ctx context.Context, txn *payments.Transaction) error {
	status, ok := w.nativeCBTypeToStatus[*txn.TradeState]
	if !ok {
		return fmt.Errorf("%w, %s", errUnknownTransactionState, *txn.TradeState)
	}
	pmt := domain.Payment{
		BizTradeNO: *txn.OutTradeNo,
		TxnID:      *txn.TransactionId,
		Status:     status,
	}
	err := w.repo.UpdatePayment(ctx, pmt)
	if err != nil {
		return err
	}
	// 把PaymentEvent 发送到消息中间件
	err1 := w.producer.ProducePaymentEvent(ctx, events.PaymentEvent{
		BizTradeNO: pmt.BizTradeNO,
		Status:     pmt.Status.AsUint8(),
	})
	if err1 != nil {
		// 记录日志 监控告警
		// 保存到本地的表里 进行发送PaymentEvent的重试
		err2 := w.repo.CreatePaymentEvent(ctx, domain.PaymentEvent{
			BizTradeNO: pmt.BizTradeNO,
			Status:     pmt.Status.AsUint8(),
		})
		// 如果是数据BizTradeNO唯一索引冲突 不需要处理 其他的需要重试 确保发到了mysql
		if err2 != nil {
			// 记录到本地 之后补偿
		}

	}
	return nil
}

// SyncWechatInfo 提供给其他模块调用 如定时任务 查询微信支付结果 核对订单状态 更新数据库订单状态
func (n *NativePaymentService) SyncWechatInfo(ctx context.Context, bizTradeNO string) error {
	panic("unimplemented")
}

// FindExpiredPayment 找到过期的支付 可能需要核对
func (n *NativePaymentService) FindExpiredPayment(ctx context.Context, offset, limit int, t time.Time) ([]domain.Payment, error) {
	panic("unimplemented")
}

// 根据Payment记录的Id 找到数据库表的payment记录
func (n *NativePaymentService) GetPayment(ctx context.Context, bizTradeId string) (domain.Payment, error) {
	panic("unimplemented")
}
