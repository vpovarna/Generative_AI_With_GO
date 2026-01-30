# Week 3-4: Microservices Architecture

## Goal
Build a production-grade microservices system demonstrating:
- **Hexagonal (Clean) Architecture**
- **REST API** + **gRPC** communication
- **PostgreSQL** with migrations
- **Elasticsearch** for search
- **Redis Streams** for async messaging
- **Docker Compose** for local development

---

## System Architecture

```
┌─────────────────┐         ┌──────────────────┐
│   API Gateway   │◄────────┤   Frontend/CLI   │
│  (REST/gRPC)    │         └──────────────────┘
└────────┬────────┘
         │
    ┌────┴─────────────────────┐
    │                          │
┌───▼──────────┐      ┌────────▼────────┐
│ Order Service│      │  Search Service │
│  (REST API)  │      │     (gRPC)      │
└───┬──────────┘      └─────────────────┘
    │                          │
    │  ┌───────────────────────┤
    │  │                       │
┌───▼──▼─────┐     ┌───────────▼────────┐
│ PostgreSQL │     │   Elasticsearch    │
└────────────┘     └────────────────────┘
    │
    │
┌───▼────────────┐
│  Redis Stream  │◄───────┐
└────────────────┘        │
         │                │
         │          ┌─────┴──────────┐
         └──────────►  Worker Service│
                    │  (Async Jobs)  │
                    └────────────────┘
```

---

## Hexagonal Architecture Structure

```
service/
├── cmd/
│   └── server/
│       └── main.go              # Entry point, DI
│
├── internal/
│   ├── domain/                  # Core Business Logic (no dependencies)
│   │   ├── entities.go          # Business entities
│   │   ├── repository.go        # Port (interface)
│   │   ├── service.go           # Port (interface)
│   │   └── errors.go            # Domain errors
│   │
│   ├── usecases/                # Application Business Rules
│   │   ├── create_order.go      # Use case implementations
│   │   ├── get_order.go
│   │   └── list_orders.go
│   │
│   ├── adapters/                # External Adapters (implements ports)
│   │   ├── postgres/
│   │   │   └── repository.go    # DB implementation
│   │   ├── elasticsearch/
│   │   │   └── repository.go
│   │   ├── redis/
│   │   │   ├── publisher.go
│   │   │   └── consumer.go
│   │   └── grpc/
│   │       └── client.go
│   │
│   └── ports/                   # Driving Adapters (HTTP, gRPC)
│       ├── http/
│       │   ├── handler.go
│       │   ├── routes.go
│       │   └── middleware.go
│       └── grpc/
│           └── server.go
│
└── migrations/                  # Database migrations
    ├── 000001_init.up.sql
    └── 000001_init.down.sql
```

---

## Service 1: Order Service (REST API)

### Requirements

**Domain Entities**:
- Order: ID, CustomerID, Items, Total, Status, Timestamps
- OrderItem: ProductID, Quantity, Price
- OrderStatus: Pending, Confirmed, Shipped, Cancelled

**Business Rules**:
1. Order must have at least one item
2. Order total calculated from items
3. Status transitions validated
4. Events published on state changes

**REST Endpoints**:
- `POST /orders` - Create order
- `GET /orders/:id` - Get order by ID
- `GET /orders?customer_id=X` - List customer orders
- `PUT /orders/:id` - Update order
- `DELETE /orders/:id` - Cancel order

**Technical Requirements**:
- HTTP server with graceful shutdown
- Request validation
- Error handling and appropriate status codes
- Structured logging
- Middleware: Logger, Recovery, RequestID, CORS
- Database connection pooling
- Transaction support

### Implementation Checklist

Domain Layer:
- [ ] Define Order and OrderItem structs
- [ ] Implement business validation methods
- [ ] Define repository interface (port)
- [ ] Define event publisher interface (port)
- [ ] Custom domain errors

Use Cases Layer:
- [ ] CreateOrder use case
- [ ] GetOrder use case
- [ ] ListOrders use case
- [ ] UpdateOrder use case
- [ ] CancelOrder use case

Adapters Layer:
- [ ] PostgreSQL repository implementation
- [ ] Redis publisher implementation
- [ ] Connection management

Ports Layer:
- [ ] HTTP handlers
- [ ] Router setup (use go-chi or gorilla/mux)
- [ ] Middleware chain
- [ ] Request/Response DTOs

Infrastructure:
- [ ] Database migrations (up/down)
- [ ] Main application with DI
- [ ] Graceful shutdown
- [ ] Environment configuration

---

## Service 2: Search Service (gRPC)

### Requirements

**Purpose**: Full-text search across all entities using Elasticsearch

**Proto Definition** (`proto/search.proto`):
```protobuf
service SearchService {
  rpc IndexDocument(IndexRequest) returns (IndexResponse);
  rpc Search(SearchRequest) returns (SearchResponse);
  rpc DeleteDocument(DeleteRequest) returns (DeleteResponse);
}
```

**Messages**:
- IndexRequest: id, type, fields (map)
- SearchRequest: query, type, limit, offset
- SearchResponse: documents[], total
- Document: id, type, fields, score

**Technical Requirements**:
- gRPC server with reflection
- Elasticsearch client
- Index management
- Full-text search
- Faceted search (bonus)
- Error handling with gRPC status codes

### Implementation Checklist

Domain Layer:
- [ ] Document entity
- [ ] Search repository interface
- [ ] Query builders

Use Cases:
- [ ] IndexDocument use case
- [ ] Search use case
- [ ] DeleteDocument use case
- [ ] BulkIndex use case (bonus)

Adapters:
- [ ] Elasticsearch repository
- [ ] Index templates
- [ ] Mapping definitions
- [ ] Redis consumer (listen for events)

Ports:
- [ ] gRPC server implementation
- [ ] Proto message conversion
- [ ] Interceptors (logging, recovery)

Infrastructure:
- [ ] gRPC server setup
- [ ] Elasticsearch connection
- [ ] Health checks
- [ ] Graceful shutdown

---

## Service 3: Worker Service (Async Processing)

### Requirements

**Purpose**: Process events from Redis Streams asynchronously

**Event Types**:
- order.created → Index in Elasticsearch
- order.updated → Update in Elasticsearch
- order.cancelled → Update status
- Send notifications (bonus)
- Generate reports (bonus)

**Technical Requirements**:
- Redis Streams consumer
- Consumer group for load balancing
- Message acknowledgment
- Error handling and retries
- Dead letter queue for failed messages
- Graceful shutdown

### Implementation Checklist

Jobs/Handlers:
- [ ] OrderCreated handler
- [ ] OrderUpdated handler
- [ ] OrderCancelled handler

Adapters:
- [ ] Redis Streams consumer
- [ ] Consumer group management
- [ ] Message acknowledgment
- [ ] Elasticsearch client (for indexing)

Infrastructure:
- [ ] Worker main loop
- [ ] Signal handling
- [ ] Retry logic
- [ ] Error logging

---

## Database Schema

### PostgreSQL (Order Service)

**orders table**:
```sql
CREATE TABLE orders (
    id UUID PRIMARY KEY,
    customer_id VARCHAR(255) NOT NULL,
    items JSONB NOT NULL,
    total DECIMAL(10, 2) NOT NULL,
    status VARCHAR(50) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_orders_customer_id ON orders(customer_id);
CREATE INDEX idx_orders_status ON orders(status);
CREATE INDEX idx_orders_created_at ON orders(created_at);
```

### Migration Tools

Use `golang-migrate/migrate`:
```bash
# Create migration
migrate create -ext sql -dir migrations -seq create_orders

# Run migrations
migrate -path migrations -database $DATABASE_URL up

# Rollback
migrate -path migrations -database $DATABASE_URL down 1
```

---

## Redis Streams

### Stream Structure

**Stream Name**: `orders:events`

**Message Format**:
```json
{
  "event_type": "order.created",
  "data": "{...order json...}",
  "timestamp": 1234567890
}
```

### Consumer Group Setup

```bash
# Create consumer group
XGROUP CREATE orders:events worker-group 0 MKSTREAM

# Read from group
XREADGROUP GROUP worker-group worker-1 STREAMS orders:events >

# Acknowledge message
XACK orders:events worker-group <message-id>
```

---

## Docker Compose Setup

### Services

**docker-compose.yml**:
- PostgreSQL (port 5432)
- Redis (port 6379)
- Elasticsearch (port 9200)
- Order Service (port 8080)
- Search Service (port 50051)
- Worker Service (no exposed ports)

### Features:
- Health checks for all infrastructure
- Service dependencies
- Volume mounts for persistence
- Environment variable configuration
- Network isolation

### Implementation Checklist

- [ ] PostgreSQL service with health check
- [ ] Redis service with health check
- [ ] Elasticsearch service with health check
- [ ] Order service with build context
- [ ] Search service with build context
- [ ] Worker service with build context
- [ ] Volumes for data persistence
- [ ] Environment variables
- [ ] Service dependencies

---

## API Examples

### REST (Order Service)

**Create Order**:
```bash
curl -X POST http://localhost:8080/orders \
  -H "Content-Type: application/json" \
  -d '{
    "customer_id": "customer-123",
    "items": [
      {"product_id": "prod-1", "quantity": 2, "price": 29.99}
    ]
  }'
```

**Get Order**:
```bash
curl http://localhost:8080/orders/<order-id>
```

### gRPC (Search Service)

Use `grpcurl` or write a Go client:
```bash
grpcurl -plaintext \
  -d '{"query": "laptop", "type": "orders", "limit": 10}' \
  localhost:50051 search.SearchService/Search
```

---

## Testing Strategy

### Unit Tests
- Domain logic (business rules)
- Use case orchestration
- Pure functions

### Integration Tests
Use `testcontainers-go`:
- Test with real PostgreSQL
- Test with real Redis
- Test with real Elasticsearch

### E2E Tests
- Full flow: Create order → Index in ES → Search
- Test via Docker Compose

### Test Checklist

Order Service:
- [ ] Unit tests for domain validation
- [ ] Unit tests for use cases (with mocks)
- [ ] Integration tests with PostgreSQL
- [ ] HTTP handler tests
- [ ] Race detector tests

Search Service:
- [ ] Unit tests for search logic
- [ ] Integration tests with Elasticsearch
- [ ] gRPC server tests

Worker Service:
- [ ] Unit tests for job handlers
- [ ] Integration tests with Redis Streams
- [ ] End-to-end event processing

---

## Key Go Libraries

### HTTP Framework
- `github.com/go-chi/chi/v5` - Lightweight router
- OR `github.com/gorilla/mux` - Powerful router

### Database
- `github.com/jackc/pgx/v5` - PostgreSQL driver (recommended)
- OR `database/sql` with `github.com/lib/pq`

### gRPC
- `google.golang.org/grpc` - gRPC framework
- `google.golang.org/protobuf` - Protocol Buffers

### Redis
- `github.com/redis/go-redis/v9` - Redis client

### Elasticsearch
- `github.com/elastic/go-elasticsearch/v8` - Official client

### Migrations
- `github.com/golang-migrate/migrate/v4` - Database migrations

### Testing
- `github.com/stretchr/testify` - Test assertions
- `github.com/testcontainers/testcontainers-go` - Integration tests

### Configuration
- `github.com/kelseyhightower/envconfig` - Environment variables
- OR `github.com/spf13/viper` - Configuration management

---

## Development Workflow

### Setup
```bash
# Start infrastructure
docker-compose up -d postgres redis elasticsearch

# Run migrations
make migrate-up

# Run service
cd order-service
go run cmd/server/main.go
```

### Testing
```bash
# Unit tests
go test ./...

# Integration tests
go test -tags=integration ./...

# Race detector
go test -race ./...

# Coverage
go test -cover ./...
```

### Proto Generation
```bash
# Generate Go code from proto files
protoc --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
       proto/*.proto
```

---

## Advanced Features (Bonus)

### Observability
- [ ] Structured logging (zerolog, zap)
- [ ] Distributed tracing (OpenTelemetry)
- [ ] Metrics (Prometheus)
- [ ] Health check endpoints

### Security
- [ ] JWT authentication
- [ ] API rate limiting
- [ ] Input validation
- [ ] SQL injection prevention

### Performance
- [ ] Connection pooling
- [ ] Caching layer (Redis)
- [ ] Batch operations
- [ ] Indexing strategies

### Reliability
- [ ] Circuit breakers
- [ ] Retry with exponential backoff
- [ ] Timeouts on all external calls
- [ ] Graceful degradation

---

## Success Criteria

By end of Week 3-4, you should have:

Architecture:
- ✅ Clean/Hexagonal architecture in all services
- ✅ Clear separation: domain, use cases, adapters, ports
- ✅ Dependency inversion principle applied

Services:
- ✅ Order service with REST API
- ✅ Search service with gRPC
- ✅ Worker service processing events
- ✅ All services containerized

Data Layer:
- ✅ PostgreSQL with migrations
- ✅ Elasticsearch integration
- ✅ Redis Streams pub/sub

Quality:
- ✅ Unit tests (>70% coverage)
- ✅ Integration tests with real dependencies
- ✅ Race detector passing
- ✅ Graceful shutdown working

DevOps:
- ✅ Docker Compose setup
- ✅ Makefile for common tasks
- ✅ Environment-based configuration
- ✅ README with setup instructions

---

## Common Patterns

### Dependency Injection
```go
// main.go
func main() {
    // Infrastructure
    db := setupDatabase()
    redis := setupRedis()
    
    // Repositories (adapters)
    orderRepo := postgres.NewOrderRepository(db)
    publisher := redis.NewPublisher(redis)
    
    // Use Cases
    createOrderUC := usecases.NewCreateOrderUseCase(orderRepo, publisher)
    
    // Handlers (ports)
    handler := httpPort.NewHandler(createOrderUC, ...)
    
    // Server
    server := setupServer(handler)
    server.ListenAndServe()
}
```

### Error Handling
```go
// Domain errors
var (
    ErrOrderNotFound = errors.New("order not found")
    ErrInvalidOrder  = errors.New("invalid order")
)

// HTTP error responses
switch {
case errors.Is(err, domain.ErrOrderNotFound):
    respondJSON(w, 404, ErrorResponse{Error: "Order not found"})
case errors.Is(err, domain.ErrInvalidOrder):
    respondJSON(w, 400, ErrorResponse{Error: "Invalid order"})
default:
    respondJSON(w, 500, ErrorResponse{Error: "Internal error"})
}
```

### Graceful Shutdown
```go
quit := make(chan os.Signal, 1)
signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
<-quit

ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

if err := server.Shutdown(ctx); err != nil {
    log.Fatal("Shutdown error:", err)
}
```

---

## Resources

- [Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
- [Hexagonal Architecture](https://alistair.cockburn.us/hexagonal-architecture/)
- [Go gRPC Tutorial](https://grpc.io/docs/languages/go/)
- [PostgreSQL Best Practices](https://wiki.postgresql.org/wiki/Don%27t_Do_This)
- [Elasticsearch Go Client](https://www.elastic.co/guide/en/elasticsearch/client/go-api/current/index.html)
- [Redis Streams](https://redis.io/topics/streams-intro)

**Target**: Build a working microservices system that demonstrates production-ready patterns!
