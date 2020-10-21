package tg

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"regexp"

	"github.com/friendsofgo/errors"
)

var (
	// https://regex101.com/r/WsJx0O/1/
	regexpJoinLink = regexp.MustCompile(`\/joinchat\/([\da-zA-Z_-]+)$`)

	regexpJoinLinkHashHex = regexp.MustCompile(`[a-fA-F\d]{32}$`)
)

// ParseJoinLink returns hash from join chat link or empty string if not found.
func ParseJoinLink(v string) string {
	if matches := regexpJoinLink.FindStringSubmatch(v); len(matches) > 0 {
		return matches[1]
	}

	return ""
}

type JoinLinkPayload struct {
	CreatorID int32
	ChatID    int32
	RandomID  int64
}

func (payload *JoinLinkPayload) BotChatID() int64 {
	return MTProtoToBotID(payload.ChatID)
}

func DecodeJoinLinkPayload(encodedPayload string) (*JoinLinkPayload, error) {
	var buf *bytes.Buffer

	if isHexJoinLinkPayload(encodedPayload) {
		decoded, err := hex.DecodeString(encodedPayload)
		if err != nil {
			return nil, errors.Wrap(err, "decode join link as hex")
		}

		buf = bytes.NewBuffer(decoded)
	} else {
		tmp := bytes.NewBufferString(encodedPayload)

		for i := 0; i < len(encodedPayload)%4; i++ {
			tmp.WriteRune('=')
		}

		data, err := base64.URLEncoding.DecodeString(tmp.String())
		if err != nil {
			return nil, errors.Wrap(err, "decode join link as base64")
		}

		buf = bytes.NewBuffer(data)
	}

	pl := &JoinLinkPayload{}

	if err := binary.Read(buf, binary.BigEndian, pl); err != nil {
		return nil, errors.Wrap(err, "decode join link payload")
	}

	return pl, nil
}

func isHexJoinLinkPayload(payload string) bool {
	return regexpJoinLinkHashHex.MatchString(payload)
}
