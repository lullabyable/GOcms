-- GoCMS Go 全新安装SQL
-- 版本: 1

CREATE TABLE IF NOT EXISTS `mac_config` (
  `config_id` int unsigned NOT NULL AUTO_INCREMENT,
  `type` varchar(64) NOT NULL DEFAULT '',
  `name` varchar(128) NOT NULL DEFAULT '',
  `value` text,
  PRIMARY KEY (`config_id`),
  KEY `idx_type` (`type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `mac_type` (
  `type_id` int unsigned NOT NULL AUTO_INCREMENT,
  `type_name` varchar(64) NOT NULL DEFAULT '',
  `type_en` varchar(64) NOT NULL DEFAULT '',
  `type_pid` int unsigned NOT NULL DEFAULT 0,
  `type_sort` int NOT NULL DEFAULT 0,
  `type_mid` smallint unsigned NOT NULL DEFAULT 1,
  `type_letter` varchar(16) NOT NULL DEFAULT '',
  `type_color` varchar(16) NOT NULL DEFAULT '',
  `type_tpl_list` varchar(128) NOT NULL DEFAULT '',
  `type_tpl_detail` varchar(128) NOT NULL DEFAULT '',
  `type_tpl_play` varchar(128) NOT NULL DEFAULT '',
  `type_tpl_down` varchar(128) NOT NULL DEFAULT '',
  `type_key` varchar(255) NOT NULL DEFAULT '',
  `type_des` varchar(255) NOT NULL DEFAULT '',
  `type_title` varchar(255) NOT NULL DEFAULT '',
  `type_jumpurl` varchar(255) NOT NULL DEFAULT '',
  `type_pic` varchar(255) NOT NULL DEFAULT '',
  `type_status` tinyint NOT NULL DEFAULT 1,
  `type_extend` text,
  `type_show_tpl` varchar(128) NOT NULL DEFAULT '',
  `type_read_tpl` varchar(128) NOT NULL DEFAULT '',
  PRIMARY KEY (`type_id`),
  KEY `idx_pid` (`type_pid`),
  KEY `idx_en` (`type_en`),
  KEY `idx_sort` (`type_sort`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `mac_vod` (
  `vod_id` int unsigned NOT NULL AUTO_INCREMENT,
  `type_id` int unsigned NOT NULL DEFAULT 0,
  `type_id_1` int unsigned NOT NULL DEFAULT 0,
  `group_id` int NOT NULL DEFAULT 0,
  `vod_name` varchar(255) NOT NULL DEFAULT '',
  `vod_sub` varchar(255) NOT NULL DEFAULT '',
  `vod_en` varchar(128) NOT NULL DEFAULT '',
  `vod_time` varchar(64) NOT NULL DEFAULT '',
  `vod_class` varchar(255) NOT NULL DEFAULT '',
  `vod_tag` varchar(255) NOT NULL DEFAULT '',
  `vod_pic` varchar(255) NOT NULL DEFAULT '',
  `vod_pic_thumb` varchar(255) NOT NULL DEFAULT '',
  `vod_pic_slide` varchar(255) NOT NULL DEFAULT '',
  `vod_pic_screenshot` varchar(255) NOT NULL DEFAULT '',
  `vod_actor` varchar(512) NOT NULL DEFAULT '',
  `vod_director` varchar(255) NOT NULL DEFAULT '',
  `vod_writer` varchar(255) NOT NULL DEFAULT '',
  `vod_blurb` varchar(512) NOT NULL DEFAULT '',
  `vod_remarks` varchar(255) NOT NULL DEFAULT '',
  `vod_pubdate` varchar(64) NOT NULL DEFAULT '',
  `vod_area` varchar(64) NOT NULL DEFAULT '',
  `vod_lang` varchar(64) NOT NULL DEFAULT '',
  `vod_year` varchar(16) NOT NULL DEFAULT '',
  `vod_version` varchar(64) NOT NULL DEFAULT '',
  `vod_state` varchar(32) NOT NULL DEFAULT '',
  `vod_author` varchar(255) NOT NULL DEFAULT '',
  `vod_jumpurl` varchar(255) NOT NULL DEFAULT '',
  `vod_letter` char(1) NOT NULL DEFAULT '',
  `vod_color` varchar(16) NOT NULL DEFAULT '',
  `vod_lock` tinyint NOT NULL DEFAULT 0,
  `vod_level` tinyint NOT NULL DEFAULT 0,
  `vod_points` int NOT NULL DEFAULT 0,
  `vod_points_play` int NOT NULL DEFAULT 0,
  `vod_points_down` int NOT NULL DEFAULT 0,
  `vod_hits` int unsigned NOT NULL DEFAULT 0,
  `vod_hits_day` int unsigned NOT NULL DEFAULT 0,
  `vod_hits_week` int unsigned NOT NULL DEFAULT 0,
  `vod_hits_month` int unsigned NOT NULL DEFAULT 0,
  `vod_duration` varchar(32) NOT NULL DEFAULT '',
  `vod_up` int unsigned NOT NULL DEFAULT 0,
  `vod_down` int unsigned NOT NULL DEFAULT 0,
  `vod_score` varchar(8) NOT NULL DEFAULT '0.0',
  `vod_score_all` int unsigned NOT NULL DEFAULT 0,
  `vod_score_num` int unsigned NOT NULL DEFAULT 0,
  `vod_content` text,
  `vod_play_from` varchar(512) NOT NULL DEFAULT '',
  `vod_play_server` varchar(255) NOT NULL DEFAULT '',
  `vod_play_note` varchar(255) NOT NULL DEFAULT '',
  `vod_play_url` text,
  `vod_down_from` varchar(512) NOT NULL DEFAULT '',
  `vod_down_server` varchar(255) NOT NULL DEFAULT '',
  `vod_down_note` varchar(255) NOT NULL DEFAULT '',
  `vod_down_url` text,
  `vod_plot` tinyint NOT NULL DEFAULT 0,
  `vod_plot_name` text,
  `vod_plot_detail` text,
  `vod_status` tinyint NOT NULL DEFAULT 1,
  `vod_time_add` int unsigned NOT NULL DEFAULT 0,
  `vod_time_hits` int unsigned NOT NULL DEFAULT 0,
  `vod_time_make` int unsigned NOT NULL DEFAULT 0,
  `vod_trysee` int NOT NULL DEFAULT 0,
  `vod_copyright` tinyint NOT NULL DEFAULT 0,
  `vod_rel_art` varchar(255) NOT NULL DEFAULT '',
  `vod_rel_vod` varchar(255) NOT NULL DEFAULT '',
  `vod_pwd` varchar(255) NOT NULL DEFAULT '',
  `vod_pwd_url` varchar(255) NOT NULL DEFAULT '',
  `vod_pwd_play` varchar(255) NOT NULL DEFAULT '',
  `vod_pwd_down` varchar(255) NOT NULL DEFAULT '',
  PRIMARY KEY (`vod_id`),
  KEY `idx_type_status` (`type_id`, `vod_status`),
  KEY `idx_type_time` (`type_id`, `vod_time`),
  KEY `idx_name` (`vod_name`),
  KEY `idx_en` (`vod_en`),
  KEY `idx_hits` (`vod_hits`),
  KEY `idx_score` (`vod_score`),
  KEY `idx_level` (`vod_level`),
  KEY `idx_letter` (`vod_letter`),
  KEY `idx_year` (`vod_year`),
  KEY `idx_area` (`vod_area`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `mac_art` (
  `art_id` int unsigned NOT NULL AUTO_INCREMENT,
  `type_id` int unsigned NOT NULL DEFAULT 0,
  `type_id_1` int unsigned NOT NULL DEFAULT 0,
  `art_name` varchar(255) NOT NULL DEFAULT '',
  `art_sub` varchar(255) NOT NULL DEFAULT '',
  `art_en` varchar(128) NOT NULL DEFAULT '',
  `art_time` varchar(64) NOT NULL DEFAULT '',
  `art_letter` char(1) NOT NULL DEFAULT '',
  `art_color` varchar(16) NOT NULL DEFAULT '',
  `art_from` varchar(255) NOT NULL DEFAULT '',
  `art_author` varchar(255) NOT NULL DEFAULT '',
  `art_tag` varchar(255) NOT NULL DEFAULT '',
  `art_class` varchar(255) NOT NULL DEFAULT '',
  `art_pic` varchar(255) NOT NULL DEFAULT '',
  `art_pic_thumb` varchar(255) NOT NULL DEFAULT '',
  `art_pic_slide` varchar(255) NOT NULL DEFAULT '',
  `art_content` text,
  `art_blurb` varchar(512) NOT NULL DEFAULT '',
  `art_remarks` varchar(255) NOT NULL DEFAULT '',
  `art_jumpurl` varchar(255) NOT NULL DEFAULT '',
  `art_lock` tinyint NOT NULL DEFAULT 0,
  `art_level` tinyint NOT NULL DEFAULT 0,
  `art_points` int NOT NULL DEFAULT 0,
  `art_hits` int unsigned NOT NULL DEFAULT 0,
  `art_hits_day` int unsigned NOT NULL DEFAULT 0,
  `art_hits_week` int unsigned NOT NULL DEFAULT 0,
  `art_hits_month` int unsigned NOT NULL DEFAULT 0,
  `art_up` int unsigned NOT NULL DEFAULT 0,
  `art_down` int unsigned NOT NULL DEFAULT 0,
  `art_status` tinyint NOT NULL DEFAULT 1,
  `art_time_add` int unsigned NOT NULL DEFAULT 0,
  `art_time_hits` int unsigned NOT NULL DEFAULT 0,
  `art_time_make` int unsigned NOT NULL DEFAULT 0,
  `art_note` varchar(255) NOT NULL DEFAULT '',
  PRIMARY KEY (`art_id`),
  KEY `idx_type_status` (`type_id`, `art_status`),
  KEY `idx_type_time` (`type_id`, `art_time`),
  KEY `idx_name` (`art_name`),
  KEY `idx_en` (`art_en`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `mac_actor` (
  `actor_id` int unsigned NOT NULL AUTO_INCREMENT,
  `actor_name` varchar(128) NOT NULL DEFAULT '',
  `actor_en` varchar(128) NOT NULL DEFAULT '',
  `actor_sex` tinyint NOT NULL DEFAULT 0,
  `actor_area` varchar(64) NOT NULL DEFAULT '',
  `actor_birthday` varchar(32) NOT NULL DEFAULT '',
  `actor_birtharea` varchar(64) NOT NULL DEFAULT '',
  `actor_star` varchar(16) NOT NULL DEFAULT '',
  `actor_height` varchar(16) NOT NULL DEFAULT '',
  `actor_weight` varchar(16) NOT NULL DEFAULT '',
  `actor_pic` varchar(255) NOT NULL DEFAULT '',
  `actor_blurb` text,
  `actor_content` text,
  `actor_tag` varchar(255) NOT NULL DEFAULT '',
  `actor_level` tinyint NOT NULL DEFAULT 0,
  `actor_lock` tinyint NOT NULL DEFAULT 0,
  `actor_time` varchar(64) NOT NULL DEFAULT '',
  `actor_hits` int unsigned NOT NULL DEFAULT 0,
  `actor_hits_day` int unsigned NOT NULL DEFAULT 0,
  `actor_hits_week` int unsigned NOT NULL DEFAULT 0,
  `actor_hits_month` int unsigned NOT NULL DEFAULT 0,
  `actor_status` tinyint NOT NULL DEFAULT 1,
  PRIMARY KEY (`actor_id`),
  KEY `idx_name` (`actor_name`),
  KEY `idx_en` (`actor_en`),
  KEY `idx_sex` (`actor_sex`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `mac_role` (
  `role_id` int unsigned NOT NULL AUTO_INCREMENT,
  `role_name` varchar(128) NOT NULL DEFAULT '',
  `role_en` varchar(128) NOT NULL DEFAULT '',
  `role_sex` tinyint NOT NULL DEFAULT 0,
  `role_pic` varchar(255) NOT NULL DEFAULT '',
  `role_actor` varchar(128) NOT NULL DEFAULT '',
  `role_blurb` varchar(512) NOT NULL DEFAULT '',
  `role_content` text,
  `role_sort` int NOT NULL DEFAULT 0,
  `role_level` tinyint NOT NULL DEFAULT 0,
  `role_lock` tinyint NOT NULL DEFAULT 0,
  `role_status` tinyint NOT NULL DEFAULT 1,
  PRIMARY KEY (`role_id`),
  KEY `idx_name` (`role_name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `mac_user` (
  `user_id` int unsigned NOT NULL AUTO_INCREMENT,
  `group_id` int unsigned NOT NULL DEFAULT 0,
  `user_name` varchar(64) NOT NULL DEFAULT '',
  `user_pwd` varchar(255) NOT NULL DEFAULT '',
  `user_nick_name` varchar(64) NOT NULL DEFAULT '',
  `user_email` varchar(128) NOT NULL DEFAULT '',
  `user_phone` varchar(32) NOT NULL DEFAULT '',
  `user_portrait` varchar(255) NOT NULL DEFAULT '',
  `user_points` int NOT NULL DEFAULT 0,
  `user_points_day` int NOT NULL DEFAULT 0,
  `user_status` tinyint NOT NULL DEFAULT 1,
  `user_reg_time` int unsigned NOT NULL DEFAULT 0,
  `user_reg_ip` varchar(64) NOT NULL DEFAULT '',
  `user_login_time` int unsigned NOT NULL DEFAULT 0,
  `user_login_ip` varchar(64) NOT NULL DEFAULT '',
  `user_last_time` int unsigned NOT NULL DEFAULT 0,
  `user_last_ip` varchar(64) NOT NULL DEFAULT '',
  `user_login_num` int unsigned NOT NULL DEFAULT 0,
  `expiry_time` int unsigned NOT NULL DEFAULT 0,
  PRIMARY KEY (`user_id`),
  UNIQUE KEY `uk_name` (`user_name`),
  KEY `idx_group` (`group_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `mac_group` (
  `group_id` int unsigned NOT NULL AUTO_INCREMENT,
  `group_name` varchar(64) NOT NULL DEFAULT '',
  `group_type` tinyint NOT NULL DEFAULT 0,
  `group_points` int NOT NULL DEFAULT 0,
  `group_state` tinyint NOT NULL DEFAULT 1,
  PRIMARY KEY (`group_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `mac_admin` (
  `admin_id` int unsigned NOT NULL AUTO_INCREMENT,
  `admin_name` varchar(64) NOT NULL DEFAULT '',
  `admin_pwd` varchar(255) NOT NULL DEFAULT '',
  `admin_role` tinyint NOT NULL DEFAULT 2,
  `admin_status` tinyint NOT NULL DEFAULT 1,
  `admin_last_time` int unsigned NOT NULL DEFAULT 0,
  `admin_last_ip` varchar(64) NOT NULL DEFAULT '',
  `admin_login_num` int unsigned NOT NULL DEFAULT 0,
  PRIMARY KEY (`admin_id`),
  UNIQUE KEY `uk_name` (`admin_name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `mac_comment` (
  `comment_id` int unsigned NOT NULL AUTO_INCREMENT,
  `comment_rid` int unsigned NOT NULL DEFAULT 0,
  `comment_type` tinyint NOT NULL DEFAULT 1,
  `user_id` int unsigned NOT NULL DEFAULT 0,
  `comment_content` text,
  `comment_time` int unsigned NOT NULL DEFAULT 0,
  `comment_status` tinyint NOT NULL DEFAULT 1,
  PRIMARY KEY (`comment_id`),
  KEY `idx_rid_type` (`comment_rid`, `comment_type`),
  KEY `idx_time` (`comment_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `mac_gbook` (
  `gbook_id` int unsigned NOT NULL AUTO_INCREMENT,
  `user_id` int unsigned NOT NULL DEFAULT 0,
  `gbook_content` text,
  `gbook_time` int unsigned NOT NULL DEFAULT 0,
  `gbook_status` tinyint NOT NULL DEFAULT 0,
  `gbook_reply` text,
  `gbook_reply_time` int unsigned NOT NULL DEFAULT 0,
  PRIMARY KEY (`gbook_id`),
  KEY `idx_time` (`gbook_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `mac_topic` (
  `topic_id` int unsigned NOT NULL AUTO_INCREMENT,
  `topic_name` varchar(128) NOT NULL DEFAULT '',
  `topic_en` varchar(128) NOT NULL DEFAULT '',
  `topic_sub` varchar(255) NOT NULL DEFAULT '',
  `topic_pic` varchar(255) NOT NULL DEFAULT '',
  `topic_pic_thumb` varchar(255) NOT NULL DEFAULT '',
  `topic_key` varchar(255) NOT NULL DEFAULT '',
  `topic_des` varchar(512) NOT NULL DEFAULT '',
  `topic_content` text,
  `topic_sort` int NOT NULL DEFAULT 0,
  `topic_level` tinyint NOT NULL DEFAULT 0,
  `topic_lock` tinyint NOT NULL DEFAULT 0,
  `topic_hits` int unsigned NOT NULL DEFAULT 0,
  `topic_status` tinyint NOT NULL DEFAULT 1,
  `topic_time` varchar(64) NOT NULL DEFAULT '',
  `topic_time_add` int unsigned NOT NULL DEFAULT 0,
  `topic_time_make` int unsigned NOT NULL DEFAULT 0,
  PRIMARY KEY (`topic_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `mac_link` (
  `link_id` int unsigned NOT NULL AUTO_INCREMENT,
  `link_name` varchar(128) NOT NULL DEFAULT '',
  `link_url` varchar(255) NOT NULL DEFAULT '',
  `link_logo` varchar(255) NOT NULL DEFAULT '',
  `link_type` tinyint NOT NULL DEFAULT 1,
  `link_sort` int NOT NULL DEFAULT 0,
  `link_status` tinyint NOT NULL DEFAULT 1,
  PRIMARY KEY (`link_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `mac_migrations` (
  `version` int NOT NULL DEFAULT 0
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 插入默认管理员 (admin / admin123)
INSERT INTO `mac_admin` (`admin_name`, `admin_pwd`, `admin_role`, `admin_status`) VALUES
('admin', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 1, 1);

-- 插入默认用户组
INSERT INTO `mac_group` (`group_name`, `group_type`, `group_points`, `group_state`) VALUES
('普通会员', 1, 0, 1),
('VIP会员', 1, 1000, 1);

-- 漫画表
CREATE TABLE IF NOT EXISTS `mac_manga` (
  `manga_id` int unsigned NOT NULL AUTO_INCREMENT,
  `type_id` int unsigned NOT NULL DEFAULT 0,
  `type_id_1` int unsigned NOT NULL DEFAULT 0,
  `manga_name` varchar(255) NOT NULL DEFAULT '',
  `manga_sub` varchar(255) NOT NULL DEFAULT '',
  `manga_en` varchar(128) NOT NULL DEFAULT '',
  `manga_time` varchar(64) NOT NULL DEFAULT '',
  `manga_class` varchar(255) NOT NULL DEFAULT '',
  `manga_tag` varchar(255) NOT NULL DEFAULT '',
  `manga_pic` varchar(255) NOT NULL DEFAULT '',
  `manga_pic_thumb` varchar(255) NOT NULL DEFAULT '',
  `manga_author` varchar(255) NOT NULL DEFAULT '',
  `manga_blurb` varchar(512) NOT NULL DEFAULT '',
  `manga_remarks` varchar(255) NOT NULL DEFAULT '',
  `manga_area` varchar(64) NOT NULL DEFAULT '',
  `manga_lang` varchar(64) NOT NULL DEFAULT '',
  `manga_year` varchar(16) NOT NULL DEFAULT '',
  `manga_content` text,
  `manga_play_from` varchar(512) NOT NULL DEFAULT '',
  `manga_play_url` text,
  `manga_down_from` varchar(512) NOT NULL DEFAULT '',
  `manga_down_url` text,
  `manga_hits` int unsigned NOT NULL DEFAULT 0,
  `manga_hits_day` int unsigned NOT NULL DEFAULT 0,
  `manga_score` varchar(8) NOT NULL DEFAULT '0.0',
  `manga_status` tinyint NOT NULL DEFAULT 1,
  `manga_time_add` int unsigned NOT NULL DEFAULT 0,
  `manga_time_hits` int unsigned NOT NULL DEFAULT 0,
  PRIMARY KEY (`manga_id`),
  KEY `idx_type` (`type_id`),
  KEY `idx_name` (`manga_name`),
  KEY `idx_en` (`manga_en`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 直播表
CREATE TABLE IF NOT EXISTS `mac_live` (
  `live_id` int unsigned NOT NULL AUTO_INCREMENT,
  `type_id` int unsigned NOT NULL DEFAULT 0,
  `live_name` varchar(128) NOT NULL DEFAULT '',
  `live_en` varchar(128) NOT NULL DEFAULT '',
  `live_time` varchar(64) NOT NULL DEFAULT '',
  `live_pic` varchar(255) NOT NULL DEFAULT '',
  `live_url` varchar(512) NOT NULL DEFAULT '',
  `live_from` varchar(64) NOT NULL DEFAULT '',
  `live_sort` int NOT NULL DEFAULT 0,
  `live_level` tinyint NOT NULL DEFAULT 0,
  `live_status` tinyint NOT NULL DEFAULT 1,
  PRIMARY KEY (`live_id`),
  KEY `idx_type` (`type_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 弹幕表
CREATE TABLE IF NOT EXISTS `mac_danmaku` (
  `danmaku_id` int unsigned NOT NULL AUTO_INCREMENT,
  `danmaku_rid` int unsigned NOT NULL DEFAULT 0,
  `danmaku_type` tinyint NOT NULL DEFAULT 1,
  `user_id` int unsigned NOT NULL DEFAULT 0,
  `danmaku_text` varchar(512) NOT NULL DEFAULT '',
  `danmaku_time` int unsigned NOT NULL DEFAULT 0,
  `danmaku_color` varchar(16) NOT NULL DEFAULT '#FFFFFF',
  `danmaku_mode` tinyint NOT NULL DEFAULT 1,
  `danmaku_status` tinyint NOT NULL DEFAULT 1,
  PRIMARY KEY (`danmaku_id`),
  KEY `idx_rid` (`danmaku_rid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
