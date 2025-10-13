-- Create sync_status table
CREATE TABLE IF NOT EXISTS sync_status (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    device_id VARCHAR(255) NOT NULL REFERENCES devices(device_id) ON DELETE CASCADE,
    device_name VARCHAR(255) NOT NULL,
    last_sync_time TIMESTAMP,
    last_sync_type VARCHAR(20),
    pending_count INTEGER DEFAULT 0,
    synced_count INTEGER DEFAULT 0,
    status VARCHAR(20) DEFAULT 'idle',
    error_message TEXT,
    conflicts_resolved INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT check_sync_type CHECK (last_sync_type IN ('realtime', 'batch', 'manual')),
    CONSTRAINT check_status CHECK (status IN ('idle', 'syncing', 'success', 'error')),
    UNIQUE(user_id, device_id)
);

CREATE INDEX idx_sync_status_user_id ON sync_status(user_id);
CREATE INDEX idx_sync_status_device_id ON sync_status(device_id);
