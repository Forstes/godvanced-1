CREATE TABLE IF NOT EXISTS users (
    id bigserial PRIMARY KEY,
    role smallint NOT NULL DEFAULT 0, /* enum { user = 0, admin = 1 } */
    email text NOT NULL,
    name text NOT NULL,
    password text NOT NULL,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS activities (
    id bigserial PRIMARY KEY,
    user_id bigint NOT NULL,
    name text NOT NULL,
    answer_points smallint[] NOT NULL DEFAULT '{}',
    answers_sum smallint NOT NULL DEFAULT 0,
    status smallint NOT NULL DEFAULT 2 /* enum { ikigai = 0, tool = 1, trash = 2 } */
);

CREATE TABLE IF NOT EXISTS questions (
    id serial PRIMARY KEY,
    title text NOT NULL,
    video_url text NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS answers (
    id serial PRIMARY KEY,
    question_id int NOT NULL,
    title text NOT NULL,
    points smallint NOT NULL
);