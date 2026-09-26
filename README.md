# Students API

A REST API for students
**"Build a Fast & Scalable REST API with GoLang"**


It uses Go's standard `net/http` router, **raw SQL** with `database/sql` and
SQLite, a YAML config file, request validation, structured logging (`slog`),
and graceful shutdown.

> **One difference from the video:** the video uses the `mattn/go-sqlite3`
> driver, which needs a C compiler (gcc). This project uses the pure-Go
> `modernc.org/sqlite` driver instead. Only the import and the driver name
> (`"sqlite"` instead of `"sqlite3"`) in `internal/storage/sqlite/sqlite.go` differ.

## Bug fixes compared with the video

| # | Problem in the video's code | Fix |
|---|---|---|
| 1 | Ctrl+C printed `failed to start server` and exited before the graceful shutdown finished | ignore `http.ErrServerClosed` in `main.go` |
| 2 | `GET /api/students/{id}` for a missing ID returned **500** | new `storage.ErrNotFound`, handler returns **404** |
| 3 | database errors were sent as `{}` | send `response.GeneralError(err)` so the message is visible |
| 4 | "user created successfully" was logged even when the insert failed | log only after the error check |

## Run it

```bash
go run ./cmd/students-api -config config/local.yaml
# or:  CONFIG_PATH=config/local.yaml go run ./cmd/students-api
```

The server starts on http://localhost:8082 and stores data in `storage/storage.db`.
Press Ctrl+C to stop it.

## Project structure

```
example-project/
├── cmd/students-api/main.go              ← entry point: config → storage → router → server → graceful shutdown
├── config/local.yaml                     ← settings (env, database path, server address)
└── internal/
    ├── config/config.go                  ← loads the YAML into a Config struct (cleanenv)
    ├── types/types.go                    ← the Student struct (+ validation tags)
    ├── storage/storage.go                ← the Storage INTERFACE (what the database layer must do)
    ├── storage/sqlite/sqlite.go          ← the SQLite implementation (raw SQL)
    ├── http/handlers/student/student.go  ← the APIs (handlers)
    └── utils/response/response.go        ← JSON response + error helpers
```

A request flows like this:

```
HTTP request → handler (student.go) → Storage interface → sqlite.go → storage/storage.db
```

## The APIs

| Method | URL | What it does |
|---|---|---|
| POST | `/api/students` | create a student |
| GET | `/api/students/{id}` | get one student |
| GET | `/api/students` | list all students |

## Try it

```bash
curl -X POST localhost:8082/api/students -d '{"name":"Ravi","email":"ravi@example.com","age":25}'
curl localhost:8082/api/students/1
curl localhost:8082/api/students
```

## Libraries

| Library | Used for |
|---|---|
| `github.com/ilyakaznacheev/cleanenv` | reading `config/local.yaml` |
| `github.com/go-playground/validator/v10` | `validate:"required"` checks |
| `modernc.org/sqlite` | the SQLite database driver |

## The old advanced version

The `_advanced/` folder holds older practice code. Go ignores folders whose
name starts with `_`, so it doesn't affect this project.
