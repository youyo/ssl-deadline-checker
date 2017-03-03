
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
ALTER TABLE `hostnames` DROP INDEX `id`, ADD UNIQUE `idx_id`(`id`);
ALTER TABLE `hostnames` DROP INDEX `timelimit`, ADD INDEX `idx_timelimit`(`timelimit`);
ALTER TABLE `hostnames` DROP INDEX `remaining_days`, ADD INDEX `idx_remaining_days`(`remaining_days`);
ALTER TABLE `hostnames` ADD `notification_days`int(11) unsigned NOT NULL DEFAULT 45 AFTER `remaining_days`;

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
ALTER TABLE `hostnames` DROP `notification_days`;
ALTER TABLE `hostnames` DROP INDEX `idx_remaining_days`, ADD INDEX `remaining_days`(`remaining_days`);
ALTER TABLE `hostnames` DROP INDEX `idx_timelimit`, ADD INDEX `timelimit`(`timelimit`);
ALTER TABLE `hostnames` DROP INDEX `idx_id`, ADD UNIQUE `id`(`id`);
