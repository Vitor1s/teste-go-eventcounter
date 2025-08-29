package main

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "os"
    "strings"
    "sync"
    "time"

    amqp091 "github.com/rabbitmq/amqp091-go"
    eventcounter "github.com/reb-felipe/eventcounter/pkg"
)


//Aqui vou criar a estrutura das mensagems processadas com type e body
    type ProcessoMensagens struct {
		UserID   string 
		EventType eventcounter.EventType
		MessageID string


}

//Aqui a implementaçao do Consumer
type EventConsumer struct{
	counters  map[eventcounter.EventType]map[string]int
	mu        sync.Mutex
	processed map[string]bool  //Com isso evita duplicidade de mensagens
}

// Aqui vou criar a funçao que vai gerar os eventos do Consumer 

func NewEventConsumer() *EventConsumer {
	return &EventConsumer{
		counters: map[eventcounter.EventType]map[strings]int{
			eventcounter.EventCreated: make(map[string]int),
			eventcounter.EventDeleted: make(map[string]int),
			eventcounter.EventUpdated: make(map[string]int),
		},
		processed: make(map[string]bool),
	}
}


// vamos criar as funçoes que vao consumir as mensagens

func (ec *Consumer) Created(ctx context.Context, uid string) error {
	return ec.incrementCounter(eventcounter.EventCreated, uid)
}

func (ec *Consumer) Deleted(ctx context.Context, uid string) error {
	return ec.incrementCounter(eventcounter.EventDeleted, uid)
}

func (ec *Consumer) Updated(ctx context.Context, uid string) error {
	return ec.incrementCounter(eventcounter.EventUpdated, uid)
}

func (ec *Consumer) incrementCounter(eventType eventcounter.EventType, userID string) error {
    ec.mu.Lock()
    defer ec.mu.Unlock()
    
    if ec.counters[eventType] == nil {
        ec.counters[eventType] = make(map[string]int)
    }
    
    ec.counters[eventType][userID]++
    log.Printf("Incrementado contador: %s para usuário %s = %d", 
        eventType, userID, ec.counters[eventType][userID])
    
    return nil
}


func parseRoutingKey

func main() {
	

}