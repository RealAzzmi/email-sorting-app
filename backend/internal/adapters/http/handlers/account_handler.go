package handlers

import (
	"net/http"
	"strconv"

	"github.com/email-sorting-app/internal/usecases"
	"github.com/gin-gonic/gin"
)

type AccountHandler struct {
	accountUsecase *usecases.AccountUsecase
}

func NewAccountHandler(accountUsecase *usecases.AccountUsecase) *AccountHandler {
	return &AccountHandler{
		accountUsecase: accountUsecase,
	}
}

func (h *AccountHandler) GetAccounts(c *gin.Context) {
	accounts, err := h.accountUsecase.GetAllAccounts(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch accounts"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"accounts": accounts})
}

func (h *AccountHandler) DeleteAccount(c *gin.Context) {
	accountID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid account ID"})
		return
	}

	err = h.accountUsecase.DeleteAccount(c.Request.Context(), accountID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete account"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Account deleted successfully"})
}
