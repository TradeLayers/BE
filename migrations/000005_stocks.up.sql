CREATE TABLE IF NOT EXISTS stocks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    stock_name VARCHAR(128) NOT NULL,
    symbol VARCHAR(32) NOT NULL UNIQUE
);