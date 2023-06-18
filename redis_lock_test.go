package goredislock

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
)

// 只有一个go程能获得 锁
func TestRedisLock(t *testing.T) {
	//实例化全局redis
	NewRedisClient("127.0.0.1:6379", 0, "", "")

	for i := 0; i < 10; i++ {
		go func(num int) {
			// 默认10秒过期，
			locker, ok := NewLocker("test_lock_key").Lock()
			defer locker.Unlock()
			fmt.Println(num, ok)
		}(i)
	}
	time.Sleep(time.Second * 3)
}

func TestRedisLockWithContext(t *testing.T) {
	//实例化全局redis
	NewRedisClient("127.0.0.1:6379", 0, "", "")
	locker, ok := NewLocker("test_lock_key", WithContext(context.Background())).Lock()
	defer locker.Unlock()
	fmt.Println(ok)
}

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
func TestRedisLockWithExpire(t *testing.T) {
	//实例化全局redis
	NewRedisClient("127.0.0.1:6379", 0, "", "")
	locker, ok := NewLocker("test_lock_key", WithExpire(time.Second*2)).Lock()
	fmt.Println(ok)
	time.Sleep(time.Second * 10)
	defer locker.Unlock()
}

// WithRedisClient 测试
func TestRedisLockWithRedisClient(t *testing.T) {

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
	locker, ok := NewLocker("test_lock_key", WithExpire(time.Second*2), WithRedisClient(redisClient)).Lock()
	fmt.Println(ok)
	defer locker.Unlock()
}

func TestSetGlobalRedisClient(t *testing.T) {
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
	SetGlobalRedisClient(redisClient)
	locker, ok := NewLocker("test_lock_key", WithExpire(time.Second*2)).Lock()
	defer locker.Unlock()
	fmt.Println(ok)
}
