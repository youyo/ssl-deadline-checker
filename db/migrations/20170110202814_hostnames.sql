
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
CREATE TABLE `hostnames` (
	`id` int(11) unsigned NOT NULL AUTO_INCREMENT,
	`hostname` varchar(191) NOT NULL DEFAULT '',
	`timelimit` date NOT NULL,
	`remaining_days` int(11) unsigned NOT NULL,
	`created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
	`updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
	PRIMARY KEY (`hostname`),
	UNIQUE KEY `id` (`id`),
	KEY `timelimit` (`timelimit`),
	KEY `remaining_days` (`remaining_days`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
DROP TABLE `hostnames`;
