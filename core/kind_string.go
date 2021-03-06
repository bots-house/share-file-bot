// Code generated by "stringer -type Kind -trimprefix Kind"; DO NOT EDIT.

package core

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[KindUnknown-0]
	_ = x[KindDocument-1]
	_ = x[KindAnimation-2]
	_ = x[KindAudio-3]
	_ = x[KindPhoto-4]
	_ = x[KindVideo-5]
	_ = x[KindVoice-6]
}

const _Kind_name = "UnknownDocumentAnimationAudioPhotoVideoVoice"

var _Kind_index = [...]uint8{0, 7, 15, 24, 29, 34, 39, 44}

func (i Kind) String() string {
	if i < 0 || i >= Kind(len(_Kind_index)-1) {
		return "Kind(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _Kind_name[_Kind_index[i]:_Kind_index[i+1]]
}
