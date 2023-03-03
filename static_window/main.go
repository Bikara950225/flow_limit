package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

type StaticWindow struct {
	limit     int           // 限制次数
	count     int           // 当前次数
	startTime time.Duration // 开始计时的时间
	cycle     time.Duration // 重新计数的周期
	lock      sync.Mutex
}

func NewStaticWindow(limit int, cycle time.Duration) *StaticWindow {
	return &StaticWindow{
		limit:     limit,
		count:     0,
		startTime: time.Duration(time.Now().Unix()),
		cycle:     cycle,
	}
}

func (s *StaticWindow) Allow() bool {
	s.lock.Lock()
	defer s.lock.Unlock()

	now := time.Duration(time.Now().Unix())
	if now-s.startTime > s.cycle {
		s.startTime = now
		s.count = 0
	}

	if s.count+1 <= s.limit {
		s.count += 1
		return true
	} else {
		return false
	}
}

func main() {
	wg := sync.WaitGroup{}
	wg.Add(10)

	sw := NewStaticWindow(60, time.Minute)
	var count int64
	var failCount int64

	for i := 0; i < 10; i++ {
		go func() {
			defer wg.Done()

			for j := 0; j < 100; j++ {
				if sw.Allow() {
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
