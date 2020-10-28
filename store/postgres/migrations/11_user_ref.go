package migrations

func init() {
	include(11, query(`
		alter table "user"
			add column 
				ref varchar(60) default 'unknown';
		alter table "user"
			alter column ref drop default;
    `), query(``))
}
