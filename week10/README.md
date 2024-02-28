分布式任务调度 

可能存在多个实例执行热榜计算任务，同一时刻只需要一个实例进行热榜计算，
引入分布式锁，多个实例来竞争分布式锁，同一时刻获得锁的实力进行热榜计算 
热榜计算完成后，持有分布式锁的实例如果释放了分布式锁，则该分布式锁可能会立即被另一个实例获取，又马上执行了一次任务 
所以只在实例启动的时候竞争分布式锁，之后该实例一直持有该分布式锁，
且分布式锁设置过期时间，如果获得分布式锁的实例宕机，可以通过过期时间，让其他实例有机会获得锁

增加考虑实例负载，那么在执行分布式任务，获取分布式锁的时候，在实例负载高的时候，执行完任务需要释放分布式锁

