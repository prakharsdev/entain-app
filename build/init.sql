CREATE TABLE IF NOT EXISTS users (
    id BIGINT PRIMARY KEY,
    balance NUMERIC(12, 2) NOT NULL DEFAULT 0.00
);

CREATE TABLE IF NOT EXISTS transactions (
    transaction_id TEXT PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id),
    amount NUMERIC(12, 2) NOT NULL,
    state TEXT NOT NULL CHECK (state IN ('win', 'lose')),
    source_type TEXT NOT NULL CHECK (source_type IN ('game', 'server', 'payment')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Seed users with zero balance
INSERT INTO users (id, balance) VALUES
(1, 0.00),
(2, 0.00),
(3, 0.00)
ON CONFLICT (id) DO NOTHING;
