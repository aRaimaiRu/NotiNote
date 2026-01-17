-- Create device_type enum for user devices
CREATE TYPE device_type AS ENUM ('web', 'android', 'ios');

-- Create repeat_type enum for reminders
CREATE TYPE repeat_type AS ENUM ('once', 'daily', 'weekly', 'monthly');

-- Create notification_status enum for notification logs
CREATE TYPE notification_status AS ENUM ('pending', 'sent', 'failed', 'cancelled');

-- User devices for FCM token management
CREATE TABLE user_devices (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    device_token TEXT NOT NULL,
    device_type device_type NOT NULL,
    device_name VARCHAR(255),
    browser_info VARCHAR(255),
    is_active BOOLEAN DEFAULT true,
    last_used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,

    -- Unique constraint: same token cannot be registered twice for same user
    UNIQUE(user_id, device_token)
);

-- Note reminders (scheduled notifications)
CREATE TABLE note_reminders (
    id BIGSERIAL PRIMARY KEY,
    note_id BIGINT NOT NULL REFERENCES notes(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    message TEXT,
    scheduled_at TIMESTAMPTZ NOT NULL,

    -- Repeat configuration
    repeat_type repeat_type DEFAULT 'once',
    repeat_config JSONB, -- {"days": [1,3,5]} for weekly, {"day": 15} for monthly
    repeat_end_at TIMESTAMPTZ, -- When to stop repeating (null = forever)

    -- Status tracking
    is_enabled BOOLEAN DEFAULT true,
    next_trigger_at TIMESTAMPTZ NOT NULL,
    last_triggered_at TIMESTAMPTZ,
    trigger_count INT DEFAULT 0,

    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- Notification delivery log
CREATE TABLE notification_logs (
    id BIGSERIAL PRIMARY KEY,
    reminder_id BIGINT REFERENCES note_reminders(id) ON DELETE SET NULL,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    device_id BIGINT REFERENCES user_devices(id) ON DELETE SET NULL,

    title VARCHAR(255) NOT NULL,
    body TEXT,
    data JSONB, -- Additional payload data

    status notification_status NOT NULL DEFAULT 'pending',
    error_message TEXT,
    fcm_message_id VARCHAR(255),

    scheduled_at TIMESTAMPTZ,
    sent_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for user_devices
-- Find active devices for a user (most common query for sending notifications)
CREATE INDEX idx_user_devices_user_id ON user_devices(user_id) WHERE is_active = true;

-- Find device by token (for token updates/deactivation)
CREATE INDEX idx_user_devices_token ON user_devices(device_token);

-- Indexes for note_reminders
-- Find due reminders (critical for scheduler performance)
CREATE INDEX idx_note_reminders_next_trigger ON note_reminders(next_trigger_at)
    WHERE is_enabled = true;

-- Find reminders for a user (listing user's reminders)
CREATE INDEX idx_note_reminders_user_id ON note_reminders(user_id);

-- Find reminders for a note
CREATE INDEX idx_note_reminders_note_id ON note_reminders(note_id);

-- Composite index for finding enabled reminders for a user
CREATE INDEX idx_note_reminders_user_enabled ON note_reminders(user_id, is_enabled)
    WHERE is_enabled = true;

-- Indexes for notification_logs
-- Find logs for a user (history)
CREATE INDEX idx_notification_logs_user_id ON notification_logs(user_id);

-- Find logs by status (for retry processing)
CREATE INDEX idx_notification_logs_status ON notification_logs(status)
    WHERE status = 'pending';

-- Find logs by reminder (debugging/audit)
CREATE INDEX idx_notification_logs_reminder_id ON notification_logs(reminder_id);

-- Recent logs query
CREATE INDEX idx_notification_logs_created_at ON notification_logs(created_at DESC);

-- Create trigger for user_devices updated_at
CREATE TRIGGER update_user_devices_updated_at
    BEFORE UPDATE ON user_devices
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Create trigger for note_reminders updated_at
CREATE TRIGGER update_note_reminders_updated_at
    BEFORE UPDATE ON note_reminders
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Add documentation comments
COMMENT ON TABLE user_devices IS 'Stores user device registrations for push notifications (FCM tokens)';
COMMENT ON COLUMN user_devices.id IS 'Unique device registration identifier';
COMMENT ON COLUMN user_devices.user_id IS 'Owner of the device';
COMMENT ON COLUMN user_devices.device_token IS 'FCM device token for push notifications';
COMMENT ON COLUMN user_devices.device_type IS 'Platform type: web, android, or ios';
COMMENT ON COLUMN user_devices.device_name IS 'User-friendly device name';
COMMENT ON COLUMN user_devices.browser_info IS 'Browser/app info for web devices';
COMMENT ON COLUMN user_devices.is_active IS 'Whether device is currently active for receiving notifications';
COMMENT ON COLUMN user_devices.last_used_at IS 'Last time the device received a notification';

COMMENT ON TABLE note_reminders IS 'Stores scheduled reminders for notes with support for recurring notifications';
COMMENT ON COLUMN note_reminders.id IS 'Unique reminder identifier';
COMMENT ON COLUMN note_reminders.note_id IS 'Associated note';
COMMENT ON COLUMN note_reminders.user_id IS 'Owner of the reminder';
COMMENT ON COLUMN note_reminders.title IS 'Notification title';
COMMENT ON COLUMN note_reminders.message IS 'Notification body/message';
COMMENT ON COLUMN note_reminders.scheduled_at IS 'Original scheduled time';
COMMENT ON COLUMN note_reminders.repeat_type IS 'Repeat pattern: once, daily, weekly, monthly';
COMMENT ON COLUMN note_reminders.repeat_config IS 'JSON config for repeat pattern (days for weekly, day for monthly)';
COMMENT ON COLUMN note_reminders.repeat_end_at IS 'When to stop repeating (null = forever)';
COMMENT ON COLUMN note_reminders.is_enabled IS 'Whether reminder is currently active';
COMMENT ON COLUMN note_reminders.next_trigger_at IS 'Next scheduled trigger time';
COMMENT ON COLUMN note_reminders.last_triggered_at IS 'Last time the reminder was triggered';
COMMENT ON COLUMN note_reminders.trigger_count IS 'Number of times the reminder has been triggered';

COMMENT ON TABLE notification_logs IS 'Audit log for all notification delivery attempts';
COMMENT ON COLUMN notification_logs.id IS 'Unique log entry identifier';
COMMENT ON COLUMN notification_logs.reminder_id IS 'Associated reminder (null if reminder deleted)';
COMMENT ON COLUMN notification_logs.user_id IS 'User who received the notification';
COMMENT ON COLUMN notification_logs.device_id IS 'Device that received the notification';
COMMENT ON COLUMN notification_logs.title IS 'Notification title that was sent';
COMMENT ON COLUMN notification_logs.body IS 'Notification body that was sent';
COMMENT ON COLUMN notification_logs.data IS 'Additional payload data sent with notification';
COMMENT ON COLUMN notification_logs.status IS 'Delivery status: pending, sent, failed, cancelled';
COMMENT ON COLUMN notification_logs.error_message IS 'Error details if delivery failed';
COMMENT ON COLUMN notification_logs.fcm_message_id IS 'FCM message ID for successful deliveries';
COMMENT ON COLUMN notification_logs.scheduled_at IS 'When notification was scheduled to be sent';
COMMENT ON COLUMN notification_logs.sent_at IS 'When notification was actually sent';
