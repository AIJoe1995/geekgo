package domain

type Payment struct {
	Amt         Amount
	BizTradeNO  string // 业务生成 交易序号
	Description string // 订单的描述
	Status      PaymentStatus
	TxnID       string // transaction 第三方返回的交易id
}

type Amount struct {
	Currency string
	Total    int64 // int64 记录分数 10分是1角， 对于不同的货币Total含义不同
}

type PaymentStatus uint8

func (p PaymentStatus) AsUint8() uint8 {
	return uint8(p)
}

const (
	PaymentStatusUnknown = iota
	PaymentStatusInit
	PaymentStatusSuccess
	PaymentStatusFailed
	PaymentStatusRefund
)

type PaymentEvent struct {
	Id         int64
	BizTradeNO string
	Status     uint8 // PaymentEvent的status
	Sent       uint8 // 1 表示发送到kafka成功 0表示失败

}
