package dao

// Basic imports
import (
	"context"
	"geekgo/week8/webook/internal/repository/dao/article"
	ijwt "geekgo/week8/webook/internal/web/jwt"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"testing"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type InteractiveDAOTestSuite struct {
	suite.Suite
	server *gin.Engine
	db     *gorm.DB
}

func initTestDB() *gorm.DB {
	db, err := gorm.Open(mysql.Open("root:1234@tcp(localhost:3306)/webook?parseTime=true&&charset=utf8"), &gorm.Config{
		//DryRun: true,
	})
	if err != nil {
		panic(err)
	}
	InitTable(db)
	return db
}

func (suite *InteractiveDAOTestSuite) SetupSuite() {
	suite.server = gin.Default()
	suite.server.Use(func(ctx *gin.Context) {
		ctx.Set("user", ijwt.UserClaims{Id: 1})
		ctx.Next()
	})
	suite.db = initTestDB()

}

// Make sure that db
// before each test
func (suite *InteractiveDAOTestSuite) SetupTest() {
	// 向Interactive 插入数据
	data := []*Interactive{}
	arts := []*article.PublishedArticle{}
	for i := int64(1); i <= 100; i++ {
		data = append(data, &Interactive{
			Biz:        "article",
			BizId:      i,
			LikeCnt:    i,
			CollectCnt: i,
			ReadCnt:    i,
		})
		arts = append(arts, &article.PublishedArticle{
			Id: i,
		})
	}

	suite.db.Model(&Interactive{}).Create(data)
	suite.db.Model(&article.PublishedArticle{}).Create(arts)

}

func (suite *InteractiveDAOTestSuite) TearDownTest() {
	err := suite.db.Exec("TRUNCATE TABLE `interactives`").Error
	assert.NoError(suite.T(), err)
	err = suite.db.Exec("TRUNCATE TABLE `published_articles`").Error
	assert.NoError(suite.T(), err)
}

// All methods that begin with "Test" are run as tests within a
// suite.
func (suite *InteractiveDAOTestSuite) TestTopNLiked() {
	t := suite.T()
	testCases := []struct {
		name    string
		before  func(t *testing.T)
		after   func(t *testing.T)
		topn    int64
		biz     string
		wantRes []Interactive
	}{
		{
			name: "select top 5 liked article",
			before: func(t *testing.T) {

			},
			after: func(t *testing.T) {},
			topn:  int64(5),
			biz:   "article",
			wantRes: []Interactive{
				{BizId: 100, LikeCnt: 100},
				{BizId: 99, LikeCnt: 99},
				{BizId: 98, LikeCnt: 98},
				{BizId: 97, LikeCnt: 97},
				{BizId: 96, LikeCnt: 96},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before(t)
			dao := GORMInteractiveDAO{db: suite.db}
			ctx := context.Background()
			data, err := dao.TopNLike(ctx, tc.biz, tc.topn)
			require.NoError(t, err)
			for i, row := range data {
				assert.Equal(t, tc.wantRes[i].BizId, row.BizId)
				assert.Equal(t, tc.wantRes[i].LikeCnt, row.LikeCnt)
			}
			tc.after(t)
		})
	}

}

func (suite *InteractiveDAOTestSuite) TestGetPublishedByIds() {
	t := suite.T()
	testCases := []struct {
		name    string
		ids     []int64
		before  func(t *testing.T)
		after   func(t *testing.T)
		biz     string
		wantRes []article.PublishedArticle
	}{
		{
			name: "select top 5 liked article",
			ids:  []int64{1, 3, 5, 8},
			before: func(t *testing.T) {

			},
			after: func(t *testing.T) {},
			biz:   "article",
			wantRes: []article.PublishedArticle{
				{Id: 1},
				{Id: 3},
				{Id: 5},
				{Id: 8},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before(t)
			dao := article.NewGORMArticleDAO(suite.db)
			ctx := context.Background()
			data, err := dao.GetPublishedByIds(ctx, tc.ids)
			require.NoError(t, err)
			for i, row := range data {
				assert.Equal(t, tc.wantRes[i].Id, row.Id)
			}
			tc.after(t)
		})
	}
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestInteractiveDAOTestSuite(t *testing.T) {
	suite.Run(t, new(InteractiveDAOTestSuite))
}
