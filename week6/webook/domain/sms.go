package domain

type SMS struct {
	Id      int64
	TplId   string
	Args    []string
	Numbers []string
	// 设置可以重试三次
	RetryMax int
}
