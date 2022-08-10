package pkg

import (
	"fmt"
	"sync"
)

type ListenersQueue struct {
	mutex sync.Mutex
	left  []*Listener
	right []*Listener
}

func (c *ListenersQueue) Append(l *Listener) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.left = append(c.left, l)
}

func (c *ListenersQueue) Pop() *Listener {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	fmt.Printf("ListenersQueue Return: left = %v, right = %v \n", len(c.left), len(c.right))

	if len(c.left) == 0 && len(c.right) == 0 {
		return nil
	}

	if len(c.right) > 0 {
		lastVal := c.right[len(c.right)-1]
		c.right = c.right[:len(c.right)-1]
		fmt.Printf("ListenersQueue Return: %s\n", lastVal)
		return lastVal
	}

	for i := len(c.left) - 1; i >= 0; i-- {
		c.right = append(c.right, c.left[i])
	}
	c.left = make([]*Listener, 0)

	if len(c.right) == 0 {
		return nil
	}
	fmt.Printf("ListenersQueue Return: %s\n", c.right[len(c.right)-1])
	return c.right[len(c.right)-1]
}

func (c *ListenersQueue) Last() *Listener {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	fmt.Printf("ListenersQueue Last: left = %v, right = %v \n", len(c.left), len(c.right))
	if len(c.left) == 0 && len(c.right) == 0 {
		return nil
	}

	if len(c.right) > 0 {
		fmt.Printf("ListenersQueue Last: %s \n", c.right[len(c.right)-1])
		return c.right[len(c.right)-1]
	}

	for i := len(c.left) - 1; i >= 0; i-- {
		c.right = append(c.right, c.left[i])
	}
	c.left = make([]*Listener, 0)

	if len(c.right) == 0 {
		return nil
	}

	fmt.Printf("ListenersQueue Last: %s \n", c.right[len(c.right)-1])
	return c.right[len(c.right)-1]
}

type Listener struct {
	mutex   sync.Mutex
	channel chan<- string
	quit    <-chan struct{}
}

func InitListener(channel chan string, quit chan struct{}) *Listener {
	return &Listener{
		mutex:   sync.Mutex{},
		channel: channel,
		quit:    quit,
	}
}

func (l *Listener) Send(s string) bool {
	fmt.Printf("ListenersQueue Send: %s \n", s)
	l.mutex.Lock()
	defer l.mutex.Unlock()

	select {
	case l.channel <- s:
		return true
	case <-l.quit:
		return false
	}
}
