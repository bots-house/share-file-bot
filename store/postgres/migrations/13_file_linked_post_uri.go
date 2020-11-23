package migrations

func init() {
	include(13, query(`
		alter table "file" add column linked_post_uri text; 
    `), query(``))
}
