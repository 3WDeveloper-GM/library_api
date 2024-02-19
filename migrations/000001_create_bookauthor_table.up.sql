CREATE TABLE IF NOT EXISTS books (
   id serial PRIMARY KEY,
   book_id text UNIQUE NOT NULL,
   created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
   title text NOT NULL,
   publisher text NOT NULL, 
   year integer NOT NULL, 
   page_count integer NOT NULL, 
   genres text[] NOT NULL,
   version integer NOT NULL DEFAULT 1
);

CREATE TABLE IF NOT EXISTS authors (
   id serial PRIMARY KEY,
   author_id text UNIQUE NOT NULL,
   name text NOT NULL,
   books_authored integer NOT NULL DEFAULT 1,
   created_at timestamp(0) with time zone NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS book_author_link (
   id serial PRIMARY KEY,
   book_id text NOT NULL REFERENCES books(book_id) ON DELETE CASCADE ON UPDATE CASCADE,
   author_id text NOT NULL REFERENCES authors(author_id) ON DELETE CASCADE ON UPDATE CASCADE
);