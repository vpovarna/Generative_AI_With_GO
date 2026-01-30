# Week 5: AI Agents with MCP and AWS Bedrock

## Goal
Build production-grade AI agents using:
- **AWS Bedrock** - for LLM inference (Claude 3.5 Sonnet)
- **Model Context Protocol (MCP)** - for tool integrations
- **Agent Patterns** - ReAct, Tool Calling, RAG
- **Production Patterns** - Error handling, retries, streaming

---

## Architecture Overview

```
┌─────────────────────────────────────────┐
│          Agent Orchestrator             │
│  (Handles conversation & tool calling)  │
└────────┬───────────────────────┬────────┘
         │                       │
    ┌────▼────┐            ┌─────▼──────┐
    │ Bedrock │            │ MCP Server │
    │ Claude  │            │  (Tools)   │
    └─────────┘            └──────┬─────┘
                                  │
                    ┌─────────────┼─────────────┐
                    │             │             │
              ┌─────▼─────┐ ┌────▼────┐  ┌────▼────┐
              │  Database │ │   Web   │  │  Files  │
              │   Tool    │ │ Search  │  │  Tool   │
              └───────────┘ └─────────┘  └─────────┘
```

---

## Project Structure

```
04-ai-agents/
├── cmd/
│   ├── agent/
│   │   └── main.go          # Interactive agent CLI
│   ├── rag-indexer/
│   │   └── main.go          # RAG document indexer
│   └── server/
│       └── main.go          # Agent API server
│
├── internal/
│   ├── llm/
│   │   └── bedrock/
│   │       ├── client.go         # Bedrock API client
│   │       ├── models.go         # Request/Response types
│   │       └── streaming.go      # Streaming support
│   │
│   ├── mcp/
│   │   ├── server.go             # MCP server implementation
│   │   └── types.go              # MCP types
│   │
│   ├── agent/
│   │   ├── simple.go             # Simple conversational agent
│   │   ├── react.go              # ReAct pattern agent
│   │   ├── rag.go                # RAG agent
│   │   └── types.go              # Agent interfaces
│   │
│   ├── tools/
│   │   ├── database.go           # SQL query tool
│   │   ├── calculator.go         # Math calculations
│   │   ├── web_search.go         # Web search integration
│   │   ├── file_system.go        # File operations
│   │   └── types.go              # Tool interfaces
│   │
│   └── vectorstore/
│       ├── vectorstore.go        # Vector search interface
│       ├── opensearch.go         # OpenSearch implementation
│       └── embeddings.go         # Bedrock embeddings
│
├── go.mod
└── README.md
```

---

## Phase 1: AWS Bedrock Integration

### Requirements

**Bedrock Client**: Wrapper around AWS SDK
- Initialize with AWS credentials
- Support Claude 3.5 Sonnet model
- Handle API rate limits
- Retry with exponential backoff

**Request Types**:
- Messages: role (user/assistant), content
- System prompt
- Max tokens, temperature, top_p
- Stop sequences
- Tool definitions
- Tool choice (auto, any, specific tool)

**Response Types**:
- Message ID
- Content blocks (text or tool_use)
- Stop reason (end_turn, tool_use, max_tokens)
- Token usage (input/output)

**Streaming Support**:
- Stream responses token-by-token
- Handle partial JSON
- Reconnect on errors

### Implementation Checklist

- [ ] AWS SDK v2 configuration
- [ ] Bedrock Runtime client
- [ ] InvokeModel wrapper
- [ ] InvokeModelWithResponseStream wrapper
- [ ] Request/response type definitions
- [ ] Error handling and retries
- [ ] Context timeout support
- [ ] Token usage tracking

### Claude 3.5 Sonnet Capabilities
- Context window: 200K tokens
- Tool calling: Native support
- JSON mode: Structured outputs
- Vision: Image understanding (bonus)
- Multi-turn conversations

---

## Phase 2: Model Context Protocol (MCP)

### Requirements

**MCP Server**: Manages tools/resources for agents

**Core Interfaces**:
- Tool: name, description, input schema, execution function
- Resource: URI-based content access
- Prompt: Reusable prompt templates

**Tool Definition**:
- Name: Unique identifier
- Description: What the tool does
- Input Schema: JSON Schema for parameters
- Handler: Execution function

**Tool Execution**:
- Parse arguments
- Validate against schema
- Execute with timeout
- Return result or error

### Implementation Checklist

- [ ] MCP server struct
- [ ] Tool registration
- [ ] Tool listing
- [ ] Tool execution
- [ ] Input validation
- [ ] Error handling
- [ ] Timeout support
- [ ] Tool discovery

### Tool Registry Pattern
```
Register("tool_name", Tool{
    Description: "What it does",
    Schema: {...},
    Handler: func(ctx, args) (result, error)
})
```

---

## Phase 3: Tool Implementations

### Tool 1: Database Query Tool

**Purpose**: Execute read-only SQL queries

**Requirements**:
- Accept SQL query as input
- Only allow SELECT statements
- Return results as JSON
- Limit result count
- Timeout protection

**Schema**:
- Input: query (string)
- Output: array of row objects

**Safety**:
- SQL injection prevention
- Read-only user
- Query timeout
- Result size limits

### Tool 2: Calculator Tool

**Purpose**: Perform mathematical calculations

**Requirements**:
- Evaluate math expressions
- Support common functions (sqrt, pow, log, etc.)
- Handle decimals and large numbers
- Return precise results

**Schema**:
- Input: expression (string)
- Output: result (number)

**Use Library**: `github.com/expr-lang/expr` or similar

### Tool 3: Web Search Tool

**Purpose**: Search the web for information

**Requirements**:
- Search query
- Return top N results
- Extract title, snippet, URL
- Cache results (optional)

**Schema**:
- Input: query (string), limit (int)
- Output: array of search results

**Integration Options**:
- DuckDuckGo API (free)
- Google Custom Search API
- Bing Search API

### Tool 4: File System Tool

**Purpose**: Read/write files safely

**Requirements**:
- Read file contents
- Write to files
- List directory contents
- Sandbox to specific directory

**Schema**:
- read: path → content
- write: path, content → success
- list: path → []filenames

**Safety**:
- Path traversal prevention
- Size limits
- Allowed extensions

### Implementation Checklist

Per Tool:
- [ ] Handler function
- [ ] Input schema definition
- [ ] Validation logic
- [ ] Error handling
- [ ] Unit tests
- [ ] Integration tests
- [ ] Documentation

---

## Phase 4: Agent Patterns

### Pattern 1: Simple Conversational Agent

**Purpose**: Basic chat without tools

**Requirements**:
- Maintain conversation history
- Send user message + history to LLM
- Receive and display response
- Track token usage

**Features**:
- System prompt configuration
- Multi-turn conversations
- History management
- Clear/reset functionality

**Implementation**:
- Agent struct with history
- Chat method
- History trimming (sliding window)
- Token counting

### Pattern 2: ReAct Agent (Reasoning + Acting)

**Purpose**: Agent that uses tools to answer questions

**Requirements**:
- Reasoning Loop:
  1. Receive question
  2. LLM decides to use tool or answer
  3. If tool → execute tool → add result to context
  4. Repeat until final answer
  5. Return answer

**Features**:
- Tool calling in loop
- Max iterations (prevent infinite loops)
- Intermediate step logging
- Thought process visible

**Algorithm**:
```
while not done and iterations < max:
    response = llm.invoke(messages + tools)
    
    if response.stop_reason == "tool_use":
        for tool_call in response.tool_calls:
            result = execute_tool(tool_call)
            add_tool_result_to_messages(result)
    else:
        return extract_answer(response)
```

**Implementation Checklist**:
- [ ] Agent struct with LLM client and MCP server
- [ ] Prepare tool definitions for Bedrock
- [ ] Execute tool calls
- [ ] Format tool results
- [ ] Iteration loop
- [ ] Max iteration limit
- [ ] Extract final answer

### Pattern 3: RAG Agent (Retrieval-Augmented Generation)

**Purpose**: Answer questions using external knowledge base

**Requirements**:
- Vector Store:
  - Store document embeddings
  - Similarity search
  - Return top K relevant docs

- Embeddings:
  - Use Bedrock Titan Embeddings
  - Embed query
  - Embed documents

- RAG Pipeline:
  1. Receive question
  2. Embed question
  3. Search vector store
  4. Retrieve relevant docs
  5. Build context from docs
  6. Query LLM with context
  7. Return answer with sources

**Vector Store Options**:
- OpenSearch (AWS)
- Pinecone
- Weaviate
- In-memory (for testing)

**Implementation Checklist**:
- [ ] Embedding client (Bedrock Titan)
- [ ] Vector store interface
- [ ] Vector store implementation
- [ ] Document indexer
- [ ] Similarity search
- [ ] RAG agent
- [ ] Context building
- [ ] Source attribution

---

## Phase 5: Production Features

### Streaming Responses

**Requirements**:
- Stream tokens as they're generated
- Display in real-time
- Handle errors mid-stream
- Reconnect on failure

**Implementation**:
- Use Bedrock streaming API
- Channel-based streaming
- Error handling in goroutine

### Error Handling

**Categories**:
- API errors (rate limit, auth, quota)
- Tool errors (execution failures)
- Parsing errors (invalid JSON)
- Network errors (timeouts, connection)

**Patterns**:
- Wrap errors with context
- Retry transient errors
- Log all errors
- User-friendly messages

### Rate Limiting

**Requirements**:
- Respect Bedrock limits
- Exponential backoff
- Token bucket algorithm
- Per-model limits

### Conversation Memory

**Types**:
- Short-term: Recent messages
- Long-term: Semantic search over history
- Working memory: Current task context

**Implementation**:
- Sliding window for short-term
- Vector store for long-term
- Token limit management

---

## Example Use Cases

### Use Case 1: SQL Assistant

**Scenario**: Help users query database

**System Prompt**:
```
You are a SQL assistant for an e-commerce database.
Help users query the orders, customers, and products tables.
When asked a question, use the query_database tool.
Always explain results in plain language.
```

**Tools**: database query

**Flow**:
1. User: "Show me top 5 customers by revenue"
2. Agent: Uses query_database tool
3. Agent: Explains results

### Use Case 2: Research Assistant

**Scenario**: Answer questions with web research

**System Prompt**:
```
You are a research assistant.
When you don't know something, use web_search.
Cite sources in your answers.
```

**Tools**: web_search, calculator

**Flow**:
1. User: "What's the market cap of Tesla?"
2. Agent: Searches web
3. Agent: Returns answer with source

### Use Case 3: Document Q&A

**Scenario**: Answer questions about uploaded documents

**System Prompt**:
```
Answer questions based only on the provided context.
If information isn't in the context, say so.
```

**Tools**: RAG with vector store

**Flow**:
1. Index documents (preprocessing)
2. User: "What is the refund policy?"
3. Agent: Searches relevant docs
4. Agent: Answers with page references

---

## AWS Setup

### Prerequisites

```bash
# Install AWS CLI
brew install awscli

# Configure credentials
aws configure
# Enter AWS Access Key ID
# Enter AWS Secret Access Key
# Default region: us-east-1

# Verify access
aws sts get-caller-identity
```

### Enable Bedrock Models

1. Go to AWS Console → Bedrock
2. Navigate to "Model access"
3. Request access to:
   - Claude 3.5 Sonnet v2
   - Titan Embeddings G1

### Test Bedrock Access

```bash
# Test Claude 3.5 Sonnet
aws bedrock-runtime invoke-model \
    --model-id anthropic.claude-3-5-sonnet-20241022-v2:0 \
    --body '{"messages":[{"role":"user","content":"Hello"}],"max_tokens":100}' \
    --region us-east-1 \
    response.json

cat response.json
```

---

## Testing Strategy

### Unit Tests

**LLM Client**:
- Mock Bedrock API
- Test request formatting
- Test response parsing
- Test error handling

**MCP Server**:
- Test tool registration
- Test tool execution
- Test schema validation

**Tools**:
- Test each tool independently
- Mock external dependencies
- Test error cases

### Integration Tests

**Bedrock Integration**:
- Test real API calls (with test key)
- Test streaming
- Test rate limiting

**Agent Tests**:
- Test ReAct loop
- Test tool calling
- Test max iterations

### E2E Tests

**Full Scenarios**:
- User asks question → Agent uses tools → Returns answer
- Test with real LLM (expensive!)
- Use short test cases

---

## Key Go Libraries

```go
// AWS
github.com/aws/aws-sdk-go-v2/aws
github.com/aws/aws-sdk-go-v2/config
github.com/aws/aws-sdk-go-v2/service/bedrockruntime

// MCP (check latest version)
github.com/modelcontextprotocol/go-sdk

// Utilities
github.com/expr-lang/expr          // Calculator
github.com/lib/pq                  // PostgreSQL
```

---

## Success Criteria

By end of Week 5, you should have:

Bedrock Integration:
- ✅ Working Bedrock client
- ✅ Request/response handling
- ✅ Streaming support
- ✅ Error handling and retries

MCP & Tools:
- ✅ MCP server implementation
- ✅ At least 3 tools registered
- ✅ Tool execution working
- ✅ Schema validation

Agents:
- ✅ Simple conversational agent
- ✅ ReAct agent with tool calling
- ✅ RAG agent (bonus)
- ✅ Conversation history management

Production Ready:
- ✅ Error handling throughout
- ✅ Logging and observability
- ✅ Unit tests (>70% coverage)
- ✅ Integration tests with mocks
- ✅ CLI interface working

---

## Advanced Features (Bonus)

### Multi-Agent Systems
- Orchestrator agent that delegates to specialized agents
- SQL agent, Search agent, Analysis agent
- Agent routing based on query type

### Prompt Templates
- Template management
- Variable substitution
- Few-shot examples

### Agent Evaluation
- Test suite of questions
- Compare answers to ground truth
- Measure accuracy and token usage

### Caching
- Cache LLM responses
- Cache tool results
- Cache embeddings

---

## Common Pitfalls

### 1. Token Limits
**Problem**: Exceeding context window
**Solution**: Sliding window, summarization

### 2. Infinite Loops
**Problem**: Agent keeps calling tools
**Solution**: Max iteration limit

### 3. Tool Errors
**Problem**: Tool fails, agent confused
**Solution**: Return error message to agent, let it decide

### 4. Cost Management
**Problem**: Expensive API calls
**Solution**: Caching, streaming, smaller models for testing

### 5. Prompt Injection
**Problem**: User manipulates agent behavior
**Solution**: Input validation, system prompt protection

---

## Development Workflow

```bash
# Setup
export AWS_REGION=us-east-1
export AWS_PROFILE=default
export DATABASE_URL="postgres://..."

# Run agent
cd 04-ai-agents
go run cmd/agent/main.go

# Run tests
go test ./...

# Run with race detector
go test -race ./...

# Run specific agent
go run cmd/agent/main.go --agent=react

# Index documents for RAG
go run cmd/rag-indexer/main.go --dir=./docs
```

---

## Resources

- [AWS Bedrock Documentation](https://docs.aws.amazon.com/bedrock/)
- [Claude API Reference](https://docs.anthropic.com/claude/reference)
- [Model Context Protocol](https://modelcontextprotocol.io/)
- [ReAct Paper](https://arxiv.org/abs/2210.03629)
- [RAG Paper](https://arxiv.org/abs/2005.11401)
- [Prompt Engineering Guide](https://www.promptingguide.ai/)

---

## Deliverables

1. **Working Agent CLI** - Interactive command-line interface
2. **At Least 3 Tools** - Database, Calculator, Web Search
3. **ReAct Agent** - Full reasoning loop implementation
4. **Tests** - Unit and integration tests
5. **Documentation** - README with setup and examples

**Target**: Build a production-ready AI agent system that demonstrates enterprise patterns!
