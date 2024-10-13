-- Initialize the database with a schema to receive the Events from the main file

CREATE TABLE IF NOT EXISTS events (
    event_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    test_id uuid,
    type text,
    domain text,
    payload jsonb,
    body text,
    created_at timestamp with time zone DEFAULT now()
);