package log

import (
	"github.com/go-kit/kit/log/level"
	"github.com/go-kit/kit/log/term"
)

func loggerColorFn(keyvals ...interface{}) term.FgBgColor {
	for i := 0; i < len(keyvals)-1; i += 2 {
		if keyvals[i] != "level" {
			continue
		}
		switch keyvals[i+1] {
		case level.InfoValue():
			return term.FgBgColor{Fg: term.Blue}
		case level.WarnValue():
			return term.FgBgColor{Fg: term.Yellow}
		case level.ErrorValue():
			return term.FgBgColor{Fg: term.Red}
		default:
			return term.FgBgColor{}
		}
	}
	return term.FgBgColor{}
}
