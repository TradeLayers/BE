ALTER TABLE watchlist
ADD COLUMN IF NOT EXISTS threshold_reached BOOLEAN NOT NULL DEFAULT FALSE;

CREATE TABLE IF NOT EXISTS threshold_notifications (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id TEXT NOT NULL,
    stock_id UUID NOT NULL,
    symbol TEXT NOT NULL,
    threshold_price DOUBLE PRECISION NOT NULL,
    trigger_price DOUBLE PRECISION NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    read_at TIMESTAMP NULL,

    CONSTRAINT fk_threshold_notifications_user
        FOREIGN KEY (user_id)
        REFERENCES users(firebase_id)
        ON DELETE CASCADE,

    CONSTRAINT fk_threshold_notifications_stock
        FOREIGN KEY (stock_id)
        REFERENCES stocks(id)
        ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_threshold_notifications_user_read_created
    ON threshold_notifications (user_id, read_at, created_at DESC);