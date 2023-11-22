package logger

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/frasnym/go-expense-telebot/common/ctxdata"
)

// TODO: Use Library
func Error(ctx context.Context, err error) {
	printToConsole(ctx, LogLevelError, err.Error())
}

func Warn(ctx context.Context, msg string) {
	printToConsole(ctx, LogLevelWarn, msg)
}

func Info(ctx context.Context, msg string) {
	printToConsole(ctx, LogLevelInfo, msg)
}

func Debug(ctx context.Context, msg string) {
	printToConsole(ctx, LogLevelDebug, msg)
}

func printToConsole(ctx context.Context, level LogLevel, msg string) {
	log := ConsoleLog{
		Level:         level,
		CorrelationID: ctxdata.GetCorrelationID(ctx),
		Message:       msg,
	}
	logByte, _ := json.Marshal(log)
	fmt.Println(string(logByte))
}

func LogService(ctx context.Context, serviceName string, err error) {
	if err != nil {
		Error(ctx, fmt.Errorf("%s: %w", serviceName, err))
		return
	}

	Info(ctx, fmt.Sprintf("%s success", serviceName))
}
