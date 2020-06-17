package migrations

func init() {
	include(1, query(`
        create table "user" (
            id integer primary key,
            first_name text not null,
            last_name text,
            username text,
            language_code text not null,
            is_admin bool not null,
            joined_at timestamp with time zone not null,
            updated_at timestamp with time zone
        );
    `), query(`
        drop table "user";
    `))
}
