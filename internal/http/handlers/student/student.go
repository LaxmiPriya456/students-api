// Package student has the HTTP handlers (the APIs) for students.
//
// JAVA EQUIVALENT: a @RestController. There is no service layer in this
// project: the handlers validate the input and call storage directly.
package student

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog" // structured logging, built into Go (like SLF4J)
	"net/http"
	"strconv"

	"example.com/students-api/internal/storage"
	"example.com/students-api/internal/types"
	"example.com/students-api/internal/utils/response"

	"github.com/go-playground/validator/v10"
)

// New handles POST /api/students.
//
// PATTERN: this function does not handle the request itself. It RETURNS a
// handler function (http.HandlerFunc). The inner function is a CLOSURE: it
// "remembers" the storage variable from the outer function. This is how the
// dependency is injected without a struct or @Autowired.
//
// The parameter type is the storage.Storage INTERFACE, not *sqlite.Sqlite,
// so any database implementation can be passed in.
func New(storage storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// slog.Info writes a log line with a level and time, like log.info(...)
		slog.Info("creating a student")

		var student types.Student

		// Decode the JSON body into the struct (@RequestBody).
		err := json.NewDecoder(r.Body).Decode(&student)

		// io.EOF ("end of file") = the body was completely empty.
		if errors.Is(err, io.EOF) {
			response.WriteJson(w, http.StatusBadRequest, response.GeneralError(fmt.Errorf("empty body")))
			return
		}

		if err != nil {
			response.WriteJson(w, http.StatusBadRequest, response.GeneralError(err))
			return
		}

		// Check the `validate:"required"` tags (like @Valid in Spring).
		if err := validator.New().Struct(student); err != nil {

			// TYPE ASSERTION: err is a plain `error` interface; .(validator.ValidationErrors)
			// says "I know the real type is ValidationErrors, give me that".
			// Like a cast in Java: (ValidationErrors) err
			validateErrs := err.(validator.ValidationErrors)
			response.WriteJson(w, http.StatusBadRequest, response.ValidationError(validateErrs))
			return
		}

		lastId, err := storage.CreateStudent(
			student.Name,
			student.Email,
			student.Age,
		)

		if err != nil {
			// FIX: send response.GeneralError(err), not err itself.
			// An error value has no exported fields, so encoding it directly
			// to JSON produces just {} and the client never sees the message.
			response.WriteJson(w, http.StatusInternalServerError, response.GeneralError(err))
			return
		}

		// FIX: log success only AFTER checking err. Before, this line ran even
		// when the insert failed, logging "created successfully userId=0".
		//
		// slog.String adds a key/value pair to the log line: userId=1
		slog.Info("user created successfully", slog.String("userId", fmt.Sprint(lastId)))

		// 201 Created with {"id": 1}
		response.WriteJson(w, http.StatusCreated, map[string]int64{"id": lastId})
	}
}

// GetById handles GET /api/students/{id}.
//
// The parameter is called `store` here, not `storage` like in the other
// handlers. A variable named `storage` would SHADOW (hide) the storage
// package inside this function, and we need the package to reach
// storage.ErrNotFound below.
func GetById(store storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id") // @PathVariable (always a string)
		slog.Info("getting a student", slog.String("id", id))

		// ParseInt(text, base 10, 64-bit) = Long.parseLong(id)
		intId, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			response.WriteJson(w, http.StatusBadRequest, response.GeneralError(err))
			return
		}

		student, err := store.GetStudentById(intId)

		// FIX: "not found" is the CLIENT asking for something that doesn't
		// exist, so it is 404 Not Found, not 500 Internal Server Error.
		if errors.Is(err, storage.ErrNotFound) {
			response.WriteJson(w, http.StatusNotFound, response.GeneralError(err))
			return
		}

		if err != nil {
			slog.Error("error getting user", slog.String("id", id))
			response.WriteJson(w, http.StatusInternalServerError, response.GeneralError(err))
			return
		}

		response.WriteJson(w, http.StatusOK, student)
	}
}

// GetList handles GET /api/students.
func GetList(storage storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slog.Info("getting all students")

		students, err := storage.GetStudents()
		if err != nil {
			// FIX: GeneralError(err) instead of err (see the note in New).
			response.WriteJson(w, http.StatusInternalServerError, response.GeneralError(err))
			return
		}

		response.WriteJson(w, http.StatusOK, students)
	}
}
