CREATE TABLE IF NOT EXISTS processed_webhooks (
    webhook_id TEXT PRIMARY KEY NOT NULL,
    event_name TEXT NOT NULL,
    processed_at TIMESTAMP NOT NULL
);
