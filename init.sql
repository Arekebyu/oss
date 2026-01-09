CREATE TABLE IF NOT EXISTS pages (
    id SERIAL PRIMARY KEY,
    url TEXT UNIQUE NOT NULL,
    title TEXT,
    crawled_at TIMESTAMP with TIME ZONE
);

CREATE TABLE IF NOT EXISTS sections (
    id SERIAL PRIMARY KEY,
    page_id INTEGER REFERENCES pages(id) ON DELETE CASCADE,
    section_type TEXT,  
    content TEXT,
    language TEXT,      
    sort_order INTEGER  
);