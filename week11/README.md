在mysql中创建一张表，存入需要定时执行的分布式任务
抢到任务的结点可以执行任务，只有一个节点能抢到任务
抢到任务的结点需要上报health 
如果health上报超时或多次未上报认定结点失败，把任务释放，其他节点可以抢占任务 
如果任务执行结束，需要向数据库更新下次执行的时间，定时任务下次不需要执行了的就删除记录


CronJobService 服务提供 添加任务 抢占任务 设置下次运行时间的接口 

终止任务 可以通过修改mysql中任务的Status 并停止上报health, 
终止任务的函数放在domain.Job结构体里开放给调度任务的模块


任务的调度: 拿到domain.Job{Cfg, Expression, Executor}之后，需要从注册到schedular里面的Executor map里面找到任务对应的执行器 
抢占 -> 执行 -> 释放(job.CancelFunc) 

