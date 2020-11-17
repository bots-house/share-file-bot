package tg

import (
	"errors"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func IsChatNotFoundError(err error) bool {
	err2, ok := err.(tgbotapi.Error)
	return ok && err2.Message == "Bad Request: chat not found"
}

func IsMemberListIsInaccessible(err error) bool {
	err2, ok := err.(tgbotapi.Error)
	return ok && err2.Message == "Bad Request: member list is inaccessible"
}

func IsBotIsNotMember(err error) bool {
	return IsBotIsNotMemberOfChannel(err) || IsBotIsNotMemberOfSupergroup(err)
}

func IsBotIsNotMemberOfChannel(err error) bool {
	err2, ok := err.(tgbotapi.Error)
	return ok && err2.Message == "Forbidden: bot is not a member of the channel chat"
}

func IsNotEnoughRightsToExportChatInviteLink(err error) bool {
	return isTelegramErr(err, "Bad Request: not enough rights to export chat invite link")
}

func isTelegramErr(err error, msg string) bool {
	var tgErr tgbotapi.Error
	if errors.As(err, &tgErr) {
		return tgErr.Message == msg
	}
	return false
}

func IsBotIsNotMemberOfSupergroup(err error) bool {
	err2, ok := err.(tgbotapi.Error)
	return ok && err2.Message == "Forbidden: bot is not a member of the supergroup chat"
}

func IsCantCheckChatMember(err error) bool {
	return IsChatNotFoundError(err) ||
		IsMemberListIsInaccessible(err) ||
		IsBotIsNotMember(err) ||
		IsNotEnoughRightsToExportChatInviteLink(err)
}
