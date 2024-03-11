1. 引入本地消息表 确保PaymentEvent 一定会发送成功 
这里将发送到kafka失败的Payment记录 存放在数据库表PaymentEvent, 引入了一个任务，不断从PaymentEvent中拿到未成功发送的消息，发送成功标记为已发送


2. 记账的幂等性：可以通过Biz+bizId建立唯一索引，确定同一个biz+bizId不会重复记账，redis去重，布隆过滤器等
这里将biz+bizId放在mysql数据库表里 新建一条记录， 如果新建失败 则不入账

3. 服务治理

4. Reward Payment  Account  对账机制 
   冻结某一刻 三个表 snapshot 
   Reward中取出记录，拿到BizTradeNO, 取出对应Payment记录, 进行分账，计算出一张AccountReplication副本, 将Account和AccountReplication表作对比。



web前端接口提供打赏入口, 调用打赏模块建立打赏订单。
打赏订单需要调用支付模块的支付功能，收到支付结果后，需要更新账户余额，通知用户。

打赏模块负责：生成打赏单(调用支付模块)，处理打赏结果(根据支付回调结果)
账户模块负责：更新账户信息
支付模块负责：调用第三方支付，处理回调结果 

账户模块数据库表： Account, AccountActivity
打赏模块数据库表：Reward 打赏的编号不能重复，
支付模块数据库表：支付订单信息 Payment

打赏-支付-账户 三者间的对账功能 定时任务 


