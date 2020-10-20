package migrations

func init() {
	include(8, query(`
		create type chat_type as enum (
			'Group',
			'SuperGroup',
			'Channel'
		);

		create table chat (
			id serial primary key not null, 
			telegram_id bigint not null, 
			title varchar(255) not null, 
			type chat_type not null, 
			owner_id integer not null references "user"(id) on delete cascade,
			linked_at timestamptz not null, 
			updated_at timestamptz
		);
    `), query(`
        drop table chat;
    `))
}
