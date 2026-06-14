-- 自定义采集规则表
CREATE TABLE IF NOT EXISTS `mac_cj_rule` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(100) NOT NULL DEFAULT '',
  `source_url` varchar(500) NOT NULL DEFAULT '',
  `list_rule` text,
  `detail_rule` text,
  `type_id` int NOT NULL DEFAULT 0,
  `status` tinyint NOT NULL DEFAULT 0,
  `last_run` bigint NOT NULL DEFAULT 0,
  `created_at` bigint NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 资源站采集源表
CREATE TABLE IF NOT EXISTS `mac_collect_source` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(100) NOT NULL DEFAULT '',
  `api_url` varchar(500) NOT NULL DEFAULT '',
  `up_rule` varchar(10) NOT NULL DEFAULT '',
  `status` tinyint NOT NULL DEFAULT 0,
  `priority` int NOT NULL DEFAULT 0,
  `last_sync` bigint NOT NULL DEFAULT 0,
  `created_at` bigint NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
