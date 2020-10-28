package tg

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"

func IsChatNotFoundError(err error) bool {
	err2, ok := err.(tgbotapi.Error)
	return ok && err2.Message == "Bad Request: chat not found"
}

func IsMemberListIsInaccessible(err error) bool {
	err2, ok := err.(tgbotapi.Error)
	return ok && err2.Message == "Bad Request: member list is inaccessible"
}
