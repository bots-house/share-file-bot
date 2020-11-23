package bot

import (
	"testing"

	tgbotapi "github.com/bots-house/telegram-bot-api"
	"github.com/stretchr/testify/assert"
)

func TestExtractRefFromMsg(t *testing.T) {
	for _, test := range []struct {
		Name         string
		Text         string
		ExceptedText string
		Ref          string
	}{
		{
			Name:         "StartWithFile",
			Text:         "/start LlOiBaweeonab_xFZ3xnVD1XxLZCjHPhlgeMCMMWyBSpocXcaf",
			ExceptedText: "/start LlOiBaweeonab_xFZ3xnVD1XxLZCjHPhlgeMCMMWyBSpocXcaf",
			Ref:          "",
		},
		{
			Name:         "StartWithRef",
			Text:         "/start ref_tgstat_1",
			ExceptedText: "/start",
			Ref:          "tgstat_1",
		},
		{
			Name:         "StartWithFileAndRef",
			Text:         "/start ref_tgstat_1-LlOiBaweeonab_xFZ3xnVD1XxLZCjHPhlgeMCMMWyBSpocXcaf",
			ExceptedText: "/start LlOiBaweeonab_xFZ3xnVD1XxLZCjHPhlgeMCMMWyBSpocXcaf",
			Ref:          "tgstat_1",
		},
		{
			Name:         "JustStart",
			Text:         "/start",
			ExceptedText: "/start",
			Ref:          "",
		},
	} {
		test := test

		t.Run(test.Name, func(t *testing.T) {
			msg := &tgbotapi.Message{Text: test.Text, Entities: &[]tgbotapi.MessageEntity{
				{
					Type:   "bot_command",
					Offset: 0,
					Length: 6,
				},
			}}

			ref := extractRefFromMsg(msg)

			assert.Equal(t, test.Ref, ref)
			assert.Equal(t, test.ExceptedText, msg.Text)
		})
	}
}
