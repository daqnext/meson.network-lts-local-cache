package go_fast_cache

import (
	"github.com/daqnext/go-fast-cache/sortedset"
	"github.com/daqnext/go-fast-cache/ttltype"
	"sync"
	"time"
)

const (
	MaxTTLSecond      = 7200
	DefaultCountLimit = 1000000
	MinCountLimit     = 10000

	MaxDeleteExpireIntervalSecond     = 300
	DefaultDeleteExpireIntervalSecond = 5

	DefaultDeleteOverLimitRate = 0.15
)

type localCache struct {
	s          *sortedset.SortedSet
	countLimit int64
	lock       sync.Mutex
}

// New Instance of localCache, the interval of scheduleDeleteExpire job use the default value 5 seconds
func New() *localCache {
	cache := &localCache{
		s:          sortedset.Make(),
		countLimit: DefaultCountLimit,
	}
	cache.scheduleDeleteExpire(5)
	cache.scheduleDeleteOverLimit()
	return cache
}

// NewWithInterval Instance of localCache, param intervalSecond defines the interval of scheduleDeleteExpire job, if intervalSecond <=0,it will use the default value 5 seconds
func NewWithInterval(intervalSecond int) *localCache {
	if intervalSecond > MaxDeleteExpireIntervalSecond {
		intervalSecond = MaxDeleteExpireIntervalSecond
	}
	if intervalSecond < 1 {
		intervalSecond = DefaultDeleteExpireIntervalSecond
	}
	cache := &localCache{
		s:          sortedset.Make(),
		countLimit: DefaultCountLimit,
	}
	cache.scheduleDeleteExpire(intervalSecond)
	cache.scheduleDeleteOverLimit()
	return cache
}

// SetCountLimit Key count limit,default is 1000000. The 15% of the keys with the most recent expiration time will be deleted if the number of keys exceeds the limit.
func (lc *localCache) SetCountLimit(limit int64) {
	if limit < MinCountLimit {
		limit = MinCountLimit
	}
	lc.countLimit = limit
}

func (lc *localCache) Get(key string) (value interface{}, ttl int64, exist bool) {
	//check expire
	e, exist := lc.s.Get(key)
	if !exist {
		return nil, 0, false
	}
	nowTime := time.Now().Unix()
	if e.Score <= nowTime {
		return nil, 0, false
	}
	return e.Value, e.Score - nowTime, true
}

// Set Set key value with expire time, ttl.Keep or second. If key not exist and set ttl ttl.Keep,it will use default ttl 30sec
func (lc *localCache) Set(key string, value interface{}, ttlSecond int64) {
	if ttlSecond < 0 {
		return
	}

	if ttlSecond > MaxTTLSecond {
		ttlSecond = MaxTTLSecond
	}
	var expireTime int64

	if ttlSecond == ttltype.Keep {
		//keep
		ttlLeft, exist := lc.ttl(key)
		if !exist {
			ttlLeft = 30
		}
		expireTime = time.Now().Unix() + ttlLeft
	} else {
		//new expire
		expireTime = time.Now().Unix() + ttlSecond
	}
	lc.s.Add(key, expireTime, value)
}

// TTL get ttl of a key with second
func (lc *localCache) ttl(key string) (int64, bool) {
	e, exist := lc.s.Get(key)
	if !exist {
		return 0, false
	}
	ttl := e.Score - time.Now().Unix()
	if ttl <= 0 {
		return 0, false
	}
	return ttl, true
}

func (lc *localCache) scheduleDeleteOverLimit() {
	go func() {
		time.Sleep(500 * time.Millisecond)
		for {
			time.Sleep(1 * time.Second)
			//log.Println("scheduleDeleteOverLimit start")
			if lc.s.Len() >= lc.countLimit {
				deleteCount := float64(lc.countLimit) * DefaultDeleteOverLimitRate
				lc.s.RemoveByRank(0, int64(deleteCount))
			}
		}
	}()
}

// ScheduleDeleteExpire delete expired keys
func (lc *localCache) scheduleDeleteExpire(intervalSecond int) {
	go func() {
		for {
			time.Sleep(time.Duration(intervalSecond) * time.Second)
			//log.Println("scheduleDeleteExpire start")
			max := time.Now().Unix()
			//remove expired keys
			lc.s.RemoveByScore(max)
		}
	}()
}

func (lc *localCache) GetLen() int64 {
	return lc.s.Len()
}
