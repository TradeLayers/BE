DROP INDEX IF EXISTS idx_threshold_notifications_user_read_created;

DROP TABLE IF EXISTS threshold_notifications;

ALTER TABLE watchlist
DROP COLUMN IF EXISTS threshold_reached;