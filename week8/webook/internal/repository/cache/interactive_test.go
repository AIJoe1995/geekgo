package cache

import (
	"context"
	_ "embed"
	"fmt"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"testing"
)

type InteractiveCacheTestSuite struct {
	suite.Suite
	client redis.Cmdable
	cache  InteractiveCacheV1
}

func (s *InteractiveCacheTestSuite) SetupSuite() {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	s.client = client
	s.cache = NewInteractiveCacheV1(client)
}

func (s *InteractiveCacheTestSuite) SetupTest() {
	//likeCntZSKey like_cnt
	biz := "article"
	key := fmt.Sprintf("interactive:%s:%s", biz, "like_cnt")

	// 向redis中插入数据
	for i := 1; i <= 100; i++ {
		bizId := i
		like_cnt := i * 2
		batchNo := i % 10
		batchKey := fmt.Sprintf("%s:%d", key, batchNo)
		s.client.ZAdd(
			context.Background(),
			batchKey,
			redis.Z{Score: float64(like_cnt),
				Member: bizId,
			},
		)
	}
}

func (s *InteractiveCacheTestSuite) TearDownTest() {
	//likeCntZSKey like_cnt
	biz := "article"
	key := fmt.Sprintf("interactive:%s:%s", biz, "like_cnt")
	for i := 0; i < 10; i++ {
		batchKey := fmt.Sprintf("%s:%d", key, i)
		s.client.Del(context.Background(), batchKey)
	}
}

func (s *InteractiveCacheTestSuite) TestZADD() {
	t := s.T()
	testCases := []struct {
		name   string
		before func(t *testing.T)
		after  func(t *testing.T)
	}{
		{
			name:   "check setupsuite data preparation",
			before: func(t *testing.T) {},
			after: func(t *testing.T) {
				// 从redis 取出数据
				biz := "article"
				key := fmt.Sprintf("interactive:%s:%s", biz, "like_cnt")
				for i := 1; i < 100; i++ {
					bizId := i
					batchKey := fmt.Sprintf("%s:%d", key, i%10)
					res, err := s.client.ZScore(context.Background(), batchKey, fmt.Sprintf("%d", bizId)).Result()
					require.NoError(t, err)
					fmt.Println(res)
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before(t)
			tc.after(t)
		})
	}
}

func (s *InteractiveCacheTestSuite) TestZINCRBY() {
	t := s.T()
	testCases := []struct {
		name   string
		before func(t *testing.T)
		after  func(t *testing.T)
	}{
		{
			name: "check setupsuite zincrby",
			before: func(t *testing.T) {
				// 从redis 取出数据
				biz := "article"
				key := fmt.Sprintf("interactive:%s:%s", biz, "like_cnt")
				for i := 1; i < 100; i++ {
					bizId := i
					batchKey := fmt.Sprintf("%s:%d", key, i%10)
					_, err := s.client.ZIncrBy(context.Background(), batchKey, 1, fmt.Sprintf("%d", bizId)).Result()
					require.NoError(t, err)
					//fmt.Println(res)
				}
			},
			after: func(t *testing.T) {
				// 从redis 取出数据
				biz := "article"
				key := fmt.Sprintf("interactive:%s:%s", biz, "like_cnt")
				for i := 1; i < 100; i++ {
					bizId := i
					batchKey := fmt.Sprintf("%s:%d", key, i%10)
					res, err := s.client.ZScore(context.Background(), batchKey, fmt.Sprintf("%d", bizId)).Result()
					require.NoError(t, err)
					assert.Equal(t, float64(2*i+1), res)
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before(t)
			tc.after(t)
		})
	}
}

func (s *InteractiveCacheTestSuite) TestZRANGE() {
	t := s.T()
	testCases := []struct {
		name   string
		before func(t *testing.T)
		after  func(t *testing.T)
	}{
		{
			name: "check setupsuite zrange",
			before: func(t *testing.T) {
			},
			after: func(t *testing.T) {
				// 从redis 取出数据
				biz := "article"
				key := fmt.Sprintf("interactive:%s:%s", biz, "like_cnt")
				for i := 0; i < 10; i++ {
					//bizId := i
					batchKey := fmt.Sprintf("%s:%d", key, i%10)

					res, err := s.client.ZRevRange(context.Background(), batchKey, 0, 0).Result()
					require.NoError(t, err)
					fmt.Println(res)
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before(t)
			tc.after(t)
		})
	}
}

//var (
//	//go:embed lua/interative_incr_cnt_zs.lua
//	luaIncrCntZS string
//)

func (s *InteractiveCacheTestSuite) TestLuaZINCRBY() {
	t := s.T()
	testCases := []struct {
		name   string
		before func(t *testing.T)
		after  func(t *testing.T)
	}{
		{
			name: "check setupsuite lua zincrby",
			before: func(t *testing.T) {
				// 从redis 取出数据
				biz := "article"
				key := fmt.Sprintf("interactive:%s:%s", biz, "like_cnt")
				for i := 1; i < 100; i++ {
					bizId := i
					batchKey := fmt.Sprintf("%s:%d", key, i%10)
					//_, err := s.client.ZIncrBy(context.Background(), batchKey, 1, fmt.Sprintf("%d", bizId)).Result()
					err := s.client.Eval(context.Background(), luaIncrCntZS, []string{batchKey}, bizId, 1).Err()
					require.NoError(t, err)
					//fmt.Println(res)
				}
			},
			after: func(t *testing.T) {
				// 从redis 取出数据
				biz := "article"
				key := fmt.Sprintf("interactive:%s:%s", biz, "like_cnt")
				for i := 1; i < 100; i++ {
					bizId := i
					batchKey := fmt.Sprintf("%s:%d", key, i%10)
					res, err := s.client.ZScore(context.Background(), batchKey, fmt.Sprintf("%d", bizId)).Result()
					require.NoError(t, err)
					assert.Equal(t, float64(2*i+1), res)
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before(t)
			tc.after(t)
		})
	}
}

func (s *InteractiveCacheTestSuite) TestLuaZRANGE() {
	t := s.T()
	testCases := []struct {
		name   string
		before func(t *testing.T)
		after  func(t *testing.T)
	}{
		{
			name: "check setupsuite lua zrange",
			before: func(t *testing.T) {
			},
			after: func(t *testing.T) {
				// 从redis 取出数据
				biz := "article"
				key := fmt.Sprintf("interactive:%s:%s", biz, "like_cnt")
				for i := 0; i < 10; i++ {
					//bizId := i
					batchKey := fmt.Sprintf("%s:%d", key, i%10)

					res := s.client.Eval(context.Background(), luaTopNZS, []string{batchKey}, 1)
					res_i := res.Val().([]interface{})
					keyids := []string{}
					for _, k := range res_i {
						keyids = append(keyids, k.(string))
					}

					// 怎么把member和分数一起返回
					//require.NoError(t, err)
					//fmt.Println(res) // []interface {} ([]interface {}{"92"})
					fmt.Println(keyids)
					//assert.Equal(t, keyids, []string{string(i)})

				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before(t)
			tc.after(t)
		})
	}
}

func (s *InteractiveCacheTestSuite) TestTopNLiked() {
	t := s.T()
	testCases := []struct {
		name   string
		before func(t *testing.T)
		after  func(t *testing.T)
	}{
		{
			name: "test topn liked",
			before: func(t *testing.T) {
			},
			after: func(t *testing.T) {
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before(t)
			cache := NewInteractiveCacheV1(s.client)
			res, err := cache.TopNLike(context.Background(), "article", 5)
			//
			require.NoError(t, err)
			for i, e := range res {
				// i = 0, 应该是 artId 100 score 200
				// i = 1 artid 99 socre 198
				assert.Equal(t, int64(100-i), e.ArtId)
				assert.Equal(t, int64((100-i)*2), e.LikeCnt)
			}
			tc.after(t)
		})
	}
}

func TestInteractiveCacheTestSuite(t *testing.T) {
	suite.Run(t, new(InteractiveCacheTestSuite))
}
