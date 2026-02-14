## Scope 

Create a KG Agent. It should use Bedrock / Claude for reasoning.
The communication method should be API base. This should be an application which can start and accept messages through an API POST request. 
Is should have tools, connect to DB or to another API for fetching similarities. Performing hybrid search. 
Maybe should have other tools.
I would also like to add guardrails and query rewrite as tasks in the flow. 

## Phases
Phase 1: Foundation - AWS/Claude Connection & Basic LLM  
  - Set up AWS Bedrock client in Go
  - Implement basic Claude API integration
  - Create simple prompt/response flow
  - Test basic reasoning capabilities
  - Deliverable: A CLI or simple function that can send a prompt to Claude and get a response

Phase 2: API Layer  
  - Build HTTP server with POST endpoint for queries
  - Request/response models for the agent
  - Basic error handling and logging
  - Deliverable: REST API that accepts documentation questions and returns Claude responses

Phase 3: Query Write logic
  - Add query rewriting for better retrieval

Phase 4: Knowledge Base & Vector Search  
Contains three sub-phases

  Phase 4A: Database + Ingestion
  - Setup Postgres + pgvector (docker-compose)
  - Create schema (documents, chunks, embeddings)
  - Build embedding service (Bedrock Titan)
  - Build chunker
  - Build ingestion CLI
  - Load sample documents

  Phase 4B: Search Service
  - Create cmd/search-api/ - separate service
  - Implement semantic search endpoint
  - Implement keyword search endpoint
  - Implement hybrid search endpoint
  - Add ranking (RRF)

  Phase 4C: Integration
  - Call search service from agent
  - Format results as context
  - Pass context to Claude
  - Add simple caching (in-memory map)
  - Test end-to-end

Phase 5: Advanced
 - Tool calling (let Claude decide when to search)
 - Redis caching
 - Multi-hop reasoning
 - Conversation memory

Phase 6: Extra Features  
  - Implement guardrails (input/output validation, safety checks)
  - Add conversation memory/history
  - Deliverable: Production-ready documentation agent