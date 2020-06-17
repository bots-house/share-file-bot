package migrations

func init() {
	include(3, query(`
        create table "download" (
            id serial primary key,
            document_id integer not null references document(id) on delete cascade,
            user_id integer not null references "user"(id) on delete set null,
            at timestamp with time zone not null
        );

        create index idx_download_document_id on download(document_id);
    `), query(`
        drop table "download";
    `))
}
