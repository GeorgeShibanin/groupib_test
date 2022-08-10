package main

import (
	"group-IB/pkg"
	"log"
	"net/http"
	"strings"
	"time"
)

type HTTPHandler struct {
	broker *pkg.Broker
}

func InitHandler() *HTTPHandler {
	return &HTTPHandler{broker: pkg.InitBroker()}
}

func (h *HTTPHandler) HandlePutQueue(rw http.ResponseWriter, r *http.Request) {
	log.Println("HandlePutQueue")
	queryName := strings.Split(r.URL.Path, "/")
	value := r.URL.Query().Get("v")

	if len(queryName) < 2 || value == "" {
		http.Error(rw, "invalid query params", http.StatusBadRequest)
		return
	}

	h.broker.SendMessage(queryName[1], value)
	return
}

func (h *HTTPHandler) HandleGetQueue(rw http.ResponseWriter, r *http.Request) {
	queryName := strings.Split(r.URL.Path, "/")
	timeout := r.URL.Query().Get("timeout")
	if timeout == "" {
		timeout = "0"
	}

	seconds, err := time.ParseDuration(timeout + "s")
	if err != nil {
		http.Error(rw, "invalid timeout", http.StatusBadRequest)
		return
	}

	if len(queryName) < 2 {
		http.Error(rw, "unexpected query", http.StatusBadRequest)
		return
	}

	message, err := h.broker.GetMessage(queryName[1], seconds)
	if err != nil {
		rw.WriteHeader(http.StatusNotFound)
		return
	}
	_, err = rw.Write([]byte(message))
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
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
	handler := InitHandler()
	http.HandleFunc("/", handler.HandleQueue)
	//port := os.Getenv("SERVER_PORT")
	port := "8080"
	err := http.ListenAndServe("localhost:"+port, nil)
	if err != nil {
		return
	}
}
