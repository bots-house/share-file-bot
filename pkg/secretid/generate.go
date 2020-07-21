package secretid

import gonanoid "github.com/matoous/go-nanoid"

const (
	alphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ_abcdefghijklmnopqrstuvwxyz"

	defaultLength = 5
	longLength    = 50
)

func IsLong(id string) bool {
	return len(id) > defaultLength
}

func Generate(long bool) string {
	var length int

	if long {
		length = longLength
	} else {
		length = defaultLength
	}

	id, err := gonanoid.Generate(alphabet, length)
	if err != nil {
		panic("generate secret id")
	}

	return id
}
