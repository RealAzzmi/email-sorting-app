package handlers

import (
	"net/http"
	"strconv"

	"github.com/email-sorting-app/internal/domain/repositories"
	"github.com/email-sorting-app/internal/usecases"
	"github.com/gin-gonic/gin"
)

type EmailHandler struct {
	emailUsecase *usecases.EmailUsecase
}

func NewEmailHandler(emailUsecase *usecases.EmailUsecase) *EmailHandler {
	return &EmailHandler{
		emailUsecase: emailUsecase,
	}
}

func (h *EmailHandler) GetAccountEmails(c *gin.Context) {
	accountID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid account ID"})
		return
	}

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	// Validate pagination parameters
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	params := repositories.PaginationParams{
		Page:     page,
		PageSize: pageSize,
	}

	paginatedEmails, err := h.emailUsecase.GetAccountEmailsPaginated(c.Request.Context(), accountID, params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, paginatedEmails)
}

func (h *EmailHandler) RefreshAccountEmails(c *gin.Context) {
	accountID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid account ID"})
		return
	}

	err = h.emailUsecase.RefreshAccountEmails(c.Request.Context(), accountID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Emails refreshed successfully"})
}