package main

import (
	"fmt"
	"net/http"

	handler "github.com/frasnym/go-expense-telebot/api"
	"github.com/frasnym/go-expense-telebot/config"
)

func main() {
	cfg := config.GetConfig()

	http.HandleFunc("/", handler.IndexHandler)
	http.HandleFunc("/webhook", handler.WebhookHandler)
	fmt.Printf("Server is running on port %s...\n", cfg.Port)
	http.ListenAndServe(fmt.Sprint(":", cfg.Port), nil)
}
