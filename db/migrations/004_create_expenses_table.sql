-- Create expenses table
CREATE TABLE IF NOT EXISTS expenses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    amount DECIMAL(12, 2) NOT NULL,
    category_id UUID NOT NULL REFERENCES categories(id),
    account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    date TIMESTAMP NOT NULL,
    description TEXT,
    source VARCHAR(20) DEFAULT 'manual',
    merchant_id UUID,
    merchant_name VARCHAR(255),
    location_id UUID,
    raw_data TEXT,
    verified BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT check_source CHECK (source IN ('manual', 'auto'))
);

CREATE INDEX idx_expenses_user_id ON expenses(user_id);
CREATE INDEX idx_expenses_category_id ON expenses(category_id);
CREATE INDEX idx_expenses_account_id ON expenses(account_id);
CREATE INDEX idx_expenses_merchant_id ON expenses(merchant_id);
CREATE INDEX idx_expenses_date ON expenses(date);
CREATE INDEX idx_expenses_created_at ON expenses(created_at);
