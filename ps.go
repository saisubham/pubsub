package main

import (
	"fmt"
	"sync"
	"time"
)

type Pubsub struct {
	mu     sync.Mutex
	subs   map[string][]chan string
	closed bool
}

func NewPubsub() *Pubsub {
	return &Pubsub{subs: make(map[string][]chan string)}
}

func (ps *Pubsub) Subscribe(topic string) <-chan string {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	ch := make(chan string, 1)
	ps.subs[topic] = append(ps.subs[topic], ch)
	return ch
}

func (ps *Pubsub) Publish(topic string, msg string) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if ps.closed {
		return
	}

	for _, ch := range ps.subs[topic] {
		go func(ch chan string) {
			ch <- msg
		}(ch)
	}
}

func (ps *Pubsub) Close() {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if !ps.closed {
		ps.closed = true
		for _, subs := range ps.subs {
			for _, ch := range subs {
				close(ch)
			}
		}
	}
}

func main() {
	ps := NewPubsub()
	ch1 := ps.Subscribe("foo")
	ch2 := ps.Subscribe("foo")
	ch3 := ps.Subscribe("bar")

	listener := func(name string, ch <-chan string) {
		for i := range ch {
			fmt.Printf("[%s] got %s\n", name, i)
			time.Sleep(50 * time.Millisecond)
		}
	}
	go listener("1", ch1)
	go listener("2", ch2)
	go listener("3", ch3)

	pub := func(topic string, msg string) {
		fmt.Printf("Publishing @%s: %s\n", topic, msg)
		ps.Publish(topic, msg)
		time.Sleep(1 * time.Millisecond)
	}

	time.Sleep(50 * time.Millisecond)
	pub("foo", "foo 1")
	pub("buzz", "buzz 1")
	pub("foo", "foo 2")
	pub("bar", "bar 1")
	pub("bar", "bar 2")
	pub("foo", "foo 3")

	time.Sleep(50 * time.Millisecond)
	ps.Close()
	time.Sleep(50 * time.Millisecond)
}
