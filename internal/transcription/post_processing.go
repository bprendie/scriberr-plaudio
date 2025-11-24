package transcription

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"scriberr/internal/database"
	"scriberr/internal/llm"
	"scriberr/internal/models"
	"scriberr/internal/rag"
	"scriberr/internal/transcription/interfaces"
)

// LLMService interface for post-processing
type LLMService interface {
	ChatCompletion(ctx context.Context, model string, messages []llm.ChatMessage, temperature float64) (*llm.ChatResponse, error)
}

// PostProcessingHook handles post-transcription processing (summarization, RAG storage)
type PostProcessingHook struct {
	ragService *rag.RAGService
	llmService LLMService
	llmModel   string
}

// NewPostProcessingHook creates a new post-processing hook
func NewPostProcessingHook(ragService *rag.RAGService, llmService LLMService, llmModel string) *PostProcessingHook {
	return &PostProcessingHook{
		ragService: ragService,
		llmService: llmService,
		llmModel:   llmModel,
	}
}

// OnTranscriptionCompleted is called when a transcription job completes
func (h *PostProcessingHook) OnTranscriptionCompleted(jobID string) {
	if h.ragService == nil {
		log.Printf("[post-processing] RAG service not initialized, skipping")
		return
	}

	// Get the completed job
	var job models.TranscriptionJob
	if err := database.DB.Where("id = ?", jobID).First(&job).Error; err != nil {
		log.Printf("[post-processing] Failed to get job %s: %v", jobID, err)
		return
	}

	if job.Status != models.StatusCompleted {
		log.Printf("[post-processing] Job %s not completed, status: %s", jobID, job.Status)
		return
	}

	if job.Transcript == nil || *job.Transcript == "" {
		log.Printf("[post-processing] Job %s has no transcript", jobID)
		return
	}

	ctx := context.Background()
	transcriptJSON := *job.Transcript
	
	// Extract text from JSON transcript
	transcriptText, err := h.extractTextFromTranscript(transcriptJSON)
	if err != nil {
		log.Printf("[post-processing] Failed to extract text from transcript for job %s: %v", jobID, err)
		// Fallback: use raw transcript if JSON parsing fails
		transcriptText = transcriptJSON
	}

	if strings.TrimSpace(transcriptText) == "" {
		log.Printf("[post-processing] Job %s has empty transcript text", jobID)
		return
	}

	log.Printf("[post-processing] Processing job %s, transcript length: %d chars", jobID, len(transcriptText))

	// Generate summary using LLM (but don't fail if this fails)
	var summary string
	summary, err = h.generateSummary(ctx, transcriptText)
	if err != nil {
		log.Printf("[post-processing] Failed to generate summary for job %s: %v (will store transcript without summary)", jobID, err)
		summary = "" // Empty summary, but we'll still store the transcript
	} else {
		log.Printf("[post-processing] Generated summary for job %s, length: %d", jobID, len(summary))
		
		// Update job with summary
		job.Summary = &summary
		if err := database.DB.Save(&job).Error; err != nil {
			log.Printf("[post-processing] Failed to save summary for job %s: %v", jobID, err)
		}
	}

	// Store in vector database for RAG (even if summary failed)
	if err := h.ragService.StoreSummary(jobID, summary, transcriptText); err != nil {
		log.Printf("[post-processing] Failed to store in vector DB for job %s: %v", jobID, err)
		return
	}

	log.Printf("[post-processing] Successfully stored job %s in RAG (summary: %v)", jobID, summary != "")
}

// extractTextFromTranscript extracts the text content from a JSON transcript
func (h *PostProcessingHook) extractTextFromTranscript(transcriptJSON string) (string, error) {
	// Try to parse as TranscriptResult JSON
	var result interfaces.TranscriptResult
	if err := json.Unmarshal([]byte(transcriptJSON), &result); err == nil {
		// If we have text, use it
		if result.Text != "" {
			return result.Text, nil
		}
		// Otherwise, reconstruct from segments
		if len(result.Segments) > 0 {
			var textBuilder strings.Builder
			for _, segment := range result.Segments {
				if segment.Text != "" {
					if textBuilder.Len() > 0 {
						textBuilder.WriteString(" ")
					}
					textBuilder.WriteString(segment.Text)
				}
			}
			return textBuilder.String(), nil
		}
		return "", fmt.Errorf("no text found in transcript result")
	}
	
	// If JSON parsing fails, try to extract text from a simple JSON structure
	var simpleResult struct {
		Text string `json:"text"`
	}
	if err := json.Unmarshal([]byte(transcriptJSON), &simpleResult); err == nil && simpleResult.Text != "" {
		return simpleResult.Text, nil
	}
	
	// Last resort: if it's not JSON, assume it's plain text
	if !strings.HasPrefix(strings.TrimSpace(transcriptJSON), "{") {
		return transcriptJSON, nil
	}
	
	return "", fmt.Errorf("unable to extract text from transcript")
}

// generateSummary generates a summary using the LLM
func (h *PostProcessingHook) generateSummary(ctx context.Context, transcriptText string) (string, error) {
	// Limit transcript length for summary generation (to avoid token limits)
	maxTranscriptLength := 10000
	textForSummary := transcriptText
	if len(transcriptText) > maxTranscriptLength {
		textForSummary = transcriptText[:maxTranscriptLength] + "... [truncated]"
	}
	
	// Create summary prompt
	prompt := fmt.Sprintf("Please provide a concise summary of the following transcription:\n\n%s", textForSummary)
	
	messages := []llm.ChatMessage{
		{Role: "user", Content: prompt},
	}

	response, err := h.llmService.ChatCompletion(ctx, h.llmModel, messages, 0.7)
	if err != nil {
		return "", fmt.Errorf("LLM call failed: %w", err)
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("no response from LLM")
	}

	return response.Choices[0].Message.Content, nil
}
