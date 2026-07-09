package app

import (
	"testing"
	"time"
)

// The CSV upload/import pipeline, converted to playsql. None of it had coverage,
// which is how failUpload shipped with its arguments transposed.

func mkUploadSession(t *testing.T, f *fixture, id string) {
	t.Helper()
	if err := f.s.storeUploadSession(&UploadSession{
		ID:        id,
		UserID:    int64(f.user.Id),
		Filename:  "items.csv",
		FileSize:  2048,
		Delimiter: ",",
		Encoding:  "utf8",
		Status:    "uploading",
	}); err != nil {
		t.Fatalf("storeUploadSession: %v", err)
	}
}

// mkImport inserts an import row directly: storeImport reads the user off the form's
// auth context, which is not worth assembling here.
func mkImport(t *testing.T, f *fixture) string {
	t.Helper()
	var id string
	err := f.s.db.QueryRow(`
		INSERT INTO imports (id, upload_id, user_id, source, status)
		VALUES (gen_random_uuid(), gen_random_uuid(), gen_random_uuid(), 'items', 'queued')
		RETURNING id::text`).Scan(&id)
	if err != nil {
		t.Fatalf("insert import: %v", err)
	}
	return id
}

// TestStoreAndFindUploadSession: the string primary key is inserted as given
// (playsql treats a non-integer key as non-incrementing), and the timestamps are
// stamped where the raw INSERT wrote NOW().
func TestStoreAndFindUploadSession(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	mkUploadSession(t, f, "sess-1")

	sess, err := s.findUploadSession("sess-1")
	is.NoErr(err)
	is.Equal(sess.ID, "sess-1")
	is.Equal(sess.Filename, "items.csv")
	is.Equal(sess.Status, "uploading")
	is.Equal(int(sess.FileSize), 2048)
	is.Equal(sess.UploadedChunks, 0)
	is.True(!sess.TotalChunks.Valid, "total_chunks starts null")
	is.True(!sess.ErrorMessage.Valid, "error_message starts null")
	is.True(!sess.CreatedAt.IsZero(), "created_at should be stamped")

	// A missing session is (nil, nil), not an error.
	missing, err := s.findUploadSession("nope")
	is.NoErr(err)
	is.True(missing == nil, "an unknown session returns nil, nil")
}

// TestFailUpload: the statement this replaced passed `message, id` against a query
// whose $1 was the id — so it matched `WHERE id = <the error message>`, updated zero
// rows and returned nil. An upload failure was never recorded.
func TestFailUpload(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	mkUploadSession(t, f, "sess-fail")
	is.NoErr(s.failUpload("sess-fail", "boom: bad header"))

	sess, err := s.findUploadSession("sess-fail")
	is.NoErr(err)
	is.Equal(sess.Status, "failed")
	is.True(sess.ErrorMessage.Valid, "the failure message must be recorded")
	is.Equal(sess.ErrorMessage.String, "boom: bad header")

	// And the message was never mistaken for an id.
	other, err := s.findUploadSession("boom: bad header")
	is.NoErr(err)
	is.True(other == nil, "the message must not be treated as a session id")
}

func TestUpdateUploadStatusAndChunks(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	mkUploadSession(t, f, "sess-2")

	is.NoErr(s.updateUploadStatus("sess-2", "assembling"))
	sess, err := s.findUploadSession("sess-2")
	is.NoErr(err)
	is.Equal(sess.Status, "assembling")

	// total_chunks is write-once: the WhereNull guard is the old `AND total_chunks IS NULL`.
	is.NoErr(s.updateTotalChunks("sess-2", 7))
	sess, _ = s.findUploadSession("sess-2")
	is.True(sess.TotalChunks.Valid, "total_chunks should be set")
	is.Equal(int(sess.TotalChunks.Int64), 7)

	is.NoErr(s.updateTotalChunks("sess-2", 99))
	sess, _ = s.findUploadSession("sess-2")
	is.Equal(int(sess.TotalChunks.Int64), 7) // unchanged

	// The increment stays raw SQL.
	is.NoErr(s.incrementUploadedChunks("sess-2"))
	is.NoErr(s.incrementUploadedChunks("sess-2"))
	sess, _ = s.findUploadSession("sess-2")
	is.Equal(sess.UploadedChunks, 2)
}

// TestUpdateUploadSession_BumpsUpdatedAt: playsql stamps updated_at because
// uploadSessionRead maps it, where the raw statements said `updated_at = NOW()`.
func TestUpdateUploadSession_BumpsUpdatedAt(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	mkUploadSession(t, f, "sess-3")
	before, err := s.findUploadSession("sess-3")
	is.NoErr(err)

	time.Sleep(2 * time.Millisecond)
	is.NoErr(s.updateUploadStatus("sess-3", "assembling"))

	after, err := s.findUploadSession("sess-3")
	is.NoErr(err)
	is.True(after.UpdatedAt.After(before.UpdatedAt), "updateUploadStatus should bump updated_at")
}

func TestImportLifecycle(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	id := mkImport(t, f)

	imp, err := s.findImportByID(id)
	is.NoErr(err)
	is.Equal(imp.Status, "queued")
	is.True(imp.StartedAt == nil, "not started yet")
	is.Equal(*imp.TotalRows, 0) // the row counters default to 0, not null

	is.NoErr(s.markStarted(id))
	is.NoErr(s.updateTotalRows(id, 120))

	imp, err = s.findImportByID(id)
	is.NoErr(err)
	is.Equal(imp.Status, "processing")
	is.True(imp.StartedAt != nil, "started_at should be stamped")
	is.Equal(*imp.TotalRows, 120)

	s.completeImport(imp)
	imp, err = s.findImportByID(id)
	is.NoErr(err)
	is.Equal(imp.Status, "completed")
	is.True(imp.FinishedAt != nil, "finished_at should be stamped")
}

func TestFailImport(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	id := mkImport(t, f)
	s.failImport(id, "unreadable csv")

	imp, err := s.findImportByID(id)
	is.NoErr(err)
	is.Equal(imp.Status, "failed")
	is.True(imp.ErrorMEssage != nil, "the failure message should be recorded")
	is.Equal(*imp.ErrorMEssage, "unreadable csv")
	is.True(imp.FinishedAt != nil, "finished_at should be stamped")
}

// TestUpdateProgressAndIssues: updateProgress and the two issue writers all run on
// the caller's transaction. storeImportIssues replaced hand-built placeholders with
// a single InsertMany.
func TestUpdateProgressAndIssues(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	id := mkImport(t, f)

	tx, err := s.db.Begin()
	is.NoErr(err)

	is.NoErr(s.updateProgress(tx, id, 10, 8, 2, 1))
	is.NoErr(s.saveRowIssue(tx, id, ImportIssue{
		Row: 3, Column: "price", Level: IssueLevel.Error, Message: "not a number", Value: "abc",
	}))
	is.NoErr(s.storeImportIssues(tx, id, []ImportIssue{
		{Row: 4, Column: "sku", Level: IssueLevel.Warning, Message: "blank", Value: ""},
		{Row: 5, Column: "name", Level: IssueLevel.Warning, Message: "truncated", Value: "xxx"},
	}))
	is.NoErr(tx.Commit())

	imp, err := s.findImportByID(id)
	is.NoErr(err)
	is.Equal(*imp.ProcessedRows, 10)
	is.Equal(*imp.SuccessRows, 8)
	is.Equal(*imp.FailedRows, 2)
	is.Equal(*imp.WarningRows, 1)

	is.Equal(scalarInt(t, s.db, `SELECT count(*) FROM import_row_issues WHERE import_id = $1::uuid`, id), 3)
	is.Equal(scalarString(t, s.db,
		`SELECT message FROM import_row_issues WHERE import_id = $1::uuid AND row_number = 3`, id), "not a number")

	// An empty batch is a no-op, not an error.
	tx2, err := s.db.Begin()
	is.NoErr(err)
	is.NoErr(s.storeImportIssues(tx2, id, nil))
	is.NoErr(tx2.Commit())
}
