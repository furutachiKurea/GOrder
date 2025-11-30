package main

import (
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

type PaymentHandler struct {
}

func NewPaymentHandler() *PaymentHandler {
	return &PaymentHandler{}
}

func (h PaymentHandler) RegisterRoutes(router *gin.Engine) {
	router.POST("/api/webhook", h.handleWebhook)
}

func (h PaymentHandler) handleWebhook(c *gin.Context) {
	log.Info().Msg("Got webhook from stripe")
}
