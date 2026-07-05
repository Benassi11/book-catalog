# Library API

A small REST API for managing a book catalog, written in Go with an embedded SQLite database (via `modernc.org/sqlite`, no CGO required).

## Project structure

```
.
├── cmd/api/main.go       # entrypoint: HTTP routes, server bootstrap
├── internal/book/book.go # Book model, validation, and SQLite-backed Store
├── web/static/           # static frontend (served at "/")
├── bruno/                # Bruno API collection for manual testing
├── data/                 # SQLite database file (gitignored, created at runtime)
├── Dockerfile
├── docker-compose.yml
└── go.mod
```

This follows the common Go project layout: `cmd/` holds the binary entrypoint, `internal/` holds code that must not be imported by other modules, and `web/` holds web assets.

## Requirements

- Go 1.25+

## Running locally

```bash
go run ./cmd/api
```

The server listens on `http://localhost:8080`.

- Static frontend is served at `/`.
- API is served under `/api/books`.

### Environment variables

| Env var     | Default        | Description                      |
| ----------- | -------------- | -------------------------------- |
| `DB_PATH` | `library.db` | Path to the SQLite database file |

## API

All request/response bodies are JSON.

| Method | Path                | Description         |
| ------ | ------------------- | ------------------- |
| GET    | `/api/books`      | List all books      |
| POST   | `/api/books`      | Create a book       |
| GET    | `/api/books/{id}` | Get a book by ID    |
| PUT    | `/api/books/{id}` | Update a book by ID |
| DELETE | `/api/books/{id}` | Delete a book by ID |

### Book object

```json
{
  "id": 1,
  "title": "The Great Gatsby",
  "author": "F. Scott Fitzgerald",
  "year": 1925
}
```

`title`, `author`, and `year` are required. `year` must be a valid year not in the future.

Status codes

- `200 OK` — successful GET/PUT
- `201 Created` — successful POST
- `204 No Content` — successful DELETE
- `400 Bad Request` — invalid body/id or failed validation
- `404 Not Found` — book does not exist

A ready-to-use [Bruno](https://www.usebruno.com/) collection with all these requests is available in the `bruno/` directory.

## Running with Docker

```bash
docker compose up --build
```

This builds the Go binary in a multi-stage Docker build, copies the static assets, and starts the container on port `8080`. The `data/` directory is mounted as a volume so the SQLite database persists across restarts, and `DB_PATH` is set to `/app/data/library.db`.

## Notes

- No test suite yet.
- No `pkg/` directory — all shared code currently lives under `internal/book`, which is intentional since nothing here is meant to be imported by external modules.
