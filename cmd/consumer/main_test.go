package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	eventcounter "github.com/reb-felipe/eventcounter/pkg"
)

func TestNewEventConsumer(t *testing.T) {
	consumer := NewEventConsumer()
	
	if consumer == nil {
		t.Fatal("NewEventConsumer retornou nil")
	}
	
	if consumer.contadores == nil {
		t.Fatal("Contadores não foram inicializados")
	}
	
	if consumer.processado == nil {
		t.Fatal("Processado map não foi inicializado")
	}
	
	expectedTypes := []eventcounter.EventType{
		eventcounter.EventCreated,
		eventcounter.EventUpdated,
		eventcounter.EventDeleted,
	}
	
	for _, eventType := range expectedTypes {
		if _, exists := consumer.contadores[eventType]; !exists {
			t.Errorf("Contador para evento %s não foi inicializado", eventType)
		}
	}
}

func TestEventConsumerCreated(t *testing.T) {
	consumer := NewEventConsumer()
	ctx := context.Background()
	userID := "test_user"
	
	err := consumer.Created(ctx, userID)
	if err != nil {
		t.Fatalf("Erro ao processar evento Created: %v", err)
	}
	
	count := consumer.contadores[eventcounter.EventCreated][userID]
	if count != 1 {
		t.Errorf("Esperado contador = 1, obtido = %d", count)
	}
	
	err = consumer.Created(ctx, userID)
	if err != nil {
		t.Fatalf("Erro ao processar segundo evento Created: %v", err)
	}
	
	count = consumer.contadores[eventcounter.EventCreated][userID]
	if count != 2 {
		t.Errorf("Esperado contador = 2, obtido = %d", count)
	}
}

func TestEventConsumerUpdated(t *testing.T) {
	consumer := NewEventConsumer()
	ctx := context.Background()
	userID := "test_user"
	
	err := consumer.Updated(ctx, userID)
	if err != nil {
		t.Fatalf("Erro ao processar evento Updated: %v", err)
	}
	
	count := consumer.contadores[eventcounter.EventUpdated][userID]
	if count != 1 {
		t.Errorf("Esperado contador = 1, obtido = %d", count)
	}
}

func TestEventConsumerDeleted(t *testing.T) {
	consumer := NewEventConsumer()
	ctx := context.Background()
	userID := "test_user"
	
	err := consumer.Deleted(ctx, userID)
	if err != nil {
		t.Fatalf("Erro ao processar evento Deleted: %v", err)
	}
	
	count := consumer.contadores[eventcounter.EventDeleted][userID]
	if count != 1 {
		t.Errorf("Esperado contador = 1, obtido = %d", count)
	}
}

func TestEventConsumerMultipleUsers(t *testing.T) {
	consumer := NewEventConsumer()
	ctx := context.Background()
	
	users := []string{"user1", "user2", "user3"}
	
	// Adicionar eventos para diferentes usuários
	for _, user := range users {
		consumer.Created(ctx, user)
		consumer.Updated(ctx, user)
		consumer.Deleted(ctx, user)
	}
	
	// Verificar contadores
	for _, user := range users {
		for _, eventType := range []eventcounter.EventType{
			eventcounter.EventCreated,
			eventcounter.EventUpdated,
			eventcounter.EventDeleted,
		} {
			count := consumer.contadores[eventType][user]
			if count != 1 {
				t.Errorf("Usuário %s, evento %s: esperado = 1, obtido = %d", user, eventType, count)
			}
		}
	}
}

func TestParseRoutingKey(t *testing.T) {
	tests := []struct {
		routingKey    string
		expectedUser  string
		expectedEvent eventcounter.EventType
		shouldError   bool
	}{
		{"user123.event.created", "user123", eventcounter.EventCreated, false},
		{"user456.event.updated", "user456", eventcounter.EventUpdated, false},
		{"user789.event.deleted", "user789", eventcounter.EventDeleted, false},
		{"invalid.format", "", "", true},
		{"user.wrong.created", "", "", true},
		{"user.event.invalid", "", "", true},
		{"user.event", "", "", true},
		{"", "", "", true},
	}
	
	for _, test := range tests {
		userID, eventType, err := parseRoutingKey(test.routingKey)
		
		if test.shouldError {
			if err == nil {
				t.Errorf("Esperado erro para routing key '%s', mas não houve erro", test.routingKey)
			}
		} else {
			if err != nil {
				t.Errorf("Erro inesperado para routing key '%s': %v", test.routingKey, err)
				continue
			}
			
			if userID != test.expectedUser {
				t.Errorf("Routing key '%s': esperado user = '%s', obtido = '%s'", 
					test.routingKey, test.expectedUser, userID)
			}
			
			if eventType != test.expectedEvent {
				t.Errorf("Routing key '%s': esperado event = '%s', obtido = '%s'", 
					test.routingKey, test.expectedEvent, eventType)
			}
		}
	}
}

func TestSaveCountersToFiles(t *testing.T) {
	consumer := NewEventConsumer()
	ctx := context.Background()
	
	// Criar dados de teste
	consumer.Created(ctx, "user1")
	consumer.Created(ctx, "user2")
	consumer.Updated(ctx, "user1")
	consumer.Deleted(ctx, "user3")
	
	// Criar diretório temporário
	tempDir := filepath.Join(os.TempDir(), "test_output")
	defer os.RemoveAll(tempDir)
	
	// Salvar arquivos
	err := consumer.saveCountersToFiles(tempDir)
	if err != nil {
		t.Fatalf("Erro ao salvar arquivos: %v", err)
	}
	
	// Verificar se os arquivos foram criados
	expectedFiles := []string{
		filepath.Join(tempDir, "created.json"),
		filepath.Join(tempDir, "updated.json"),
		filepath.Join(tempDir, "deleted.json"),
	}
	
	for _, file := range expectedFiles {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			t.Errorf("Arquivo %s não foi criado", file)
		}
	}
}

func TestLoadConfig(t *testing.T) {
	// Testar valores padrão
	config := loadConfig()
	
	if config.AMQPUrl == "" {
		t.Error("AMQPUrl não pode estar vazio")
	}
	
	if config.Queue == "" {
		t.Error("Queue não pode estar vazio")
	}
	
	if config.Exchange == "" {
		t.Error("Exchange não pode estar vazio")
	}
	
	if config.OutputDir == "" {
		t.Error("OutputDir não pode estar vazio")
	}
}

func TestGetEnv(t *testing.T) {
	// Testar valor padrão
	result := getEnv("NONEXISTENT_VAR", "default")
	if result != "default" {
		t.Errorf("Esperado 'default', obtido '%s'", result)
	}
	
	// Testar com variável definida
	os.Setenv("TEST_VAR", "test_value")
	defer os.Unsetenv("TEST_VAR")
	
	result = getEnv("TEST_VAR", "default")
	if result != "test_value" {
		t.Errorf("Esperado 'test_value', obtido '%s'", result)
	}
}

func TestProcessedMessageDeduplication(t *testing.T) {
	consumer := NewEventConsumer()
	messageID := "test_message_123"
	
	// Primeira vez - deve processar
	consumer.mu.Lock()
	wasProcessed := consumer.processado[messageID]
	if !wasProcessed {
		consumer.processado[messageID] = true
	}
	consumer.mu.Unlock()
	
	if wasProcessed {
		t.Error("Mensagem não deveria estar marcada como processada inicialmente")
	}
	
	// Segunda vez - deve estar processada
	consumer.mu.Lock()
	wasProcessed = consumer.processado[messageID]
	consumer.mu.Unlock()
	
	if !wasProcessed {
		t.Error("Mensagem deveria estar marcada como processada")
	}
}

// Benchmark para testar performance
func BenchmarkEventConsumerCreated(b *testing.B) {
	consumer := NewEventConsumer()
	ctx := context.Background()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		consumer.Created(ctx, "test_user")
	}
}

func BenchmarkParseRoutingKey(b *testing.B) {
	routingKey := "user123.event.created"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = parseRoutingKey(routingKey)
	}
}
