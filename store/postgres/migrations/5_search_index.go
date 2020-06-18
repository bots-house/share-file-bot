package migrations

func init() {
	include(4, query(`
    create index document_search_idx document USING GIN(
        setweight(to_tsvector(name), 'A') ||
        setweight(to_tsvector(caption), 'B')) ||
        setweight(to_tsvector(caption), 'C'))
    );

    `), query(`

    `))
}
