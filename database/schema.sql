

-- Table schemas 

CREATE TABLE IF NOT EXISTS providers(
    provider_id SERIAL PRIMARY KEY, 
    last_name TEXT,
    first_name TEXT, 
    phone_number TEXT,
    email TEXT
);


CREATE TABLE IF NOT EXISTS courses(
    course_id SERIAL PRIMARY KEY,
    course_name TEXT,
    course_description TEXT,
    provider_id INTEGER NOT NULL REFERENCES providers(provider_id)
);

CREATE TABLE IF NOT EXISTS chapters(
    chapter_id SERIAL PRIMARY KEY,
    chapter_name TEXT, 
    chapter_num INTEGER,
    parent_chapter_id INTEGER REFERENCES chapters(chapter_id), -- NULL signals its the parent chapter of the current one
    course_id INTEGER NOT NULL REFERENCES courses(course_id),
    provider_id INTEGER NOT NULL REFERENCES providers(provider_id)
);

CREATE TABLE IF NOT EXISTS lessons(
    lesson_id SERIAL PRIMARY KEY, 
    lesson_name TEXT, 
    chapter_id INTEGER NOT NULL REFERENCES chapters(chapter_id),
    provider_id INTEGER NOT NULL REFERENCES providers(provider_id)
);

CREATE TABLE IF NOT EXISTS video_metadata(
    video_id SERIAL PRIMARY KEY,
    title TEXT,
    video_description TEXT,
    video_file_id INTEGER,
    subtitle_text TEXT,
    subtitle_tsv tsvector GENERATED ALWAYS AS (to_tsvector('simple', coalesce(subtitle_text, ''))) STORED,
    lesson_id INTEGER NOT NULL REFERENCES lessons(lesson_id),
    provider_id INTEGER NOT NULL REFERENCES providers(provider_id)
);

CREATE INDEX IF NOT EXISTS idx_video_metadata_subtitle_tsv
    ON video_metadata USING GIN (subtitle_tsv);

