-- Create merchant_patterns table
CREATE TABLE IF NOT EXISTS merchant_patterns (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    merchant_name VARCHAR(255) NOT NULL,
    category_id UUID NOT NULL REFERENCES categories(id),
    match_type VARCHAR(20) DEFAULT 'contains',
    is_active BOOLEAN DEFAULT TRUE,
    use_count INTEGER DEFAULT 0,
    last_used_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT check_match_type CHECK (match_type IN ('exact', 'contains')),
    UNIQUE(user_id, merchant_name)
);

CREATE INDEX idx_merchant_patterns_user_id ON merchant_patterns(user_id);
CREATE INDEX idx_merchant_patterns_is_active ON merchant_patterns(is_active);
CREATE INDEX idx_merchant_patterns_merchant_name ON merchant_patterns(merchant_name);
