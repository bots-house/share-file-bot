package migrations

func init() {
	include(12, query(`
		alter table "file" add column is_violates_copyright boolean; 
    `), query(``))
}
