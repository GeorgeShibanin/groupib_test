package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

type Broker struct {
	mutex     sync.RWMutex
	selfQueue map[string]*SelfQueue
}

type SelfQueue struct {
	mutex sync.RWMutex
	left  []string
	right []string
}

type HTTPHandler struct {
	broker *Broker
}

func (q *SelfQueue) Append(s string) {
	q.mutex.Lock()
	q.left = append(q.left, s)
	q.mutex.Unlock()
}

func (q *SelfQueue) Return() string {
	if len(q.left) == 0 && len(q.right) == 0 {
		return ""
	}
	if len(q.right) > 0 {
		lastVal := q.right[len(q.right)-1]
		q.right = q.right[:len(q.right)-1]
		return lastVal
	}
	q.mutex.Lock()
	defer q.mutex.Unlock()
	for i := len(q.left) - 1; i >= 0; i-- {
		q.right = append(q.right, q.left[i])
	}
	q.left = make([]string, 0)
	return q.Return()
}

func InitSelfQueue() *SelfQueue {
	return &SelfQueue{
		left:  make([]string, 0),
		right: make([]string, 0),
	}
}

func (b *Broker) ApplyQuery(q string, wait time.Duration) (string, error) {
	b.mutex.RLock()
	defer b.mutex.RUnlock()
	ctx, cancel := context.WithTimeout(context.Background(), wait)
	defer cancel()
	var val string
	for {
		if b.selfQueue[q] == nil {
			if ctx.Err() != nil {
				return "", fmt.Errorf("wrong query name")
			}
			continue
		}
		val = b.selfQueue[q].Return()
		if val != "" || ctx.Err() != nil {
			break
		}
	}
	log.Println("breaks here", val)
	return val, nil
}

func (b *Broker) ApplySelfQueue(v string, q string) {
	log.Println(v, q)
	b.mutex.RLock()
	queue, ok := b.selfQueue[q]
	b.mutex.RUnlock()
	if !ok || queue == nil {
		fmt.Println("self query initialization:", q)
		b.mutex.Lock()
		queue = InitSelfQueue()
		b.selfQueue[q] = queue
		b.mutex.Unlock()
	}
	b.selfQueue[q].Append(v)
	return
}

func (h *HTTPHandler) HandlePutQueue(rw http.ResponseWriter, r *http.Request) {
	log.Println("HandlePutQueue")
	rw.Header().Set("Content-Type", "application/json")
	queryName := strings.Split(r.URL.Path, "/")
	value := r.URL.Query().Get("v")

	if len(queryName) < 2 || value == "" {
		http.Error(rw, "", http.StatusBadRequest)
		return
	}

	h.broker.ApplySelfQueue(value, queryName[1])
}

func (h *HTTPHandler) HandleGetQueue(rw http.ResponseWriter, r *http.Request) {
	queryName := strings.Split(r.URL.Path, "/")
	timeout := r.URL.Query().Get("timeout")
	if timeout == "" {
		timeout = "0"
	}
	seconds, err := time.ParseDuration(timeout + "s")
	if err != nil {
		http.Error(rw, "", http.StatusBadRequest)
		return
	}

	if len(queryName) < 2 {
		http.Error(rw, "", http.StatusBadRequest)
		return
	}

	resultFromQuery, err := h.broker.ApplyQuery(queryName[1], seconds)
	if resultFromQuery == "" || err != nil {
		rw.WriteHeader(404)
		return
	}
	_, err = rw.Write([]byte(resultFromQuery))
	if err != nil {
		rw.WriteHeader(404)
		return
	}
}

func (h *HTTPHandler) HandleQueue(rw http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		h.HandleGetQueue(rw, r)
		return
	}
	if r.Method == "PUT" {
		h.HandlePutQueue(rw, r)
		return
	}
	http.Error(rw, "", http.StatusNotFound)
}

func main() {
	handler := HTTPHandler{
		broker: &Broker{
			selfQueue: make(map[string]*SelfQueue),
		},
	}
	http.HandleFunc("/", handler.HandleQueue)

	err := http.ListenAndServe("localhost:8080", nil)
	if err != nil {
		return
	}
}
