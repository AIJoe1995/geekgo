package fixer

import (
	"context"
	"geekgo/week13/migrator"
	"geekgo/week13/migrator/events"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// fix
// base有 target无 查出base记录 insert到target
// base与target不一样 查出base记录 update到target
// upsert
// target有 base无 删除target

type OverrideFixer[T migrator.Entity] struct {
	base    *gorm.DB
	target  *gorm.DB
	columns []string
}

func NewOverrideFixer[T migrator.Entity](base *gorm.DB, target *gorm.DB) (*OverrideFixer[T], error) {
	var t T
	rows, err := target.Model(&t).Limit(1).Rows()
	if err != nil {
		return nil, err
	}
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	return &OverrideFixer[T]{
		base:    base,
		target:  target,
		columns: columns,
	}, nil
}

func (o *OverrideFixer[T]) Fix(ctx context.Context, id int64, typ string) error {
	var src T
	// 找出数据
	switch typ {
	case events.InconsistentEventTypeTargetMissing, events.InconsistentEventTypeNotEqual:
		err := o.base.WithContext(ctx).Where("id = ?", id).
			First(&src).Error
		switch err {
		// 找到了数据
		case nil:
			return o.target.Clauses(&clause.OnConflict{
				// 我们需要 Entity 告诉我们，修复哪些数据
				DoUpdates: clause.AssignmentColumns(o.columns),
			}).Create(&src).Error
		case gorm.ErrRecordNotFound:
			// base没有这条记录 从target删除
			return o.target.Delete("id = ?", id).Error
		default:
			return err
		}
	case events.InconsistentEventTypeBaseMissing:
		return o.target.Delete("id = ?", id).Error
	}
	return nil
}