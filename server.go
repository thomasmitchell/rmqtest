package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

type ServerConfig struct {
	Port                     uint16
	RabbitMQConnectionString string
}

func StartServer(conf *ServerConfig) error {
	queue, err := NewQueue(conf.RabbitMQConnectionString)
	if err != nil {
		return fmt.Errorf("Error setting up queue: %s", err.Error())
	}

	router := mux.NewRouter()
	messageAPI := NewAPIMessage(queue)
	router.Handle("/messages", messageAPI).Methods("GET", "POST")

	return http.ListenAndServe(fmt.Sprintf(":%s", conf.Port), router)
}

type APIMessage struct {
	queue *Queue
}

func NewAPIMessage(queue *Queue) *APIMessage {
	return &APIMessage{
		queue: queue,
	}
}

func (a *APIMessage) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		a.get(w, r)
	case http.MethodPost:
		a.post(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (a *APIMessage) get(w http.ResponseWriter, r *http.Request) {
	delivery, err := a.queue.Dequeue()
	if delivery == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	message := delivery.Body
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(message)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not return message in response")
		if err := Nack(delivery); err != nil {
			panic("Could not nack message")
		}
		return
	}

	err = Ack(delivery)
	if err != nil {
		panic("Could not ack message")
	}
}

func (a *APIMessage) post(w http.ResponseWriter, r *http.Request) {
	requestBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	r.Body.Close()

	err = a.queue.Enqueue(requestBody)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
