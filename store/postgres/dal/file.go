// Code generated by SQLBoiler 4.2.0 (https://github.com/volatiletech/sqlboiler). DO NOT EDIT.
// This file is meant to be re-generated in place and/or deleted at any time.

package dal

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/friendsofgo/errors"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"github.com/volatiletech/sqlboiler/v4/queries/qmhelper"
	"github.com/volatiletech/sqlboiler/v4/types"
	"github.com/volatiletech/strmangle"
)

// File is an object representing the database table.
type File struct {
	ID        int         `boil:"id" json:"id" toml:"id" yaml:"id"`
	FileID    string      `boil:"file_id" json:"file_id" toml:"file_id" yaml:"file_id"`
	Caption   null.String `boil:"caption" json:"caption,omitempty" toml:"caption" yaml:"caption,omitempty"`
	MimeType  null.String `boil:"mime_type" json:"mime_type,omitempty" toml:"mime_type" yaml:"mime_type,omitempty"`
	Size      int         `boil:"size" json:"size" toml:"size" yaml:"size"`
	Name      string      `boil:"name" json:"name" toml:"name" yaml:"name"`
	OwnerID   int         `boil:"owner_id" json:"owner_id" toml:"owner_id" yaml:"owner_id"`
	CreatedAt time.Time   `boil:"created_at" json:"created_at" toml:"created_at" yaml:"created_at"`
	PublicID  string      `boil:"public_id" json:"public_id" toml:"public_id" yaml:"public_id"`
	Kind      string      `boil:"kind" json:"kind" toml:"kind" yaml:"kind"`
	Metadata  types.JSON  `boil:"metadata" json:"metadata" toml:"metadata" yaml:"metadata"`

	R *fileR `boil:"-" json:"-" toml:"-" yaml:"-"`
	L fileL  `boil:"-" json:"-" toml:"-" yaml:"-"`
}

var FileColumns = struct {
	ID        string
	FileID    string
	Caption   string
	MimeType  string
	Size      string
	Name      string
	OwnerID   string
	CreatedAt string
	PublicID  string
	Kind      string
	Metadata  string
}{
	ID:        "id",
	FileID:    "file_id",
	Caption:   "caption",
	MimeType:  "mime_type",
	Size:      "size",
	Name:      "name",
	OwnerID:   "owner_id",
	CreatedAt: "created_at",
	PublicID:  "public_id",
	Kind:      "kind",
	Metadata:  "metadata",
}

// Generated where

type whereHelpernull_String struct{ field string }

func (w whereHelpernull_String) EQ(x null.String) qm.QueryMod {
	return qmhelper.WhereNullEQ(w.field, false, x)
}
func (w whereHelpernull_String) NEQ(x null.String) qm.QueryMod {
	return qmhelper.WhereNullEQ(w.field, true, x)
}
func (w whereHelpernull_String) IsNull() qm.QueryMod    { return qmhelper.WhereIsNull(w.field) }
func (w whereHelpernull_String) IsNotNull() qm.QueryMod { return qmhelper.WhereIsNotNull(w.field) }
func (w whereHelpernull_String) LT(x null.String) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.LT, x)
}
func (w whereHelpernull_String) LTE(x null.String) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.LTE, x)
}
func (w whereHelpernull_String) GT(x null.String) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.GT, x)
}
func (w whereHelpernull_String) GTE(x null.String) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.GTE, x)
}

type whereHelpertypes_JSON struct{ field string }

func (w whereHelpertypes_JSON) EQ(x types.JSON) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.EQ, x)
}
func (w whereHelpertypes_JSON) NEQ(x types.JSON) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.NEQ, x)
}
func (w whereHelpertypes_JSON) LT(x types.JSON) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.LT, x)
}
func (w whereHelpertypes_JSON) LTE(x types.JSON) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.LTE, x)
}
func (w whereHelpertypes_JSON) GT(x types.JSON) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.GT, x)
}
func (w whereHelpertypes_JSON) GTE(x types.JSON) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.GTE, x)
}

var FileWhere = struct {
	ID        whereHelperint
	FileID    whereHelperstring
	Caption   whereHelpernull_String
	MimeType  whereHelpernull_String
	Size      whereHelperint
	Name      whereHelperstring
	OwnerID   whereHelperint
	CreatedAt whereHelpertime_Time
	PublicID  whereHelperstring
	Kind      whereHelperstring
	Metadata  whereHelpertypes_JSON
}{
	ID:        whereHelperint{field: "\"file\".\"id\""},
	FileID:    whereHelperstring{field: "\"file\".\"file_id\""},
	Caption:   whereHelpernull_String{field: "\"file\".\"caption\""},
	MimeType:  whereHelpernull_String{field: "\"file\".\"mime_type\""},
	Size:      whereHelperint{field: "\"file\".\"size\""},
	Name:      whereHelperstring{field: "\"file\".\"name\""},
	OwnerID:   whereHelperint{field: "\"file\".\"owner_id\""},
	CreatedAt: whereHelpertime_Time{field: "\"file\".\"created_at\""},
	PublicID:  whereHelperstring{field: "\"file\".\"public_id\""},
	Kind:      whereHelperstring{field: "\"file\".\"kind\""},
	Metadata:  whereHelpertypes_JSON{field: "\"file\".\"metadata\""},
}

// FileRels is where relationship names are stored.
var FileRels = struct {
	Owner     string
	Downloads string
}{
	Owner:     "Owner",
	Downloads: "Downloads",
}

// fileR is where relationships are stored.
type fileR struct {
	Owner     *User         `boil:"Owner" json:"Owner" toml:"Owner" yaml:"Owner"`
	Downloads DownloadSlice `boil:"Downloads" json:"Downloads" toml:"Downloads" yaml:"Downloads"`
}

// NewStruct creates a new relationship struct
func (*fileR) NewStruct() *fileR {
	return &fileR{}
}

// fileL is where Load methods for each relationship are stored.
type fileL struct{}

var (
	fileAllColumns            = []string{"id", "file_id", "caption", "mime_type", "size", "name", "owner_id", "created_at", "public_id", "kind", "metadata"}
	fileColumnsWithoutDefault = []string{"file_id", "caption", "mime_type", "size", "name", "owner_id", "created_at", "public_id", "kind"}
	fileColumnsWithDefault    = []string{"id", "metadata"}
	filePrimaryKeyColumns     = []string{"id"}
)

type (
	// FileSlice is an alias for a slice of pointers to File.
	// This should generally be used opposed to []File.
	FileSlice []*File

	fileQuery struct {
		*queries.Query
	}
)

// Cache for insert, update and upsert
var (
	fileType                 = reflect.TypeOf(&File{})
	fileMapping              = queries.MakeStructMapping(fileType)
	filePrimaryKeyMapping, _ = queries.BindMapping(fileType, fileMapping, filePrimaryKeyColumns)
	fileInsertCacheMut       sync.RWMutex
	fileInsertCache          = make(map[string]insertCache)
	fileUpdateCacheMut       sync.RWMutex
	fileUpdateCache          = make(map[string]updateCache)
	fileUpsertCacheMut       sync.RWMutex
	fileUpsertCache          = make(map[string]insertCache)
)

var (
	// Force time package dependency for automated UpdatedAt/CreatedAt.
	_ = time.Second
	// Force qmhelper dependency for where clause generation (which doesn't
	// always happen)
	_ = qmhelper.Where
)

// One returns a single file record from the query.
func (q fileQuery) One(ctx context.Context, exec boil.ContextExecutor) (*File, error) {
	o := &File{}

	queries.SetLimit(q.Query, 1)

	err := q.Bind(ctx, exec, o)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "dal: failed to execute a one query for file")
	}

	return o, nil
}

// All returns all File records from the query.
func (q fileQuery) All(ctx context.Context, exec boil.ContextExecutor) (FileSlice, error) {
	var o []*File

	err := q.Bind(ctx, exec, &o)
	if err != nil {
		return nil, errors.Wrap(err, "dal: failed to assign all query results to File slice")
	}

	return o, nil
}

// Count returns the count of all File records in the query.
func (q fileQuery) Count(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)

	err := q.Query.QueryRowContext(ctx, exec).Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "dal: failed to count file rows")
	}

	return count, nil
}

// Exists checks if the row exists in the table.
func (q fileQuery) Exists(ctx context.Context, exec boil.ContextExecutor) (bool, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)
	queries.SetLimit(q.Query, 1)

	err := q.Query.QueryRowContext(ctx, exec).Scan(&count)
	if err != nil {
		return false, errors.Wrap(err, "dal: failed to check if file exists")
	}

	return count > 0, nil
}

// Owner pointed to by the foreign key.
func (o *File) Owner(mods ...qm.QueryMod) userQuery {
	queryMods := []qm.QueryMod{
		qm.Where("\"id\" = ?", o.OwnerID),
	}

	queryMods = append(queryMods, mods...)

	query := Users(queryMods...)
	queries.SetFrom(query.Query, "\"user\"")

	return query
}

// Downloads retrieves all the download's Downloads with an executor.
func (o *File) Downloads(mods ...qm.QueryMod) downloadQuery {
	var queryMods []qm.QueryMod
	if len(mods) != 0 {
		queryMods = append(queryMods, mods...)
	}

	queryMods = append(queryMods,
		qm.Where("\"download\".\"file_id\"=?", o.ID),
	)

	query := Downloads(queryMods...)
	queries.SetFrom(query.Query, "\"download\"")

	if len(queries.GetSelect(query.Query)) == 0 {
		queries.SetSelect(query.Query, []string{"\"download\".*"})
	}

	return query
}

// LoadOwner allows an eager lookup of values, cached into the
// loaded structs of the objects. This is for an N-1 relationship.
func (fileL) LoadOwner(ctx context.Context, e boil.ContextExecutor, singular bool, maybeFile interface{}, mods queries.Applicator) error {
	var slice []*File
	var object *File

	if singular {
		object = maybeFile.(*File)
	} else {
		slice = *maybeFile.(*[]*File)
	}

	args := make([]interface{}, 0, 1)
	if singular {
		if object.R == nil {
			object.R = &fileR{}
		}
		args = append(args, object.OwnerID)

	} else {
	Outer:
		for _, obj := range slice {
			if obj.R == nil {
				obj.R = &fileR{}
			}

			for _, a := range args {
				if a == obj.OwnerID {
					continue Outer
				}
			}

			args = append(args, obj.OwnerID)

		}
	}

	if len(args) == 0 {
		return nil
	}

	query := NewQuery(
		qm.From(`user`),
		qm.WhereIn(`user.id in ?`, args...),
	)
	if mods != nil {
		mods.Apply(query)
	}

	results, err := query.QueryContext(ctx, e)
	if err != nil {
		return errors.Wrap(err, "failed to eager load User")
	}

	var resultSlice []*User
	if err = queries.Bind(results, &resultSlice); err != nil {
		return errors.Wrap(err, "failed to bind eager loaded slice User")
	}

	if err = results.Close(); err != nil {
		return errors.Wrap(err, "failed to close results of eager load for user")
	}
	if err = results.Err(); err != nil {
		return errors.Wrap(err, "error occurred during iteration of eager loaded relations for user")
	}

	if len(resultSlice) == 0 {
		return nil
	}

	if singular {
		foreign := resultSlice[0]
		object.R.Owner = foreign
		if foreign.R == nil {
			foreign.R = &userR{}
		}
		foreign.R.OwnerFiles = append(foreign.R.OwnerFiles, object)
		return nil
	}

	for _, local := range slice {
		for _, foreign := range resultSlice {
			if local.OwnerID == foreign.ID {
				local.R.Owner = foreign
				if foreign.R == nil {
					foreign.R = &userR{}
				}
				foreign.R.OwnerFiles = append(foreign.R.OwnerFiles, local)
				break
			}
		}
	}

	return nil
}

// LoadDownloads allows an eager lookup of values, cached into the
// loaded structs of the objects. This is for a 1-M or N-M relationship.
func (fileL) LoadDownloads(ctx context.Context, e boil.ContextExecutor, singular bool, maybeFile interface{}, mods queries.Applicator) error {
	var slice []*File
	var object *File

	if singular {
		object = maybeFile.(*File)
	} else {
		slice = *maybeFile.(*[]*File)
	}

	args := make([]interface{}, 0, 1)
	if singular {
		if object.R == nil {
			object.R = &fileR{}
		}
		args = append(args, object.ID)
	} else {
	Outer:
		for _, obj := range slice {
			if obj.R == nil {
				obj.R = &fileR{}
			}

			for _, a := range args {
				if queries.Equal(a, obj.ID) {
					continue Outer
				}
			}

			args = append(args, obj.ID)
		}
	}

	if len(args) == 0 {
		return nil
	}

	query := NewQuery(
		qm.From(`download`),
		qm.WhereIn(`download.file_id in ?`, args...),
	)
	if mods != nil {
		mods.Apply(query)
	}

	results, err := query.QueryContext(ctx, e)
	if err != nil {
		return errors.Wrap(err, "failed to eager load download")
	}

	var resultSlice []*Download
	if err = queries.Bind(results, &resultSlice); err != nil {
		return errors.Wrap(err, "failed to bind eager loaded slice download")
	}

	if err = results.Close(); err != nil {
		return errors.Wrap(err, "failed to close results in eager load on download")
	}
	if err = results.Err(); err != nil {
		return errors.Wrap(err, "error occurred during iteration of eager loaded relations for download")
	}

	if singular {
		object.R.Downloads = resultSlice
		for _, foreign := range resultSlice {
			if foreign.R == nil {
				foreign.R = &downloadR{}
			}
			foreign.R.File = object
		}
		return nil
	}

	for _, foreign := range resultSlice {
		for _, local := range slice {
			if queries.Equal(local.ID, foreign.FileID) {
				local.R.Downloads = append(local.R.Downloads, foreign)
				if foreign.R == nil {
					foreign.R = &downloadR{}
				}
				foreign.R.File = local
				break
			}
		}
	}

	return nil
}

// SetOwner of the file to the related item.
// Sets o.R.Owner to related.
// Adds o to related.R.OwnerFiles.
func (o *File) SetOwner(ctx context.Context, exec boil.ContextExecutor, insert bool, related *User) error {
	var err error
	if insert {
		if err = related.Insert(ctx, exec, boil.Infer()); err != nil {
			return errors.Wrap(err, "failed to insert into foreign table")
		}
	}

	updateQuery := fmt.Sprintf(
		"UPDATE \"file\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, []string{"owner_id"}),
		strmangle.WhereClause("\"", "\"", 2, filePrimaryKeyColumns),
	)
	values := []interface{}{related.ID, o.ID}

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, updateQuery)
		fmt.Fprintln(writer, values)
	}
	if _, err = exec.ExecContext(ctx, updateQuery, values...); err != nil {
		return errors.Wrap(err, "failed to update local table")
	}

	o.OwnerID = related.ID
	if o.R == nil {
		o.R = &fileR{
			Owner: related,
		}
	} else {
		o.R.Owner = related
	}

	if related.R == nil {
		related.R = &userR{
			OwnerFiles: FileSlice{o},
		}
	} else {
		related.R.OwnerFiles = append(related.R.OwnerFiles, o)
	}

	return nil
}

// AddDownloads adds the given related objects to the existing relationships
// of the file, optionally inserting them as new records.
// Appends related to o.R.Downloads.
// Sets related.R.File appropriately.
func (o *File) AddDownloads(ctx context.Context, exec boil.ContextExecutor, insert bool, related ...*Download) error {
	var err error
	for _, rel := range related {
		if insert {
			queries.Assign(&rel.FileID, o.ID)
			if err = rel.Insert(ctx, exec, boil.Infer()); err != nil {
				return errors.Wrap(err, "failed to insert into foreign table")
			}
		} else {
			updateQuery := fmt.Sprintf(
				"UPDATE \"download\" SET %s WHERE %s",
				strmangle.SetParamNames("\"", "\"", 1, []string{"file_id"}),
				strmangle.WhereClause("\"", "\"", 2, downloadPrimaryKeyColumns),
			)
			values := []interface{}{o.ID, rel.ID}

			if boil.IsDebug(ctx) {
				writer := boil.DebugWriterFrom(ctx)
				fmt.Fprintln(writer, updateQuery)
				fmt.Fprintln(writer, values)
			}
			if _, err = exec.ExecContext(ctx, updateQuery, values...); err != nil {
				return errors.Wrap(err, "failed to update foreign table")
			}

			queries.Assign(&rel.FileID, o.ID)
		}
	}

	if o.R == nil {
		o.R = &fileR{
			Downloads: related,
		}
	} else {
		o.R.Downloads = append(o.R.Downloads, related...)
	}

	for _, rel := range related {
		if rel.R == nil {
			rel.R = &downloadR{
				File: o,
			}
		} else {
			rel.R.File = o
		}
	}
	return nil
}

// SetDownloads removes all previously related items of the
// file replacing them completely with the passed
// in related items, optionally inserting them as new records.
// Sets o.R.File's Downloads accordingly.
// Replaces o.R.Downloads with related.
// Sets related.R.File's Downloads accordingly.
func (o *File) SetDownloads(ctx context.Context, exec boil.ContextExecutor, insert bool, related ...*Download) error {
	query := "update \"download\" set \"file_id\" = null where \"file_id\" = $1"
	values := []interface{}{o.ID}
	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, query)
		fmt.Fprintln(writer, values)
	}
	_, err := exec.ExecContext(ctx, query, values...)
	if err != nil {
		return errors.Wrap(err, "failed to remove relationships before set")
	}

	if o.R != nil {
		for _, rel := range o.R.Downloads {
			queries.SetScanner(&rel.FileID, nil)
			if rel.R == nil {
				continue
			}

			rel.R.File = nil
		}

		o.R.Downloads = nil
	}
	return o.AddDownloads(ctx, exec, insert, related...)
}

// RemoveDownloads relationships from objects passed in.
// Removes related items from R.Downloads (uses pointer comparison, removal does not keep order)
// Sets related.R.File.
func (o *File) RemoveDownloads(ctx context.Context, exec boil.ContextExecutor, related ...*Download) error {
	var err error
	for _, rel := range related {
		queries.SetScanner(&rel.FileID, nil)
		if rel.R != nil {
			rel.R.File = nil
		}
		if _, err = rel.Update(ctx, exec, boil.Whitelist("file_id")); err != nil {
			return err
		}
	}
	if o.R == nil {
		return nil
	}

	for _, rel := range related {
		for i, ri := range o.R.Downloads {
			if rel != ri {
				continue
			}

			ln := len(o.R.Downloads)
			if ln > 1 && i < ln-1 {
				o.R.Downloads[i] = o.R.Downloads[ln-1]
			}
			o.R.Downloads = o.R.Downloads[:ln-1]
			break
		}
	}

	return nil
}

// Files retrieves all the records using an executor.
func Files(mods ...qm.QueryMod) fileQuery {
	mods = append(mods, qm.From("\"file\""))
	return fileQuery{NewQuery(mods...)}
}

// FindFile retrieves a single record by ID with an executor.
// If selectCols is empty Find will return all columns.
func FindFile(ctx context.Context, exec boil.ContextExecutor, iD int, selectCols ...string) (*File, error) {
	fileObj := &File{}

	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from \"file\" where \"id\"=$1", sel,
	)

	q := queries.Raw(query, iD)

	err := q.Bind(ctx, exec, fileObj)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "dal: unable to select from file")
	}

	return fileObj, nil
}

// Insert a single record using an executor.
// See boil.Columns.InsertColumnSet documentation to understand column list inference for inserts.
func (o *File) Insert(ctx context.Context, exec boil.ContextExecutor, columns boil.Columns) error {
	if o == nil {
		return errors.New("dal: no file provided for insertion")
	}

	var err error

	nzDefaults := queries.NonZeroDefaultSet(fileColumnsWithDefault, o)

	key := makeCacheKey(columns, nzDefaults)
	fileInsertCacheMut.RLock()
	cache, cached := fileInsertCache[key]
	fileInsertCacheMut.RUnlock()

	if !cached {
		wl, returnColumns := columns.InsertColumnSet(
			fileAllColumns,
			fileColumnsWithDefault,
			fileColumnsWithoutDefault,
			nzDefaults,
		)

		cache.valueMapping, err = queries.BindMapping(fileType, fileMapping, wl)
		if err != nil {
			return err
		}
		cache.retMapping, err = queries.BindMapping(fileType, fileMapping, returnColumns)
		if err != nil {
			return err
		}
		if len(wl) != 0 {
			cache.query = fmt.Sprintf("INSERT INTO \"file\" (\"%s\") %%sVALUES (%s)%%s", strings.Join(wl, "\",\""), strmangle.Placeholders(dialect.UseIndexPlaceholders, len(wl), 1, 1))
		} else {
			cache.query = "INSERT INTO \"file\" %sDEFAULT VALUES%s"
		}

		var queryOutput, queryReturning string

		if len(cache.retMapping) != 0 {
			queryReturning = fmt.Sprintf(" RETURNING \"%s\"", strings.Join(returnColumns, "\",\""))
		}

		cache.query = fmt.Sprintf(cache.query, queryOutput, queryReturning)
	}

	value := reflect.Indirect(reflect.ValueOf(o))
	vals := queries.ValuesFromMapping(value, cache.valueMapping)

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, cache.query)
		fmt.Fprintln(writer, vals)
	}

	if len(cache.retMapping) != 0 {
		err = exec.QueryRowContext(ctx, cache.query, vals...).Scan(queries.PtrsFromMapping(value, cache.retMapping)...)
	} else {
		_, err = exec.ExecContext(ctx, cache.query, vals...)
	}

	if err != nil {
		return errors.Wrap(err, "dal: unable to insert into file")
	}

	if !cached {
		fileInsertCacheMut.Lock()
		fileInsertCache[key] = cache
		fileInsertCacheMut.Unlock()
	}

	return nil
}

// Update uses an executor to update the File.
// See boil.Columns.UpdateColumnSet documentation to understand column list inference for updates.
// Update does not automatically update the record in case of default values. Use .Reload() to refresh the records.
func (o *File) Update(ctx context.Context, exec boil.ContextExecutor, columns boil.Columns) (int64, error) {
	var err error
	key := makeCacheKey(columns, nil)
	fileUpdateCacheMut.RLock()
	cache, cached := fileUpdateCache[key]
	fileUpdateCacheMut.RUnlock()

	if !cached {
		wl := columns.UpdateColumnSet(
			fileAllColumns,
			filePrimaryKeyColumns,
		)

		if len(wl) == 0 {
			return 0, errors.New("dal: unable to update file, could not build whitelist")
		}

		cache.query = fmt.Sprintf("UPDATE \"file\" SET %s WHERE %s",
			strmangle.SetParamNames("\"", "\"", 1, wl),
			strmangle.WhereClause("\"", "\"", len(wl)+1, filePrimaryKeyColumns),
		)
		cache.valueMapping, err = queries.BindMapping(fileType, fileMapping, append(wl, filePrimaryKeyColumns...))
		if err != nil {
			return 0, err
		}
	}

	values := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), cache.valueMapping)

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, cache.query)
		fmt.Fprintln(writer, values)
	}
	var result sql.Result
	result, err = exec.ExecContext(ctx, cache.query, values...)
	if err != nil {
		return 0, errors.Wrap(err, "dal: unable to update file row")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "dal: failed to get rows affected by update for file")
	}

	if !cached {
		fileUpdateCacheMut.Lock()
		fileUpdateCache[key] = cache
		fileUpdateCacheMut.Unlock()
	}

	return rowsAff, nil
}

// UpdateAll updates all rows with the specified column values.
func (q fileQuery) UpdateAll(ctx context.Context, exec boil.ContextExecutor, cols M) (int64, error) {
	queries.SetUpdate(q.Query, cols)

	result, err := q.Query.ExecContext(ctx, exec)
	if err != nil {
		return 0, errors.Wrap(err, "dal: unable to update all for file")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "dal: unable to retrieve rows affected for file")
	}

	return rowsAff, nil
}

// UpdateAll updates all rows with the specified column values, using an executor.
func (o FileSlice) UpdateAll(ctx context.Context, exec boil.ContextExecutor, cols M) (int64, error) {
	ln := int64(len(o))
	if ln == 0 {
		return 0, nil
	}

	if len(cols) == 0 {
		return 0, errors.New("dal: update all requires at least one column argument")
	}

	colNames := make([]string, len(cols))
	args := make([]interface{}, len(cols))

	i := 0
	for name, value := range cols {
		colNames[i] = name
		args[i] = value
		i++
	}

	// Append all of the primary key values for each column
	for _, obj := range o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), filePrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := fmt.Sprintf("UPDATE \"file\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, colNames),
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), len(colNames)+1, filePrimaryKeyColumns, len(o)))

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args...)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "dal: unable to update all in file slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "dal: unable to retrieve rows affected all in update all file")
	}
	return rowsAff, nil
}

// Upsert attempts an insert using an executor, and does an update or ignore on conflict.
// See boil.Columns documentation for how to properly use updateColumns and insertColumns.
func (o *File) Upsert(ctx context.Context, exec boil.ContextExecutor, updateOnConflict bool, conflictColumns []string, updateColumns, insertColumns boil.Columns) error {
	if o == nil {
		return errors.New("dal: no file provided for upsert")
	}

	nzDefaults := queries.NonZeroDefaultSet(fileColumnsWithDefault, o)

	// Build cache key in-line uglily - mysql vs psql problems
	buf := strmangle.GetBuffer()
	if updateOnConflict {
		buf.WriteByte('t')
	} else {
		buf.WriteByte('f')
	}
	buf.WriteByte('.')
	for _, c := range conflictColumns {
		buf.WriteString(c)
	}
	buf.WriteByte('.')
	buf.WriteString(strconv.Itoa(updateColumns.Kind))
	for _, c := range updateColumns.Cols {
		buf.WriteString(c)
	}
	buf.WriteByte('.')
	buf.WriteString(strconv.Itoa(insertColumns.Kind))
	for _, c := range insertColumns.Cols {
		buf.WriteString(c)
	}
	buf.WriteByte('.')
	for _, c := range nzDefaults {
		buf.WriteString(c)
	}
	key := buf.String()
	strmangle.PutBuffer(buf)

	fileUpsertCacheMut.RLock()
	cache, cached := fileUpsertCache[key]
	fileUpsertCacheMut.RUnlock()

	var err error

	if !cached {
		insert, ret := insertColumns.InsertColumnSet(
			fileAllColumns,
			fileColumnsWithDefault,
			fileColumnsWithoutDefault,
			nzDefaults,
		)
		update := updateColumns.UpdateColumnSet(
			fileAllColumns,
			filePrimaryKeyColumns,
		)

		if updateOnConflict && len(update) == 0 {
			return errors.New("dal: unable to upsert file, could not build update column list")
		}

		conflict := conflictColumns
		if len(conflict) == 0 {
			conflict = make([]string, len(filePrimaryKeyColumns))
			copy(conflict, filePrimaryKeyColumns)
		}
		cache.query = buildUpsertQueryPostgres(dialect, "\"file\"", updateOnConflict, ret, update, conflict, insert)

		cache.valueMapping, err = queries.BindMapping(fileType, fileMapping, insert)
		if err != nil {
			return err
		}
		if len(ret) != 0 {
			cache.retMapping, err = queries.BindMapping(fileType, fileMapping, ret)
			if err != nil {
				return err
			}
		}
	}

	value := reflect.Indirect(reflect.ValueOf(o))
	vals := queries.ValuesFromMapping(value, cache.valueMapping)
	var returns []interface{}
	if len(cache.retMapping) != 0 {
		returns = queries.PtrsFromMapping(value, cache.retMapping)
	}

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, cache.query)
		fmt.Fprintln(writer, vals)
	}
	if len(cache.retMapping) != 0 {
		err = exec.QueryRowContext(ctx, cache.query, vals...).Scan(returns...)
		if err == sql.ErrNoRows {
			err = nil // Postgres doesn't return anything when there's no update
		}
	} else {
		_, err = exec.ExecContext(ctx, cache.query, vals...)
	}
	if err != nil {
		return errors.Wrap(err, "dal: unable to upsert file")
	}

	if !cached {
		fileUpsertCacheMut.Lock()
		fileUpsertCache[key] = cache
		fileUpsertCacheMut.Unlock()
	}

	return nil
}

// Delete deletes a single File record with an executor.
// Delete will match against the primary key column to find the record to delete.
func (o *File) Delete(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if o == nil {
		return 0, errors.New("dal: no File provided for delete")
	}

	args := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), filePrimaryKeyMapping)
	sql := "DELETE FROM \"file\" WHERE \"id\"=$1"

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args...)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "dal: unable to delete from file")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "dal: failed to get rows affected by delete for file")
	}

	return rowsAff, nil
}

// DeleteAll deletes all matching rows.
func (q fileQuery) DeleteAll(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if q.Query == nil {
		return 0, errors.New("dal: no fileQuery provided for delete all")
	}

	queries.SetDelete(q.Query)

	result, err := q.Query.ExecContext(ctx, exec)
	if err != nil {
		return 0, errors.Wrap(err, "dal: unable to delete all from file")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "dal: failed to get rows affected by deleteall for file")
	}

	return rowsAff, nil
}

// DeleteAll deletes all rows in the slice, using an executor.
func (o FileSlice) DeleteAll(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if len(o) == 0 {
		return 0, nil
	}

	var args []interface{}
	for _, obj := range o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), filePrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "DELETE FROM \"file\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, filePrimaryKeyColumns, len(o))

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "dal: unable to delete all from file slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "dal: failed to get rows affected by deleteall for file")
	}

	return rowsAff, nil
}

// Reload refetches the object from the database
// using the primary keys with an executor.
func (o *File) Reload(ctx context.Context, exec boil.ContextExecutor) error {
	ret, err := FindFile(ctx, exec, o.ID)
	if err != nil {
		return err
	}

	*o = *ret
	return nil
}

// ReloadAll refetches every row with matching primary key column values
// and overwrites the original object slice with the newly updated slice.
func (o *FileSlice) ReloadAll(ctx context.Context, exec boil.ContextExecutor) error {
	if o == nil || len(*o) == 0 {
		return nil
	}

	slice := FileSlice{}
	var args []interface{}
	for _, obj := range *o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), filePrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "SELECT \"file\".* FROM \"file\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, filePrimaryKeyColumns, len(*o))

	q := queries.Raw(sql, args...)

	err := q.Bind(ctx, exec, &slice)
	if err != nil {
		return errors.Wrap(err, "dal: unable to reload all in FileSlice")
	}

	*o = slice

	return nil
}

// FileExists checks if the File row exists.
func FileExists(ctx context.Context, exec boil.ContextExecutor, iD int) (bool, error) {
	var exists bool
	sql := "select exists(select 1 from \"file\" where \"id\"=$1 limit 1)"

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, iD)
	}
	row := exec.QueryRowContext(ctx, sql, iD)

	err := row.Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "dal: unable to check if file exists")
	}

	return exists, nil
}
