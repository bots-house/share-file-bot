package migrations

func init() {
	include(7, query(`
		--- rename document -> file
		alter table document rename to file; 
		alter sequence document_id_seq rename to file_id_seq;
		alter index document_pkey rename to file_pkey;
		alter index document_public_id_key rename to file_public_id_key;
		alter table file rename constraint document_owner_id_fkey to file_owner_id_fkey;
		alter index idx_document_owner_id rename to document_owner_id_idx;
		alter index document_owner_id_idx rename to file_owner_id_idx;

		--- rename in related download
		alter table download rename column document_id to file_id; 
		alter index idx_download_document_id rename to download_file_id_idx;
		alter table download rename constraint download_document_id_fkey to download_file_id_fkey;
		-- rename in related user
    `), query(`
        alter tabel "user" drop column settings;
    `))
}
