package handler

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/frasnym/go-expense-telebot/common"
	"github.com/frasnym/go-expense-telebot/common/ctxdata"
	"github.com/frasnym/go-expense-telebot/common/logger"
	"github.com/frasnym/go-expense-telebot/config"
	"github.com/frasnym/go-expense-telebot/pkg/gsheet"
	"github.com/frasnym/go-expense-telebot/pkg/session"
	"github.com/frasnym/go-expense-telebot/pkg/telebot"
	"github.com/frasnym/go-expense-telebot/repository"
	"github.com/frasnym/go-expense-telebot/service"
)

// WebhookHandler handles incoming HTTP requests for a Telegram bot's webhook.
// It processes updates, handles commands, and takes appropriate actions based on the received messages.
// If any errors occur during the process, they are logged.
// After processing the request, it writes a "Webhook OK" message to the response writer (w).
func WebhookHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	ctx := ctxdata.EnsureCorrelationIDExist(r)

	// Log any errors and write "Webhook OK" as the API response
	defer func() {
		logger.LogService(ctx, "WebhookHandler", err)
		fmt.Fprint(w, "Webhook OK")
	}()

	// Create a new bot repository with the application's configuration and Telegram bot
	cfg := config.GetConfig()
	botRepo := repository.NewBotRepository(cfg, telebot.GetBot())
	gsheetRepo := repository.NewGSheetRepository(cfg, gsheet.GetService())
	spendeeSvc := service.NewSpendeeService(&botRepo, &gsheetRepo)

	// Get the update from the request body
	update, err := botRepo.GetUpdate(ctx, r.Body)
	if err != nil {
		err = fmt.Errorf("err botRepo.GetUpdate: %w", err)
		return
	}

	// Handle messages and commands
	if update.Message != nil {
		userID := update.Message.From.ID
		chatID := update.Message.Chat.ID

		// Handle commands
		if update.Message.IsCommand() {
			switch update.Message.Command() {
			case common.CommandUploadSpendee:
				if err = spendeeSvc.Request(ctx, userID, chatID); err != nil {
					err = fmt.Errorf("err spendeeSvc.Request: %w", err)
				}
				return
			default:
				err = fmt.Errorf("invalid command: %s", update.Message.Command())
				return
			}
		}

		// Get the user's current action
		action, errSession := session.GetAction(userID)
		if errSession != nil {
			logger.Warn(ctx, fmt.Sprintf("session.GetAction: %s", errSession.Error()))
			return
		}

		// Handle requests based on the user's current action
		switch action {
		case common.CommandUploadSpendee:
			if update.Message.Document != nil {
				fileID := update.Message.Document.FileID
				if err = spendeeSvc.Processor(ctx, userID, fileID); err != nil {
					err = fmt.Errorf("err spendeeSvc.Processor: %w", err)
				}
				return
			} else {
				// TODO: Handle not file
				err = errors.New("no file uploaded")
			}

			return
		}

		err = fmt.Errorf("unprocessable text: %s", update.Message.Text)
	}
}
