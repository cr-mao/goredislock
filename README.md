## go redis分布锁实现


### 功能
- redis客户端可传，默认用全局实例
- 实现自动过期
- 实现自动续租 （续租时间为过期时间的2/3)
  - lua实现获得和过期时间设置原子操作



### download

```shell
go get github.com/cr-mao/goredislock@v1.0.0
```

### test

```shell
go test -v ./...
````

### demo

```go
package main

import (
  "context"
  "fmt"
  "time"
  
  "github.com/cr-mao/goredislock"
)

/*
续租测试，

2秒过期时间，续租时间大概是 1.33秒，10进行了7次续租，复合要求
2023/06/17 17:37:08 PONG
true
2023/06/17 17:37:10 key=test_lock_key ,续期结果:<nil>,1
2023/06/17 17:37:11 key=test_lock_key ,续期结果:<nil>,1
2023/06/17 17:37:12 key=test_lock_key ,续期结果:<nil>,1
2023/06/17 17:37:14 key=test_lock_key ,续期结果:<nil>,1
2023/06/17 17:37:15 key=test_lock_key ,续期结果:<nil>,1
2023/06/17 17:37:16 key=test_lock_key ,续期结果:<nil>,1
2023/06/17 17:37:18 key=test_lock_key ,续期结果:<nil>,1
*/
func main() {
	// 实例化全局redisclient, 分布式锁则会用这个redisClient
  goredislock.NewRedisClient("127.0.0.1:6379")
  // 1.33秒左右就会续租
  locker, ok := goredislock.NewLocker("test_lock_key", goredislock.WithContext(context.Background()), goredislock.WithExpire(time.Second*2)).Lock()
  fmt.Println(ok)
  time.Sleep(time.Second*10)
  defer locker.Unlock()
}
```

**with your self redisClient**

```go
package main

import (
  "context"
  "fmt"
  "time"

  redis "github.com/go-redis/redis/v8"

  "github.com/cr-mao/goredislock"
)

func main() {
  var redisClient = redis.NewClient(&redis.Options{
    Network:  "tcp",
    Addr:     "127.0.0.1:6379",
    Password: "", //密码
    DB:       0,  // redis数据库

    //连接池容量及闲置连接数量
    PoolSize:     15, // 连接池数量
    MinIdleConns: 10, //好比最小连接数
    //超时
    DialTimeout:  5 * time.Second, //连接建立超时时间
    ReadTimeout:  3 * time.Second, //读超时，默认3秒， -1表示取消读超时
    WriteTimeout: 3 * time.Second, //写超时，默认等于读超时
    PoolTimeout:  4 * time.Second, //当所有连接都处在繁忙状态时，客户端等待可用连接的最大等待时长，默认为读超时+1秒。

    //闲置连接检查包括IdleTimeout，MaxConnAge
    IdleCheckFrequency: 60 * time.Second, //闲置连接检查的周期，默认为1分钟，-1表示不做周期性检查，只在客户端获取连接时对闲置连接进行处理。
    IdleTimeout:        5 * time.Minute,  //闲置超时
    MaxConnAge:         0 * time.Second,  //连接存活时长，从创建开始计时，超过指定时长则关闭连接，默认为0，即不关闭存活时长较长的连接

    //命令执行失败时的重试策略
    MaxRetries:      0,                      // 命令执行失败时，最多重试多少次，默认为0即不重试
    MinRetryBackoff: 8 * time.Millisecond,   //每次计算重试间隔时间的下限，默认8毫秒，-1表示取消间隔
    MaxRetryBackoff: 512 * time.Millisecond, //每次计算重试间隔时间的上限，默认512毫秒，-1表示取消间隔

  })

  locker, ok := goredislock.NewLocker("test_lock_key", goredislock.WithContext(context.Background()), goredislock.WithExpire(time.Second*2), goredislock.WithRedisClient(redisClient)).Lock()
  fmt.Println(ok)
  defer locker.Unlock()
}


```







