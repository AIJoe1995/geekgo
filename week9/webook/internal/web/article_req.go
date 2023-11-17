package web

type EditReq struct {
	Id      int64  `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

type LikeReq struct {
	Id   int64 `json:"id"`
	Like bool  `json:"like"`
}
type CollectReq struct {
	Id  int64 `json:"id"`
	Cid int64 `json:"cid"`
}
