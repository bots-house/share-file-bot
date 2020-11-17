package postgres

import (
	"github.com/friendsofgo/errors"
	"github.com/jackc/pgconn"
)

func isConstraintErr(err error, constraint string) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.ConstraintName == constraint
	}

	return false
}

func isFilePublicIDCollisionErr(err error) bool {
	return isConstraintErr(err, "file_public_id_key")
}

func isChatAlreadyConnectedError(err error) bool {
	return isConstraintErr(err, "chat_owner_id_telegram_id_key")
}
