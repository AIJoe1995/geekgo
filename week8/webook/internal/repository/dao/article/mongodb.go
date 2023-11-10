package article

import (
	"context"
	"errors"
	"github.com/bwmarrin/snowflake"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

type MongodbArticleDAO struct {
	col     *mongo.Collection
	livecol *mongo.Collection
	node    *snowflake.Node
}

func (m *MongodbArticleDAO) UpdateById(ctx context.Context, art Article) error {
	now := time.Now().UnixMilli()
	filter := bson.M{"id": art.Id, "author_id": art.AuthorId}
	update := bson.D{bson.E{"$set", bson.M{
		"title":   art.Title,
		"content": art.Content,
		"status":  art.Status,
		"utime":   now,
	}}}
	res, err := m.col.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if res.ModifiedCount == 0 {
		return errors.New("数据更新失败")
	}
	return nil
}

func (m *MongodbArticleDAO) Insert(ctx context.Context, art Article) (int64, error) {
	//
	id := m.node.Generate().Int64()
	art.Id = id
	_, err := m.col.InsertOne(ctx, art)
	//res.InsertedID
	return id, err
}

func (m *MongodbArticleDAO) Sync(ctx context.Context, art Article) (int64, error) {
	//
	now := time.Now().UnixMilli()
	var id int64
	var err error

	if art.Id == 0 {
		id, err = m.Insert(ctx, art)
	} else {
		err = m.UpdateById(ctx, art)
	}
	if err != nil {
		return 0, err
	}

	//update := bson.E{"$set", art}
	upsert := bson.D{bson.E{Key: "$set", Value: art},
		bson.E{Key: "$setOnInsert",
			Value: bson.D{bson.E{Key: "ctime", Value: now}}}}

	filter := bson.D{bson.E{Key: "id", Value: art.Id},
		bson.E{Key: "author_id", Value: art.AuthorId}}
	art.Utime = now
	_, err = m.livecol.UpdateOne(ctx, filter, upsert)
	return id, err

}

func (m *MongodbArticleDAO) GetPublishedById(ctx context.Context, id int64) (PublishedArticle, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MongodbArticleDAO) SyncStatus(ctx context.Context, author_id int64, article_id int64, status uint8) error {
	filter := bson.D{bson.E{Key: "id", Value: article_id},
		bson.E{Key: "author_id", Value: author_id}}
	sets := bson.D{bson.E{Key: "$set",
		Value: bson.D{bson.E{Key: "status", Value: status}}}}
	res, err := m.col.UpdateOne(ctx, filter, sets)
	if err != nil {
		return err
	}
	if res.ModifiedCount != 1 {
		return ErrPossibleIncorrectAuthor
	}
	return nil
}

func NewMongodbArticleDAO(col *mongo.Collection, livecol *mongo.Collection, node *snowflake.Node) *MongodbArticleDAO {
	return &MongodbArticleDAO{
		col:     col,
		livecol: livecol,
		node:    node,
	}
}

// mongodb创建索引
func InitCollections(db *mongo.Database) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	index := []mongo.IndexModel{
		{
			Keys:    bson.D{bson.E{Key: "id", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{bson.E{Key: "author_id", Value: 1},
				bson.E{Key: "ctime", Value: 1},
			},
			Options: options.Index(),
		},
	}
	_, err := db.Collection("articles").Indexes().
		CreateMany(ctx, index)
	if err != nil {
		return err
	}
	_, err = db.Collection("published_articles").Indexes().
		CreateMany(ctx, index)
	return err
}
