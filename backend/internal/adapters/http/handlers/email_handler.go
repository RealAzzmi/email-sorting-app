package handlers

import (
	"net/http"
	"strconv"

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

	emails, err := h.emailUsecase.GetAccountEmails(c.Request.Context(), accountID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"emails": emails})
}