-- Create categories table
CREATE TABLE IF NOT EXISTS categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    color VARCHAR(7) NOT NULL,
    is_default BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_categories_user_id ON categories(user_id);
CREATE INDEX idx_categories_is_default ON categories(is_default);

-- Insert default system categories
INSERT INTO categories (id, user_id, name, color, is_default) VALUES
    ('11111111-1111-1111-1111-111111111111', NULL, 'Groceries', '#4CAF50', TRUE),
    ('22222222-2222-2222-2222-222222222222', NULL, 'Transport', '#2196F3', TRUE),
    ('33333333-3333-3333-3333-333333333333', NULL, 'Shopping', '#FFC107', TRUE),
    ('44444444-4444-4444-4444-444444444444', NULL, 'Bills', '#F44336', TRUE),
    ('55555555-5555-5555-5555-555555555555', NULL, 'Entertainment', '#9C27B0', TRUE),
    ('66666666-6666-6666-6666-666666666666', NULL, 'Food & Dining', '#FF9800', TRUE),
    ('77777777-7777-7777-7777-777777777777', NULL, 'Health', '#E91E63', TRUE),
    ('88888888-8888-8888-8888-888888888888', NULL, 'Travel', '#00BCD4', TRUE),
    ('99999999-9999-9999-9999-999999999999', NULL, 'Other', '#9E9E9E', TRUE);
