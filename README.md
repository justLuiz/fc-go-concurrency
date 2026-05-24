# Auction System — Auto-Close

Sistema de leilões em Go com fechamento automático via Goroutines. Ao criar um leilão, uma goroutine é iniciada em background e atualiza o status para `Closed` ao expirar o tempo configurado.

## Pré-requisitos

- Docker e Docker Compose

## Execução

```bash
docker-compose up --build
```

A API ficará disponível em `http://localhost:8080`.

## Variáveis de Ambiente

Definidas em `cmd/auction/.env`:

| Variável | Descrição | Exemplo |
|---|---|---|
| `AUCTION_DURATION` | Tempo até o fechamento automático do leilão (formato Go duration) | `20s`, `5m`, `1h` |
| `BATCH_INSERT_INTERVAL` | Intervalo para inserção em lote de lances | `20s` |
| `MAX_BATCH_SIZE` | Tamanho máximo do lote de lances | `4` |
| `MONGODB_URL` | URL de conexão com o MongoDB | `mongodb://admin:admin@mongodb:27017/auctions?authSource=admin` |
| `MONGODB_DB` | Nome do banco de dados | `auctions` |

### Alterar a duração do leilão

Edite `cmd/auction/.env` e ajuste `AUCTION_DURATION`:

```env
AUCTION_DURATION=2m
```

Aceita qualquer valor no formato de duração do Go: `30s`, `2m`, `1h30m`.

## Testes

Os testes de integração requerem uma instância do MongoDB em execução. Suba o banco antes de rodar:

```bash
docker-compose up -d mongodb
go test ./internal/infra/database/auction/... -v
```

O teste `TestAuctionAutoClose` cria um leilão com `AUCTION_DURATION=1s`, aguarda 2 segundos e verifica que o status foi alterado para `Closed` automaticamente, sem intervenção manual.

## Endpoints

| Método | Rota | Descrição |
|---|---|---|
| `POST` | `/auction` | Cria um novo leilão |
| `GET` | `/auction` | Lista leilões |
| `GET` | `/auction/:auctionId` | Busca leilão por ID |
| `GET` | `/auction/winner/:auctionId` | Retorna o lance vencedor |
| `POST` | `/bid` | Cria um lance |
| `GET` | `/bid/:auctionId` | Lista lances de um leilão |
| `GET` | `/user/:userId` | Busca usuário por ID |
