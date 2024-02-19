package handlers

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/3WDeveloper-GM/library_app/backend/config"
	"github.com/3WDeveloper-GM/library_app/backend/internal"
	books "github.com/3WDeveloper-GM/library_app/backend/internal/Books"
	"github.com/3WDeveloper-GM/library_app/backend/internal/validator"
)

func DeleteEntryHandlerDelete(app *config.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		n, err := app.ReadIDParams(r)
		if err != nil {
			app.ServerErrorResponse(w, r, err)
			return
		}

		var deleteID = &books.DeleteID{
			ID: n,
		}

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		tx, err := app.Models.Create.DB.BeginTx(ctx, nil)
		if err != nil {
			app.ServerErrorResponse(w, r, err)
			return
		}
		defer tx.Rollback()

		err = app.Models.Delete.DeleteBook(ctx, tx, deleteID)
		if err != nil {
			switch {
			case errors.Is(err, books.ErrEditConflict):
				app.EditConflictResponse(w, r)
			default:
				app.ServerErrorResponse(w, r, err)
			}
			return
		}

		if err := tx.Commit(); err != nil {
			app.ServerErrorResponse(w, r, err)
			return
		}

		err = app.WriteJson(w, r, http.StatusOK, config.Envelope{"message": "entry deleted succesfully"}, nil)
		if err != nil {
			app.ServerErrorResponse(w, r, err)
		}
	}
}

func InsertEntryHandlerPost(app *config.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		read := &books.ReadEntry{}

		var input struct {
			Book struct {
				ID        int64    `json:"id"`
				Title     string   `json:"title"`
				Publisher string   `json:"publisher"`
				Year      int32    `json:"year"`
				PageCount int32    `json:"page_count"`
				Genres    []string `json:"genres"`
			} `json:"book"`
			Authors struct {
				List []string `json:"names"`
			} `json:"authors"`
		}

		err := app.ReadJSON(w, r, &input)
		if err != nil {
			app.ServerErrorResponse(w, r, err)
			return
		}

		book := &books.Book{
			Title:     input.Book.Title,
			Publisher: input.Book.Publisher,
			Year:      input.Book.Year,
			PageCount: input.Book.PageCount,
			Genres:    input.Book.Genres,
		}

		authors := &books.Authors{
			List: input.Authors.List,
		}

		inputEntry := &books.CreateBookEntry{
			Book:    book,
			Authors: authors,
		}

		read.Authors = make([]books.ReadAuthor, len(authors.List))

		v := validator.NewValidator()
		if !inputEntry.ValidateEntry(v) {
			app.FailedValidationResponse(w, r, v.Errors)
			return
		}

		//sanitize author names
		for index, value := range inputEntry.Authors.List {
			inputEntry.Authors.List[index] = app.ToUpper(value)
		}

		books.HashEntries(inputEntry.Book, inputEntry.Authors)

		app.Log.Info().Interface("entry", inputEntry).Send()

		ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
		defer cancel()

		tx, err := app.Models.Create.DB.BeginTx(ctx, nil)
		if err != nil {
			app.ServerErrorResponse(w, r, err)
			return
		}

		defer tx.Rollback()

		err = app.Models.Create.Insert(ctx, tx, inputEntry, read)
		if err != nil {
			app.ServerErrorResponse(w, r, err)
			return
		}

		if err := tx.Commit(); err != nil {
			app.ServerErrorResponse(w, r, err)
			return
		}

		read.Convert()

		err = app.WriteJson(w, r, http.StatusCreated, config.Envelope{
			"message": "entry created!",
			"entry":   read,
		},
			nil)
		if err != nil {
			app.ServerErrorResponse(w, r, err)
		}

	}
}

func FetchAuthorEntryHandlerGet(app *config.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		n, err := app.ReadIDParams(r)
		if err != nil {
			app.NotFoundResponse(w, r)
			return
		}

		readAuthorEntry := &books.ReadAuthorEntry{
			BookList: books.ReadBookList{},
			Author: books.ReadAuthor{
				ID: n,
			},
			BookDisplay: []books.ReadBook{},
		}

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		err = app.Models.Read.AuthorGet(ctx, nil, readAuthorEntry)
		if err != nil {
			app.ServerErrorResponse(w, r, err)
			return
		}

		readAuthorEntry.BookDisplay = make([]books.ReadBook, len(readAuthorEntry.BookList.Hash))
		readAuthorEntry.Convert()

		err = app.WriteJson(w, r, http.StatusOK, config.Envelope{
			"message": "found entry!",
			"entry":   readAuthorEntry,
		}, nil)

		if err != nil {
			app.ServerErrorResponse(w, r, err)
		}
	}
}

func FetchEntryHandlerGet(app *config.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		n, err := app.ReadIDParams(r)
		if err != nil {
			app.NotFoundResponse(w, r)
			return
		}

		readEntry := &books.ReadEntry{
			Book: books.ReadBook{
				ID: n,
			},
			Authors: []books.ReadAuthor{},
		}

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		err = app.Models.Read.Get(ctx, nil, readEntry)
		if err != nil {
			app.ServerErrorResponse(w, r, err)
			return
		}

		readEntry.Authors = make([]books.ReadAuthor, len(readEntry.List.Name))
		readEntry.Convert()

		err = app.WriteJson(w, r, http.StatusOK, config.Envelope{
			"entry":   readEntry,
			"message": "entry found",
		}, nil)
		if err != nil {
			app.ServerErrorResponse(w, r, err)
		}
	}
}

func UpdateEntriesHandlerPatch(app *config.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		n, err := app.ReadIDParams(r)
		if err != nil {
			app.NotFoundResponse(w, r)
			return
		}

		ctx1, cancel := context.WithTimeout(context.Background(), 1200*time.Millisecond)
		defer cancel()

		old_entry := &books.ReadEntry{
			Book: books.ReadBook{
				ID: n,
			},
			Authors: []books.ReadAuthor{},
		}
		err = app.Models.Read.Get(ctx1, nil, old_entry)
		if err != nil {
			app.ServerErrorResponse(w, r, err)
			return
		}

		var input struct {
			Book struct {
				Title     *string  `json:"title"`
				Publisher *string  `json:"publisher"`
				Year      *int32   `json:"year"`
				PageCount *int32   `json:"page_count"`
				Genres    []string `json:"genres"`
			} `json:"book"`
			Authors struct {
				Name []string `json:"names"`
			} `json:"authors"`
		}

		err = app.ReadJSON(w, r, &input)
		if err != nil {
			app.ServerErrorResponse(w, r, err)
			return
		}

		for index, value := range input.Authors.Name {
			input.Authors.Name[index] = app.ToUpper(value)
		}

		new_entry := &books.UpdateEntry{
			Book: books.UpdateBook{
				ID: &n,
			},
			Author: books.UpdateAuthors{},
		}

		if input.Book.Title == nil {
			new_entry.Book.Title = &old_entry.Book.Title
		} else {
			new_entry.Book.Title = input.Book.Title
		}
		if input.Book.Publisher == nil {
			new_entry.Book.Publisher = &old_entry.Book.Publisher
		} else {
			new_entry.Book.Publisher = input.Book.Publisher
		}
		if input.Book.Year == nil {
			new_entry.Book.Year = &old_entry.Book.Year
		} else {
			new_entry.Book.Year = input.Book.Year
		}
		if input.Book.PageCount == nil {
			new_entry.Book.PageCount = &old_entry.Book.PageCount
		} else {
			new_entry.Book.PageCount = input.Book.PageCount
		}
		if input.Book.Genres == nil {
			new_entry.Book.Genres = old_entry.Book.Genres
		} else {
			new_entry.Book.Genres = input.Book.Genres
		}
		if input.Authors.Name == nil {
			new_entry.Author.Name = old_entry.List.Name
		} else {
			new_entry.Author.Name = input.Authors.Name
		}

		v := validator.NewValidator()
		if !new_entry.ValidateEntry(v) {
			app.FailedValidationResponse(w, r, v.Errors)
			return
		}

		_, exclusiveOld, exclusiveNew := internal.DiffArrays(old_entry.List.Name, new_entry.Author.Name)

		hashOld, hashNew := internal.HashArrays(exclusiveOld, exclusiveNew)

		new_entry.HashEntries()

		ctx, cancel2 := context.WithTimeout(context.Background(), 4*time.Second)
		defer cancel2()

		tx, err := app.Models.Update.DB.BeginTx(ctx, nil)
		if err != nil {
			app.ServerErrorResponse(w, r, err)
			return
		}

		defer tx.Rollback()

		err = app.Models.Update.Update(ctx, tx, new_entry, old_entry, hashOld, exclusiveNew, hashNew)
		if err != nil {
			app.ServerErrorResponse(w, r, err)
			return
		}

		if err := tx.Commit(); err != nil {
			app.ServerErrorResponse(w, r, err)
			return
		}

		err = app.WriteJson(w, r, http.StatusOK, config.Envelope{
			"entry":   old_entry,
			"message": "succesfully updated",
		}, nil)
		if err != nil {
			app.ServerErrorResponse(w, r, err)
		}

	}
}
