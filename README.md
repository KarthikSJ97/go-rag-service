# Local-First Go RAG Service

A high-performance, fully local Retrieval-Augmented Generation (RAG) pipeline built with **Go (Fiber)**, **Qdrant**, and **Ollama**. 

This service allows you to ingest text documents, extract vector embeddings concurrently using local LLM models, store vector chunks with metadata in Qdrant, and perform context-aware filtered querying with fast API responses.

---

## 🛠️ Tech Stack

* **Language & Web Framework:** Go (v1.22+), [Fiber v2](https://gofiber.io/)
* **Vector Database:** [Qdrant](https://qdrant.tech/) (v1.11+) via gRPC
* **LLM & Embeddings Engine:** [Ollama](https://ollama.com/) (Local embeddings & generation)
* **Containerization & Orchestration:** Docker & Docker Compose

---

## 📁 Project Structure

```text
.
├── docker-compose.yml       # Orchestrates app, Qdrant, and Ollama services
├── Dockerfile               # Multi-stage Go build container definition
├── go.mod                   # Go module dependencies
├── go.sum                   # Go checksums
├── LICENSE                  # License file
├── README.md                # Documentation
└── src
    ├── config               # Environment configuration and validation
    │   └── config.go
    ├── document             # File parsing and text chunking logic
    │   └── loader.go
    ├── llm                  # Ollama API client (Embeddings & Generation)
    │   └── client.go
    ├── main.go              # Application entry point
    ├── routes               # REST API endpoints (/ingest, /query, /documents)
    │   └── routes.go
    ├── service              # Dependency injection container
    │   └── container.go
    └── vectorstore          # Qdrant client & vector search/management operations
        └── qdrant.go
```

---

## 🔒 Why Local-First Architecture Matters

1. **Complete Data Privacy & Zero Data Leakage:** No proprietary documents, text chunks, or queries are ever sent to third-party cloud APIs (like OpenAI or Anthropic). Everything stays on your local machine or private infrastructure.
2. **Zero API Expenses:** Running local embedding models (`bge-small-en-v1.5`) and LLMs (`llama3`) via Ollama eliminates per-token API costs.
3. **Low Latency & Offline Capability:** Communication between Fiber, Qdrant (via gRPC), and Ollama happens within the isolated Docker bridge network, ensuring fast execution without internet connectivity dependencies.

---

## 🚀 Getting Started

### Prerequisites
* [Docker](https://docs.docker.com/get-docker/) installed and running.
* [Git](https://git-scm.com/) installed.

---

### Step 1: Clone the Repository

```bash
git clone <your-repository-url>
cd go-rag-service
```

### Step 2: Configure Environment Variables

Create a `.env` file in the root directory and populate with the required configuration key names:

```bash
PORT=PORT
MAX_PAYLOAD_SIZE_IN_MB=MAX_PAYLOAD_SIZE_IN_MB
QDRANT_HOST=QDRANT_HOST
QDRANT_PORT=QDRANT_PORT
QDRANT_COLLECTION_NAME=QDRANT_COLLECTION_NAME
OLLAMA_HOST=OLLAMA_HOST
OLLAMA_EMB_MODEL=OLLAMA_EMB_MODEL
OLLAMA_LLM_MODEL=OLLAMA_LLM_MODEL
CHUNK_SIZE=CHUNK_SIZE
OVERLAP=OVERLAP
VECTOR_SIZE=VECTOR_SIZE
```

### Step 3: Run the Application Stack

Start all services (App, Qdrant, Ollama) in detached mode using Docker Compose:

```bash
docker-compose up --build -d
```

### Step 4: Verify Running Containers

Check that all container services are healthy and running:

```bash
docker ps
```

### Step 5: Shut Down

To stop all services and tear down containers:

```bash
docker-compose down
```

To also delete persisted Qdrant volume data during shutdown, run:
```bash
docker volume rm qdrant_data
```

---

## 📡 API Endpoints Summary

| Method | Endpoint | Description |
| :--- | :--- | :--- |
| `GET` | `/health` | Service health check |
| `POST` | `/api/v1/ingest` | Upload a text document (`.txt` or `.pdf`) to chunk, embed, and store in Qdrant |
| `POST` | `/api/v1/query` | RAG query endpoint with optional `document_ids` filter |
| `GET` | `/api/v1/documents` | List ingested documents with chunk counts, filenames, and file sizes |
| `DELETE` | `/api/v1/documents/:id` | Delete all vectors and metadata associated with a document ID |