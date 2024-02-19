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

type UpdateBook struct {
	ID        *int64   `json:"id"`
	Hash      *string  `json:"hash"`
	Title     *string  `json:"title"`
	Publisher *string  `json:"publisher"`
	Year      *int32   `json:"year"`
	PageCount *int32   `json:"page_count"`
	Genres    []string `json:"genres"`
}

type UpdateAuthors struct {
	Name []string `json:"names"`
	Hash []string `json:"hash"`
}

type UpdateEntry struct {
	Book   UpdateBook    `json:"book"`
	Author UpdateAuthors `json:"authors"`
}

func (u *UpdateEntry) HashEntries() {
	//Hash Book ID
	hash := sha1.New()
	hash.Write([]byte(*u.Book.Title))
	hashInBytes := hash.Sum(nil)

	Hashed := hex.EncodeToString(hashInBytes)

	u.Book.Hash = &Hashed

	//Hash author entries
	hashed := make([]string, len(u.Author.Name))

	for index, value := range u.Author.Name {
		hash := sha1.New()
		hash.Write([]byte(value))
		hashInBytes := hash.Sum(nil)

		hashed[index] = hex.EncodeToString(hashInBytes)
	}

	u.Author.Hash = hashed
}

type UpdateEntryModel struct {
	DB *sql.DB
}

func (u *UpdateEntryModel) Update(ctx context.Context, tx *sql.Tx, entry *UpdateEntry, read *ReadEntry, hashOld, exclusiveNew, hashNew []string) error {
	query := `
		UPDATE books
		SET title = $1, publisher = $2, year = $3, page_count = $4, genres = $5, version = version + 1
		WHERE id = $6
		RETURNING book_id, title, publisher, year, page_count, genres, version
	`

	args := []interface{}{
		entry.Book.Title,
		entry.Book.Publisher,
		entry.Book.Year,
		entry.Book.PageCount,
		pq.Array(entry.Book.Genres),
		entry.Book.ID,
	}

	err := tx.QueryRowContext(ctx, query, args...).Scan(
		&read.Book.Hash,
		&read.Book.Title,
		&read.Book.Publisher,
		&read.Book.Year,
		&read.Book.PageCount,
		pq.Array(&read.Book.Genres),
		&read.Book.Version,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}

	if len(exclusiveNew) != 0 {
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

		err = tx.QueryRowContext(ctx, query,
			pq.Array(exclusiveNew),
			pq.Array(hashNew)).
			Scan(
				pq.Array(read.List.ID),
				pq.Array(read.List.Name),
				pq.Array(read.List.Identifier),
				pq.Array(read.List.Books_authored),
			)

		if err != nil {
			return errors.New(err.Error() + " 2")
		}
	}

	if len(hashOld) != 0 {
		query = `
		UPDATE authors
		SET books_authored = books_authored - 1
		WHERE authors.author_id = ANY(
			SELECT * FROM UNNEST($1::TEXT[]) as author_id
		)
		RETURNING books_authored
	`

		err = tx.QueryRowContext(ctx, query, pq.Array(hashOld)).Scan(
			pq.Array(read.List.Books_authored),
		)

		if err != nil {
			switch {
			case errors.Is(err, sql.ErrNoRows):
				return ErrEditConflict
			default:
				return err
			}
		}
	}

	query = `
		DELETE FROM book_author_link
		WHERE book_id = $1
	`

	args2 := []interface{}{entry.Book.Hash}
	result, err := tx.ExecContext(ctx, query, args2...)

	if err != nil {
		return errors.New(err.Error() + " 4")
	}

	rows, err := result.RowsAffected()

	if err != nil {
		return errors.New(err.Error() + " 5")
	}

	if rows == 0 {
		return errors.New(sql.ErrNoRows.Error() + " 6")
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

	args3 := []interface{}{pq.Array([]string{*entry.Book.Hash}), pq.Array(entry.Author.Hash)}

	return tx.QueryRowContext(ctx, query, args3...).Scan(
		pq.Array(read.List.Name),
		pq.Array(read.List.Identifier),
	)

}

func (b *UpdateEntry) ValidateEntry(v *validator.Validator) bool {

	//book validation
	var section = "title"
	var mustbeProvidedMsg = "%s field must be provided"
	var maxTitleBytes = 300

	v.Check(*b.Book.Title != "", section, fmt.Sprintf(
		mustbeProvidedMsg, section,
	))
	v.Check(len(*b.Book.Title) <= maxTitleBytes, section, fmt.Sprintf(
		"%s field must have a length no longer than %d bytes", section, maxTitleBytes,
	))

	section = "publisher"
	var maxPublisherByteAmount = 300
	v.Check(*b.Book.Publisher != "", section, fmt.Sprintf(
		mustbeProvidedMsg, section,
	))
	v.Check(len(*b.Book.Publisher) <= maxPublisherByteAmount, section, fmt.Sprintf(
		"%s field must have less than %d bytes", section, maxPublisherByteAmount,
	))

	section = "year"
	var PresentDate = int32(time.Now().Year())
	var minimumYear int32 = 1900

	v.Check(*b.Book.Year != 0, section, fmt.Sprintf(
		mustbeProvidedMsg, section,
	))
	v.Check(*b.Book.Year <= PresentDate, section, fmt.Sprintf(
		"%s field must not be set in the future", section,
	))
	v.Check(*b.Book.Year >= minimumYear, section, fmt.Sprintf(
		"%s field must not be set before %d", section, minimumYear,
	))

	section = "page_count"
	var maxPageCount int32 = 10000
	var minPageCount int32 = 1

	v.Check(*b.Book.PageCount != 0, section, fmt.Sprintf(
		mustbeProvidedMsg, section,
	))
	v.Check(*b.Book.PageCount <= maxPageCount, section, fmt.Sprintf(
		"%s field must not be more than %d pages", section, maxPageCount,
	))
	v.Check(*b.Book.PageCount >= minPageCount, section, fmt.Sprintf(
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

	v.Check(b.Author.Name != nil, section, fmt.Sprintf(
		mustbeProvidedMsg, section,
	))
	v.Check(len(b.Author.Name) <= maxAuthorCount, section, fmt.Sprintf(
		"%s field must not contain more than %d authors", section, maxAuthorCount,
	))
	v.Check(len(b.Author.Name) >= minAuthorCount, section, fmt.Sprintf(
		"%s field must contain at least %d authors", section, minAuthorCount,
	))

	if !v.Valid() {
		return false
	}

	section = "author_items"
	var maxItemByteAmount = 100
	for _, value := range b.Author.Name {
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
