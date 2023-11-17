package web

import (
	"errors"
	"geekgo/week9/webook/internal/domain"
	"geekgo/week9/webook/internal/service"
	ijwt "geekgo/week9/webook/internal/web/jwt"
	"geekgo/week9/webook/pkgs/ginx"
	"github.com/gin-gonic/gin"
	"golang.org/x/sync/errgroup"
	"net/http"
	"strconv"
	"time"
)

type ArticleHandler struct {
	svc service.ArticleService
	// articleHandler 还需要组合 interactiveService 在点赞阅读等场景需要使用这个service
	intrSvc service.InteractiveService
	biz     string
}

const biz = "article"

func NewArticleHandler(svc service.ArticleService, intrSvc service.InteractiveService) *ArticleHandler {
	return &ArticleHandler{
		svc:     svc,
		intrSvc: intrSvc,
		biz:     biz,
	}
}

func (ah *ArticleHandler) RegisterRoutes(server *gin.Engine) {
	g := server.Group("articles")
	//g.POST("/edit", ah.Edit)
	//g.POST("/publish", ah.Publish)
	//g.POST("/withdraw", ah.Withdraw)
	//g.GET("/detail/:id", ah.Detail)
	//pub := g.Group("/pub")
	//pub.GET("/:id", ah.PubDetail)
	//pub.POST("/like", ah.Like)
	//pub.POST("/collect", ah.Collect)

	g.POST("/edit", ginx.WrapReq[EditReq](ah.EditV1))
	g.POST("/publish", ginx.WrapReq[EditReq](ah.PublishV1))
	g.POST("/withdraw", ginx.WrapReq[EditReq](ah.WithdrawV1))
	g.GET("/detail/:id", ginx.WrapReq[struct{}](ah.DetailV1))
	pub := g.Group("/pub")
	pub.GET("/:id", ginx.WrapReq[struct{}](ah.PubDetailV1))
	pub.POST("/like", ginx.WrapReq[LikeReq](ah.LikeV1))
	pub.POST("/collect", ginx.WrapReq[CollectReq](ah.CollectV1))

}

func (ah *ArticleHandler) Edit(ctx *gin.Context) {
	type EditReq struct {
		Id      int64  `json:"id"`
		Title   string `json:"title"`
		Content string `json:"content"`
	}
	var req EditReq
	if err := ctx.Bind(&req); err != nil {
		return
	}
	// 取出jwttoken里面保存的userid  频繁使用 抽取出公共逻辑
	uc, ok := ctx.Get("claims")
	if !ok {
		// 应该是系统错误
		ctx.String(http.StatusOK, "系统错误")
		//ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	userClaims, ok := uc.(ijwt.UserClaim)
	if !ok {
		// 应该是系统错误
		ctx.String(http.StatusOK, "系统错误")
		//ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	uid := userClaims.Uid
	aid, err := ah.svc.Save(ctx, domain.Article{
		Id:      req.Id,
		Title:   req.Title,
		Content: req.Content,
		Author: domain.Author{
			Id: uid,
		},
		Status: domain.ArticleStatusUnpublished,
	})

	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Data: aid,
	})

}

func (ah *ArticleHandler) Publish(ctx *gin.Context) {
	// publish 需要 同时更新制作库和线上库
	type EditReq struct {
		Id      int64  `json:"id"`
		Title   string `json:"title"`
		Content string `json:"content"`
	}
	var req EditReq
	if err := ctx.Bind(&req); err != nil {
		return
	}
	// 取出jwttoken里面保存的userid  频繁使用 抽取出公共逻辑
	uc, ok := ctx.Get("claims")
	if !ok {
		// 应该是系统错误
		ctx.String(http.StatusOK, "系统错误")
		//ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	userClaims, ok := uc.(ijwt.UserClaim)
	if !ok {
		// 应该是系统错误
		ctx.String(http.StatusOK, "系统错误")
		//ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	uid := userClaims.Uid
	aid, err := ah.svc.Publish(ctx, domain.Article{
		Id:      req.Id,
		Title:   req.Title,
		Content: req.Content,
		Author: domain.Author{
			Id: uid,
		},
		Status: domain.ArticleStatusPublished,
	})

	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Data: aid,
	})
}

func (ah *ArticleHandler) Withdraw(ctx *gin.Context) {
	type EditReq struct {
		Id      int64  `json:"id"`
		Title   string `json:"title"`
		Content string `json:"content"`
	}
	var req EditReq
	if err := ctx.Bind(&req); err != nil {
		return
	}
	// 取出jwttoken里面保存的userid  频繁使用 抽取出公共逻辑
	uc, ok := ctx.Get("claims")
	if !ok {
		// 应该是系统错误
		ctx.String(http.StatusOK, "系统错误")
		//ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	userClaims, ok := uc.(ijwt.UserClaim)
	if !ok {
		// 应该是系统错误
		ctx.String(http.StatusOK, "系统错误")
		//ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	uid := userClaims.Uid

	err := ah.svc.Withdraw(ctx, req.Id, uid)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	ctx.String(http.StatusOK, "成功设置自己可见")
}

func (ah *ArticleHandler) Detail(ctx *gin.Context) {
	// detail 返回制作库的文章数据
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	art, err := ah.svc.GetById(ctx, id)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	uc, ok := ctx.Get("claims")
	if !ok {
		// 应该是系统错误
		ctx.String(http.StatusOK, "系统错误")
		//ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	userClaims, ok := uc.(ijwt.UserClaim)
	if !ok {
		// 应该是系统错误
		ctx.String(http.StatusOK, "系统错误")
		//ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	uid := userClaims.Uid
	if art.Author.Id != uid {
		ctx.String(http.StatusOK, "非法访问文章")
		//ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	ctx.JSON(http.StatusOK,
		Result{
			Data: ArticleVO{
				Id:    art.Id,
				Title: art.Title,
				// 不需要这个摘要信息
				//Abstract: art.Abstract(),
				Status:  art.Status.ToUint8(),
				Content: art.Content,
				// 这个是创作者看自己的文章列表，也不需要这个字段
				//Author: art.Author
				Ctime: art.Ctime.Format(time.DateTime),
				Utime: art.Utime.Format(time.DateTime),
			},
		})

}

func getUidFromCtxClaims(ctx *gin.Context) (int64, error) {
	uc, ok := ctx.Get("claims")
	if !ok {
		return 0, errors.New("获取uid出错")
	}
	userClaims, ok := uc.(ijwt.UserClaim)
	if !ok {
		// 应该是系统错误
		return 0, errors.New("获取uid出错")
	}
	uid := userClaims.Uid
	return uid, nil
}

func (ah *ArticleHandler) PubDetail(ctx *gin.Context) {
	// detail 返回制作库的文章数据
	// 设为Private的文章 非创作者不能访问， 应该在哪里控制？
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	uc, ok := ctx.Get("claims")
	if !ok {
		// 应该是系统错误
		ctx.String(http.StatusOK, "系统错误")
		//ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	userClaims, ok := uc.(ijwt.UserClaim)
	if !ok {
		// 应该是系统错误
		ctx.String(http.StatusOK, "系统错误")
		//ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	uid := userClaims.Uid

	var eg errgroup.Group
	var art domain.Article

	eg.Go(func() error {
		art, err = ah.svc.GetPublishedById(ctx, id, uid)
		return err
	})

	var intr domain.Interactive
	eg.Go(func() error {
		intr, err = ah.intrSvc.Get(ctx, ah.biz, id, uid)
		return err
	})

	err = eg.Wait()
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
	// 读取文章的同时 增加阅读计数 把阅读文章的事件发送到kafka 然后等关注阅读事件的消费者消费
	// 在svc层GetPublishedById 调用kafka把消息发出去

	// 显示文章内容之外 还需要显示 like_cnt 该用户是否like 。。。

	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}

	ctx.JSON(http.StatusOK, Result{
		Data: ArticleVO{
			Id:      art.Id,
			Title:   art.Title,
			Status:  art.Status.ToUint8(),
			Content: art.Content,
			// 要把作者信息带出去
			Author:     art.Author.Name, // 这个是从哪里聚合进来的？
			Ctime:      art.Ctime.Format(time.DateTime),
			Utime:      art.Utime.Format(time.DateTime),
			Liked:      intr.Liked,
			Collected:  intr.Collected,
			LikeCnt:    intr.LikeCnt,
			ReadCnt:    intr.ReadCnt,
			CollectCnt: intr.CollectCnt,
		},
	})

}

func (ah *ArticleHandler) Like(ctx *gin.Context) {
	type LikeReq struct {
		Id   int64 `json:"id"`
		Like bool  `json:"like"`
	}
	var req LikeReq
	if err := ctx.Bind(&req); err != nil {
		return
	}
	var err error
	uid, err := getUidFromCtxClaims(ctx)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	if req.Like {
		err = ah.intrSvc.Like(ctx, ah.biz, req.Id, uid)
	} else {
		err = ah.intrSvc.CancelLike(ctx, ah.biz, req.Id, uid)
	}
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Msg: "ok",
	})
}

func (ah *ArticleHandler) Collect(ctx *gin.Context) {
	// collect
	type CollectReq struct {
		Id  int64 `json:"id"`
		Cid int64 `json:"cid"`
	}
	var req CollectReq
	var err error
	if err = ctx.Bind(&req); err != nil {
		return
	}
	uid, err := getUidFromCtxClaims(ctx)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	err = ah.intrSvc.Collect(ctx, ah.biz, req.Id, req.Cid, uid)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Msg: "ok",
	})
}

func (ah *ArticleHandler) EditV1(ctx *gin.Context, req EditReq) (ginx.Result, error) {

	// 取出jwttoken里面保存的userid  频繁使用 抽取出公共逻辑
	uc, ok := ctx.Get("claims")
	if !ok {
		return ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		}, errors.New("get claims error")

	}
	userClaims, ok := uc.(ijwt.UserClaim)
	if !ok {
		return ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		}, errors.New("assert user claims error")
	}
	uid := userClaims.Uid
	aid, err := ah.svc.Save(ctx, domain.Article{
		Id:      req.Id,
		Title:   req.Title,
		Content: req.Content,
		Author: domain.Author{
			Id: uid,
		},
		Status: domain.ArticleStatusUnpublished,
	})

	if err != nil {
		return ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		}, err

	}
	return ginx.Result{
		Data: aid,
	}, nil

}

func (ah *ArticleHandler) PublishV1(ctx *gin.Context, req EditReq) (ginx.Result, error) {

	// 取出jwttoken里面保存的userid  频繁使用 抽取出公共逻辑
	uc, ok := ctx.Get("claims")
	if !ok {
		return ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		}, errors.New("get claims error")
	}
	userClaims, ok := uc.(ijwt.UserClaim)
	if !ok {
		return ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		}, errors.New("assert userclaim error")
	}
	uid := userClaims.Uid
	aid, err := ah.svc.Publish(ctx, domain.Article{
		Id:      req.Id,
		Title:   req.Title,
		Content: req.Content,
		Author: domain.Author{
			Id: uid,
		},
		Status: domain.ArticleStatusPublished,
	})

	if err != nil {
		return ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		}, err
	}

	return ginx.Result{
		Data: aid,
	}, nil
}

func (ah *ArticleHandler) WithdrawV1(ctx *gin.Context, req EditReq) (ginx.Result, error) {

	// 取出jwttoken里面保存的userid  频繁使用 抽取出公共逻辑
	uc, ok := ctx.Get("claims")
	if !ok {
		return ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		}, errors.New("get claims error")
	}
	userClaims, ok := uc.(ijwt.UserClaim)
	if !ok {
		// 应该是系统错误
		return ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		}, errors.New("assert userclaim error")
	}
	uid := userClaims.Uid

	err := ah.svc.Withdraw(ctx, req.Id, uid)
	if err != nil {
		return ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		}, err
	}
	return ginx.Result{

		Msg: "成功设置自己可见",
	}, nil
}

func (ah *ArticleHandler) DetailV1(ctx *gin.Context, req struct{}) (ginx.Result, error) {
	// detail 返回制作库的文章数据
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		}, err
	}
	art, err := ah.svc.GetById(ctx, id)
	if err != nil {
		return ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		}, err
	}
	uc, ok := ctx.Get("claims")
	if !ok {
		return ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		}, errors.New("get claims error")
	}
	userClaims, ok := uc.(ijwt.UserClaim)
	if !ok {
		return ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		}, errors.New("assert userclaim error")
	}
	uid := userClaims.Uid
	if art.Author.Id != uid {
		return ginx.Result{
			Code: 4,
			Msg:  "非法访问文章",
		}, errors.New("非法访问文章")

	}
	return ginx.Result{
		Data: ArticleVO{
			Id:    art.Id,
			Title: art.Title,
			// 不需要这个摘要信息
			//Abstract: art.Abstract(),
			Status:  art.Status.ToUint8(),
			Content: art.Content,
			// 这个是创作者看自己的文章列表，也不需要这个字段
			//Author: art.Author
			Ctime: art.Ctime.Format(time.DateTime),
			Utime: art.Utime.Format(time.DateTime),
		}}, nil

}

func (ah *ArticleHandler) PubDetailV1(ctx *gin.Context, req struct{}) (ginx.Result, error) {
	// detail 返回制作库的文章数据
	// 设为Private的文章 非创作者不能访问， 应该在哪里控制？
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		}, err
	}
	uc, ok := ctx.Get("claims")
	if !ok {
		return ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		}, errors.New("get claims error")
	}
	userClaims, ok := uc.(ijwt.UserClaim)
	if !ok {
		return ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		}, errors.New("assert userclaim error")
	}
	uid := userClaims.Uid

	var eg errgroup.Group
	var art domain.Article

	eg.Go(func() error {
		art, err = ah.svc.GetPublishedById(ctx, id, uid)
		return err
	})

	var intr domain.Interactive
	eg.Go(func() error {
		intr, err = ah.intrSvc.Get(ctx, ah.biz, id, uid)
		return err
	})

	err = eg.Wait()
	if err != nil {
		return ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		}, err
	}
	// 读取文章的同时 增加阅读计数 把阅读文章的事件发送到kafka 然后等关注阅读事件的消费者消费
	// 在svc层GetPublishedById 调用kafka把消息发出去

	// 显示文章内容之外 还需要显示 like_cnt 该用户是否like 。。。

	if err != nil {
		return ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		}, err
	}

	return ginx.Result{
		Data: ArticleVO{
			Id:      art.Id,
			Title:   art.Title,
			Status:  art.Status.ToUint8(),
			Content: art.Content,
			// 要把作者信息带出去
			Author:     art.Author.Name, // 这个是从哪里聚合进来的？
			Ctime:      art.Ctime.Format(time.DateTime),
			Utime:      art.Utime.Format(time.DateTime),
			Liked:      intr.Liked,
			Collected:  intr.Collected,
			LikeCnt:    intr.LikeCnt,
			ReadCnt:    intr.ReadCnt,
			CollectCnt: intr.CollectCnt,
		}}, nil

}

func (ah *ArticleHandler) LikeV1(ctx *gin.Context, req LikeReq) (ginx.Result, error) {

	var err error
	uid, err := getUidFromCtxClaims(ctx)
	if err != nil {
		return ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		}, err
	}
	if req.Like {
		err = ah.intrSvc.Like(ctx, ah.biz, req.Id, uid)
	} else {
		err = ah.intrSvc.CancelLike(ctx, ah.biz, req.Id, uid)
	}
	if err != nil {
		return ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		}, err
	}
	return ginx.Result{
		Msg: "ok",
	}, nil
}

func (ah *ArticleHandler) CollectV1(ctx *gin.Context, req CollectReq) (ginx.Result, error) {
	// collect

	uid, err := getUidFromCtxClaims(ctx)
	if err != nil {
		return ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		}, err
	}
	err = ah.intrSvc.Collect(ctx, ah.biz, req.Id, req.Cid, uid)
	if err != nil {
		return ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		}, err
	}

	return ginx.Result{
		Msg: "ok",
	}, nil
}
