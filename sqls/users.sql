CREATE TABLE IF NOT EXISTS users (
                                     id SERIAL PRIMARY KEY,
                                     login TEXT UNIQUE NOT NULL,
                                     password TEXT NOT NULL
);