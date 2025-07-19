package handlers

import (
	"net/http"
	"strconv"

	"github.com/email-sorting-app/internal/usecases"
	"github.com/gin-gonic/gin"
)

type CategoryHandler struct {
	categoryUsecase *usecases.CategoryUsecase
}

func NewCategoryHandler(categoryUsecase *usecases.CategoryUsecase) *CategoryHandler {
	return &CategoryHandler{
		categoryUsecase: categoryUsecase,
	}
}

func (h *CategoryHandler) GetAccountCategories(c *gin.Context) {
	accountID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid account ID"})
		return
	}

	categories, err := h.categoryUsecase.GetAccountCategories(c.Request.Context(), accountID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"categories": categories})
}