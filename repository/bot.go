package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/frasnym/go-expense-telebot/common"
	"github.com/frasnym/go-expense-telebot/common/logger"
	"github.com/frasnym/go-expense-telebot/config"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// BotRepository is an interface for managing interactions with a Telegram bot.
type BotRepository interface {
	SetWebhook(ctx context.Context) error
	GetUpdate(ctx context.Context, r io.Reader) (*tgbotapi.Update, error)
	SendMessage(ctx context.Context, c tgbotapi.Chattable) (*tgbotapi.Message, error)
	SendTextMessage(ctx context.Context, chatID int64, text string) (*tgbotapi.Message, error)
	DeleteMessage(ctx context.Context, chatID int64, messageID int) (*tgbotapi.Message, error)
	GetFileURL(ctx context.Context, fileID string) (string, error)
}

type botRepo struct {
	bot *tgbotapi.BotAPI
	cfg *config.Config
}

// SendMessage sends a message using the Telegram bot.
func (s *botRepo) SendMessage(ctx context.Context, c tgbotapi.Chattable) (*tgbotapi.Message, error) {
	var err error
	defer func() {
		logger.LogService(ctx, "BotSendMessage", err)
	}()

	msg, err := s.bot.Send(c)
	if err != nil {
		err = fmt.Errorf("err bot.Send: %w", err)
		return nil, err
	}

	return &msg, nil
}

// GetUpdate decodes an update from the provided reader.
func (*botRepo) GetUpdate(ctx context.Context, r io.Reader) (*tgbotapi.Update, error) {
	var err error
	defer func() {
		logger.LogService(ctx, "BotGetUpdate", err)
	}()

	update := tgbotapi.Update{}
	if err := json.NewDecoder(r).Decode(&update); err != nil {
		err = fmt.Errorf("err json.NewDecoder(r).Decode: %w", err)
		return nil, err
	}

	return &update, nil
}

// SendTextMessage sends a text message to a specific chat.
func (s *botRepo) SendTextMessage(ctx context.Context, chatID int64, text string) (*tgbotapi.Message, error) {
	var err error
	defer func() {
		logger.LogService(ctx, "BotSendTextMessage", err)
	}()

	stringMsg := tgbotapi.NewMessage(chatID, text)
	msg, err := s.bot.Send(stringMsg)
	if err != nil {
		err = fmt.Errorf("err bot.Send: %w", err)
		return nil, err
	}

	return &msg, nil
}

// SetWebhook sets up the bot's webhook for receiving updates.
func (s *botRepo) SetWebhook(ctx context.Context) error {
	var err error
	defer func() {
		logger.LogService(ctx, "BotSetWebhook", err)
	}()

	webhookURL := fmt.Sprintf("https://%s/webhook", s.cfg.VercelUrl)

	info, err := s.bot.GetWebhookInfo()
	if err != nil {
		err = fmt.Errorf("err bot.GetWebhookInfo: %w", err)
		return err
	}
	if info.URL == webhookURL {
		return common.ErrNoChanges
	}

	_, err = s.bot.SetWebhook(tgbotapi.NewWebhook(webhookURL))
	if err != nil {
		err = fmt.Errorf("err bot.SetWebhook: %w", err)
		return err
	}

	return nil
}

func (r *botRepo) DeleteMessage(ctx context.Context, chatID int64, messageID int) (*tgbotapi.Message, error) {
	var err error
	defer func() {
		logger.LogService(ctx, "BotDeleteMessage", err)
	}()

	del := tgbotapi.NewDeleteMessage(chatID, messageID)
	msg, err := r.SendMessage(ctx, del)
	if err != nil {
		err = fmt.Errorf("err SendMessage: %w", err)
		return nil, err
	}

	return msg, nil
}

func (r *botRepo) GetFileURL(ctx context.Context, fileID string) (string, error) {
	var err error
	defer func() {
		logger.LogService(ctx, "BotGetFileURL", err)
	}()

	// Get information about the file
	file, err := r.bot.GetFile(tgbotapi.FileConfig{FileID: fileID})
	if err != nil {
		return "", nil
	}

	// Build the URL to download the file
	fileURL := "https://api.telegram.org/file/bot" + r.bot.Token + "/" + file.FilePath

	return fileURL, nil

	// fmt.Printf("fileURL: %v\n", fileURL)

	// // Download the file
	// resp, err := http.Get(fileURL)
	// if err != nil {
	// 	return nil
	// }
	// defer resp.Body.Close()

	// // Read the content of the file
	// fileContent, err := ioutil.ReadAll(resp.Body)
	// if err != nil {
	// 	return nil
	// }

	// // Process the file content (in this example, simply print it)
	// log.Printf("Received file content:\n%s", string(fileContent))
	// return nil

}

// NewBotRepository creates a new BotRepository using the provided configuration and Telegram bot.
func NewBotRepository(cfg *config.Config, bot *tgbotapi.BotAPI) BotRepository {
	return &botRepo{cfg: cfg, bot: bot}
}
