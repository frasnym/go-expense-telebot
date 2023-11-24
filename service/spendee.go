package service

import (
	"context"
	"encoding/csv"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/frasnym/go-expense-telebot/common"
	"github.com/frasnym/go-expense-telebot/common/logger"
	"github.com/frasnym/go-expense-telebot/common/notification"
	"github.com/frasnym/go-expense-telebot/pkg/session"
	"github.com/frasnym/go-expense-telebot/repository"
)

// SpendeeService is an interface for managing Spendee-related actions.
type SpendeeService interface {
	Request(ctx context.Context, userID int, chatID int64) error
	Processor(ctx context.Context, userID int, input string) error
}

type spendeeSvc struct {
	botRepo    repository.BotRepository
	gsheetRepo repository.GSheetRepository

	notificationClient notification.NotificationClient
}

func (s *spendeeSvc) Request(ctx context.Context, userID int, chatID int64) error {
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
func (s *spendeeSvc) Processor(ctx context.Context, userID int, fileID string) error {
	var err error
	var result []string
	defer func() {
		logger.LogService(ctx, "SpendeeProcessor", err)
	}()

	if session.IsInteractionTimedOut(userID) {
		s.notificationClient.NotifySendToChat(ctx, userID, "Request timeout")
		session.DeleteUserSession(userID)
		return err
	}

	defer func() {
		// Notify result
		resultMsg := "Finished"
		for _, v := range result {
			resultMsg = fmt.Sprintf("%s\n- %s", resultMsg, v)
		}
		resultMsg = fmt.Sprintf("%s\n\nURL: %s", resultMsg, "TBA")

		s.notificationClient.NotifySendToChat(ctx, userID, resultMsg)
		session.DeleteUserSession(userID)
	}()

	// Get file url
	fileUrl, errDoc := s.botRepo.GetFileURL(ctx, fileID)
	if errDoc != nil {
		err = fmt.Errorf("err botRepo.GetFileURL: %w", errDoc)
		return err
	}

	// Reject if not csv
	if !strings.HasSuffix(strings.ToLower(fileUrl), ".csv") {
		s.notificationClient.NotifySendToChat(ctx, userID, "File must be csv, please upload again")

		err = session.ResetTimer(userID)
		if err != nil {
			err = fmt.Errorf("err session.ResetTimer: %w", err)
		}
		return err
	}

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
			msg := fmt.Sprintf("can only process ended month: %s", expenseDate.Format("2006-01-02"))
			result = append(result, msg)
			logger.Warn(ctx, msg)
			break
		}

		dateMonthFormat := expenseDate.Format("01")
		gsheetInputMap[dateMonthFormat] = append(
			gsheetInputMap[dateMonthFormat],
			[]any{
				expenseDate.Format("2006-01-02 15:04:05"), // Date
				record[3],                              // Category
				strings.Replace(record[4], "-", "", 1), // Amount
				record[6],                              // Note
				record[7],                              // Label
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
			msg := fmt.Sprintf("data already written: %s, skipping...", targetRange)
			result = append(result, msg)
			logger.Warn(ctx, msg)
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
func NewSpendeeService(botRepo *repository.BotRepository, gsheetRepo *repository.GSheetRepository, notificationClient *notification.NotificationClient) SpendeeService {
	return &spendeeSvc{botRepo: *botRepo, gsheetRepo: *gsheetRepo, notificationClient: *notificationClient}
}
