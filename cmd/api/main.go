package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"strconv"

	"golang/internal/book"
)

func main() {
	os.Exit(run())
}

func run() int {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "library.db"
	}

	store, err := book.NewStore(dbPath)
	if err != nil {
		log.Println(err)
		return 1
	}
	defer store.Close()

	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/books", func(w http.ResponseWriter, _ *http.Request) {
		list, err := store.List()
		if err != nil {
			respondError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, list)
	})

	mux.HandleFunc("POST /api/books", func(w http.ResponseWriter, r *http.Request) {
		var receivedBook book.Book
		if err := json.NewDecoder(r.Body).Decode(&receivedBook); err != nil {
			http.Error(w, "invalid body", http.StatusBadRequest)
			return
		}

		createdBook, err := store.Create(receivedBook)
		if err != nil {
			respondError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, createdBook)
	})

	mux.HandleFunc("GET /api/books/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			http.Error(w, "invalid id", http.StatusBadRequest)
			return
		}
		foundBook, err := store.Get(id)
		if err != nil {
			respondError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, foundBook)
	})

	mux.HandleFunc("PUT /api/books/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			http.Error(w, "invalid id", http.StatusBadRequest)
			return
		}
		var receivedBook book.Book
		if err := json.NewDecoder(r.Body).Decode(&receivedBook); err != nil {
			http.Error(w, "invalid body", http.StatusBadRequest)
			return
		}
		updatedBook, err := store.Update(id, receivedBook)
		if err != nil {
			respondError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, updatedBook)
	})

	mux.HandleFunc("DELETE /api/books/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			http.Error(w, "invalid id", http.StatusBadRequest)
			return
		}
		if err := store.Delete(id); err != nil {
			respondError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})

	mux.Handle("GET /", http.FileServer(http.Dir("web/static")))

	log.Println("server running at http://localhost:8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Println(err)
		return 1
	}
	return 0
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		log.Println(err)
	}
}

func respondError(w http.ResponseWriter, err error) {
	if errors.Is(err, book.ErrNotFound) {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	if errors.Is(err, book.ErrValidation) {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	http.Error(w, err.Error(), http.StatusInternalServerError)
}
