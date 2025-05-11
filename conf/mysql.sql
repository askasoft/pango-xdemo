---------------------------------;
CREATE TABLE `SCHEMA`.`files` (
	`id` varchar(255) NOT NULL,
	`name` longtext NOT NULL,
	`ext` longtext NOT NULL,
	`time` datetime(3) NOT NULL,
	`size` bigint NOT NULL,
	`data` longblob NOT NULL,
	PRIMARY KEY (`id`)
);
---------------------------------;
CREATE TABLE `SCHEMA`.`jobs` (
	`id` bigint AUTO_INCREMENT NOT NULL,
	`cid` bigint NOT NULL,
	`rid` bigint NOT NULL,
	`name` varchar(250) NOT NULL,
	`status` varchar(1) NOT NULL,
	`locale` varchar(20) NOT NULL,
	`param` longtext NOT NULL,
	`state` longtext NOT NULL,
	`result` longtext NOT NULL,
	`error` longtext NOT NULL,
	`created_at` datetime(3) NOT NULL,
	`updated_at` datetime(3) NOT NULL,
	PRIMARY KEY (`id`),
	INDEX `idx_jobs_name` (`name`)
);
---------------------------------;
CREATE TABLE `SCHEMA`.`job_logs` (
	`id` bigint AUTO_INCREMENT NOT NULL,
	`jid` bigint NOT NULL,
	`time` datetime(3) NOT NULL,
	`level` varchar(1) NOT NULL,
	`message` longtext NOT NULL,
	PRIMARY KEY (`id`),
	INDEX `idx_job_logs_jid` (`jid`)
);
---------------------------------;
CREATE TABLE `SCHEMA`.`job_chains` (
	`id` bigint AUTO_INCREMENT NOT NULL,
	`name` varchar(250) NOT NULL,
	`status` varchar(1) NOT NULL,
	`states` longtext NOT NULL,
	`created_at` datetime(3) NOT NULL,
	`updated_at` datetime(3) NOT NULL,
	PRIMARY KEY (`id`),
	INDEX `idx_job_chains_name` (`name`)
);
---------------------------------;
CREATE TABLE `SCHEMA`.`users` (
	`id` bigint AUTO_INCREMENT NOT NULL,
	`name` varchar(100) NOT NULL,
	`email` varchar(200) NOT NULL,
	`password` varchar(200) NOT NULL,
	`role` varchar(1) NOT NULL,
	`status` varchar(1) NOT NULL,
	`cidr` longtext NOT NULL,
	`secret` bigint NOT NULL,
	`created_at` datetime(3) NOT NULL,
	`updated_at` datetime(3) NOT NULL,
	PRIMARY KEY (`id`),
	UNIQUE INDEX `idx_users_email` (`email`)
);
---------------------------------;
CREATE TABLE `SCHEMA`.`configs` (
	`name` varchar(64) NOT NULL,
	`value` longtext NOT NULL,
	`style` varchar(2) NOT NULL,
	`order` bigint NOT NULL,
	`required` boolean NOT NULL,
	`secret` boolean NOT NULL,
	`viewer` varchar(1) NOT NULL,
	`editor` varchar(1) NOT NULL,
	`validation` longtext NOT NULL,
	`created_at` datetime(3) NOT NULL,
	`updated_at` datetime(3) NOT NULL,
	PRIMARY KEY (`name`)
);
---------------------------------;
CREATE TABLE `SCHEMA`.`audit_logs` (
	`id` bigint AUTO_INCREMENT NOT NULL,
	`date` datetime(3) NOT NULL,
	`uid` bigint NOT NULL,
	`cip` varchar(40) NOT NULL,
	`role` varchar(1) NOT NULL,
	`func` varchar(32) NOT NULL,
	`action` varchar(32) NOT NULL,
	`params` json,
	PRIMARY KEY (`id`)
);
---------------------------------;
CREATE TABLE `SCHEMA`.`pets` (
	`id` bigint AUTO_INCREMENT NOT NULL,
	`name` varchar(100) NOT NULL,
	`gender` varchar(1) NOT NULL,
	`born_at` datetime(3) NOT NULL,
	`origin` varchar(10) NOT NULL,
	`temper` varchar(1) NOT NULL,
	`habits` json,
	`amount` bigint NOT NULL,
	`price` decimal(10, 2) NOT NULL,
	`shop_name` varchar(200) NOT NULL,
	`shop_address` varchar(200) NOT NULL,
	`shop_telephone` varchar(20) NOT NULL,
	`shop_link` varchar(1000) NOT NULL,
	`description` longtext NOT NULL,
	`created_at` datetime(3) NOT NULL,
	`updated_at` datetime(3) NOT NULL,
	PRIMARY KEY (`id`)
);