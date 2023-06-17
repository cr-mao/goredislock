package goredislock

import (
	"context"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
)

const defaultExpireTime = time.Second * 10

// 分布式锁
type Locker struct {
	key        string        // redis key
	unlock     bool          // 是否已经解锁 ，解锁则不用续租
	incrScript *redis.Script // lua script
	option     options       // 可选项
}

type Options func(o *options)

type options struct {
	expire      time.Duration // expiration time ,默认是10秒
	redisClient *redis.Client // redis 实例， 可以传， 不传就得用全局的
	ctx         context.Context
}

// 续租的时候 获得和设置过期原子操作.
const incrLua = `
if redis.call('get', KEYS[1]) == ARGV[1] then
  return redis.call('expire', KEYS[1],ARGV[2]) 				
 else
   return '0' 					
end`

func NewLocker(key string, opts ...Options) *Locker {
	var lock = &Locker{
		key:        key,
		incrScript: redis.NewScript(incrLua),
	}
	for _, opt := range opts {
		opt(&lock.option)
	}
	// 没设置过期时间
	if lock.option.expire == 0 {
		lock.option.expire = defaultExpireTime
	}
	// 未设置redis 实例
	if lock.option.redisClient == nil {
		lock.option.redisClient = GlobalRedisClient
	}
	// 未设置context
	if lock.option.ctx == nil {
		lock.option.ctx = context.Background()
	}
	return lock
}

// 过期选项
func WithExpire(expire time.Duration) Options {
	return func(o *options) {
		o.expire = expire
	}
}

// redisClient 选项,可以每次都传，如果没传，就用全局都
func WithRedisClient(redisClient *redis.Client) Options {
	return func(o *options) {
		o.redisClient = redisClient
	}
}

func WithContext(ctx context.Context) Options {
	return func(o *options) {
		o.ctx = ctx
	}
}

// 第一个返回：返回锁，方便链式操作
// 第二个 返回结果
func (this *Locker) Lock() (*Locker, bool) {
	boolcmd := this.option.redisClient.SetNX(context.Background(), this.key, "1", this.option.expire)
	if ok, err := boolcmd.Result(); err != nil || !ok {
		return this, false
	}
	this.expandLockTime()
	return this, true
}

// 续租
func (this *Locker) expandLockTime() {
	sleepTime := this.option.expire * 2 / 3
	go func() {
		for {
			time.Sleep(sleepTime)
			if this.unlock {
				break
			}
			this.resetExpire()
		}
	}()
}

// 重新设置过期时间
func (this *Locker) resetExpire() {
	cmd := this.incrScript.Run(this.option.ctx, this.option.redisClient, []string{this.key}, 1, this.option.expire.Seconds())
	v, err := cmd.Result()
	log.Printf("key=%s ,续期结果:%v,%v\n", this.key, err, v)
}

// 释放锁  干完活后释放锁
func (this *Locker) Unlock() {
	this.unlock = true
	this.option.redisClient.Del(this.option.ctx, this.key)
}
