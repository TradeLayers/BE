CREATE TABLE IF NOT EXISTS watchlist (
    user_id TEXT NOT NULL,
    stock_id UUID NOT NULL,
    added_at TIMESTAMP NOT NULL DEFAULT NOW(),

    PRIMARY KEY (user_id, stock_id),

    CONSTRAINT fk_watchlist_user
        FOREIGN KEY (user_id)
        REFERENCES users(firebase_id)
        ON DELETE CASCADE,

    CONSTRAINT fk_watchlist_stock
        FOREIGN KEY (stock_id)
        REFERENCES stocks(id)
        ON DELETE CASCADE
);
