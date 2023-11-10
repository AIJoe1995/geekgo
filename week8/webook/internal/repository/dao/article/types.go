package article

import "context"

// 抽象出ArticleDAO 提供gorm实现和mongodb实现 制作库线上库 同库不同表方案

type ArticleDAO interface {
	UpdateById(ctx context.Context, art Article) error
	Insert(ctx context.Context, art Article) (int64, error)
	Sync(ctx context.Context, art Article) (int64, error)
	GetPublishedById(ctx context.Context, id int64) (PublishedArticle, error)
	SyncStatus(ctx context.Context, author_id int64, article_id int64, status uint8) error
	GetPublishedByIds(ctx context.Context, ids []int64) ([]PublishedArticle, error)
}

//// 提供gorm实现 和 mongodb实现 在这里抽象出dao article的接口
//
//type AuthorArticle interface {
//	Create(ctx context.Context, art Article) (int64, error)
//	Update(ctx context.Context, art Article) error
//}
//
//type ReaderArticle interface {
//	Save(ctx context.Context, art PublishedArticle)
//}
