package postgres

import (
	"github.com/friendsofgo/errors"
	"github.com/lib/pq"
)

func isConstraintError(err error, constraint string) bool {
	var pgErr *pq.Error

	if errors.As(err, &pgErr) {
		return pgErr.Constraint == constraint
	}

	return false
}

func isFilePublicIDCollisionErr(err error) bool {
	return isConstraintError(err, "file_public_id_key")
}

func isChatAlreadyConnectedError(err error) bool {
	return isConstraintError(err, "chat_owner_id_telegram_id_key")
}
