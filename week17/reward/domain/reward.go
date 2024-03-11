package domain

type Reward struct {
	Id     int64
	Uid    int64  // 谁打赏
	Target Target // 打赏什么
	Amt    int64
	Status RewardStatus // 打赏状态 成功 失败
}

type Target struct {
	Biz     string
	BizId   int64
	BizName string
	Uid     int64 // 接收打赏的用户
}

type RewardStatus uint8

func (r RewardStatus) AsUint8() uint8 {
	return uint8(r)
}

const (
	RewardStatusUnknown = iota
	RewardStatusInit
	RewardStatusPayed
	RewardStatusFailed
)

type CodeURL struct {
	Rid int64
	URL string
}

func (r Reward) Completed() bool {
	return r.Status == RewardStatusFailed || r.Status == RewardStatusPayed
}
