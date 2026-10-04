CREATE TABLE IF NOT EXISTS tasks (
                                     id SERIAL PRIMARY KEY,
                                     userID INTEGER NOT NULL,
                                     taskName TEXT NOT NULL,
                                     taskDesc TEXT,
                                     isCompleted BOOLEAN NOT NULL DEFAULT FALSE,
                                     createdAt TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
                                     updatedAt TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
