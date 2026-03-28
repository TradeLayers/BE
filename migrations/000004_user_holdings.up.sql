CREATE TABLE IF NOT EXISTS users_holdings (
    user_id TEXT NOT NULL,
    stock_id UUID NOT NULL,
    quantity DECIMAL(20,2) NOT NULL,

    PRIMARY KEY (user_id, stock_id),

    CONSTRAINT fk_users_holdings_user
        FOREIGN KEY (user_id)
        REFERENCES users(firebase_id)
        ON DELETE CASCADE,

    CONSTRAINT fk_users_holdings_stock
        FOREIGN KEY (stock_id)
        REFERENCES stocks(id)
        ON DELETE CASCADE
);
