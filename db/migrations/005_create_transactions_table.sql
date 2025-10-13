-- Create transactions table
CREATE TABLE IF NOT EXISTS transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    raw_text TEXT NOT NULL,
    timestamp TIMESTAMP NOT NULL,
    sender_info VARCHAR(255),
    amount DECIMAL(12, 2),
    merchant_name VARCHAR(255),
    account_last4 VARCHAR(4),
    parsed BOOLEAN DEFAULT FALSE,
    processed BOOLEAN DEFAULT FALSE,
    expense_id UUID REFERENCES expenses(id) ON DELETE SET NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_transactions_user_id ON transactions(user_id);
CREATE INDEX idx_transactions_expense_id ON transactions(expense_id);
CREATE INDEX idx_transactions_processed ON transactions(processed);
CREATE INDEX idx_transactions_timestamp ON transactions(timestamp);
