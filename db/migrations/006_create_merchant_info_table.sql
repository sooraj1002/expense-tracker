-- Create merchant_info table
CREATE TABLE IF NOT EXISTS merchant_info (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    aliases TEXT[],
    common_category_id UUID REFERENCES categories(id),
    transaction_count INTEGER DEFAULT 0,
    total_spent DECIMAL(12, 2) DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, name)
);

CREATE INDEX idx_merchant_info_user_id ON merchant_info(user_id);
CREATE INDEX idx_merchant_info_name ON merchant_info(name);
