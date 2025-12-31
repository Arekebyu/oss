package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	Service *SearchService
}

type SearchRequest struct {
	Query string `form:"q" binding:"required"`
}

func (h *Handler) HandleSearch(c *gin.Context) {
	var req SearchRequest

	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Query parameter 'q' is required"})
		return
	}

	results, err := h.Service.SearchAndRank(c.Request.Context(), req.Query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Search failed " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"query":   req.Query,
		"count":   len(results),
		"results": results,
	})
}
