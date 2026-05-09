package handler

import "github.com/gin-gonic/gin"

type FinancesHandler struct{}

func NewFinancesHandler() FinancesHandler {
	return FinancesHandler{}
}

func (h FinancesHandler) Wallet(c *gin.Context) {
	notImplemented(c, "GET /api/wallet")
}

func (h FinancesHandler) WalletInformation(c *gin.Context) {
	notImplemented(c, "GET /api/wallet/information")
}

func (h FinancesHandler) WalletTimeline(c *gin.Context) {
	notImplemented(c, "GET /api/wallet/time_line")
}

func (h FinancesHandler) WalletStatementList(c *gin.Context) {
	notImplemented(c, "GET /api/wallet/statement_list")
}
