package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

type LeakyBucket struct {
	tick  time.Duration
	ch    chan struct{}
	limit int
}

func NewLeakyBucket(limit int, tick time.Duration) *LeakyBucket {
	lb := &LeakyBucket{
		limit: limit,
		ch:    make(chan struct{}, limit),
		tick:  tick,
	}
	// TODO cancel
	go lb.take()

	return lb
}

func (s *LeakyBucket) Allow() bool {
	select {
	case s.ch <- struct{}{}:
		return true
	default:
		return false
	}
}

func (s *LeakyBucket) take() {
	for {
		time.Sleep(s.tick)
		select {
		case <-s.ch:
		default:
		}
	}
}

// 不使用channel的实现
type LeakyBucket2 struct {
	lock     sync.Mutex
	limit    int
	count    int
	cycle    time.Duration
	lastLeak time.Time
}

func NewLeakyBucket2(limit int, cycle time.Duration) *LeakyBucket2 {
	return &LeakyBucket2{
		limit:    limit,
		count:    0,
		cycle:    cycle,
		lastLeak: time.Now(),
	}
}

func (s *LeakyBucket2) Allow() bool {
	s.lock.Lock()
	defer s.lock.Unlock()

	// 先计算距离上一次调用经过的时间，计算出需要减少多少的count
	now := time.Now()
	subLeaky := now.Sub(s.lastLeak) / s.cycle
	count := s.count - int(subLeaky)
	if count < 0 {
		count = 0
	}
	s.count = count

	if s.count+1 <= s.limit {
		s.count++
		s.lastLeak = now
		return true
	} else {
		return false
	}
}

func main() {
	//lb := NewLeakyBucket(60, 20*time.Millisecond)
	lb := NewLeakyBucket2(60, 20*time.Millisecond)
	wg := sync.WaitGroup{}
	wg.Add(1000)

	var count int64
	var failCount int64

	for i := 0; i < 1000; i++ {
		go func() {
			defer wg.Done()

			for j := 0; j < 100; j++ {
				if lb.Allow() {
					atomic.AddInt64(&count, 1)
				} else {
					atomic.AddInt64(&failCount, 1)
				}
			}
		}()
	}
	wg.Wait()

	fmt.Println(count)
	fmt.Println(failCount)
}
