ALTER TABLE book_inventory 
ADD CONSTRAINT unique_library_book_shelf UNIQUE (library_id, book_id, bookshelf_id);