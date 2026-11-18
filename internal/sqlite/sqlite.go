package sqlite

/*
#cgo LDFLAGS: -lsqlite3
#include <sqlite3.h>
#include <stdlib.h>

static int bind_text(sqlite3_stmt *stmt, int idx, const char *val, int n) {
    return sqlite3_bind_text(stmt, idx, val, n, SQLITE_TRANSIENT);
}
*/
import "C"

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"io"
	"sync"
	"unsafe"
)

func init() {
	sql.Register("sqlite3", &Driver{})
}

type Driver struct{}

type Conn struct {
	db *C.sqlite3
	mu sync.Mutex
}

type Stmt struct {
	conn  *Conn
	stmt  *C.sqlite3_stmt
	query string
}

type Result struct {
	lastID       int64
	rowsAffected int64
}

type Rows struct {
	stmt *Stmt
	cols []string
	done bool
}

type Tx struct {
	conn *Conn
}

func (d *Driver) Open(dsn string) (driver.Conn, error) {
	cDSN := C.CString(dsn)
	defer C.free(unsafe.Pointer(cDSN))

	var db *C.sqlite3
	flags := C.int(C.SQLITE_OPEN_READWRITE | C.SQLITE_OPEN_CREATE | C.SQLITE_OPEN_FULLMUTEX | C.SQLITE_OPEN_URI)
	rc := C.sqlite3_open_v2(cDSN, &db, flags, nil)
	if rc != C.SQLITE_OK {
		errMsg := "sqlite open error"
		if db != nil {
			errMsg = C.GoString(C.sqlite3_errmsg(db))
			C.sqlite3_close(db)
		}
		return nil, fmt.Errorf("sqlite3 open (%d): %s", int(rc), errMsg)
	}

	C.sqlite3_busy_timeout(db, 5000)
	conn := &Conn{db: db}
	return conn, nil
}

func (c *Conn) lastError() error {
	if c.db == nil {
		return errors.New("sqlite3: connection closed")
	}
	return fmt.Errorf("sqlite3: %s", C.GoString(C.sqlite3_errmsg(c.db)))
}

func (c *Conn) Prepare(query string) (driver.Stmt, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.prepareLocked(query)
}

func (c *Conn) prepareLocked(query string) (driver.Stmt, error) {
	if c.db == nil {
		return nil, driver.ErrBadConn
	}
	cQuery := C.CString(query)
	defer C.free(unsafe.Pointer(cQuery))

	var stmt *C.sqlite3_stmt
	rc := C.sqlite3_prepare_v2(c.db, cQuery, C.int(len(query)), &stmt, nil)
	if rc != C.SQLITE_OK {
		return nil, c.lastError()
	}
	return &Stmt{conn: c, stmt: stmt, query: query}, nil
}

func (c *Conn) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.db == nil {
		return nil
	}
	rc := C.sqlite3_close_v2(c.db)
	if rc != C.SQLITE_OK {
		return c.lastError()
	}
	c.db = nil
	return nil
}

func (c *Conn) Begin() (driver.Tx, error) {
	if _, err := c.ExecContext(context.Background(), "BEGIN", nil); err != nil {
		return nil, err
	}
	return &Tx{conn: c}, nil
}

func (c *Conn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(args) == 0 {
		cQuery := C.CString(query)
		defer C.free(unsafe.Pointer(cQuery))
		var errMsg *C.char
		rc := C.sqlite3_exec(c.db, cQuery, nil, nil, &errMsg)
		if rc != C.SQLITE_OK {
			msg := C.GoString(errMsg)
			C.sqlite3_free(unsafe.Pointer(errMsg))
			return nil, fmt.Errorf("sqlite3 exec: %s", msg)
		}
		return &Result{
			lastID:       int64(C.sqlite3_last_insert_rowid(c.db)),
			rowsAffected: int64(C.sqlite3_changes(c.db)),
		}, nil
	}

	s, err := c.prepareLocked(query)
	if err != nil {
		return nil, err
	}
	stmt := s.(*Stmt)
	defer stmt.closeLocked()

	vals := make([]driver.Value, len(args))
	for i, a := range args {
		vals[i] = a.Value
	}
	return stmt.execLocked(vals)
}

func (c *Conn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	s, err := c.prepareLocked(query)
	if err != nil {
		return nil, err
	}
	stmt := s.(*Stmt)

	vals := make([]driver.Value, len(args))
	for i, a := range args {
		vals[i] = a.Value
	}
	rows, err := stmt.queryLocked(vals)
	if err != nil {
		stmt.closeLocked()
		return nil, err
	}
	return rows, nil
}

func (tx *Tx) Commit() error {
	_, err := tx.conn.ExecContext(context.Background(), "COMMIT", nil)
	return err
}

func (tx *Tx) Rollback() error {
	_, err := tx.conn.ExecContext(context.Background(), "ROLLBACK", nil)
	return err
}

func (s *Stmt) Close() error {
	s.conn.mu.Lock()
	defer s.conn.mu.Unlock()
	return s.closeLocked()
}

func (s *Stmt) closeLocked() error {
	if s.stmt == nil {
		return nil
	}
	C.sqlite3_finalize(s.stmt)
	s.stmt = nil
	return nil
}

func (s *Stmt) NumInput() int {
	if s.stmt == nil {
		return -1
	}
	return int(C.sqlite3_bind_parameter_count(s.stmt))
}

func (s *Stmt) bindLocked(args []driver.Value) error {
	C.sqlite3_reset(s.stmt)
	C.sqlite3_clear_bindings(s.stmt)

	for i, v := range args {
		idx := C.int(i + 1)
		var rc C.int
		switch val := v.(type) {
		case nil:
			rc = C.sqlite3_bind_null(s.stmt, idx)
		case int64:
			rc = C.sqlite3_bind_int64(s.stmt, idx, C.sqlite3_int64(val))
		case float64:
			rc = C.sqlite3_bind_double(s.stmt, idx, C.double(val))
		case bool:
			b := C.int(0)
			if val {
				b = 1
			}
			rc = C.sqlite3_bind_int(s.stmt, idx, b)
		case string:
			cStr := C.CString(val)
			rc = C.bind_text(s.stmt, idx, cStr, C.int(len(val)))
			C.free(unsafe.Pointer(cStr))
		case []byte:
			str := string(val)
			cStr := C.CString(str)
			rc = C.bind_text(s.stmt, idx, cStr, C.int(len(str)))
			C.free(unsafe.Pointer(cStr))
		default:
			str := fmt.Sprintf("%v", val)
			cStr := C.CString(str)
			rc = C.bind_text(s.stmt, idx, cStr, C.int(len(str)))
			C.free(unsafe.Pointer(cStr))
		}
		if rc != C.SQLITE_OK {
			return s.conn.lastError()
		}
	}
	return nil
}

func (s *Stmt) Exec(args []driver.Value) (driver.Result, error) {
	s.conn.mu.Lock()
	defer s.conn.mu.Unlock()
	return s.execLocked(args)
}

func (s *Stmt) execLocked(args []driver.Value) (driver.Result, error) {
	if err := s.bindLocked(args); err != nil {
		return nil, err
	}
	rc := C.sqlite3_step(s.stmt)
	if rc != C.SQLITE_DONE && rc != C.SQLITE_ROW {
		return nil, s.conn.lastError()
	}
	return &Result{
		lastID:       int64(C.sqlite3_last_insert_rowid(s.conn.db)),
		rowsAffected: int64(C.sqlite3_changes(s.conn.db)),
	}, nil
}

func (s *Stmt) Query(args []driver.Value) (driver.Rows, error) {
	s.conn.mu.Lock()
	defer s.conn.mu.Unlock()
	return s.queryLocked(args)
}

func (s *Stmt) queryLocked(args []driver.Value) (driver.Rows, error) {
	if err := s.bindLocked(args); err != nil {
		return nil, err
	}
	colCount := int(C.sqlite3_column_count(s.stmt))
	cols := make([]string, colCount)
	for i := 0; i < colCount; i++ {
		cols[i] = C.GoString(C.sqlite3_column_name(s.stmt, C.int(i)))
	}
	return &Rows{stmt: s, cols: cols}, nil
}

func (r *Result) LastInsertId() (int64, error) {
	return r.lastID, nil
}

func (r *Result) RowsAffected() (int64, error) {
	return r.rowsAffected, nil
}

func (r *Rows) Columns() []string {
	return r.cols
}

func (r *Rows) Close() error {
	if r.stmt == nil {
		return nil
	}
	err := r.stmt.Close()
	r.stmt = nil
	return err
}

func (r *Rows) Next(dest []driver.Value) error {
	if r.done || r.stmt == nil || r.stmt.stmt == nil {
		return io.EOF
	}
	r.stmt.conn.mu.Lock()
	defer r.stmt.conn.mu.Unlock()

	rc := C.sqlite3_step(r.stmt.stmt)
	if rc == C.SQLITE_DONE {
		r.done = true
		return io.EOF
	}
	if rc != C.SQLITE_ROW {
		return r.stmt.conn.lastError()
	}

	for i := range dest {
		colIdx := C.int(i)
		colType := C.sqlite3_column_type(r.stmt.stmt, colIdx)
		switch colType {
		case C.SQLITE_INTEGER:
			dest[i] = int64(C.sqlite3_column_int64(r.stmt.stmt, colIdx))
		case C.SQLITE_FLOAT:
			dest[i] = float64(C.sqlite3_column_double(r.stmt.stmt, colIdx))
		case C.SQLITE_TEXT, C.SQLITE_BLOB:
			ptr := C.sqlite3_column_text(r.stmt.stmt, colIdx)
			n := C.sqlite3_column_bytes(r.stmt.stmt, colIdx)
			dest[i] = C.GoStringN((*C.char)(unsafe.Pointer(ptr)), n)
		case C.SQLITE_NULL:
			dest[i] = nil
		default:
			ptr := C.sqlite3_column_text(r.stmt.stmt, colIdx)
			dest[i] = C.GoString((*C.char)(unsafe.Pointer(ptr)))
		}
	}
	return nil
}
