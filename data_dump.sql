--
-- PostgreSQL database dump
--

\restrict AgCbRfel0MkmuSqm9j6a0nuPZvYI7vhpPMIj6FVauQNJORqWNd8YqOdYIlg3g3I

-- Dumped from database version 15.4
-- Dumped by pg_dump version 16.14

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- Data for Name: genres; Type: TABLE DATA; Schema: public; Owner: postgres
--

INSERT INTO public.genres (id, name) VALUES (1, 'Uncategorized');
INSERT INTO public.genres (id, name) VALUES (2, 'Fiction');
INSERT INTO public.genres (id, name) VALUES (3, 'Non-Fiction');
INSERT INTO public.genres (id, name) VALUES (4, 'Sci-Fi');
INSERT INTO public.genres (id, name) VALUES (5, 'Fantasy');
INSERT INTO public.genres (id, name) VALUES (6, 'Mystery');


--
-- Data for Name: books; Type: TABLE DATA; Schema: public; Owner: postgres
--

INSERT INTO public.books (id, title, author, isbn, created_at, genre_id) VALUES (2, 'The Hobbit', 'J.R.R. Tolkien', '9780261103344', '2026-07-19 12:21:37.794213+05', 1);
INSERT INTO public.books (id, title, author, isbn, created_at, genre_id) VALUES (3, '184', 'George Orwell', '9780451524935', '2026-07-19 12:21:37.794213+05', 1);
INSERT INTO public.books (id, title, author, isbn, created_at, genre_id) VALUES (4, 'Animal Farm', 'George Orwell', '9780451526342', '2026-07-19 12:21:37.794213+05', 1);
INSERT INTO public.books (id, title, author, isbn, created_at, genre_id) VALUES (6, 'Foundation', 'Isaac Asimov', '9780553293357', '2026-07-19 12:21:37.794213+05', 1);
INSERT INTO public.books (id, title, author, isbn, created_at, genre_id) VALUES (7, 'Neuromancer', 'William Gibson', '9780441569595', '2026-07-19 12:21:37.794213+05', 1);
INSERT INTO public.books (id, title, author, isbn, created_at, genre_id) VALUES (8, 'Snow Crash', 'Neal Stephenson', '9780553380958', '2026-07-19 12:21:37.794213+05', 1);
INSERT INTO public.books (id, title, author, isbn, created_at, genre_id) VALUES (9, 'Brave New World', 'Aldous Huxley', '9780060850524', '2026-07-19 12:21:37.794213+05', 1);
INSERT INTO public.books (id, title, author, isbn, created_at, genre_id) VALUES (10, 'Fahrenheit 451', 'Ray Bradbury', '9781451673319', '2026-07-19 12:21:37.794213+05', 1);
INSERT INTO public.books (id, title, author, isbn, created_at, genre_id) VALUES (11, 'The Catcher in the Rye', 'J.D. Salinger', '9780316769174', '2026-07-19 12:21:37.794213+05', 1);
INSERT INTO public.books (id, title, author, isbn, created_at, genre_id) VALUES (12, 'The Great Gatsby', 'F. Scott Fitzgerald', '9780743273565', '2026-07-19 12:21:37.794213+05', 1);
INSERT INTO public.books (id, title, author, isbn, created_at, genre_id) VALUES (13, 'To Kill a Mockingbird', 'Harper Lee', '9780060935467', '2026-07-19 12:21:37.794213+05', 1);
INSERT INTO public.books (id, title, author, isbn, created_at, genre_id) VALUES (14, 'Moby Dick', 'Herman Melville', '9780142000083', '2026-07-19 12:21:37.794213+05', 1);
INSERT INTO public.books (id, title, author, isbn, created_at, genre_id) VALUES (15, 'Crime and Punishment', 'Fyodor Dostoevsky', '9780140449136', '2026-07-19 12:21:37.794213+05', 1);
INSERT INTO public.books (id, title, author, isbn, created_at, genre_id) VALUES (16, 'The Brothers Karamazov', 'Fyodor Dostoevsky', '9780374528379', '2026-07-19 12:21:37.794213+05', 1);
INSERT INTO public.books (id, title, author, isbn, created_at, genre_id) VALUES (17, 'Pride and Prejudice', 'Jane Austen', '9780141439518', '2026-07-19 12:21:37.794213+05', 1);
INSERT INTO public.books (id, title, author, isbn, created_at, genre_id) VALUES (18, 'Wuthering Heights', 'Emily Brontë', '9780141439556', '2026-07-19 12:21:37.794213+05', 1);
INSERT INTO public.books (id, title, author, isbn, created_at, genre_id) VALUES (21, 'Clean Architecture', 'Robert C. Martin', '9780134494166', '2026-07-19 13:34:47.366894+05', 1);
INSERT INTO public.books (id, title, author, isbn, created_at, genre_id) VALUES (22, 'The Go Programming Language', 'Alan Donovan', '9780134190440', '2026-07-19 19:43:53.379767+05', 1);
INSERT INTO public.books (id, title, author, isbn, created_at, genre_id) VALUES (19, 'My Book', 'Baigozha Bakdaulet', '9780134190441', '2026-07-19 12:21:37.794213+05', 1);
INSERT INTO public.books (id, title, author, isbn, created_at, genre_id) VALUES (23, 'Designing Data-Intensive Applications', 'Martin Kleppmann', '9781449373320', '2026-07-22 15:50:38.31724+05', 2);
INSERT INTO public.books (id, title, author, isbn, created_at, genre_id) VALUES (1, 'Designing Data-Intensive Applications (2nd Edition)', 'Martin Kleppmann', '9781449373329', '2026-07-19 12:21:37.794213+05', 2);


--
-- Data for Name: libraries; Type: TABLE DATA; Schema: public; Owner: postgres
--

INSERT INTO public.libraries (id, name, address, created_at) VALUES (2, 'Library 1', '100 Innovation Way, Tech City', '2026-07-22 15:48:27.966338+05');
INSERT INTO public.libraries (id, name, address, created_at) VALUES (1, 'Library 2', '200 Innovation Way, Tech City', '2026-07-21 12:37:19.508583+05');


--
-- Data for Name: bookshelves; Type: TABLE DATA; Schema: public; Owner: postgres
--

INSERT INTO public.bookshelves (id, library_id, code, capacity, empty_space, created_at) VALUES (3, 1, 'B-101', 50, 41, '2026-07-27 12:36:34.126201+05');
INSERT INTO public.bookshelves (id, library_id, code, capacity, empty_space, created_at) VALUES (1, 1, 'A-101', 50, 40, '2026-07-26 14:10:30.629481+05');


--
-- Data for Name: book_inventory; Type: TABLE DATA; Schema: public; Owner: postgres
--

INSERT INTO public.book_inventory (library_id, book_id, available_copies, bookshelf_id) VALUES (1, 2, 9, 3);
INSERT INTO public.book_inventory (library_id, book_id, available_copies, bookshelf_id) VALUES (1, 2, 10, 1);


--
-- Data for Name: members; Type: TABLE DATA; Schema: public; Owner: postgres
--

INSERT INTO public.members (id, email, joined_at, password, role) VALUES (15, 'admin', '2026-07-27 17:39:42.973562+05', '$2a$10$ApnKFfC5rihWh1yCfRuhO.pQFoO1UdecV4w3AGiDVRXRxIF2/icna', 'admin');
INSERT INTO public.members (id, email, joined_at, password, role) VALUES (16, 'client1', '2026-07-28 09:48:46.739074+05', '$2a$10$YzIRmisNLhmYQSE06IQWd.kIkiCZz8fS7nvE6TIh6tD.Xy6plfDgO', 'client');
INSERT INTO public.members (id, email, joined_at, password, role) VALUES (18, 'client2', '2026-07-28 09:54:54.446179+05', '$2a$10$9L2Z3S4pibGnTM50LNln8.8xZmBjp8NxzzBvd14R4nsgNgJVATiLS', 'client');
INSERT INTO public.members (id, email, joined_at, password, role) VALUES (28, 'employee1@gmail.com', '2026-07-28 14:24:40.203131+05', '$2a$10$n1oXvLHlNPcMubM0kUvJV.HqEQ/.3lEovZOyOqe0oHP16WGKsSHGu', 'employee');
INSERT INTO public.members (id, email, joined_at, password, role) VALUES (30, 'employee2@gmail.com', '2026-07-28 15:12:16.551119+05', '$2a$10$eRiEBx6jbpo6GtFQ324LPOsiXr5GGt4fxn6wqXAB4KTkG7ceTy4Uq', 'employee');
INSERT INTO public.members (id, email, joined_at, password, role) VALUES (32, 'client3', '2026-07-28 15:59:17.59678+05', '$2a$10$H8VKFsKzfL/m/ej44OGrSuE7cCr3/SH678U76.fUvH1G5Mq.32Jxe', 'client');


--
-- Data for Name: employees; Type: TABLE DATA; Schema: public; Owner: postgres
--

INSERT INTO public.employees (id, member_id, "position", salary, library_id) VALUES (10, 28, 'Senior Librarian', 45000.00, 1);
INSERT INTO public.employees (id, member_id, "position", salary, library_id) VALUES (11, 30, 'Senior Librarian', 45000.00, 2);


--
-- Data for Name: library_employees; Type: TABLE DATA; Schema: public; Owner: postgres
--

INSERT INTO public.library_employees (library_id, member_id) VALUES (2, 30);


--
-- Data for Name: loans; Type: TABLE DATA; Schema: public; Owner: postgres
--

INSERT INTO public.loans (id, book_id, member_id, borrowed_at, returned_at, borrowed_library_id, returned_library_id) VALUES (2, 2, 16, '2026-07-28 10:35:31.498519+05', NULL, 1, NULL);
INSERT INTO public.loans (id, book_id, member_id, borrowed_at, returned_at, borrowed_library_id, returned_library_id) VALUES (3, 2, 32, '2026-07-28 16:27:59.356382+05', '2026-07-28 17:53:16.750836+05', 1, 1);


--
-- Data for Name: returned_books; Type: TABLE DATA; Schema: public; Owner: postgres
--



--
-- Data for Name: schema_migrations; Type: TABLE DATA; Schema: public; Owner: postgres
--

INSERT INTO public.schema_migrations (version, dirty) VALUES (13, false);


--
-- Name: books_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.books_id_seq', 23, true);


--
-- Name: bookshelves_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.bookshelves_id_seq', 3, true);


--
-- Name: employees_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.employees_id_seq', 11, true);


--
-- Name: genres_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.genres_id_seq', 1, false);


--
-- Name: libraries_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.libraries_id_seq', 4, true);


--
-- Name: loans_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.loans_id_seq', 3, true);


--
-- Name: members_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.members_id_seq', 32, true);


--
-- Name: returned_books_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.returned_books_id_seq', 1, true);


--
-- PostgreSQL database dump complete
--

\unrestrict AgCbRfel0MkmuSqm9j6a0nuPZvYI7vhpPMIj6FVauQNJORqWNd8YqOdYIlg3g3I

