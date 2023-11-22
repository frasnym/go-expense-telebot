package service

import (
	"context"
	"encoding/csv"
	"fmt"
	"net/http"
	"time"

	"github.com/frasnym/go-expense-telebot/common"
	"github.com/frasnym/go-expense-telebot/common/logger"
	"github.com/frasnym/go-expense-telebot/pkg/session"
	"github.com/frasnym/go-expense-telebot/repository"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// SpendeeService is an interface for managing Spendee-related actions.
type SpendeeService interface {
	Request(ctx context.Context, userID int, chatID int64) error
	Processor(ctx context.Context, userID int, input string) error
}

type whatsappSvc struct {
	botRepo    repository.BotRepository
	gsheetRepo repository.GSheetRepository
}

func (s *whatsappSvc) Request(ctx context.Context, userID int, chatID int64) error {
	var err error
	defer func() {
		logger.LogService(ctx, "SpendeeRequest", err)
	}()

	// Start a new session for the user
	session.NewSession(userID, chatID, common.CommandUploadSpendee)

	// Send a request for the document
	replyTxt := "Please upload your Spendee CSV document"
	msg, err := s.botRepo.SendTextMessage(ctx, chatID, replyTxt)
	if err != nil {
		err = fmt.Errorf("error sending text message: %w", err)
		return err
	}

	// Set the message ID in the user's session
	if err := session.SetMessageID(userID, msg.MessageID); err != nil {
		err = fmt.Errorf("error setting message ID: %w", err)
		return err
	}
	return nil
}

// Processor processes the user's input (document) for expense report.
func (s *whatsappSvc) Processor(ctx context.Context, userID int, fileID string) error {
	var err error
	defer func() {
		logger.LogService(ctx, "SpendeeProcessor", err)
	}()

	if session.IsInteractionTimedOut(userID) {
		err = s.notifyError(ctx, userID, "Request Timeout")
		if err != nil {
			err = fmt.Errorf("err notifyError: %w", err)
		}

		session.DeleteUserSession(userID)
		return err
	}

	// Get file url
	fileUrl, errDoc := s.botRepo.GetFileURL(ctx, fileID)
	if errDoc != nil {
		err = fmt.Errorf("err botRepo.GetFileURL: %w", errDoc)
		return err
	}

	// TODO: Reject if not csv

	// Get file content
	resp, errDoc := http.Get(fileUrl)
	if errDoc != nil {
		err = fmt.Errorf("err http.Get: %w", errDoc)
		return err
	}
	defer resp.Body.Close()

	// Parse CSV content line by line
	gsheetInputMap := map[string][][]any{}
	reader := csv.NewReader(resp.Body)
	for {
		// Read one line
		record, errDoc := reader.Read()
		if errDoc != nil {
			err = fmt.Errorf("err reader.Read: %w", errDoc)
			logger.Warn(ctx, err.Error())
			break
		}

		// Skip Header
		if record[0] == "Date" {
			continue
		}

		// Only process ended month
		currentYear, currentMonth, _ := time.Now().Date()
		thisBeginningMonth := time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, time.UTC)
		expenseDate := common.ParseSpendeeDate(record[0])

		if expenseDate.After(thisBeginningMonth) {
			logger.Warn(ctx, fmt.Sprintf("can only process ended month: %s", expenseDate.Format("2006-01-02")))
			break
		}

		dateMonthFormat := expenseDate.Format("01")
		gsheetInputMap[dateMonthFormat] = append(
			gsheetInputMap[dateMonthFormat],
			[]any{
				expenseDate.Format("2006-01-02 15:04:05"), // Date
				record[3], // Category
				record[4], // Amount
				record[6], // Note
				record[7], // Label
			},
		)
	}

	// Insert header
	for k, v := range gsheetInputMap {
		gsheetInputMap[k] = common.InsertAndShift[[]any](v, []any{"date", "category", "amount", "note", "label"})
	}

	// Write to gsheet
	for period, v := range gsheetInputMap {
		// Check, don't write if already available
		targetRange := fmt.Sprintf("%s!A1", period)
		gsheetValues, errDoc := s.gsheetRepo.GetValues(ctx, targetRange)
		if errDoc != nil {
			err = fmt.Errorf("err gsheetRepo.GetValues: %w", errDoc)
			return err
		}

		if len(gsheetValues.Values) > 0 {
			logger.Warn(ctx, fmt.Sprintf("data already written: %s, skipping...", targetRange))
			continue
		}

		// Write
		err = s.gsheetRepo.AppendRow(ctx, period, v)
		if err != nil {
			err = fmt.Errorf("err gsheetRepo.AppendRow: %w", err)
			return err
		}
	}

	err = nil
	return nil
}

// NewSpendeeService creates a new SpendeeService using the provided bot repository.
func NewSpendeeService(botRepo *repository.BotRepository, gsheetRepo *repository.GSheetRepository) SpendeeService {
	return &whatsappSvc{botRepo: *botRepo, gsheetRepo: *gsheetRepo}
}

func (s *whatsappSvc) notifyError(ctx context.Context, userID int, msg string) error {
	var err error
	defer func() {
		logger.LogService(ctx, "SpendeeNotifyError", err)
	}()

	chatID, err := session.GetChatID(userID)
	if err != nil {
		err = fmt.Errorf("err session.GetChatID: %w", err)
		return err
	}

	messageID, err := session.GetMessageID(userID)
	if err != nil {
		err = fmt.Errorf("err session.GetMessageID: %w", err)
		return err
	}

	editMsg := tgbotapi.NewEditMessageText(chatID, messageID, msg)
	_, err = s.botRepo.SendMessage(ctx, editMsg)
	if err != nil {
		err = fmt.Errorf("err botRepo.SendMessage: %w", err)
		return err
	}

	return nil
}
