-- Drop triggers first
DROP TRIGGER IF EXISTS update_note_reminders_updated_at ON note_reminders;
DROP TRIGGER IF EXISTS update_user_devices_updated_at ON user_devices;

-- Drop indexes
DROP INDEX IF EXISTS idx_notification_logs_created_at;
DROP INDEX IF EXISTS idx_notification_logs_reminder_id;
DROP INDEX IF EXISTS idx_notification_logs_status;
DROP INDEX IF EXISTS idx_notification_logs_user_id;
DROP INDEX IF EXISTS idx_note_reminders_user_enabled;
DROP INDEX IF EXISTS idx_note_reminders_note_id;
DROP INDEX IF EXISTS idx_note_reminders_user_id;
DROP INDEX IF EXISTS idx_note_reminders_next_trigger;
DROP INDEX IF EXISTS idx_user_devices_token;
DROP INDEX IF EXISTS idx_user_devices_user_id;

-- Drop tables (in correct order due to foreign key constraints)
DROP TABLE IF EXISTS notification_logs;
DROP TABLE IF EXISTS note_reminders;
DROP TABLE IF EXISTS user_devices;

-- Drop enum types
DROP TYPE IF EXISTS notification_status;
DROP TYPE IF EXISTS repeat_type;
DROP TYPE IF EXISTS device_type;
