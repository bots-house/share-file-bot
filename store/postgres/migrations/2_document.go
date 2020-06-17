package migrations

func init() {
	include(2, query(`
        create table "document" (
            id serial primary key,
            file_id text not null,
            caption text,
            mime_type text,
            size integer not null,
            name text not null,
            owner_id integer not null references "user"(id) on delete cascade,
            created_at timestamp with time zone not null
        );
    `), query(`
        drop table "document";
    `))
}
