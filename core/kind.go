package core

import (
	"errors"
)

//go:generate stringer -trimprefix -type Kind

// Kind define enum type for kind of files.
type Kind int8

const (
	KindUnknown Kind = iota
	KindDocument
	KindAnimation
	KindAudio
	KindPhoto
	KindVideo
	KindVoice
)

var (
	ErrInvalidKind = errors.New("kind is invalid")
)

func ParseKind(v string) (Kind, error) {
	switch v {
	case "Document":
		return KindDocument, nil
	case "Animation":
		return KindAnimation, nil
	case "Audio":
		return KindAudio, nil
	case "Video":
		return KindVideo, nil
	case "Voice":
		return KindVoice, nil
	case "Photo":
		return KindPhoto, nil
	default:
		return KindUnknown, ErrInvalidKind
	}
}
