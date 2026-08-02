DROP INDEX IF EXISTS idx_notification_deliveries_notif_channel_device;
ALTER TABLE notification_deliveries DROP COLUMN IF EXISTS token_suffix;
ALTER TABLE notification_deliveries DROP COLUMN IF EXISTS platform;
ALTER TABLE notification_deliveries DROP COLUMN IF EXISTS device_id;
