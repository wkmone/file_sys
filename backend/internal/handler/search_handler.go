package handler

import (
	"file_sys/backend/internal/middleware"
	"file_sys/backend/internal/service"
	"file_sys/backend/internal/util"

	"github.com/gin-gonic/gin"
)

type SearchHandler struct {
	searchService *service.SearchService
}

func NewSearchHandler(searchService *service.SearchService) *SearchHandler {
	return &SearchHandler{searchService: searchService}
}

func (h *SearchHandler) Search(c *gin.Context) {
	q := c.Query("q")
	if q == "" {
		util.ValidationError(c, "missing search query")
		return
	}

	files, folders, err := h.searchService.Search(c.Request.Context(), middleware.GetUserID(c), q)
	if err != nil {
		util.InternalError(c, "search failed")
		return
	}

	util.Success(c, gin.H{"files": files, "folders": folders})
}
