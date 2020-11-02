package postgres

import (
	"github.com/friendsofgo/errors"
	"github.com/lib/pq"
)

func isFilePublicIDCollisionErr(err error) bool {
	err2, ok := errors.Cause(err).(*pq.Error)
	return ok && err2.Constraint == "file_public_id_key"
}

func isChatAlreadyConnectedError(err error) bool {
	err2, ok := errors.Cause(err).(*pq.Error)
	return ok && err2.Constraint == "chat_owner_id_telegram_id_key"
}
