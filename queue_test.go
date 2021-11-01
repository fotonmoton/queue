package queue

import (
	"log"
	"sync"
	"testing"
	"time"
)

func TestPut(t *testing.T) {
	q := NewQueue()
	q.Put(1)
	if q.Get().(int) != 1 {
		t.Fatal("wrong item in queue")
	}
}

func TestPutMultiple(t *testing.T) {
	q := NewQueue()
	q.Put(1)
	q.Put(2)
	first, second := q.Get().(int), q.Get().(int)
	if first != 1 || second != 2 {
		t.Fatal("wrong item in queue or wrong order")
	}
}

func TestEmptyGet(t *testing.T) {

	result := func() chan Item {
		q := NewQueue()
		c := make(chan Item)
		go func() { c <- q.Get() }()
		return c
	}

	select {
	case <-result():
		log.Fatal("empty queue should block")
	case <-time.After(time.Millisecond):
	}
}

func TestGetMany(t *testing.T) {
	q := NewQueue()

	result2 := func() chan []Item {
		c := make(chan []Item)
		go func() { c <- q.GetMany(2) }()
		return c
	}

	q.Put(1)

	select {
	case <-result2():
		log.Fatal("GetMany should block if not enough items in queue")
	case <-time.After(time.Millisecond):
	}

	// this call unblocks first GetMany call and empties queue
	q.Put(2)
	// Put enough items in queue for result2 not to block
	q.Put(3)
	q.Put(4)

	select {
	case res := <-result2():
		third, fourth := res[0].(int), res[1].(int)
		if third != 3 || fourth != 4 {
			t.Fatal("wrong item in queue or wrong order")
		}
	case <-time.After(time.Millisecond):
		log.Fatal("GetMany shouldn't block when queue has enough items")
	}
}

func TestConcurrent(t *testing.T) {
	q := NewQueue()
	wg := sync.WaitGroup{}
	sum := 0

	wg.Add(2000)

	for i := 0; i < 1000; i++ {
		go func(i int) {
			defer wg.Done()
			q.Put(i)
		}(i)
	}

	for i := 0; i < 1000; i++ {
		go func() {
			defer wg.Done()
			sum += q.Get().(int)
		}()
	}

	wg.Wait()

	if sum != 1000*(0+999)/2 {
		log.Fatalf("data race. Sum: %v", sum)
	}
}
