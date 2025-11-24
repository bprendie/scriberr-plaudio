package api

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// RAGChatRequest represents a RAG chat request
type RAGChatRequest struct {
	Query     string  `json:"query" binding:"required"`
	Model     string  `json:"model" binding:"required"`
	Temperature float64 `json:"temperature,omitempty"`
}

// RAGChat handles RAG-enhanced chat queries
// @Summary RAG chat query
// @Description Query across all transcriptions using RAG
// @Tags rag
// @Accept json
// @Produce json
// @Param request body RAGChatRequest true "RAG chat request"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security ApiKeyAuth
// @Security BearerAuth
// @Router /api/v1/rag/chat [post]
func (h *Handler) RAGChat(c *gin.Context) {
	var req RAGChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if h.ragService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "RAG service not initialized"})
		return
	}

	if req.Temperature == 0 {
		req.Temperature = 0.7
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Minute)
	defer cancel()

	response, err := h.ragService.Chat(ctx, req.Query, req.Model, req.Temperature)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"response": response,
		"query":    req.Query,
	})
}

// RAGStats returns statistics about the RAG system
// @Summary Get RAG statistics
// @Description Get statistics about transcripts stored in RAG
// @Tags rag
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]string
// @Security ApiKeyAuth
// @Security BearerAuth
// @Router /api/v1/rag/stats [get]
func (h *Handler) RAGStats(c *gin.Context) {
	if h.ragService == nil {
		c.JSON(http.StatusOK, gin.H{
			"status":         "inactive",
			"transcript_count": 0,
			"message":       "RAG service not initialized",
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	stats, err := h.ragService.GetStats(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}
