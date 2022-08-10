package pkg

import (
	"fmt"
	"log"
	"sync"
	"time"
)

type Broker struct {
	mutex  sync.RWMutex
	topics map[string]*Topic
}

func InitBroker() *Broker {
	return &Broker{
		mutex:  sync.RWMutex{},
		topics: make(map[string]*Topic),
	}
}

func (b *Broker) initTopic(topicName, message string) {
	fmt.Println("initTopic: ", topicName)
	b.mutex.Lock()
	defer b.mutex.Unlock()

	topic, ok := b.topics[topicName]
	if ok && topic != nil {
		b.topics[topicName].AppendMessage(message)
		return
	}

	topic = InitTopic()
	b.topics[topicName] = topic
	b.topics[topicName].AppendMessage(message)
}

func (b *Broker) SendMessage(topicName, message string) {
	log.Printf("SendMessage: %s:%s", topicName, message)
	b.mutex.RLock()

	queue, ok := b.topics[topicName]
	if !ok || queue == nil {
		b.mutex.RUnlock()
		b.initTopic(topicName, message)
		return
	}
	b.topics[topicName].AppendMessage(message)
	b.mutex.RUnlock()
}

func (b *Broker) GetMessage(queueName string, wait time.Duration) (string, error) {
	log.Printf("GetMessage: %s", queueName)
	b.mutex.RLock()
	defer b.mutex.RUnlock()

	queue, ok := b.topics[queueName]
	if !ok {
		return "", fmt.Errorf("query not found")
	}

	channel := make(chan string)
	quit := make(chan struct{})
	listener := InitListener(channel, quit)
	queue.AddListener(listener)

	select {
	case message := <-channel:
		return message, nil

	case <-time.After(wait):
		close(quit)
		return "", fmt.Errorf("timeout exceed")
	}
}
