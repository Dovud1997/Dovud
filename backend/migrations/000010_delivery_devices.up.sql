ALTER TABLE notification_deliveries
    ADD COLUMN IF NOT EXISTS device_id VARCHAR(128) NULL,
    ADD COLUMN IF NOT EXISTS platform VARCHAR(32) NULL,
    ADD COLUMN IF NOT EXISTS token_suffix VARCHAR(16) NULL;

CREATE INDEX IF NOT EXISTS idx_notification_deliveries_notif_channel_device
    ON notification_deliveries (notification_id, channel, device_id);
