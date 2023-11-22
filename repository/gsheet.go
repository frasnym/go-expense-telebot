package repository

import (
	"context"
	"fmt"

	"github.com/frasnym/go-expense-telebot/common/logger"
	"github.com/frasnym/go-expense-telebot/config"
	"google.golang.org/api/sheets/v4"
)

type GSheetRepository interface {
	AppendRow(ctx context.Context, sheetName string, input [][]any) error
	GetValues(ctx context.Context, valueRange string) (*sheets.ValueRange, error)
}

type gsheetRepo struct {
	cfg     *config.Config
	service *sheets.Service
}

// AppendRow implements GSheetRepository.
func (repo *gsheetRepo) AppendRow(ctx context.Context, sheetName string, input [][]any) error {
	var err error
	defer func() {
		logger.LogService(ctx, "GSheetAppendRow", err)
	}()

	values := &sheets.ValueRange{
		Values: input,
	}

	_, err = repo.service.Spreadsheets.Values.
		Append(repo.cfg.GsheetID, fmt.Sprintf("%s!A:E", sheetName), values).ValueInputOption("USER_ENTERED").Do()
	if err != nil {
		err = fmt.Errorf("err repo.service.Spreadsheets.Values.Append: %w", err)
		return err
	}

	return nil
}

func (repo *gsheetRepo) GetValues(ctx context.Context, valueRange string) (*sheets.ValueRange, error) {
	var err error
	defer func() {
		logger.LogService(ctx, "GSheetGetValues", err)
	}()

	// Make the API call to get values from the specified range.
	resp, err := repo.service.Spreadsheets.Values.Get(repo.cfg.GsheetID, valueRange).Do()
	if err != nil {
		err = fmt.Errorf("err repo.service.Spreadsheets.Values.Get: %w", err)
		return nil, err
	}

	return resp, nil
}

func NewGSheetRepository(cfg *config.Config, service *sheets.Service) GSheetRepository {
	return &gsheetRepo{cfg: cfg, service: service}
}
