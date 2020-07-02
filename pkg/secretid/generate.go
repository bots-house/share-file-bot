package secretid

import gonanoid "github.com/matoous/go-nanoid"

const (
	alphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ_abcdefghijklmnopqrstuvwxyz"
	lenght   = 5
)

func Generate() string {
	id, err := gonanoid.Generate(alphabet, lenght)
	if err != nil {
		panic("generate secret id")
	}
	return id
}
