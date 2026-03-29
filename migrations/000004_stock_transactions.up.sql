CREATE TYPE transaction_type AS ENUM ('bought', 'sold');

CREATE TABLE IF NOT EXISTS stock_transactions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id TEXT NOT NULL,
    stock_id UUID NOT NULL,
    price DECIMAL(20,2) NOT NULL,
    quantity DECIMAL(20,2) NOT NULL,
    transaction_date TIMESTAMP NOT NULL DEFAULT NOW(),
    transaction_type transaction_type NOT NULL,

    CONSTRAINT fk_stock_transactions_user
        FOREIGN KEY (user_id)
        REFERENCES users(firebase_id)
        ON DELETE CASCADE,

    CONSTRAINT fk_stock_transactions_stock
        FOREIGN KEY (stock_id)
        REFERENCES stocks(id)
        ON DELETE CASCADE
);
