package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractDeepLinksPayload(t *testing.T) {
	for _, test := range []struct {
		Username string
		URIes    []string
		Result   []string
		Error    error
	}{
		{
			Username: "cleepy_bot",
			URIes: []string{
				"https://t.me/cleepy_bot?start=ref_teleblog",
				"https://t.me/cleepy_bot?start=dVQK8",
				"https://t.me/cleepy_bot?start=buJ9U30UdIMl1On6c0eUrxQ3UPKkinE1xcSGQPLT2BEcsDlVN9",
				"https://t.me/cleepy_bot?start=ref_crosser-HndVA",
			},
			Result: []string{
				"dVQK8",
				"buJ9U30UdIMl1On6c0eUrxQ3UPKkinE1xcSGQPLT2BEcsDlVN9",
				"HndVA",
			},
		},
	} {
		result, err := ExtractDeepLinkPublicID(test.Username, test.URIes)
		assert.Equal(t, test.Result, result)
		assert.Equal(t, test.Error, err)
	}
}
