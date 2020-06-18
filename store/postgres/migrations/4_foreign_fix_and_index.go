package migrations

func init() {
	include(4, query(`
        create index idx_document_owner_id on document(owner_id);

        alter table download
            drop constraint download_document_id_fkey,
            add constraint download_document_id_fkey
                foreign key (document_id)
                references document(id) on delete set null,
            alter column user_id drop not null,
            alter column document_id drop not null;
    `), query(`

    `))
}
