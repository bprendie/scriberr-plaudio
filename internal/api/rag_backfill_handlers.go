package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"scriberr/internal/database"
	"scriberr/internal/models"
	"scriberr/internal/transcription/interfaces"

	"github.com/gin-gonic/gin"
)

// BackfillRAG processes all completed transcriptions and stores them in RAG
// @Summary Backfill RAG with existing transcriptions
// @Description Process all completed transcriptions and store them in the RAG system
// @Tags rag
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]string
// @Security ApiKeyAuth
// @Security BearerAuth
// @Router /api/v1/rag/backfill [post]
func (h *Handler) BackfillRAG(c *gin.Context) {
	if h.ragService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "RAG service not initialized"})
		return
	}

	// Get all completed transcriptions
	var jobs []models.TranscriptionJob
	if err := database.DB.Where("status = ?", models.StatusCompleted).
		Where("transcript IS NOT NULL AND transcript != ''").
		Find(&jobs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch transcriptions"})
		return
	}

	processed := 0
	failed := 0

	// Process each job
	for _, job := range jobs {
		if job.Transcript == nil || *job.Transcript == "" {
			continue
		}

		// Extract text from JSON transcript
		transcriptText, err := extractTextFromTranscript(*job.Transcript)
		if err != nil {
			// Fallback: use raw transcript if JSON parsing fails
			transcriptText = *job.Transcript
		}

		if strings.TrimSpace(transcriptText) == "" {
			failed++
			continue
		}

		// Get summary if available
		summary := ""
		if job.Summary != nil {
			summary = *job.Summary
		}

		// Store in RAG
		if err := h.ragService.StoreSummary(job.ID, summary, transcriptText); err != nil {
			failed++
			continue
		}
		processed++
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Backfill completed",
		"total":    len(jobs),
		"processed": processed,
		"failed":   failed,
	})
}

// extractTextFromTranscript extracts the text content from a JSON transcript (same logic as post-processing)
func extractTextFromTranscript(transcriptJSON string) (string, error) {
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
