package pkg

import (
	"fmt"
	"sync"
)

type MessagesQueue struct {
	mutex sync.Mutex
	left  []string
	right []string
}

func (q *MessagesQueue) Append(s string) {
	fmt.Printf("MessagesQueue Append: %s \n", s)
	q.mutex.Lock()
	defer q.mutex.Unlock()

	q.left = append(q.left, s)
}

func (q *MessagesQueue) Last() string {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	return q.last()
}

func (q *MessagesQueue) last() string {
	fmt.Printf("MessagesQueue Last: left = %v, right = %v \n", len(q.left), len(q.right))
	if len(q.left) == 0 && len(q.right) == 0 {
		return ""
	}

	if len(q.right) > 0 {
		fmt.Printf("MessagesQueue Last: %s \n", q.right[len(q.right)-1])
		return q.right[len(q.right)-1]
	}

	for i := len(q.left) - 1; i >= 0; i-- {
		q.right = append(q.right, q.left[i])
	}
	q.left = make([]string, 0)

	if len(q.right) == 0 {
		return ""
	}
	return q.right[len(q.right)-1]
}

func (q *MessagesQueue) Pop() string {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	lastVal := q.last()
	if lastVal == "" {
		return ""
	}
	q.right = q.right[:len(q.right)-1]
	fmt.Printf("MessagesQueue Return: %s \n", lastVal)
	return lastVal
}
