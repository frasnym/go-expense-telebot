package notification

import (
	"context"
	"fmt"

	"github.com/frasnym/go-expense-telebot/common/logger"
	"github.com/frasnym/go-expense-telebot/pkg/session"
	"github.com/frasnym/go-expense-telebot/repository"
)

type NotificationClient interface {
	NotifySendToChat(ctx context.Context, userID int, msg string) error
}

type notificationClient struct {
	botRepo repository.BotRepository
}

func (c *notificationClient) NotifySendToChat(ctx context.Context, userID int, msg string) error {
	var err error
	defer func() {
		logger.LogService(ctx, "SpendeeNotifyError", err)
	}()

	chatID, err := session.GetChatID(userID)
	if err != nil {
		err = fmt.Errorf("err session.GetChatID: %w", err)
		return err
	}

	_, err = c.botRepo.SendTextMessage(ctx, chatID, msg)
	if err != nil {
		err = fmt.Errorf("err botRepo.SendMessage: %w", err)
		return err
	}

	return nil
}

func New(botRepo repository.BotRepository) NotificationClient {
	return &notificationClient{botRepo: botRepo}
}
