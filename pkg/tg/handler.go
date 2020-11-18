package tg

import (
	"context"

	tgbotapi "github.com/bots-house/telegram-bot-api"
)

type Handler interface {
	HandleUpdate(ctx context.Context, update *tgbotapi.Update) error
}

type HandlerFunc func(ctx context.Context, update *tgbotapi.Update) error

func (hf HandlerFunc) HandleUpdate(ctx context.Context, update *tgbotapi.Update) error {
	return hf(ctx, update)
}

type Middleware func(next Handler) Handler
