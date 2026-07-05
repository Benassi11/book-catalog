package book

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "modernc.org/sqlite" // registers the "sqlite" driver used by sql.Open
)

var ErrNotFound = errors.New("book not found")

var ErrValidation = errors.New("validation failed")

type Book struct {
	ID     int    `json:"id"`
	Title  string `json:"title"`
	Author string `json:"author"`
	Year   int    `json:"year"`
}

func (b Book) validate() error {
	if b.Title == "" {
		return fmt.Errorf("%w: title is required", ErrValidation)
	}
	if b.Author == "" {
		return fmt.Errorf("%w: author is required", ErrValidation)
	}
	if b.Year == 0 {
		return fmt.Errorf("%w: year is required", ErrValidation)
	}
	if b.Year < 0 || b.Year > time.Now().Year() {
		return fmt.Errorf("%w: year must be a valid year", ErrValidation)
	}
	return nil
}

type Store struct {
	db *sql.DB
}

func NewStore(dbPath string) (*Store, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	const createTable = `
		CREATE TABLE IF NOT EXISTS books (
			id     INTEGER PRIMARY KEY AUTOINCREMENT,
			title  TEXT NOT NULL,
			author TEXT NOT NULL,
			year   INTEGER NOT NULL
		)`
	if _, err := db.Exec(createTable); err != nil {
		db.Close()
		return nil, err
	}

	return &Store{db: db}, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) List() ([]Book, error) {
	rows, err := s.db.Query("SELECT id, title, author, year FROM books")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]Book, 0)
	for rows.Next() {
		var b Book
		if err := rows.Scan(&b.ID, &b.Title, &b.Author, &b.Year); err != nil {
			return nil, err
		}
		list = append(list, b)
	}
	return list, rows.Err()
}

func (s *Store) Get(id int) (Book, error) {
	var b Book
	row := s.db.QueryRow("SELECT id, title, author, year FROM books WHERE id = ?", id)
	if err := row.Scan(&b.ID, &b.Title, &b.Author, &b.Year); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Book{}, ErrNotFound
		}
		return Book{}, err
	}
	return b, nil
}

func (s *Store) Create(book Book) (Book, error) {
	if err := book.validate(); err != nil {
		return Book{}, err
	}

	result, err := s.db.Exec(
		"INSERT INTO books (title, author, year) VALUES (?, ?, ?)",
		book.Title, book.Author, book.Year,
	)
	if err != nil {
		return Book{}, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return Book{}, err
	}

	book.ID = int(id)
	return book, nil
}

func (s *Store) Update(id int, book Book) (Book, error) {
	if err := book.validate(); err != nil {
		return Book{}, err
	}

	result, err := s.db.Exec(
		"UPDATE books SET title = ?, author = ?, year = ? WHERE id = ?",
		book.Title, book.Author, book.Year, id,
	)
	if err != nil {
		return Book{}, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return Book{}, err
	}
	if rowsAffected == 0 {
		return Book{}, ErrNotFound
	}

	book.ID = id
	return book, nil
}

func (s *Store) Delete(id int) error {
	result, err := s.db.Exec("DELETE FROM books WHERE id = ?", id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}
