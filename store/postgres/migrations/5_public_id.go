package migrations

import (
	"database/sql"
	"os"

	"github.com/pkg/errors"
	"github.com/speps/go-hashids"
)

func init() {
	include(5, func(tx *sql.Tx) error {
		salt := os.Getenv("SFB_SECRET_ID_SALT")

		data := hashids.NewData()
		data.Salt = salt

		encoder, err := hashids.NewWithData(data)
		if err != nil {
			return errors.Wrap(err, "init backend")
		}

		if _, err := tx.Exec(`
            alter table
                document
            add column
                public_id varchar(8) unique;
        `); err != nil {
			return errors.Wrap(err, "add column public id")
		}

		rows, err := tx.Query(`
            select
                id
            from
                document
        `)
		if err != nil {
			return errors.Wrap(err, "query rows")
		}
		defer rows.Close()

		var ids []int

		for rows.Next() {
			var id int

			if err := rows.Scan(&id); err != nil {
				return errors.Wrap(err, "scan")
			}

			ids = append(ids, id)
		}

		if err := rows.Err(); err != nil {
			return errors.Wrap(err, "rows check")
		}

		kv := make(map[int]string, len(ids))

		for _, id := range ids {
			v, err := encoder.Encode([]int{id})
			if err != nil {
				return errors.Wrap(err, "encode")
			}
			kv[id] = v
		}

		for k, v := range kv {
			_, err := tx.Exec(`update document set public_id = $2 where id = $1`, k, v)
			if err != nil {
				return errors.Wrap(err, "update")
			}
		}

		if _, err := tx.Exec(`
            alter table document
            alter column public_id set not null;
        `); err != nil {
			return errors.Wrap(err, "add column public id")
		}

		return nil
	}, query(`

    `))
}
