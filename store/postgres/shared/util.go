package shared

import (
	"errors"
	"time"

	"github.com/volatiletech/null/v8"
)

var ErrTooManyAffectedRows = errors.New("too many affected rows")

func ToNullString(v string) null.String {
	if v != "" {
		return null.StringFrom(v)
	} else {
		return null.String{}
	}
}

func ToNullTime(v time.Time) null.Time {
	if !v.IsZero() {
		return null.TimeFrom(v)
	} else {
		return null.Time{}
	}
}

func ToNullJSONB(v []byte) null.JSON {
	if len(v) != 0 {
		return null.JSONFrom(v)
	} else {
		return null.JSON{}
	}
}

func ToNullInt(v int) null.Int {
	if v != 0 {
		return null.IntFrom(v)
	} else {
		return null.Int{}
	}
}
