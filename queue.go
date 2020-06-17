package main

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/streadway/amqp"
)

//This doesn't currently handle when a connection error occurs. Like, at all.

type Queue struct {
	conn    *amqp.Connection
	ch      *amqp.Channel
	queue   amqp.Queue
	pubChan chan amqp.Confirmation
	//Locking to using one channel, which is inefficient, but this is just a test
	lock sync.Mutex
}

func NewQueue(connect string) (*Queue, error) {
	const queueName = "rmq-test-app-queue"
	conn, err := amqp.Dial(connect)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	err = ch.Confirm(false)
	if err != nil {
		return nil, err
	}

	pubChan := make(chan amqp.Confirmation, 1)
	pubChan = ch.NotifyPublish(pubChan)

	queue, err := ch.QueueDeclare(queueName, true, false, false, false, nil)
	if err != nil {
		return nil, err
	}

	return &Queue{
		conn:    conn,
		ch:      ch,
		queue:   queue,
		pubChan: pubChan,
	}, nil
}

func (q *Queue) Enqueue(message []byte) error {
	q.lock.Lock()
	defer q.lock.Unlock()
	err := q.ch.Publish("", q.queue.Name, false, false, amqp.Publishing{
		ContentType:  "text/plain",
		Timestamp:    time.Now(),
		DeliveryMode: amqp.Persistent,
		Body:         message,
	})

	if err != nil {
		return err
	}

	confirmation := <-q.pubChan
	if !confirmation.Ack {
		return fmt.Errorf("Message was not ack'ed by remote")
	}

	return nil
}

//Dequeue returns nil if there is no message waiting on the queue
func (q *Queue) Dequeue() (*amqp.Delivery, error) {
	delivery, ok, err := q.ch.Get(q.queue.Name, false)
	if err != nil {
		return nil, err
	}

	if !ok {
		return nil, nil
	}

	return &delivery, nil
}

func retry(fn func() error) error {
	numAttempt := 0
	const maxAttempt = 5
	currentWait := 500 * time.Millisecond
	var err error
	for {
		err = fn()
		if err == nil {
			return nil
		}

		numAttempt++
		if numAttempt == maxAttempt {
			break
		}

		fmt.Fprintf(os.Stderr, "Acknowledgement failed, trying again in %s", currentWait.String())
		time.Sleep(currentWait)
		currentWait *= 2
	}

	return err
}

func Ack(delivery *amqp.Delivery) error {
	return retry(func() error { return delivery.Ack(false) })
}

func Nack(delivery *amqp.Delivery) error {
	return retry(func() error { return delivery.Nack(false, true) })
}
