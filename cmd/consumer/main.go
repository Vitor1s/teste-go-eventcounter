package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	amqp091 "github.com/rabbitmq/amqp091-go"
	eventcounter "github.com/reb-felipe/eventcounter/pkg"
)

// Configurações
type Config struct {
	AMQPUrl   string
	Queue     string
	Exchange  string
	OutputDir string
}

// Função para carregar configurações
func loadConfig() *Config {
	var (
		amqpUrl   = flag.String("amqp-url", getEnv("AMQP_URL", "amqp://guest:guest@localhost:5672"), "URL do RabbitMQ")
		queue     = flag.String("queue", getEnv("QUEUE_NAME", "eventcountertest"), "Nome da fila")
		exchange  = flag.String("exchange", getEnv("EXCHANGE_NAME", "eventcountertest"), "Nome do exchange")
		outputDir = flag.String("output-dir", getEnv("OUTPUT_DIR", "."), "Diretório de saída dos arquivos JSON")
	)
	flag.Parse()

	return &Config{
		AMQPUrl:   *amqpUrl,
		Queue:     *queue,
		Exchange:  *exchange,
		OutputDir: *outputDir,
	}
}

// Função auxiliar para obter variáveis de ambiente com valor padrão
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

type MensagemProcessada struct {
    UserID    string
    EventType eventcounter.EventType
    MessageID string
}

// Implementação do Consumer
type EventConsumer struct {
    contadores map[eventcounter.EventType]map[string]int
    mu         sync.Mutex
    processado map[string]bool
}

func NewEventConsumer() *EventConsumer {
    return &EventConsumer{
        contadores: map[eventcounter.EventType]map[string]int{
            eventcounter.EventCreated: make(map[string]int),
            eventcounter.EventUpdated: make(map[string]int),
            eventcounter.EventDeleted: make(map[string]int),
        },
        processado: make(map[string]bool),
    }
}

// Implementação da interface Consumer
func (ec *EventConsumer) Created(ctx context.Context, uid string) error {
    return ec.incrementCounter(eventcounter.EventCreated, uid)
}

func (ec *EventConsumer) Updated(ctx context.Context, uid string) error {
    return ec.incrementCounter(eventcounter.EventUpdated, uid)
}

func (ec *EventConsumer) Deleted(ctx context.Context, uid string) error {
    return ec.incrementCounter(eventcounter.EventDeleted, uid)
}

func (ec *EventConsumer) incrementCounter(eventType eventcounter.EventType, userID string) error {
    ec.mu.Lock()
    defer ec.mu.Unlock()
    
    if ec.contadores[eventType] == nil {
        ec.contadores[eventType] = make(map[string]int)
    }
    
    ec.contadores[eventType][userID]++
    log.Printf("Incrementado contador: %s para usuário %s = %d", 
        eventType, userID, ec.contadores[eventType][userID])
    
    return nil
}

// Parseia routing key no formato: user_id.event.event_type
func parseRoutingKey(routingKey string) (userID string, eventType eventcounter.EventType, err error) {
    parts := strings.Split(routingKey, ".")
    if len(parts) != 3 || parts[1] != "event" {
        return "", "", fmt.Errorf("formato de routing key inválido: %s", routingKey)
    }
    
    userID = parts[0]
    eventType = eventcounter.EventType(parts[2])
    
    // Validar se é um tipo de evento válido
    switch eventType {
    case eventcounter.EventCreated, eventcounter.EventUpdated, eventcounter.EventDeleted:
        return userID, eventType, nil
    default:
        return "", "", fmt.Errorf("tipo de evento inválido: %s", eventType)
    }
}

func (ec *EventConsumer) saveCountersToFiles(outputDir string) error {
    ec.mu.Lock()
    defer ec.mu.Unlock()
    
    if err := os.MkdirAll(outputDir, 0755); err != nil {
        return fmt.Errorf("erro ao criar diretório %s: %v", outputDir, err)
    }
    
    for eventType, userCounts := range ec.contadores {
        if len(userCounts) == 0 {
            continue
        }
        
        filename := fmt.Sprintf("%s/%s.json", outputDir, eventType)
        file, err := os.Create(filename)
        if err != nil {
            log.Printf("Erro ao criar arquivo %s: %v", filename, err)
            continue
        }
        
        jsonData, err := json.MarshalIndent(userCounts, "", "  ")
        if err != nil {
            log.Printf("Erro ao serializar dados para %s: %v", filename, err)
            file.Close()
            continue
        }
        
        _, err = file.Write(jsonData)
        file.Close()
        
        if err != nil {
            log.Printf("Erro ao escrever arquivo %s: %v", filename, err)
        } else {
            log.Printf("Arquivo %s salvo com sucesso", filename)
        }
    }
    
    return nil
}

func main() {
    // Carregar configurações
    config := loadConfig()
    
    log.Printf("Iniciando consumer com configurações:")
    log.Printf("AMQP URL: %s", config.AMQPUrl)
    log.Printf("Queue: %s", config.Queue)
    log.Printf("Exchange: %s", config.Exchange)
    log.Printf("Output Dir: %s", config.OutputDir)
    
    // Conectar ao RabbitMQ
    conn, err := amqp091.Dial(config.AMQPUrl)
    if err != nil {
        log.Fatalf("Erro ao conectar RabbitMQ: %v", err)
    }
    defer conn.Close()
    
    channel, err := conn.Channel()
    if err != nil {
        log.Fatalf("Erro ao abrir canal: %v", err)
    }
    defer channel.Close()
    
    // Configurar consumer
    msgs, err := channel.Consume(
        config.Queue,       // queue
        "",                 // consumer
        false,              // auto-ack (false para fazer ack manual)
        false,              // exclusive
        false,              // no-local
        false,              // no-wait
        nil,                // args
    )
    if err != nil {
        log.Fatalf("Erro ao configurar consumer: %v", err)
    }
    
    // Criar consumer de eventos
    eventConsumer := NewEventConsumer()
    
    // Criar canais para cada tipo de evento
    canalCriado := make(chan MensagemProcessada, 100)
    canalAtualizado := make(chan MensagemProcessada, 100)
    canalDeletado := make(chan MensagemProcessada, 100)
    
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    
    var wg sync.WaitGroup
    
    wg.Add(1)
    go func() {
        defer wg.Done()
        for {
            select {
            case msg := <-canalCriado:
                if err := eventConsumer.Created(ctx, msg.UserID); err != nil {
                    log.Printf("Erro ao processar created: %v", err)
                }
            case <-ctx.Done():
                return
            }
        }
    }()
    
    wg.Add(1)
    go func() {
        defer wg.Done()
        for {
            select {
            case msg := <-canalAtualizado:
                if err := eventConsumer.Updated(ctx, msg.UserID); err != nil {
                    log.Printf("Erro ao processar updated: %v", err)
                }
            case <-ctx.Done():
                return
            }
        }
    }()
    
    wg.Add(1)
    go func() {
        defer wg.Done()
        for {
            select {
            case msg := <-canalDeletado:
                if err := eventConsumer.Deleted(ctx, msg.UserID); err != nil {
                    log.Printf("Erro ao processar deleted: %v", err)
                }
            case <-ctx.Done():
                return
            }
        }
    }()

    // Timer para timeout de 5 segundos
    timer := time.NewTimer(5 * time.Second)
    timer.Stop() // Para inicialmente
    
    log.Println("Consumer iniciado, aguardando mensagens...")
    
    // Loop principal
    for {
        select {
        case delivery := <-msgs:
            // Reset timer a cada nova mensagem
            if !timer.Stop() {
                select {
                case <-timer.C:
                default:
                }
            }
            timer.Reset(5 * time.Second)
            
            // Parsear routing key
            userID, eventType, err := parseRoutingKey(delivery.RoutingKey)
            if err != nil {
                log.Printf("Erro ao parsear routing key: %v", err)
                delivery.Nack(false, false) // Rejeitar mensagem
                continue
            }
            
            // Parsear body da mensagem para pegar o ID
            var msgBody map[string]string
            if err := json.Unmarshal(delivery.Body, &msgBody); err != nil {
                log.Printf("Erro ao parsear body da mensagem: %v", err)
                delivery.Nack(false, false)
                continue
            }
            
            messageID, exists := msgBody["id"]
            if !exists {
                log.Printf("ID da mensagem não encontrado")
                delivery.Nack(false, false)
                continue
            }
            
            // Verificar se já foi processada (evitar duplicatas)
            eventConsumer.mu.Lock()
            if eventConsumer.processado[messageID] {
                log.Printf("Mensagem %s já foi processada, ignorando", messageID)
                eventConsumer.mu.Unlock()
                delivery.Ack(false)
                continue
            }
            eventConsumer.processado[messageID] = true
            eventConsumer.mu.Unlock()
            
            processedMsg := MensagemProcessada{
                UserID:    userID,
                EventType: eventType,
                MessageID: messageID,
            }
            
            // Enviar para o canal apropriado
            switch eventType {
            case eventcounter.EventCreated:
                canalCriado <- processedMsg
            case eventcounter.EventUpdated:
                canalAtualizado <- processedMsg
            case eventcounter.EventDeleted:
                canalDeletado <- processedMsg
            }
            
            // ACK da mensagem
            delivery.Ack(false)
            log.Printf("Mensagem processada: %s, User: %s, Event: %s", 
                messageID, userID, eventType)
        
        case <-timer.C:
            log.Println("Timeout de 5 segundos atingido, finalizando...")
            cancel()
            
            close(canalCriado)
            close(canalAtualizado)
            close(canalDeletado)
            
            wg.Wait()
            
            if err := eventConsumer.saveCountersToFiles(config.OutputDir); err != nil {
                log.Printf("Erro ao salvar arquivos: %v", err)
            }
            
            log.Println("Consumer finalizado com sucesso!")
            return
        }
    }
}