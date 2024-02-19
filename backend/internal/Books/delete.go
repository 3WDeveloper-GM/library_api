package books

import (
	"context"
	"database/sql"
	"errors"
)

type DeleteID struct {
	ID         int64
	AuthorHash string
}

type DeleteEntryModel struct {
	DB *sql.DB
}

func (del *DeleteEntryModel) DeleteBook(ctx context.Context, tx *sql.Tx, id *DeleteID) error {

	query := `
	SELECT
		bal.author_id
	FROM
		books b
	JOIN
		book_author_link bal ON b.book_id = bal.book_id
	WHERE
		b.id = $1;
	`

	err := tx.QueryRowContext(ctx, query, id.ID).Scan(
		&id.AuthorHash,
	)
	if err != nil {
		return ErrNotFound
	}

	query = `
		UPDATE authors
		SET books_authored = books_authored - 1 
		WHERE authors.author_id = $1
	`

	result, err := tx.ExecContext(ctx, query, id.AuthorHash)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return ErrNotFound
	}

	query = `
		DELETE FROM books
		WHERE id = $1
	`

	result, err = tx.ExecContext(ctx, query, id.ID)
	if err != nil {
		return err
	}

	rows, err = result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return ErrNotFound
	}

	return nil

}
