package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

type TokenBucket struct {
	lock     sync.Mutex
	limit    int
	buckets  int
	lastLeak time.Time
	cycle    time.Duration
}

func NewTokenBucket(limit int, cycle time.Duration) *TokenBucket {
	return &TokenBucket{
		limit:    limit,
		buckets:  limit,
		lastLeak: time.Now(),
		cycle:    cycle,
	}
}

func (s *TokenBucket) Allow() bool {
	s.lock.Lock()
	defer s.lock.Unlock()

	now := time.Now()
	since := now.Sub(s.lastLeak) / s.cycle
	buckets := s.buckets + int(since)
	if buckets > s.limit {
		buckets = s.limit
	}
	s.buckets = buckets

	if s.buckets-1 >= 0 {
		s.buckets--
		s.lastLeak = now
		return true
	} else {
		return false
	}
}

type TokenBucket2 struct {
	limit int
	ch    chan struct{}
	cycle time.Duration
}

func NewTokenBucket2(limit int, cycle time.Duration) *TokenBucket2 {
	tb := &TokenBucket2{
		limit: limit,
		ch:    make(chan struct{}, limit),
		cycle: cycle,
	}
	for i := 0; i < limit; i++ {
		tb.ch <- struct{}{}
	}
	go tb.push()
	return tb
}

func (s *TokenBucket2) Allow() bool {
	select {
	case <-s.ch:
		return true
	default:
		return false
	}
}

func (s *TokenBucket2) push() {
	for {
		time.Sleep(s.cycle)
		select {
		case s.ch <- struct{}{}:
		default:
		}
	}
}

func main() {
	//tb := NewTokenBucket2(10, 29*time.Millisecond)
	tb := NewTokenBucket(10, 29*time.Millisecond)
	wg := sync.WaitGroup{}
	wg.Add(100)

	var count int64
	var failCount int64

	for i := 0; i < 100; i++ {
		go func() {
			defer wg.Done()

			if tb.Allow() {
				atomic.AddInt64(&count, 1)
			} else {
				atomic.AddInt64(&failCount, 1)
			}
		}()
	}

	wg.Wait()

	time.Sleep(5 * time.Second)
	wg.Add(100)
	for i := 0; i < 100; i++ {
		go func() {
			defer wg.Done()

			if tb.Allow() {
				atomic.AddInt64(&count, 1)
			} else {
				atomic.AddInt64(&failCount, 1)
			}
		}()
	}
	wg.Wait()

	fmt.Println(count)
	fmt.Println(failCount)
}
