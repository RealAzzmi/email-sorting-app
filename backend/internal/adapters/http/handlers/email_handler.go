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

func (h *EmailHandler) GetEmailsByCategory(c *gin.Context) {
	accountID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid account ID"})
		return
	}

	categoryID, err := strconv.ParseInt(c.Param("categoryId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category ID"})
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

	paginatedEmails, err := h.emailUsecase.GetEmailsByCategory(c.Request.Context(), accountID, categoryID, params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, paginatedEmails)
}

func (h *EmailHandler) GenerateEmailSummary(c *gin.Context) {
	emailID, err := strconv.ParseInt(c.Param("emailId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email ID"})
		return
	}

	err = h.emailUsecase.GenerateEmailSummary(c.Request.Context(), emailID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Email summary generated successfully"})
}

func (h *EmailHandler) CategorizeEmailWithAI(c *gin.Context) {
	emailID, err := strconv.ParseInt(c.Param("emailId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email ID"})
		return
	}

	err = h.emailUsecase.CategorizeEmailWithAI(c.Request.Context(), emailID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Email categorized successfully"})
}

func (h *EmailHandler) UnsubscribeFromEmail(c *gin.Context) {
	emailID, err := strconv.ParseInt(c.Param("emailId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email ID"})
		return
	}

	result, err := h.emailUsecase.UnsubscribeFromEmail(c.Request.Context(), emailID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

type BulkUnsubscribeRequest struct {
	EmailIDs []int64 `json:"email_ids" binding:"required"`
}

func (h *EmailHandler) BulkUnsubscribe(c *gin.Context) {
	var req BulkUnsubscribeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(req.EmailIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No email IDs provided"})
		return
	}

	results, err := h.emailUsecase.BulkUnsubscribe(c.Request.Context(), req.EmailIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"results": results})
}
