CREATE TABLE IF NOT EXISTS `instances` (
  `id` TEXT PRIMARY KEY,
  `execution_id` TEXT NO NULL,
  `parent_instance_id` TEXT NULL,
  `parent_schedule_event_id` INTEGER NULL,
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `completed_at` DATETIME NULL,
  `locked_until` DATETIME NULL,
  `sticky_until` DATETIME NULL,
  `worker` TEXT NULL

);

CREATE INDEX IF NOT EXISTS `idx_instances_locked_until_completed_at` ON `instances` (`locked_until`, `sticky_until`, `completed_at`, `worker`);
CREATE INDEX IF NOT EXISTS `idx_instances_parent_instance_id` ON `instances` (`parent_instance_id`);

CREATE TABLE IF NOT EXISTS `pending_events` (
  `id` TEXT PRIMARY KEY,
  `instance_id` TEXT NOT NULL,
  `event_type` INTEGER NOT NULL,
  `timestamp` DATETIME NOT NULL,
  `schedule_event_id` INT NOT NULL,
  `attributes` BLOB NOT NULL,
  `visible_at` DATETIME NULL
);

CREATE TABLE IF NOT EXISTS `history` (
  `id` TEXT PRIMARY KEY,
  `instance_id` TEXT NOT NULL,
  `event_type` INTEGER NOT NULL,
  `timestamp` DATETIME NOT NULL,
  `schedule_event_id` INT NOT NULL,
  `attributes` BLOB NOT NULL,
  `visible_at` DATETIME NULL
);

CREATE TABLE IF NOT EXISTS `activities` (
  `id` TEXT PRIMARY KEY,
  `instance_id` TEXT NOT NULL,
  `execution_id` TEXT NOT NULL,
  `event_type` INTEGER NOT NULL,
  `timestamp` DATETIME NOT NULL,
  `schedule_event_id` INT NOT NULL,
  `attributes` BLOB NOT NULL,
  `visible_at` DATETIME NULL,
  `locked_until` DATETIME NULL,
  `worker` TEXT NULL
);