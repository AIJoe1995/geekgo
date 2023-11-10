package web

import (
	"fmt"
	"geekgo/week8/webook/internal/domain"
	"geekgo/week8/webook/internal/service"
	ijwt "geekgo/week8/webook/internal/web/jwt"
	"github.com/gin-gonic/gin"
	"golang.org/x/sync/errgroup"
	"net/http"
	"strconv"
)

type ArticleHandler struct {
	articleSvc *service.ArticleService
	interSvc   service.InteractiveService
	biz        string
}

func NewArticleHandler(articleSvc *service.ArticleService, interSvc service.InteractiveService) *ArticleHandler {
	return &ArticleHandler{
		articleSvc: articleSvc,
		interSvc:   interSvc,
		biz:        "article",
	}
}

func (ah *ArticleHandler) RegisterRoutes(server *gin.Engine) {
	ag := server.Group("/articles")
	ag.POST("/topNliked", ah.TopNLike)
	ag.POST("/edit", ah.Edit)
	ag.POST("/publish", ah.Publish)
	pub := ag.Group("/pub")
	pub.GET("/:id", ah.PubDetail)
	pub.POST("/withdraw", ah.Withdraw)
	pub.POST("/like", ah.Like)
	pub.POST("/collect", ah.Collect)
}

func (ah *ArticleHandler) Withdraw(ctx *gin.Context) {
	// 撤回文章只需要修改文章对应的状态
	type Req struct {
		Id int64
	}
	var req Req
	if err := ctx.Bind(&req); err != nil {
		return
	}

	c, ok := ctx.Get("claims")
	if !ok {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	//c := ctx.MustGet("claims")
	//拿出jwt的UserClaims
	claims, ok := c.(*ijwt.UserClaims)

	if !ok {
		ctx.JSON(
			http.StatusOK, Result{
				Code: 5,
				Msg:  "系统错误",
			})
	}

	err := ah.articleSvc.Withdraw(ctx, domain.Article{
		Id: req.Id,
		Author: domain.Author{
			Id: claims.Id,
		},
	})
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Msg: "OK",
	})

}

func (ah *ArticleHandler) Edit(ctx *gin.Context) {
	type ArticleReq struct {
		Id      int64  `json:"id"`
		Title   string `json:"title"`
		Content string `json:"content"`
	}

	var req ArticleReq
	if err := ctx.Bind(&req); err != nil {
		return
	}

	// 数据检查
	if req.Title == "" {
		ctx.String(http.StatusOK, "标题不能为空")
	}

	// 从某个地方取出userid 可以从jwt里取
	c, ok := ctx.Get("claims")
	if !ok {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	claims := c.(*ijwt.UserClaims)
	uid := claims.Id

	aid, err := ah.articleSvc.Save(ctx, domain.Article{
		Id:      req.Id,
		Title:   req.Title,
		Content: req.Content,
		Author: domain.Author{
			Id: uid,
		},
		Status: domain.ArticleStatusUnpublished,
	})

	if err != nil {
		ctx.String(http.StatusOK, "保存文章失败")
		return
	}

	// aid 怎么用
	ctx.JSON(http.StatusOK, Result{
		Data: aid,
	})

}

func (ah *ArticleHandler) Publish(ctx *gin.Context) {
	// 需要把制作库同步到线上库
	type ArticleReq struct {
		Id      int64  `json:"id"`
		Title   string `json:"title"`
		Content string `json:"content"`
	}

	var req ArticleReq
	if err := ctx.Bind(&req); err != nil {
		return
	}

	// 数据检查
	if req.Title == "" {
		ctx.String(http.StatusOK, "标题不能为空")
	}

	// 从某个地方取出userid 可以从jwt里取
	c, ok := ctx.Get("claims")
	if !ok {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	claims := c.(*ijwt.UserClaims)
	uid := claims.Id

	aid, err := ah.articleSvc.Publish(ctx, domain.Article{
		Id:      req.Id,
		Title:   req.Title,
		Content: req.Content,
		Author: domain.Author{
			Id: uid,
		},
		Status: domain.ArticleStatusPublished,
	})

	if err != nil {
		ctx.String(http.StatusOK, "发布文章失败")
	}

	// aid 怎么用
	ctx.JSON(http.StatusOK, Result{
		Data: aid,
	})
}

func (ah *ArticleHandler) PubDetail(ctx *gin.Context) {
	idstr := ctx.Param("id")
	id, err := strconv.ParseInt(idstr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 4,
			Msg:  "参数错误",
		})
		return
	}

	//art, err := ah.articleSvc.GetPublishedById(ctx, id)
	//if err != nil {
	//	ctx.JSON(http.StatusOK, Result{
	//		Code: 5,
	//		Msg:  "系统错误",
	//	})
	//	return
	//}

	//// 开启一个goroutine 增加阅读计数
	//go func() {
	//	ah.interSvc.IncrReadCnt(ctx, ah.biz, art.Id)
	//}()

	// 处理文章本身信息，还希望拿到 阅读点赞收藏计数 以及读者对文章是否like collect
	// interactiveService 写一个Get方法 一站式集齐这些信息
	// 文章和文章的点赞阅读计数等信息需要一起返回， 或者先返回文章本体 后加载阅读点赞等信息 异步
	c, ok := ctx.Get("claims")
	if !ok {
		ctx.AbortWithStatus(http.StatusUnauthorized)
	}
	uid := c.(*ijwt.UserClaims).Id

	// 这里waitGroup一起返回
	var (
		eg   errgroup.Group
		art  domain.Article
		intr domain.Interactive
	)

	eg.Go(func() error {
		var er error
		art, er = ah.articleSvc.GetPublishedById(ctx, id)
		return er
	})

	eg.Go(func() error {
		var er error
		intr, er = ah.interSvc.Get(ctx, ah.biz, id, uid)
		return er
	})

	err = eg.Wait()

	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "获取文章信息失败",
		})
		return
	}
	go func() {
		ah.interSvc.IncrReadCnt(ctx, ah.biz, art.Id)
	}()

	ctx.JSON(http.StatusOK, Result{
		Data: ArticleVO{
			Id:    art.Id,
			Title: art.Title,
			//Status:  art.Status.ToUint8(),
			Content: art.Content,
			// 要把作者信息带出去
			//Author: art.Author.Name,
			//Ctime:  art.Ctime.Format(time.DateTime),
			//Utime:  art.Utime.Format(time.DateTime),

			// 点赞之类的信息
			LikeCnt:    intr.LikeCnt,
			CollectCnt: intr.CollectCnt,
			ReadCnt:    intr.ReadCnt,

			// 个人是否点赞的信息
			Liked:     intr.Liked,
			Collected: intr.Collected,
		},
	})

}

func (ah *ArticleHandler) Like(ctx *gin.Context) {
	var err error
	type LikeReq struct {
		Id   int64 `json:"id"`
		Like bool  `json:"like"`
	}
	var req LikeReq
	if err := ctx.Bind(&req); err != nil {
		return
	}
	//c := ctx.MustGet("claims")
	c, ok := ctx.Get("claims")
	if !ok {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	uc, ok := c.(*ijwt.UserClaims)
	if !ok {
		// 记录
	}
	if req.Like {
		err = ah.interSvc.Like(ctx, ah.biz, req.Id, uc.Id)
	} else {
		err = ah.interSvc.CancelLike(ctx, ah.biz, req.Id, uc.Id)
	}

	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Msg: "Ok",
	})
}

func (ah *ArticleHandler) Collect(ctx *gin.Context) {
	type CollectReq struct {
		Id  int64 `json:"id"`
		Cid int64 `json:"cid"`
	}

	var req CollectReq
	if err := ctx.Bind(&req); err != nil {
		return
	}

	c, ok := ctx.Get("claims")
	if !ok {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	uc, ok := c.(*ijwt.UserClaims) // 为什么这里一定要转成指针类型 当时存的是指针类型吗 指针类型和具体类型有什么区别
	if !ok {
		//
	}

	err := ah.interSvc.Collect(ctx, ah.biz, req.Id, req.Cid, uc.Id)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
	ctx.String(http.StatusOK, "收藏成功")
}

func (ah *ArticleHandler) TopNLike(ctx *gin.Context) {
	type TopNReq struct {
		TopN int64 `json:"top_n"`
	}
	var req TopNReq
	if err := ctx.Bind(&req); err != nil {
		return
	}

	topn, err := ah.interSvc.TopNLike(ctx, ah.biz, req.TopN) // 返回 []intr 等
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
	}
	m1 := map[int64]domain.InteractiveArticle{}
	m2 := map[int64]domain.Article{}
	ids := []int64{}
	for _, tmp := range topn {
		ids = append(ids, tmp.ArtId)
		m1[tmp.ArtId] = tmp
	}
	// topn art.Id 去数据库查询对应的art.title 等信息
	arts, err := ah.articleSvc.GetPublishedByIds(ctx, ids)
	for _, art := range arts {
		m2[art.Id] = art
	}

	fmt.Println(arts)
	avs := []ArticleVO{}
	for _, id := range ids {
		avs = append(avs, ArticleVO{
			Id:    id,
			Title: m2[id].Title,
			//Status:  art.Status.ToUint8(),
			Content: m2[id].Content,
			// 要把作者信息带出去
			//Author: art.Author.Name,
			//Ctime:  art.Ctime.Format(time.DateTime),
			//Utime:  art.Utime.Format(time.DateTime),

			// 点赞之类的信息
			LikeCnt:    m1[id].LikeCnt,
			CollectCnt: m1[id].CollectCnt,
			ReadCnt:    m1[id].ReadCnt,
		})
	}

	// 构造 avs
	ctx.JSON(http.StatusOK, Result{
		Data: avs,
	})
}
