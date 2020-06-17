package secretid

import (
	"github.com/friendsofgo/errors"
	"github.com/speps/go-hashids"
)

type HashIDs struct {
	backend *hashids.HashID
}

var _ SecretID = &HashIDs{}

func NewHashIDs(salt string) (*HashIDs, error) {
	data := hashids.NewData()
	data.Salt = salt

	backend, err := hashids.NewWithData(data)
	if err != nil {
		return nil, errors.Wrap(err, "init backend")
	}

	return &HashIDs{
		backend: backend,
	}, nil
}

func (hids *HashIDs) Encode(id int) string {
	enc, err := hids.backend.Encode([]int{id})
	if err != nil {
		panic(err)
	}
	return enc
}

func (hids *HashIDs) Decode(hash string) (int, error) {
	ids, err := hids.backend.DecodeWithError(hash)
	if err != nil {
		return 0, errors.Wrap(err, "decode id with hash ids")
	}
	return ids[0], nil
}
