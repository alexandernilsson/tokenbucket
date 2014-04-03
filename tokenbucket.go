// tokenbucket.go (c) 2014 David Rook - all rights reserved
//
// see http://en.wikipedia.org/wiki/Token_bucket
//  This was once similar to the API at http://godoc.org/github.com/juju/ratelimit
package tokenbucket

import (
	"fmt"
	"log"
	"sync"
	"time"
)

type TokenBucket struct {
	fillInterval  time.Duration
	capacity      float64
	lastCount     float64
	lastCheckTime time.Time
	lock          sync.Mutex
}

func New(fillTime time.Duration, capacity float64) *TokenBucket {
	if fillTime.Nanoseconds() <= int64(0) {
		log.Fatalf("arguments to tokenbucket.New() must be positive\n")
	}
	if capacity <= 0 {
		log.Fatalf("arguments to tokenbucket.New() must be positive\n")
	}
	var tb TokenBucket
	tb.fillInterval = fillTime
	tb.capacity = capacity
	tb.lastCount = capacity
	tb.lastCheckTime = time.Now()
	var lock sync.Mutex
	tb.lock = lock
	return &tb
}

func (t *TokenBucket) Dump() {
	t.lock.Lock()
	fmt.Printf("Interval[%v] Cap[%d] lastCount[%d] lastCheckTime[%v]\n",
		t.fillInterval, t.capacity, t.lastCount, t.lastCheckTime)
	t.lock.Unlock()
}

// Take() returns the time to wait before tokens are available
//  Calling Take commits to take them, cant be put back
func (t *TokenBucket) Take(icount int64) time.Duration {
count := float64(icount)
	now := time.Now()
	t.lock.Lock()
	t.lastCount += float64(now.Sub(t.lastCheckTime).Nanoseconds()) / float64(t.fillInterval.Nanoseconds())
	t.lastCheckTime = now
	if t.lastCount > t.capacity {
		t.lastCount = t.capacity
	}
	t.lastCount -= count
	var delay time.Duration
	if t.lastCount <= 0 {
		delay = time.Duration(-t.lastCount * float64(t.fillInterval.Nanoseconds()))
		//fmt.Printf("Take Delay = %v\n",delay)
	}
	t.lock.Unlock()
	return delay
}

// Wait() blocks until at least count tokens are available then returns -
// which could be immediately if enough tokens are available.
//  Wait() reserves tokens in spite of waiting for possession
func (t *TokenBucket) Wait(icount int64) {
count := float64(icount)
	t.lock.Lock()
	now := time.Now()
	t.lastCount += float64(now.Sub(t.lastCheckTime).Nanoseconds()) / float64(t.fillInterval.Nanoseconds())
	t.lastCheckTime = now
	if t.lastCount > t.capacity {
		t.lastCount = t.capacity
	}
	t.lastCount -= count
	if t.lastCount >= 0 {
		t.lock.Unlock()
		return
	}
	delay := time.Duration(-t.lastCount * float64(t.fillInterval.Nanoseconds()))
	t.lock.Unlock()
	//fmt.Printf("Wait delay = %v\n", delay)
	time.Sleep(delay)
}
