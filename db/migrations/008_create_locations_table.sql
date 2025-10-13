-- Create locations table
CREATE TABLE IF NOT EXISTS locations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    latitude DECIMAL(10, 8) NOT NULL,
    longitude DECIMAL(11, 8) NOT NULL,
    timestamp TIMESTAMP NOT NULL,
    address TEXT,
    accuracy DECIMAL(10, 2),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_locations_user_id ON locations(user_id);
CREATE INDEX idx_locations_timestamp ON locations(timestamp);

-- Add foreign key to expenses table for location_id
ALTER TABLE expenses ADD CONSTRAINT fk_expenses_location
    FOREIGN KEY (location_id) REFERENCES locations(id) ON DELETE SET NULL;
