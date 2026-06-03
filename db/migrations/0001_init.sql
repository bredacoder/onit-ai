-- +goose Up

CREATE TABLE users (
    id         text        PRIMARY KEY,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE tasks (
    id               text        PRIMARY KEY,
    user_id          text        NOT NULL REFERENCES users(id),
    service_type     text,
    location         text,
    time_window      text,
    budget_cap       bigint,
    task_constraints text,
    state            text        NOT NULL,
    attributes       jsonb       NOT NULL DEFAULT '{}'::jsonb,
    created_at       timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX tasks_user_id_idx ON tasks(user_id);

CREATE TABLE negotiations (
    id          text        PRIMARY KEY,
    user_id     text        NOT NULL REFERENCES users(id),
    task_id     text        NOT NULL REFERENCES tasks(id),
    provider_id text        NOT NULL,
    state       text        NOT NULL,
    created_at  timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE providers (
    id         text              PRIMARY KEY,
    user_id    text              NOT NULL REFERENCES users(id),
    name       text,
    phone      text,
    rating     double precision,
    hours      text,
    source     text,
    confidence double precision,
    evidence   text,
    created_at timestamptz       NOT NULL DEFAULT now()
);

CREATE TABLE messages (
    id             text        PRIMARY KEY,
    user_id        text        NOT NULL REFERENCES users(id),
    negotiation_id text        NOT NULL REFERENCES negotiations(id),
    direction      text,
    body           text,
    sent_at        timestamptz,
    created_at     timestamptz NOT NULL DEFAULT now()
);

-- +goose Down

DROP TABLE messages;
DROP TABLE providers;
DROP TABLE negotiations;
DROP TABLE tasks;
DROP TABLE users;
