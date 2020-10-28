package migrations

func init() {
	include(10, query(`
		alter table download
			add column 
				new_subscription boolean;

    `), query(``))
}
