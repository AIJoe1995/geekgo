package repository

import (
	"context"
	"geekgo/week11/domain"
	"geekgo/week11/repository/dao"
	"time"
)

var _ CronJobRepository = (*cronJobRepository)(nil)

type CronJobRepository interface {
	AddJob(ctx context.Context, job domain.CronJob) error
	Preempt(ctx context.Context) (domain.CronJob, error)
	UpdateNextTime(ctx context.Context, id int64, t time.Time) error
	UpdateUtime(ctx context.Context, id int64) error
	Release(ctx context.Context, id int64) error
	EndJob(ctx context.Context, id int64) error
}

type cronJobRepository struct {
	dao dao.CronJobDAO
}

func NewCronJobRepository(dao dao.CronJobDAO) CronJobRepository {
	return &cronJobRepository{dao: dao}
}

func (c *cronJobRepository) EndJob(ctx context.Context, id int64) error {
	return c.dao.EndJob(ctx, id)
}

func (c *cronJobRepository) domainToEntity(j domain.CronJob) dao.Job {
	return dao.Job{
		Id:         j.Id,
		Name:       j.Name,
		Expression: j.Expression,
		Cfg:        j.Cfg,
		Executor:   j.Executor,
		NextTime:   j.NextTime.UnixMilli(),
	}
}

func (c *cronJobRepository) entityToDomain(j dao.Job) domain.CronJob {
	return domain.CronJob{
		Id:         j.Id,
		Name:       j.Name,
		Expression: j.Expression,
		Cfg:        j.Cfg,
		Executor:   j.Executor,
		NextTime:   time.UnixMilli(j.NextTime),
	}
}

func (c *cronJobRepository) AddJob(ctx context.Context, job domain.CronJob) error {
	return c.dao.Insert(ctx, c.domainToEntity(job))
}

func (c *cronJobRepository) Preempt(ctx context.Context) (domain.CronJob, error) {
	j, err := c.dao.Preempt(ctx)
	if err != nil {
		return domain.CronJob{}, err
	}
	return c.entityToDomain(j), nil
}

func (c *cronJobRepository) UpdateNextTime(ctx context.Context, id int64, t time.Time) error {
	return c.dao.UpdateNextTime(ctx, id, t)
}

func (c *cronJobRepository) UpdateUtime(ctx context.Context, id int64) error {
	return c.dao.UpdateUtime(ctx, id)
}

func (c *cronJobRepository) Release(ctx context.Context, id int64) error {
	return c.dao.Release(ctx, id)
}
