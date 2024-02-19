package books

import (
	"context"
	"database/sql"

	"github.com/lib/pq"
)

type ReadBook struct {
	ID        int64    `json:"id"`
	Hash      string   `json:"hash"`
	Title     string   `json:"title"`
	Publisher string   `json:"publisher"`
	Year      int32    `json:"year"`
	PageCount int32    `json:"page_count"`
	Genres    []string `json:"genres,omitempty"`
	Version   int32    `json:"version,omitempty"`
}

type ReadAuthor struct {
	ID             int64  `json:"id"`
	Name           string `json:"name"`
	Identifier     string `json:"identifier"`
	Books_authored int32  `json:"books_authored"`
}

type ReadAuthorList struct {
	ID             []int64
	Name           []string
	Identifier     []string
	Books_authored []int32
}

type ReadEntry struct {
	Book    ReadBook       `json:"Book"`
	Authors []ReadAuthor   `json:"authors"`
	List    ReadAuthorList `json:"-"`
}

func (r *ReadEntry) Convert() {
	for index := range r.List.Name {
		r.Authors[index] = ReadAuthor{
			ID:             r.List.ID[index],
			Name:           r.List.Name[index],
			Identifier:     r.List.Identifier[index],
			Books_authored: r.List.Books_authored[index],
		}
	}
}

type ReadEntryModel struct {
	DB *sql.DB
}

func (r *ReadEntryModel) Get(ctx context.Context, tx *sql.Tx, read *ReadEntry) error {
	query := `
	SELECT b.book_id,b.title,b.publisher,b.year,b.page_count,b.genres, array_agg(a.name), array_agg(a.author_id), array_agg(a.id), array_agg(a.books_authored) AS authors
	FROM books b
	JOIN book_author_link bal ON b.book_id = bal.book_id
	JOIN authors a ON bal.author_id = a.author_id
	WHERE b.id = $1
	GROUP BY b.id;
	`

	if tx != nil {
		return tx.QueryRowContext(ctx, query, read.Book.ID).Scan(
			&read.Book.Hash,
			&read.Book.Title,
			&read.Book.Publisher,
			&read.Book.Year,
			&read.Book.PageCount,
			pq.Array(&read.Book.Genres),
			pq.Array(&read.List.Name),
			pq.Array(&read.List.Identifier),
			pq.Array(&read.List.ID),
			pq.Array(&read.List.Books_authored),
		)
	} else {
		return r.DB.QueryRowContext(ctx, query, read.Book.ID).Scan(
			&read.Book.Hash,
			&read.Book.Title,
			&read.Book.Publisher,
			&read.Book.Year,
			&read.Book.PageCount,
			pq.Array(&read.Book.Genres),
			pq.Array(&read.List.Name),
			pq.Array(&read.List.Identifier),
			pq.Array(&read.List.ID),
			pq.Array(&read.List.Books_authored),
		)
	}
}

type ReadAuthorEntry struct {
	BookList    ReadBookList `json:"-"`
	Author      ReadAuthor   `json:"authors"`
	BookDisplay []ReadBook   `json:"books"`
}

type ReadBookList struct {
	ID        []int64  `json:"id"`
	Hash      []string `json:"hash"`
	Title     []string `json:"title"`
	Publisher []string `json:"publisher"`
	Year      []int32  `json:"year"`
	PageCount []int32  `json:"page_count"`
}

func (r *ReadAuthorEntry) Convert() {
	for index := range r.BookList.Title {
		r.BookDisplay[index] = ReadBook{
			ID:        r.BookList.ID[index],
			Title:     r.BookList.Title[index],
			Hash:      r.BookList.Hash[index],
			Publisher: r.BookList.Publisher[index],
			Year:      r.BookList.Year[index],
			PageCount: r.BookList.PageCount[index],
		}
	}
}

func (r *ReadEntryModel) AuthorGet(ctx context.Context, tx *sql.Tx, read *ReadAuthorEntry) error {
	query := `
	SELECT
		a.author_id,
		a.name,
		a.books_authored,
		ARRAY_AGG(b.id) AS ids,
		ARRAY_AGG(b.book_id) AS book_ids,
		ARRAY_AGG(b.title) AS titles,
		ARRAY_AGG(b.publisher) AS publishers,
		ARRAY_AGG(b.year) AS years,
		ARRAY_AGG(b.page_count) AS page_counts
	FROM
		authors a
	JOIN
		book_author_link bal ON a.author_id = bal.author_id
	JOIN
		books b ON bal.book_id = b.book_id
	WHERE
		a.id = $1
	GROUP BY
		a.author_id,
		a.name,
		a.books_authored;
	`

	if tx != nil {
		return tx.QueryRowContext(ctx, query, read.Author.ID).Scan(
			&read.Author.Identifier,
			&read.Author.Name,
			&read.Author.Books_authored,
			pq.Array(&read.BookList.ID),
			pq.Array(&read.BookList.Hash),
			pq.Array(&read.BookList.Title),
			pq.Array(&read.BookList.Publisher),
			pq.Array(&read.BookList.Year),
			pq.Array(&read.BookList.PageCount),
		)
	} else {
		return r.DB.QueryRowContext(ctx, query, read.Author.ID).Scan(
			&read.Author.Identifier,
			&read.Author.Name,
			&read.Author.Books_authored,
			pq.Array(&read.BookList.ID),
			pq.Array(&read.BookList.Hash),
			pq.Array(&read.BookList.Title),
			pq.Array(&read.BookList.Publisher),
			pq.Array(&read.BookList.Year),
			pq.Array(&read.BookList.PageCount),
		)
	}
}
