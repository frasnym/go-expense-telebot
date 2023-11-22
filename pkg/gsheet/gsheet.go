package gsheet

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"

	"github.com/frasnym/go-expense-telebot/config"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

var sheetService *sheets.Service

func init() {
	cfg := config.GetConfig()

	sheetKey := Key{
		Type:           "service_account",
		ProjectID:      cfg.GsheetProjectID,
		PrivateKeyID:   cfg.GsheetUserPrivateKeyID,
		PrivateKey:     cfg.GsheetUserPrivateKey,
		ClientEmail:    cfg.GsheetUserClientEmail,
		ClientID:       cfg.GsheetUserClientID,
		AuthURI:        "https://accounts.google.com/o/oauth2/auth",
		TokenURI:       "https://oauth2.googleapis.com/token",
		AuthProvider:   "https://www.googleapis.com/oauth2/v1/certs",
		Client:         fmt.Sprintf("https://www.googleapis.com/robot/v1/metadata/x509/%s", url.QueryEscape(cfg.GsheetUserClientEmail)),
		UniverseDomain: "googleapis.com",
	}

	credential, err := json.Marshal(sheetKey)
	if err != nil {
		panic(fmt.Errorf("failed to get key: %v", err))
	}

	srv, err := sheets.NewService(context.Background(), option.WithCredentialsJSON(credential))
	if err != nil {
		panic(fmt.Errorf("unable to retrieve Sheets client: %v", err))
	}

	sheetService = srv
}

func GetService() *sheets.Service {
	if sheetService == nil {
		panic(errors.New("please init gsheet service first"))
	}

	return sheetService
}
