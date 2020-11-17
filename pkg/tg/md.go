package tg

import "strings"

var escapeMDReplacer = strings.NewReplacer(
	"_", `\_`,
	"*", `\*`,
	"[", `\[`,
	"]", `\]`,
	"(", `\(`,
	")", `\)`,
	"~", `\~`,
	"`", "\\`",
	">", `\>`,
	"#", `\#`,
	"+", `\+`,
	"-", `\-`,
	"=", `\=`,
	"|", `\|`,
	"{", `\{`,
	"}", `\}`,
	".", `\.`,
	"!", `\!`,
)

// EscapeMD escapes all markdown reserved chars.
func EscapeMD(txt string) string {
	return escapeMDReplacer.Replace(txt)
}
