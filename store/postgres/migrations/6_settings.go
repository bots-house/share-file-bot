package migrations

func init() {
	include(6, query(`
        alter table "user" add column settings jsonb not null default('{"long_ids": false, "updated_at": null}');
        alter table "document" alter column public_id type varchar(50);
    `), query(`
        alter tabel "user" drop column settings;
    `))
}
