package migrations

func init() {
	include(14, query(`
		create table bot (
			id integer primary key, 
			username varchar(32) not null,
			token text not null, 
			owner_id integer not null references "user"(id), 
			linked_at timestamptz not null
		)
    `), query(``))
}
