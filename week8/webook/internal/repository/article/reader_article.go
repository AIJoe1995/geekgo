package article

import (
	"context"
	"geekgo/week8/webook/internal/domain"
	"geekgo/week8/webook/internal/repository/dao/article"
)

type ArticleReaderRepository struct {
	dao *article.ReaderDAO
}

func (r *ArticleReaderRepository) Save(ctx context.Context, art domain.Article) error {
	return r.dao.Save(ctx, r.domainToEntity(art))
}

func NewArticleReaderRepository(dao *article.ReaderDAO) *ArticleReaderRepository {
	return &ArticleReaderRepository{
		dao: dao,
	}
}

func (r *ArticleReaderRepository) domainToEntity(art domain.Article) article.PublishedArticle {
	return article.PublishedArticle{
		Id:       art.Id,
		Title:    art.Title,
		Content:  art.Content,
		AuthorId: art.Author.Id,
	}
}
