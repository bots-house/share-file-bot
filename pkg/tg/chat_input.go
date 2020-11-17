package tg

import "regexp"

type ChatInputType int

const (
	ChatInputUnkown ChatInputType = iota
	ChatInputJoinLink
	ChatInputUsername
)

var (
	// https://regex101.com/r/k17Knt/1/
	reChatInputUsername = regexp.MustCompile(`(?:^|@|/)([a-zA-Z0-9_]{5,32})$`)

	// https://regex101.com/r/WsJx0O/1/
	reChatInputJoinLink = regexp.MustCompile(`\/joinchat\/([\da-zA-Z_-]+)$`)
)

// ParseChatInput parse user input
func ParseChatInput(query string) (qt ChatInputType, value string) {
	if matches := reChatInputJoinLink.FindStringSubmatch(query); len(matches) > 0 {
		return ChatInputJoinLink, matches[1]
	} else if matches := reChatInputUsername.FindStringSubmatch(query); len(matches) > 0 {
		return ChatInputUsername, matches[1]
	} else {
		return ChatInputUnkown, ""
	}
}
