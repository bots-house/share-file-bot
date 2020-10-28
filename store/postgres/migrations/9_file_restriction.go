package migrations

func init() {
	include(9, query(`
		alter table file
			add column 
				restrictions_chat_id integer references chat(id) on delete set null;

    `), query(``))
}
