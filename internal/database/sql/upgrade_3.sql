-- 采集资源站表（重构，兼容 maccms10 协议）
CREATE TABLE IF NOT EXISTS `mac_collect` (
  `collect_id` int unsigned NOT NULL AUTO_INCREMENT,
  `collect_name` varchar(100) NOT NULL DEFAULT '',
  `collect_url` varchar(500) NOT NULL DEFAULT '',
  `collect_type` tinyint NOT NULL DEFAULT 2 COMMENT '1=xml 2=json',
  `collect_mid` tinyint NOT NULL DEFAULT 1 COMMENT '1=vod 2=art 8=actor 9=role 12=manga',
  `collect_appid` varchar(30) NOT NULL DEFAULT '',
  `collect_appkey` varchar(30) NOT NULL DEFAULT '',
  `collect_param` varchar(200) NOT NULL DEFAULT '',
  `collect_filter` tinyint NOT NULL DEFAULT 0 COMMENT '0=不过滤 1=增改 2=仅增 3=仅改',
  `collect_filter_from` varchar(255) NOT NULL DEFAULT '',
  `collect_filter_year` varchar(255) NOT NULL DEFAULT '',
  `collect_opt` tinyint NOT NULL DEFAULT 0 COMMENT '0=增+改 1=仅增 2=仅改',
  `collect_sync_pic_opt` tinyint NOT NULL DEFAULT 0 COMMENT '0=跟随全局 1=开启 2=关闭',
  `created_at` bigint NOT NULL DEFAULT 0,
  PRIMARY KEY (`collect_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 分类绑定表（远程分类→本地分类映射）
CREATE TABLE IF NOT EXISTS `mac_collect_bind` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `collect_flag` varchar(100) NOT NULL DEFAULT '' COMMENT 'sourceId_remoteTypeId',
  `remote_type_id` int NOT NULL DEFAULT 0,
  `remote_name` varchar(100) NOT NULL DEFAULT '',
  `local_type_id` int NOT NULL DEFAULT 0,
  `local_name` varchar(100) NOT NULL DEFAULT '',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_flag` (`collect_flag`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 采集配置表（全局配置 KV 存储）
CREATE TABLE IF NOT EXISTS `mac_collect_config` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `config_key` varchar(50) NOT NULL DEFAULT '',
  `config_value` text,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_key` (`config_key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 插入默认采集配置
INSERT IGNORE INTO `mac_collect_config` (`config_key`, `config_value`) VALUES
('collect_global', '{"workers":5,"batch_size":50,"rate_limit":10,"buffer_size":500,"timeout":30,"status":1,"hits_start":0,"hits_end":0,"updown_start":0,"updown_end":0,"score_random":0,"inrule":"a","uprule":"a","filter":"","pic_sync":0}');
