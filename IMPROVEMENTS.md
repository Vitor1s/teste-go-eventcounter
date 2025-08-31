# Melhorias Implementadas

## 1. Configuração Dinâmica via Variáveis de Ambiente

O consumer agora aceita configuração através de:

### Flags de linha de comando:

```bash
./bin/consumer -amqp-url="amqp://localhost:5672" -queue="eventcountertest" -exchange="eventcountertest" -output-dir="./output"
```

### Variáveis de ambiente:

```bash
export AMQP_URL="amqp://guest:guest@localhost:5672"
export QUEUE_NAME="eventcountertest"
export EXCHANGE_NAME="eventcountertest"
export OUTPUT_DIR="./output"
./bin/consumer
```

## 2. Targets Adicionados no Makefile

Novos comandos disponíveis:

```bash
# Build tudo
make build-all

# Build apenas consumer
make build-consumer

# Rodar consumer com configurações
make run-consumer

# Rodar consumer com variáveis de ambiente
make run-consumer-env

# Rodar testes
make test

# Testes com coverage
make test-coverage

# Limpar arquivos
make clean
```

## 3. Testes Unitários 

- Testes para o consumer (`cmd/consumer/main_test.go`)
- Testes para o package (`pkg/pkg_test.go`)

Rodar com:

```bash
go test ./...
```

## Como usar:

1. **Subir ambiente**: `make env-up`
2. **Publicar mensagens**: `make generator-publish`
3. **Rodar consumer**: `make run-consumer`
4. **Testar**: `make test`
5. **Limpar**: `make clean`
