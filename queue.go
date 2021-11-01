// This is queue data structure designed
// by Bryan C. Mills (https://github.com/bcmills)
// and showcased in talk "Rethinking Classical Concurrency Patterns"
// (https://www.youtube.com/watch?v=5zXAHh5tJqQ)
package queue

type Item struct {
	Val interface{}
}

// This type represents request for content from queue.
// When n number of items become avaliable in a queue
// they will be send to ch channel for caller to consume.
type waiter struct {
	n  int
	ch chan []Item
}

// Holds all requests for items and all items
type state struct {
	waiters []waiter
	items   []Item
}

// This is the interesting part. To make this queue goroutine-safe state of the
// queue should be owned by one goroutine at a time. One option is to use locks
// but here channel used as synchronization mechanism. Queue holds state chanel
// with capacity of 1. Because first read from state channel will be nonblocking
// and all subsequent calls will wait until state value become avaliable this
// prevents data races and makes this type safe to use concurrently.
// What an ingenious design!
type Queue struct {
	state chan state
}

func NewQueue() *Queue {
	queue := Queue{state: make(chan state, 1)}
	// Shared state that will be passed
	// betwee all goroutines to synchronize access
	queue.state <- state{}
	return &queue
}

func (q *Queue) Put(item Item) {
	// This read from state channel works like mutex.Lock() call.
	// No one can modify sate until we put it back to channel
	state := <-q.state
	state.items = append(state.items, item)
	// If state has waiting requests and queue has enough items
	// for first request we send it to waiter channel
	for len(state.waiters) > 0 {
		waiter := state.waiters[0]
		if waiter.n < len(state.items) {
			break
		}
		// This channel is blocking Get* calls until we put
		// requested number of items to it
		waiter.ch <- state.items[:waiter.n:waiter.n]
		state.items = state.items[waiter.n:]
		state.waiters = state.waiters[1:]
	}
	// Release state for another goroutines to use
	q.state <- state
}

func (q *Queue) GetMany(n int) []Item {
	// Acquire exclusive right to modify state
	state := <-q.state
	// We can return items right away without creating a waiter
	if len(state.waiters) == 0 && len(state.items) >= n {
		items := state.items[:n:n]
		state.items = state.items[n:]
		q.state <- state
		return items
	}
	ch := make(chan []Item)
	state.waiters = append(state.waiters, waiter{n, ch})
	q.state <- state
	// Wait for Put call to push items to ch channel
	return <-ch
}

func (q *Queue) Get() Item {
	return q.GetMany(1)[0]
}
