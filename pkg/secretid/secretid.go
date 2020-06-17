package secretid

type SecretID interface {
	Encode(id int) string
	Decode(hash string) (int, error)
}
