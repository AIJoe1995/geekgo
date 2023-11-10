package article

import (
	"context"
	"geekgo/week8/webook/internal/domain"
	"geekgo/week8/webook/internal/repository/dao/article"
)

type ArticleAuthorRepository struct {
	dao *article.AuthorDAO
}

func NewArticleAuthorRepository(dao *article.AuthorDAO) *ArticleAuthorRepository {
	return &ArticleAuthorRepository{
		dao: dao,
	}
}

func (repo *ArticleAuthorRepository) Create(ctx context.Context, art domain.Article) (int64, error) {
	return repo.dao.Create(ctx, repo.domainToEntity(art))
}

func (repo *ArticleAuthorRepository) Update(ctx context.Context, art domain.Article) error {
	return repo.dao.Update(ctx, repo.domainToEntity(art))
}

func (repo *ArticleAuthorRepository) domainToEntity(art domain.Article) article.Article {
	return article.Article{
		Id:       art.Id,
		Title:    art.Title,
		Content:  art.Content,
		AuthorId: art.Author.Id,
	}
}
