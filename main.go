package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

type PutResponseData struct {
	Response string `json:""`
}

type Query struct {
	mutex sync.Mutex
	query chan string
}

type Broker struct {
	mutex     sync.RWMutex
	queries   map[string]*Query
	selfQueue map[string]*SelfQueue
}

type SelfQueue struct {
	mutex sync.Mutex
	left  []string
	right []string
}

type HTTPHandler struct {
	broker *Broker
}

func (q *SelfQueue) Append(s string) {
	q.left = append(q.left, s)
}

func (q *SelfQueue) Return() string {
	log.Println("len left", len(q.left), " len rignt ", len(q.right))
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
		log.Println("apeend this to right", q.left[i])
	}
	q.left = make([]string, 0)
	return q.Return()
}

func InitQuery() *Query {
	return &Query{query: make(chan string)}
}

func InitSelfQueue() *SelfQueue {
	return &SelfQueue{
		left:  make([]string, 0),
		right: make([]string, 0),
	}
}

func (q *Query) Add(v string) {
	q.query <- v
}

func (q *Query) GetChannel() chan string {
	return q.query
}

func (b *Broker) initReader(q *Query) {
	fmt.Println("initReader: ", q)
	q.mutex.Lock()
	defer q.mutex.Unlock()
	log.Println("start For in initReader")
	for {
		select {
		case s := <-q.GetChannel():
			fmt.Println("value out: ", s)
		}

	}
}

func (b *Broker) ApplyQuery(q string, wait time.Duration) (string, error) {
	b.mutex.RLock()
	defer b.mutex.RUnlock()
	ctx, cancel := context.WithTimeout(context.Background(), wait)
	defer cancel()
	query, ok := b.queries[q]
	if !ok || query == nil {
		fmt.Println("query initialization:", q)
		query = InitQuery()
		b.queries[q] = query
	}
	if !ok {
		go b.initReader(query)
	}
	var val string
	for {
		val = b.selfQueue[q].Return()
		if val != "" || ctx.Err() != nil {
			query.Add(val)
			break
		}
	}
	log.Println("breaks here", val)
	return val, nil
}

func (b *Broker) ApplySelfQueue(v string, q string) {
	log.Println(v, q)
	b.mutex.Lock()
	defer b.mutex.Unlock()
	queue, ok := b.selfQueue[q]
	if !ok || queue == nil {
		fmt.Println("self query initialization:", q)
		queue = InitSelfQueue()
		b.selfQueue[q] = queue
	}

	b.selfQueue[q].Append(v)
	log.Println("add this to SelfQueue", v)

	return
}

func (h *HTTPHandler) HandlePutQueue(rw http.ResponseWriter, r *http.Request) {
	log.Println("HandlePutQueue")
	rw.Header().Set("Content-Type", "application/json")
	queryName := strings.Split(r.URL.Path, "/")
	value := r.URL.Query().Get("v")

	if len(queryName) == 0 && value == "" {
		http.Error(rw, "invalid query params", http.StatusBadRequest)
		return
	}

	h.broker.ApplySelfQueue(value, queryName[1])

	response := PutResponseData{
		Response: "OK",
	}
	rawResponse, err := json.Marshal(response)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	rw.Header().Set("Content-Type", "application/json")
	_, err = rw.Write(rawResponse)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}
}

func (h *HTTPHandler) HandleGetQueue(rw http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(rw, "HandleGetQueue")
	queryName := strings.Split(r.URL.Path, "/")
	timeout := r.URL.Query().Get("timeout")
	if timeout == "" {
		timeout = "0"
	}
	seconds, _ := time.ParseDuration(timeout + "s")

	resultFromQuery, err := h.broker.ApplyQuery(queryName[1], seconds)
	if err != nil {
		rw.WriteHeader(404)
		return
	}

	response := PutResponseData{
		Response: resultFromQuery,
	}
	rawResponse, err := json.Marshal(response)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	rw.Header().Set("Content-Type", "application/json")
	_, err = rw.Write(rawResponse)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}
}

func (h *HTTPHandler) HandleQueue(rw http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		h.HandleGetQueue(rw, r)
	}
	if r.Method == "PUT" {
		h.HandlePutQueue(rw, r)
	}
	http.Error(rw, "unexpected method", http.StatusNotFound)
}

func main() {
	handler := HTTPHandler{
		broker: &Broker{
			queries:   make(map[string]*Query),
			selfQueue: make(map[string]*SelfQueue),
		},
	}
	http.HandleFunc("/", handler.HandleQueue)

	err := http.ListenAndServe("localhost:8080", nil)
	if err != nil {
		return
	}
}
