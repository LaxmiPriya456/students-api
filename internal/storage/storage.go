// Package storage says WHAT the database layer must be able to do,
// without saying HOW.
package storage

import (
	"errors"

	"example.com/students-api/internal/types"
)

// ErrNotFound is returned when no student has the requested ID.
//
// A package-level error like this is called a SENTINEL ERROR. Callers check
// for it with errors.Is(err, storage.ErrNotFound), like catching a specific
// exception class in Java. It lets the handler answer 404 without knowing
// anything about SQL.
var ErrNotFound = errors.New("no student found")

// Storage is an INTERFACE: a list of methods, with no code.
//
// JAVA EQUIVALENT: public interface StudentRepository { ... }
//
// The big difference from Java: there is no `implements` keyword. Any type
// that has these three methods automatically IS a Storage. The sqlite.Sqlite
// type satisfies it just by having the methods.
//
// The handlers only know about this interface, so you could swap SQLite for
// PostgreSQL by writing a new type with the same methods.
type Storage interface {
	CreateStudent(name string, email string, age int) (int64, error)
	GetStudentById(id int64) (types.Student, error)
	GetStudents() ([]types.Student, error)
}
