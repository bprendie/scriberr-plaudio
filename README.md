# Plaudio - Scriberr with RAG Extension

**Plaudio** is an extended version of [Scriberr](https://github.com/rishikanthc/Scriberr), a self-hosted audio transcription application. This fork adds **Retrieval-Augmented Generation (RAG)** capabilities, enabling semantic search and natural language querying across all transcriptions.

## 🎯 What's New

This extension adds the following features to Scriberr:

- **🔄 Auto-transcription**: Upload WAV or MP3 files and they are automatically transcribed
- **📝 Auto-summarization**: Transcriptions are automatically summarized using Ollama
- **🗄️ Vector Database Integration**: Summaries and transcripts are stored in ChromaDB for semantic search
- **💬 Global Chat**: Query across all transcriptions using natural language in a unified chat interface
- **📊 RAG Status Dashboard**: Monitor RAG system status and transcript counts in the settings

## 🏗️ Architecture

```
┌─────────────┐
│   Upload    │  WAV/MP3 files
│  Audio File │
└──────┬──────┘
       │
       ▼
┌─────────────┐
│ Auto-       │  WhisperX transcription
│ Transcribe  │
└──────┬──────┘
       │
       ▼
┌─────────────┐      ┌──────────────┐
│ Post-       │─────▶│ Generate     │
│ Processing  │      │ Summary      │
│ Hook        │      │ (Ollama)     │
└──────┬──────┘      └──────────────┘
       │
       ▼
┌─────────────┐      ┌──────────────┐
│ Extract     │─────▶│ Generate     │
│ Text        │      │ Embedding    │
└──────┬──────┘      │ (Ollama)     │
       │             └──────────────┘
       │                    │
       └──────────┬─────────┘
                  ▼
         ┌─────────────────┐
         │   ChromaDB      │
         │  (Vector Store) │
         └────────┬────────┘
                  │
                  ▼
         ┌─────────────────┐
         │  Global Chat    │
         │  (RAG Query)    │
         └─────────────────┘
```

### Technology Stack

- **Backend**: Go (Gin framework) - Extended from Scriberr
- **Frontend**: React (embedded in Go binary) - Extended from Scriberr
- **Transcription**: WhisperX (via Scriberr)
- **LLM Provider**: Ollama (configurable URL)
- **Embeddings**: `nomic-embed-text` via Ollama
- **Vector Database**: ChromaDB
- **Containerization**: Docker & Docker Compose

## 🚀 Quick Start

### Prerequisites

- Docker and Docker Compose
- Ollama server (accessible from the container)
  - Default: `http://10.0.0.50:11434`
  - Configure via `OLLAMA_URL` environment variable

### Installation

1. **Clone the repository**:
   ```bash
   git clone <your-repo-url>
   cd plaudio
   ```

2. **Ensure Ollama models are installed**:
   ```bash
   # On your Ollama server
   ollama pull nomic-embed-text
   ollama pull llama3.2  # or your preferred model
   ```

3. **Start the application**:
   ```bash
   docker-compose up -d
   ```

4. **Access the application**:
   - Web UI: `http://localhost:8080`
   - ChromaDB: `http://localhost:8000`

### Environment Variables

Configure via `docker-compose.yml` or environment variables:

```env
# Server Configuration
HOST=0.0.0.0
PORT=8080

# Storage Paths
DATABASE_PATH=/app/data/scriberr.db
UPLOAD_DIR=/app/data/uploads
WHISPERX_ENV=/app/data/whisperx-env

# RAG Configuration
OLLAMA_URL=http://10.0.0.50:11434          # Your Ollama server URL
CHROMADB_URL=http://chromadb:8000          # ChromaDB service URL
EMBEDDING_MODEL=nomic-embed-text           # Embedding model name
OLLAMA_MODEL=llama3.2                     # LLM model for summarization/chat
```

## 📖 Usage

### Uploading and Transcribing

1. **Upload Audio**: Click "Upload" or drag-and-drop WAV/MP3 files
2. **Auto-transcription**: Files are automatically queued for transcription
3. **Auto-processing**: After transcription completes:
   - Summary is generated using Ollama
   - Content is embedded and stored in ChromaDB
   - Transcript becomes searchable via Global Chat

### Global Chat (RAG Query)

Query across all your transcriptions using natural language:

1. **Navigate to Global Chat**:
   - Click the menu (☰) → "Global Chat"
   - Or click "Global Chat" button on the homepage
   - Or navigate to `/global-chat`

2. **Select a Model**: Choose your preferred Ollama model from the dropdown

3. **Ask Questions**:
   - "What was discussed about project deadlines?"
   - "Summarize all meetings from last week"
   - "What topics were covered in the technical discussions?"
   - "Find mentions of budget discussions"

The system will:
- Search the vector database for relevant transcripts
- Retrieve the most relevant context
- Generate answers using the selected Ollama model

### RAG Status Dashboard

Monitor your RAG system:

1. Go to **Settings** → **RAG** tab
2. View:
   - System status (Active/Inactive)
   - Number of transcripts in RAG
   - Collection information
3. Actions:
   - **Refresh**: Update statistics
   - **Backfill**: Process existing transcriptions into RAG

## 🔌 API Endpoints

### RAG Chat

Query the RAG system:

```bash
POST /api/v1/rag/chat
Content-Type: application/json
Authorization: Bearer <token>

{
  "query": "What was discussed about project deadlines?",
  "model": "llama3.2",
  "temperature": 0.7
}
```

### RAG Statistics

Get RAG system statistics:

```bash
GET /api/v1/rag/stats
Authorization: Bearer <token>
```

Response:
```json
{
  "status": "active",
  "transcript_count": 42,
  "collection_name": "transcriptions"
}
```

### Backfill Existing Transcriptions

Process all completed transcriptions into RAG:

```bash
POST /api/v1/rag/backfill
Authorization: Bearer <token>
```

## 🔧 Development

### Building Locally

```bash
# Build the Go binary
go build -o plaudio cmd/server/main.go

# Build frontend (if making UI changes)
cd web/frontend
npm install
npm run build
```

### Running Locally

```bash
# Start ChromaDB
docker run -d -p 8000:8000 \
  -e IS_PERSISTENT=TRUE \
  chromadb/chroma:latest

# Run the application
./plaudio
```

### Project Structure

```
plaudio/
├── cmd/server/          # Main application entry point
├── internal/
│   ├── api/            # API handlers (extended with RAG endpoints)
│   ├── rag/            # RAG service (NEW)
│   ├── embeddings/     # Embedding service (NEW)
│   ├── vectordb/       # ChromaDB client (NEW)
│   ├── llm/            # LLM service (NEW)
│   ├── transcription/  # Transcription processing (extended with post-processing hook)
│   └── ...             # Other Scriberr components
├── web/frontend/       # React frontend (extended with Global Chat)
└── docker-compose.yml  # Docker orchestration
```

## 🐛 Troubleshooting

### Transcripts Not Appearing in Global Chat

1. **Check RAG Status**: Go to Settings → RAG tab
2. **Verify Processing**: Check container logs for post-processing messages:
   ```bash
   docker compose logs scriberr | grep -i "post-processing\|rag"
   ```
3. **Backfill Existing Data**: Use the backfill endpoint to process existing transcriptions
4. **Check ChromaDB**: Ensure ChromaDB container is running:
   ```bash
   docker compose ps chromadb
   ```

### Summary Generation Failing

- Transcripts are still stored in RAG even if summary generation fails
- Check Ollama logs and ensure the model is available
- Verify `OLLAMA_MODEL` environment variable matches an installed model
- Ensure Ollama server is accessible from the container

### Connection Issues

- **Ollama**: Verify `OLLAMA_URL` is correct and accessible
- **ChromaDB**: Check `CHROMADB_URL` matches the service name in docker-compose
- **Network**: Ensure containers can communicate (check docker network)

## 📝 Differences from Scriberr

This extension adds:

1. **RAG Components**:
   - `internal/rag/` - RAG service implementation
   - `internal/embeddings/` - Ollama embedding service
   - `internal/vectordb/` - ChromaDB client
   - `internal/llm/` - Ollama LLM service

2. **Post-Processing Hook**:
   - `internal/transcription/post_processing.go` - Automatic summarization and RAG storage

3. **API Endpoints**:
   - `POST /api/v1/rag/chat` - RAG query endpoint
   - `GET /api/v1/rag/stats` - RAG statistics
   - `POST /api/v1/rag/backfill` - Backfill existing transcriptions

4. **Frontend Components**:
   - `web/frontend/src/pages/GlobalChatPage.tsx` - Global chat interface
   - `web/frontend/src/components/GlobalChatInterface.tsx` - Chat UI component
   - `web/frontend/src/components/RAGStatus.tsx` - RAG status dashboard

5. **Configuration**:
   - Extended `internal/config/config.go` with RAG settings
   - Updated `docker-compose.yml` with ChromaDB service

## 🙏 Credits

- **Original Project**: [Scriberr](https://github.com/rishikanthc/Scriberr) by [rishikanthc](https://github.com/rishikanthc)
- **RAG Extension**: This fork extends Scriberr with RAG capabilities

## 📄 License

This project maintains the same license as the original Scriberr project.

## 🔗 Related Documentation

- [RAG Setup Guide](./RAG_SETUP.md) - Detailed RAG configuration and troubleshooting
- [Scriberr Documentation](https://github.com/rishikanthc/Scriberr) - Original project documentation

## 🤝 Contributing

This is a fork of Scriberr with RAG extensions. Contributions that improve the RAG functionality or maintain compatibility with upstream Scriberr are welcome.

## 📧 Support

For issues related to:
- **RAG functionality**: Open an issue in this repository
- **Core Scriberr features**: Refer to the [original Scriberr repository](https://github.com/rishikanthc/Scriberr)

---

**Note**: This project extends Scriberr with RAG capabilities. Core transcription functionality is provided by the original Scriberr project.
