package article

import (
	"context"
	"geekgo/week8/webook/internal/domain"
	"geekgo/week8/webook/internal/repository/dao/article"
)

type ArticleRepository struct {
	dao article.ArticleDAO
}

func (r *ArticleRepository) Save(ctx context.Context, art domain.Article) (int64, error) {
	// repo 调用dao保存到制作库 如何判断是新创建还是修改后保存
	// 根据前端传来的文章id

	if art.Id > 0 {
		err := r.dao.UpdateById(ctx, r.domainToEntity(art))
		return art.Id, err
	}
	aid, err := r.dao.Insert(ctx, r.domainToEntity(art))
	return aid, err
}

func NewArticleRepository(dao article.ArticleDAO) *ArticleRepository {
	return &ArticleRepository{
		dao: dao,
	}
}

func (r *ArticleRepository) domainToEntity(art domain.Article) article.Article {
	return article.Article{
		Id:       art.Id,
		Title:    art.Title,
		Content:  art.Content,
		AuthorId: art.Author.Id,
		Status:   art.Status.ToUint8(),
	}
}

func (r *ArticleRepository) entityToDomain(art article.Article) domain.Article {
	return domain.Article{
		Id:      art.Id,
		Title:   art.Title,
		Content: art.Content,
		Author: domain.Author{
			Id: art.AuthorId,
		},
		Status: domain.ArticleStatus(art.Status),
	}
}

func (r *ArticleRepository) Publish(ctx context.Context, art domain.Article) (int64, error) {
	// 文章发表 需要调用dao 来先到制作库 再同步到线上库 在数据库那里开启事务

	return r.dao.Sync(ctx, r.domainToEntity(art))
}

func (r *ArticleRepository) GetPublishedByIds(ctx context.Context, ids []int64) ([]domain.Article, error) {
	var res []domain.Article
	res_e, err := r.dao.GetPublishedByIds(ctx, ids)
	if err != nil {
		return []domain.Article{}, err
	}
	for _, art := range res_e {
		res = append(res, domain.Article{
			Id:      art.Id,
			Title:   art.Title,
			Content: art.Content,
			Author: domain.Author{
				Id: art.AuthorId,
			},
			Status: domain.ArticleStatus(art.Status),
		})
	}
	return res, nil
}

func (r *ArticleRepository) GetPublishedById(ctx context.Context, id int64) (domain.Article, error) {
	art, err := r.dao.GetPublishedById(ctx, id)

	return domain.Article{
		Id:      art.Id,
		Title:   art.Title,
		Content: art.Content,
		Author: domain.Author{
			Id: art.AuthorId,
		},
		Status: domain.ArticleStatus(art.Status),
	}, err
}

func (r *ArticleRepository) SyncStatus(ctx context.Context, art domain.Article, status uint8) error {
	return r.dao.SyncStatus(ctx, art.Author.Id, art.Id, status)
}
