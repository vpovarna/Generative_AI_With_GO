Languages & Serialization  

| Technology       | Version | Purpose                          |
| ---------------- | ------- | -------------------------------- |
| Go               | 1.24.6  | Primary language                 |
| Protocol Buffers | v3      | Service definitions              |
| JSON/YAML        | -       | Configuration & data interchange |
| CUE              | v0.9.2  | Configuration language           |
  
⸻AI/ML Providers  

| Provider           | Package/Integration                  |
| ------------------ | ------------------------------------ |
| Anthropic (Claude) | internal/gateways/aws/claude/        |
| AWS Bedrock        | aws-sdk-go-v2/service/bedrockruntime |
| Google Vertex AI   | internal/gateways/vertexai/          |
| OpenAI             | internal/gateways/openai/            |
| AWS SageMaker      | internal/gateways/aws/sagemaker/     |
| TEI                | internal/gateways/tei/               |
  
⸻Databases & Storage  

| Technology | Purpose                       |
| ---------- | ----------------------------- |
| PostgreSQL | Guidance blocks, migrations   |
| OpenSearch | Vector store, document search |
| DynamoDB   | Key-value storage             |
| Redis      | Caching, distributed locks    |
| AWS S3     | Document/artifact storage     |
  
⸻Message Queues  

| Technology | Usage                              |
| ---------- | ---------------------------------- |
| AWS SQS    | Async agent execution              |
| Kafka      | Document indexing, event streaming |
| CSJobs     | Internal job processing system     |
  
⸻AWS Services  
  
- Bedrock - LLM inference  
- S3 - Object storage  
- SQS - Message queue  
- DynamoDB - Key-value storage  
- SES/SNS - Notifications  
- STS - Security tokens  
- SageMaker - ML model hosting  
⸻Observability  

| Category        | Technology     |
| --------------- | -------------- |
| Logging         | Zerolog        |
| Metrics         | Prometheus     |
| Tracing         | OpenTelemetry  |
| Dashboards      | Grafana        |
| Log Aggregation | Humio/LogScale |
  
⸻Testing  

| Tool           | Purpose                  |
| -------------- | ------------------------ |
| Testify        | Assertions               |
| GoMock         | Mock generation (v0.6.0) |
| Ginkgo/Gomega  | BDD testing              |
| TestContainers | Integration tests        |
| MiniRedis      | Redis mocking            |
| Pigeon         | Integration/E2E testing  |
| Faker          | Test data generation     |
  
⸻Build & CI/CD  

| Tool          | Purpose             |
| ------------- | ------------------- |
| Make          | Build orchestration |
| Goreleaser    | Release automation  |
| Buf           | Protobuf builds     |
| golangci-lint | Code linting        |
| Jenkins       | CI/CD pipeline      |
  
⸻Web Frameworks  

| Package            | Purpose                    |
| ------------------ | -------------------------- |
| go-restful         | REST API framework         |
| go-restful-openapi | Swagger/OpenAPI generation |
| Cobra              | CLI framework              |
| go-openapi         | OpenAPI client generation  |
| gRPC               | RPC framework              |
  
⸻Infrastructure  

| Technology     | Purpose             |
| -------------- | ------------------- |
| Docker         | Containerization    |
| Docker Compose | Local development   |
| Kubernetes     | Orchestration       |
| LocalStack     | Local AWS emulation |
  
⸻External Libraries  

| Library               | Purpose                  |
| --------------------- | ------------------------ |
| expr-lang/expr        | Expression evaluation    |
| itchyny/gojq          | JQ query processing      |
| slack-go/slack        | Slack API integration    |
| olivere/elastic       | Elasticsearch client     |
| modelcontextprotocol  | MCP SDK (go-sdk v1.2.0)  |
| jackc/pgx             | PostgreSQL driver        |
| dgraph-io/ristretto   | In-memory cache          |
| eapache/go-resiliency | Circuit breaker, retries |
| golang-migrate        | Database migrations      |
  
⸻Architecture Pattern  
  
Clean Architecture (Hexagonal/Ports and Adapters):  
- domain/ - Core entities and interfaces  
- usecases/ - Application business logic  
- gateways/ - External service adapters  
- services/ - Transport/API layer  
⸻Summary  
  
- 28 services/tools in cmd/  
- 17 internal packages  
- 6 LLM providers (Anthropic, Bedrock, Vertex AI, OpenAI, SageMaker, TEI)  
- Multi-cloud (AWS primary, GCP for Vertex AI)  
- Enterprise patterns: RBAC, distributed tracing, cost management, quota  
- 30+ specialized agents for different domains  
- MCP (Model Context Protocol) support  
