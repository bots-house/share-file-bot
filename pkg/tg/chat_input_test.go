package tg

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseChatInput(t *testing.T) {
	for _, test := range []struct {
		Input    string
		Excepted ChatInputType
		Value    string
	}{
		{"https://t.me/joinchat/AAAAAES_pid_l6flZONwGQ", ChatInputJoinLink, "AAAAAES_pid_l6flZONwGQ"},
		{"zzap.run/joinchat/AAAAAES_pid_l6flZONwGQ", ChatInputJoinLink, "AAAAAES_pid_l6flZONwGQ"},
		{"https://t.me/channely", ChatInputUsername, "channely"},
		{"https://zzap.run/channely", ChatInputUsername, "channely"},
		{"t.me/channely_bot", ChatInputUsername, "channely_bot"},
		{"channely", ChatInputUsername, "channely"},
		{"@channely", ChatInputUsername, "channely"},
		{},
	} {
		typ, val := ParseChatInput(test.Input)

		assert.Equal(t, test.Excepted, typ)
		assert.Equal(t, test.Value, val)
	}

}
