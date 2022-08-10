package pkg

import (
	"fmt"
	"sync"
)

type Topic struct {
	mutex     sync.Mutex
	messages  *MessagesQueue
	listeners *ListenersQueue
	locked    bool
}

func InitTopic() *Topic {
	return &Topic{
		messages: &MessagesQueue{
			mutex: sync.Mutex{},
			left:  make([]string, 0),
			right: make([]string, 0),
		},
		listeners: &ListenersQueue{
			mutex: sync.Mutex{},
			left:  make([]*Listener, 0),
			right: make([]*Listener, 0),
		},
	}
}

func (t *Topic) AddListener(listener *Listener) {
	if !t.locked {
		go t.initReader()
	}
	t.listeners.Append(listener)
}

func (t *Topic) AppendMessage(message string) {
	if !t.locked {
		go t.initReader()
	}
	t.messages.Append(message)
}

func (t *Topic) initReader() {
	t.mutex.Lock()
	t.locked = true
	fmt.Println("start iniReader")
	defer func() {
		fmt.Println("end iniReader")
		t.mutex.Unlock()
		t.locked = false
	}()
	for {
		message := t.messages.Last()
		if message == "" {
			return
		}

		for {
			listener := t.listeners.Last()
			if listener == nil {
				return
			}
			sent := listener.Send(message)
			t.listeners.Pop()
			if sent {
				break
			}
		}
		t.messages.Pop()
	}
}
