package dao

import (
	"context"
	"gorm.io/gorm"
	"time"
)

var _ CronJobDAO = (*cronJobDAO)(nil)

type CronJobDAO interface {
	Insert(ctx context.Context, j Job) error
	UpdateUtime(ctx context.Context, id int64) error
	Release(ctx context.Context, id int64) error
	UpdateNextTime(ctx context.Context, id int64, t time.Time) error
	Preempt(ctx context.Context) (Job, error)
	EndJob(ctx context.Context, id int64) error
}

type cronJobDAO struct {
	db *gorm.DB
}

func NewCronJobDAO(db *gorm.DB) CronJobDAO {
	return &cronJobDAO{db: db}
}

func (dao *cronJobDAO) EndJob(ctx context.Context, id int64) error {
	return dao.db.WithContext(ctx).Model(&Job{}).Where("id=?", id).Updates(
		map[string]interface{}{
			"utime":  time.Now().UnixMilli(),
			"status": jobStatusEnd},
	).Error

}

func (dao *cronJobDAO) Insert(ctx context.Context, j Job) error {
	now := time.Now().UnixMilli()
	j.Ctime = now
	j.Utime = now
	return dao.db.WithContext(ctx).Create(&j).Error
}

func (dao *cronJobDAO) UpdateUtime(ctx context.Context, id int64) error {
	now := time.Now().UnixMilli()
	return dao.db.WithContext(ctx).Model(&Job{}).
		Where("id = ?", id).
		Update("utime", now).Error
}

func (dao *cronJobDAO) Release(ctx context.Context, id int64) error {
	return dao.db.WithContext(ctx).Model(&Job{}).Where("id=?", id).Updates(
		map[string]interface{}{
			"status": jobStatusWaiting,
			"utime":  time.Now().UnixMilli(),
		}).Error
}

func (dao *cronJobDAO) UpdateNextTime(ctx context.Context, id int64, t time.Time) error {
	return dao.db.WithContext(ctx).Model(&Job{}).Where("id=?", id).Updates(
		map[string]interface{}{
			"next_time": t.UnixMilli(),
			"utime":     time.Now().UnixMilli(),
		}).Error
}

func (dao *cronJobDAO) Preempt(ctx context.Context) (Job, error) {
	// 循环 查找可执行的任务 直到找到一个可执行任务返回这个任务
	db := dao.db.WithContext(ctx)
	for {
		now := time.Now().UnixMilli()
		var j Job
		// 到执行时间 状态是等待执行 或者在执行中但是healthCheck不通过 执行已经终止的任务
		err := db.Where("next_time < ? and status = ?", now, jobStatusWaiting).
			Or("utime < ? and status = ?",
				time.Now().Add(-1*time.Minute).UnixMilli(), jobStatusRunning).First(&j).Error
		if err != nil {
			return Job{}, err
		}
		// 找到了记录 乐观锁 更新version
		res := db.Model(&Job{}).Where("id=? and version=?", j.Id, j.Version).Updates(
			map[string]interface{}{
				"utime":   time.Now().UnixMilli(),
				"version": gorm.Expr("version + 1"),
				"status":  jobStatusRunning,
			})
		// 如果更新不成功  如果不存在这条记录 会报ErrRecordNotFound 这时应该进入下一轮 如果是其他错误 可以直接返回
		if res.Error == gorm.ErrRecordNotFound {
			continue
		}
		if res.Error != nil {
			return Job{}, err
		}
		if res.RowsAffected == 1 {
			return j, nil
		}
	}

}

type Job struct {
	Id         int64  `gorm:"primaryKey,autoIncrement"`
	Name       string `gorm:"type:varchar(256);unique"`
	Executor   string
	Cfg        string
	Expression string
	Version    int64
	NextTime   int64 `gorm:"index"`
	Status     int
	Ctime      int64
	Utime      int64
}

const (
	jobStatusWaiting = iota
	jobStatusRunning
	jobStatusEnd
)
