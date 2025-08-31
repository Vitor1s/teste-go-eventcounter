EXCHANGE=eventcountertest
AMQP_PORT=5672
AMQP_UI_PORT=15672
CURRENT_DIR=$(shell pwd)

# Targets de build
build-all: build-generator build-consumer

build-generator:
	go build -o bin/generator cmd/generator/*.go

build-consumer:
	go build -o bin/consumer cmd/consumer/main.go

# Targets de ambiente
env-up:
	docker run -d --name evencountertest-rabbitmq -p $(AMQP_UI_PORT):15672 -p $(AMQP_PORT):5672 rabbitmq:3-management

env-down:
	docker rm -f evencountertest-rabbitmq

# Targets de teste
test:
	go test -v ./...

test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Targets do generator
generator-publish: build-generator
	bin/generator -publish=true -size=100 -amqp-url="amqp://guest:guest@localhost:$(AMQP_PORT)" -amqp-exchange="$(EXCHANGE)" -amqp-declare-queue=true

generator-publish-with-resume: build-generator
	bin/generator -publish=true -size=100 -amqp-url="amqp://guest:guest@localhost:$(AMQP_PORT)" -amqp-exchange="$(EXCHANGE)" -amqp-declare-queue=true -count=true -count-out="$(CURRENT_DIR)/data" -count=true

# Target para rodar o consumer
run-consumer: build-consumer
	bin/consumer -amqp-url="amqp://guest:guest@localhost:$(AMQP_PORT)" -queue="eventcountertest" -exchange="$(EXCHANGE)" -output-dir="./output"

# Target de limpeza
clean:
	rm -rf bin/ output/ data/ *.json coverage.out coverage.html

# Target de setup completo
up: build-all env-up

# Target com exemplo de variáveis de ambiente
run-consumer-env: build-consumer
	AMQP_URL="amqp://guest:guest@localhost:$(AMQP_PORT)" \
	QUEUE_NAME="eventcountertest" \
	EXCHANGE_NAME="$(EXCHANGE)" \
	OUTPUT_DIR="./output" \
	bin/consumer

.PHONY: build-all build-generator build-consumer env-up env-down test test-coverage generator-publish generator-publish-with-resume run-consumer clean up run-consumer-env




