package book

import (
	"errors"
	"testing"
	"time"
)

func TestBookValidate(t *testing.T) {
	futureYear := time.Now().Year() + 1

	tests := []struct {
		name    string
		book    Book
		wantErr bool
	}{
		{"valid", Book{Title: "Go", Author: "Rob", Year: 2020}, false},
		{"missing title", Book{Author: "Rob", Year: 2020}, true},
		{"missing author", Book{Title: "Go", Year: 2020}, true},
		{"missing year", Book{Title: "Go", Author: "Rob"}, true},
		{"negative year", Book{Title: "Go", Author: "Rob", Year: -1}, true},
		{"future year", Book{Title: "Go", Author: "Rob", Year: futureYear}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.book.validate()
			if tt.wantErr && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if tt.wantErr && !errors.Is(err, ErrValidation) {
				t.Fatalf("expected ErrValidation, got %v", err)
			}
		})
	}
}

func newTestStore(t *testing.T) *Store {
	t.Helper()
	s, err := NewStore(":memory:")
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

func TestStoreCreateAndGet(t *testing.T) {
	s := newTestStore(t)

	created, err := s.Create(Book{Title: "Go", Author: "Rob", Year: 2020})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if created.ID == 0 {
		t.Fatalf("expected non-zero ID")
	}

	got, err := s.Get(created.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got != created {
		t.Fatalf("got %+v, want %+v", got, created)
	}
}

func TestStoreCreateInvalid(t *testing.T) {
	s := newTestStore(t)

	_, err := s.Create(Book{Author: "Rob", Year: 2020})
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("expected ErrValidation, got %v", err)
	}
}

func TestStoreGetNotFound(t *testing.T) {
	s := newTestStore(t)

	_, err := s.Get(999)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestStoreList(t *testing.T) {
	s := newTestStore(t)

	if _, err := s.Create(Book{Title: "Go", Author: "Rob", Year: 2020}); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if _, err := s.Create(Book{Title: "Rust", Author: "Graydon", Year: 2010}); err != nil {
		t.Fatalf("Create: %v", err)
	}

	list, err := s.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 books, got %d", len(list))
	}
}

func TestStoreUpdate(t *testing.T) {
	s := newTestStore(t)

	created, err := s.Create(Book{Title: "Go", Author: "Rob", Year: 2020})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	updated, err := s.Update(created.ID, Book{Title: "Go Updated", Author: "Rob Pike", Year: 2021})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if updated.ID != created.ID {
		t.Fatalf("expected ID %d, got %d", created.ID, updated.ID)
	}

	got, err := s.Get(created.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got != updated {
		t.Fatalf("got %+v, want %+v", got, updated)
	}
}

func TestStoreUpdateNotFound(t *testing.T) {
	s := newTestStore(t)

	_, err := s.Update(999, Book{Title: "Go", Author: "Rob", Year: 2020})
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestStoreUpdateInvalid(t *testing.T) {
	s := newTestStore(t)

	created, err := s.Create(Book{Title: "Go", Author: "Rob", Year: 2020})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	_, err = s.Update(created.ID, Book{Title: "", Author: "Rob", Year: 2020})
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("expected ErrValidation, got %v", err)
	}
}

func TestStoreDelete(t *testing.T) {
	s := newTestStore(t)

	created, err := s.Create(Book{Title: "Go", Author: "Rob", Year: 2020})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := s.Delete(created.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	_, err = s.Get(created.ID)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound after delete, got %v", err)
	}
}

func TestStoreDeleteNotFound(t *testing.T) {
	s := newTestStore(t)

	err := s.Delete(999)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
