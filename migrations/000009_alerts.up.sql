CREATE TYPE alert_direction AS ENUM ('above', 'below');

CREATE TABLE IF NOT EXISTS alerts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id TEXT NOT NULL,
    stock_id UUID NOT NULL,
    threshold_price DECIMAL(20,2) NOT NULL,
    direction alert_direction NOT NULL,
    triggered_at TIMESTAMP NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_alerts_user
        FOREIGN KEY (user_id)
        REFERENCES users(firebase_id)
        ON DELETE CASCADE,

    CONSTRAINT fk_alerts_stock
        FOREIGN KEY (stock_id)
        REFERENCES stocks(id)
        ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_alerts_user_created_at ON alerts(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_alerts_stock ON alerts(stock_id);
