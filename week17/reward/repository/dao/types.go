package dao

type Reward struct {
	Id      int64  `gorm:"primaryKey,autoIncrement" bson:"id,omitempty"`
	Biz     string `gorm:"index:biz_biz_id"`
	BizId   int64  `gorm:"index:biz_biz_id"`
	BizName string
	// 被打赏的人
	TargetUid int64 `gorm:"index"`

	// 直接采用 RewardStatus 的取值
	Status uint8
	// 打赏的人
	Uid    int64
	Amount int64
	Ctime  int64
	Utime  int64
}
