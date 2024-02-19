package books

import (
	"context"
	"crypto/sha1"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/3WDeveloper-GM/library_app/backend/internal/validator"
	"github.com/lib/pq"
)

var (
	ErrNotFound     = errors.New("record not found")
	ErrEditConflict = errors.New("edit conflict")
)

type Book struct {
	ID        int64    `json:"id"`
	Hash      string   `json:"hash"`
	Title     string   `json:"title"`
	Publisher string   `json:"publisher"`
	Year      int32    `json:"year"`
	PageCount int32    `json:"page_count"`
	Genres    []string `json:"genres"`
}

type Authors struct {
	List []string `json:"names"`
	Hash []string `json:"hash"`
}

type CreateBookEntry struct {
	Book    *Book    `json:"book"`
	Authors *Authors `json:"authors"`
}

func HashEntries(b *Book, a *Authors) {
	//Hash Book ID
	hash := sha1.New()
	hash.Write([]byte(b.Title))
	hashInBytes := hash.Sum(nil)

	b.Hash = hex.EncodeToString(hashInBytes)

	//Hash author entries
	hashed := make([]string, len(a.List))

	for index, value := range a.List {
		hash := sha1.New()
		hash.Write([]byte(value))
		hashInBytes := hash.Sum(nil)

		hashed[index] = hex.EncodeToString(hashInBytes)
	}

	a.Hash = hashed
}

func (b *CreateBookEntry) ValidateEntry(v *validator.Validator) bool {

	//book validation
	var section = "title"
	var mustbeProvidedMsg = "%s field must be provided"
	var maxTitleBytes = 300

	v.Check(b.Book.Title != "", section, fmt.Sprintf(
		mustbeProvidedMsg, section,
	))
	v.Check(len(b.Book.Title) <= maxTitleBytes, section, fmt.Sprintf(
		"%s field must have a length no longer than %d bytes", section, maxTitleBytes,
	))

	section = "publisher"
	var maxPublisherByteAmount = 300
	v.Check(b.Book.Publisher != "", section, fmt.Sprintf(
		mustbeProvidedMsg, section,
	))
	v.Check(len(b.Book.Publisher) <= maxPublisherByteAmount, section, fmt.Sprintf(
		"%s field must have less than %d bytes", section, maxPublisherByteAmount,
	))

	section = "year"
	var PresentDate = int32(time.Now().Year())
	var minimumYear int32 = 1900

	v.Check(b.Book.Year != 0, section, fmt.Sprintf(
		mustbeProvidedMsg, section,
	))
	v.Check(b.Book.Year <= PresentDate, section, fmt.Sprintf(
		"%s field must not be set in the future", section,
	))
	v.Check(b.Book.Year >= minimumYear, section, fmt.Sprintf(
		"%s field must not be set before %d", section, minimumYear,
	))

	section = "page_count"
	var maxPageCount int32 = 10000
	var minPageCount int32 = 1

	v.Check(b.Book.PageCount != 0, section, fmt.Sprintf(
		mustbeProvidedMsg, section,
	))
	v.Check(b.Book.PageCount <= maxPageCount, section, fmt.Sprintf(
		"%s field must not be more than %d pages", section, maxPageCount,
	))
	v.Check(b.Book.PageCount >= minPageCount, section, fmt.Sprintf(
		"%s field must have at least %d pages", section, minPageCount,
	))

	section = "genres"
	var minGenreCount int = 1
	var maxGenreCount int = 5

	v.Check(b.Book.Genres != nil, section, fmt.Sprintf(
		mustbeProvidedMsg, section,
	))
	v.Check(len(b.Book.Genres) <= maxGenreCount, section, fmt.Sprintf(
		"%s field must not contain more than %d genres", section, maxGenreCount,
	))
	v.Check(len(b.Book.Genres) >= minGenreCount, section, fmt.Sprintf(
		"%s field must contain at least %d genres", section, minGenreCount,
	))

	if !v.Valid() {
		return false
	}

	// author list validation

	section = "authors"
	var minAuthorCount int = 1
	var maxAuthorCount int = 5

	v.Check(b.Authors.List != nil, section, fmt.Sprintf(
		mustbeProvidedMsg, section,
	))
	v.Check(len(b.Authors.List) <= maxAuthorCount, section, fmt.Sprintf(
		"%s field must not contain more than %d authors", section, maxAuthorCount,
	))
	v.Check(len(b.Authors.List) >= minAuthorCount, section, fmt.Sprintf(
		"%s field must contain at least %d authors", section, minAuthorCount,
	))

	if !v.Valid() {
		return false
	}

	section = "author_items"
	var maxItemByteAmount = 100
	for _, value := range b.Authors.List {
		v.Check(value != "", section, fmt.Sprintf(
			mustbeProvidedMsg, section,
		))
		v.Check(len(value) <= maxItemByteAmount, section, fmt.Sprintf(
			"%s field must have less than %d bytes in length", section, maxItemByteAmount,
		))
		if !v.Valid() {
			return false
		}
	}

	return v.Valid()
}

type CreateEntryModel struct {
	DB *sql.DB
}

func (c *CreateEntryModel) Insert(ctx context.Context, tx *sql.Tx, entry *CreateBookEntry, read *ReadEntry) error {
	query := `
		INSERT INTO books(book_id,title,publisher,year,page_count,genres)
		VALUES($1,$2,$3,$4,$5,$6)
		RETURNING id,book_id,title,publisher,year,page_count,genres
	`

	args := []interface{}{
		entry.Book.Hash,
		entry.Book.Title,
		entry.Book.Publisher,
		entry.Book.Year,
		entry.Book.PageCount,
		pq.Array(entry.Book.Genres)}

	err := tx.QueryRowContext(ctx, query, args...).Scan(
		&read.Book.ID,
		&read.Book.Hash,
		&read.Book.Title,
		&read.Book.Publisher,
		&read.Book.Year,
		&read.Book.PageCount,
		pq.Array(&read.Book.Genres))

	if err != nil {
		return err
	}
	query = `
		WITH filter1 AS(
			SELECT
			UNNEST($1::TEXT[]) AS name,
			UNNEST($2::TEXT[]) AS author_id,
			1 as books_authored
		), finished AS(
			INSERT INTO authors(name,author_id,books_authored)
			SELECT name, author_id, books_authored FROM filter1
			ON CONFLICT (author_id) DO UPDATE
			SET books_authored = authors.books_authored + 1
			RETURNING id, name, author_id, books_authored
		)
		SELECT array_agg(id), array_agg(name), array_agg(author_id), array_agg(books_authored) FROM finished
	`

	args = []interface{}{pq.Array(entry.Authors.List), pq.Array(entry.Authors.Hash)}
	err = tx.QueryRowContext(ctx, query, args...).Scan(
		pq.Array(&read.List.ID),
		pq.Array(&read.List.Name),
		pq.Array(&read.List.Identifier),
		pq.Array(&read.List.Books_authored),
	)

	if err != nil {
		return err
	}

	query = `
	WITH cte AS (
		SELECT book_id, author_id FROM
		UNNEST($1::TEXT[]) AS book_id,
		UNNEST($2::TEXT[]) AS author_id 
  ), finished AS (
  	INSERT INTO book_author_link(book_id, author_id)
  	SELECT book_id, author_id FROM cte
  	ON CONFLICT (book_id, author_id) DO NOTHING
  	RETURNING book_id, author_id
  )
  SELECT array_agg(book_id), array_agg(author_id) FROM finished
  
	`

	args = []interface{}{pq.Array([]string{entry.Book.Hash}), pq.Array(entry.Authors.Hash)}

	return tx.QueryRowContext(ctx, query, args...).Scan(
		pq.Array(entry.Authors.List),
		pq.Array(entry.Authors.Hash),
	)
}
