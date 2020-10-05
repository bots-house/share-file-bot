package postgres

import (
	"github.com/lib/pq"
	"github.com/pkg/errors"
)

func isFilePublicIDCollisionErr(err error) bool {
	err2, ok := errors.Cause(err).(*pq.Error)
	return ok && err2.Constraint == "file_public_id_key"
}
