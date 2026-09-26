// Package sqlite implements the storage.Storage interface with RAW SQL.
//
// JAVA EQUIVALENT: a repository written with plain JDBC
// (Connection, PreparedStatement, ResultSet), no JPA.
package sqlite

import (
	"database/sql" // Go's standard database API (like java.sql / JDBC)
	"fmt"

	"example.com/students-api/internal/config"
	"example.com/students-api/internal/types"

	// BLANK IMPORT (`_`): we never call this package directly. Importing it
	// runs its init() function, which REGISTERS the "sqlite" driver with
	// database/sql. Like loading a JDBC driver with Class.forName(...).
	//
	// NOTE: the video uses github.com/mattn/go-sqlite3 (driver name "sqlite3"),
	// which needs a C compiler (gcc). modernc.org/sqlite is pure Go and works
	// the same way; only the import and the driver name differ.
	_ "modernc.org/sqlite"
)

// Sqlite holds the database connection pool.
// It satisfies storage.Storage because it has all three methods below.
type Sqlite struct {
	Db *sql.DB // a POOL of connections, safe to use from many goroutines
}

// New opens the database and creates the table if needed.
func New(cfg *config.Config) (*Sqlite, error) {
	// sql.Open(driverName, dataSource). The video uses "sqlite3" here.
	db, err := sql.Open("sqlite", cfg.StoragePath)
	if err != nil {
		return nil, err
	}

	// Exec runs SQL that returns no rows. The backticks make a RAW STRING,
	// which can span several lines.
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS students (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT,
	email TEXT,
	age INTEGER
	)`)

	if err != nil {
		return nil, err
	}

	return &Sqlite{
		Db: db,
	}, nil
}

// CreateStudent inserts a student and returns the new ID.
func (s *Sqlite) CreateStudent(name string, email string, age int) (int64, error) {

	// Prepare = PreparedStatement. The ? placeholders are filled in by Exec,
	// which protects against SQL injection.
	stmt, err := s.Db.Prepare("INSERT INTO students (name, email, age) VALUES (?, ?, ?)")
	if err != nil {
		return 0, err
	}
	defer stmt.Close() // always close the statement when the function ends

	result, err := stmt.Exec(name, email, age)
	if err != nil {
		return 0, err
	}

	// LastInsertId = the AUTOINCREMENT id the database just created.
	lastId, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return lastId, nil
}

// GetStudentById loads one student.
func (s *Sqlite) GetStudentById(id int64) (types.Student, error) {
	stmt, err := s.Db.Prepare("SELECT id, name, email, age FROM students WHERE id = ? LIMIT 1")
	if err != nil {
		return types.Student{}, err
	}

	defer stmt.Close()

	var student types.Student

	// QueryRow runs a query that returns ONE row.
	// Scan copies the columns, in order, into the variables we point to.
	// This is the manual work an ORM like GORM does for you.
	err = stmt.QueryRow(id).Scan(&student.Id, &student.Name, &student.Email, &student.Age)
	if err != nil {
		// sql.ErrNoRows = the query matched nothing.
		if err == sql.ErrNoRows {
			return types.Student{}, fmt.Errorf("no student found with id %s", fmt.Sprint(id))
		}
		// %w WRAPS the original error inside our message.
		return types.Student{}, fmt.Errorf("query error: %w", err)
	}

	return student, nil
}

// GetStudents loads every student.
func (s *Sqlite) GetStudents() ([]types.Student, error) {
	stmt, err := s.Db.Prepare("SELECT id, name, email, age FROM students")
	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	// Query (not QueryRow) returns MANY rows, like a JDBC ResultSet.
	rows, err := stmt.Query()
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var students []types.Student

	// rows.Next() moves to the next row and returns false at the end,
	// exactly like while (resultSet.next()) { ... } in JDBC.
	for rows.Next() {
		var student types.Student

		err := rows.Scan(&student.Id, &student.Name, &student.Email, &student.Age)
		if err != nil {
			return nil, err
		}

		students = append(students, student)
	}

	return students, nil
}
