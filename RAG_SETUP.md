# RAG Setup Guide

## Overview

The RAG (Retrieval-Augmented Generation) system automatically stores all transcriptions in a vector database (ChromaDB) for semantic search. This allows you to query across all your transcriptions using natural language in the Global Chat interface.

## How It Works

### Automatic Processing

1. **Upload & Transcribe**: When you upload a WAV/MP3 file, it's automatically transcribed
2. **Auto-Summarization**: After transcription completes, the system:
   - Extracts the text from the JSON transcript
   - Generates a summary using Ollama (if available)
   - Stores both summary and transcript in ChromaDB
3. **Vector Storage**: The content is embedded using `nomic-embed-text` and stored for semantic search

### Components

- **ChromaDB**: Vector database for storing embeddings
- **Ollama Embeddings**: Uses `nomic-embed-text` model for generating embeddings
- **Ollama LLM**: Uses configured model (default: `llama3.2`) for summarization and chat
- **Post-Processing Hook**: Automatically processes completed transcriptions

## Configuration

The system uses these environment variables (set in `docker-compose.yml`):

```env
OLLAMA_URL=http://10.0.0.50:11434          # Your Ollama server
CHROMADB_URL=http://chromadb:8000          # ChromaDB service URL
EMBEDDING_MODEL=nomic-embed-text           # Embedding model name
OLLAMA_MODEL=llama3.2                     # LLM model for summarization/chat
```

## Prerequisites

Ensure Ollama has the required models installed:

```bash
# On your Ollama server (10.0.0.50)
ollama pull nomic-embed-text
ollama pull llama3.2  # or your preferred model
```

## Using Global Chat

1. Navigate to Global Chat:
   - Click the menu (☰) → "Global Chat"
   - Or click "Global Chat" button on the homepage
   - Or go to `/global-chat`

2. Select a model from the dropdown

3. Ask questions about your transcriptions:
   - "What was discussed about project deadlines?"
   - "Summarize all meetings from last week"
   - "What topics were covered in the technical discussions?"

The system will:
- Search the vector database for relevant transcripts
- Retrieve the most relevant context
- Generate answers using the selected Ollama model

## Backfilling Existing Transcriptions

If you have existing transcriptions that weren't automatically processed, you can backfill them:

```bash
curl -X POST http://localhost:8080/api/v1/rag/backfill \
  -H "Authorization: Bearer YOUR_TOKEN"
```

This will process all completed transcriptions and store them in the RAG system.

## Troubleshooting

### Transcripts Not Appearing in Search

1. **Check logs**: Look for `[post-processing]` messages in container logs:
   ```bash
   docker compose logs scriberr | grep post-processing
   ```

2. **Verify RAG is initialized**: Check startup logs for:
   ```
   RAG services initialized
   ```

3. **Check ChromaDB**: Ensure ChromaDB container is running:
   ```bash
   docker compose ps chromadb
   ```

4. **Verify Ollama connection**: Ensure Ollama is accessible from the container:
   ```bash
   docker compose exec scriberr curl http://10.0.0.50:11434/api/tags
   ```

### Summary Generation Failing

- Transcripts are still stored in RAG even if summary generation fails
- Check Ollama logs and ensure the model is available
- Verify `OLLAMA_MODEL` environment variable matches an installed model

### No Results in Global Chat

- Ensure transcriptions have been completed and processed
- Check that ChromaDB has data:
  ```bash
  docker compose exec chromadb curl http://localhost:8000/api/v1/collections
  ```
- Try backfilling existing transcriptions

## Architecture

```
┌─────────────┐
│   Upload    │
│  WAV/MP3    │
└──────┬──────┘
       │
       ▼
┌─────────────┐
│ Auto-       │
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

## API Endpoints

- `POST /api/v1/rag/chat` - Query RAG system
- `POST /api/v1/rag/backfill` - Backfill existing transcriptions

## Notes

- Transcripts are stored even if summary generation fails
- The system extracts text from JSON transcripts automatically
- Long transcripts are truncated for summary generation (10k chars) but full transcript is stored
- Each transcription is stored as a single vector document (summary + transcript)
