CREATE TABLE `config` (
                          `id` int NOT NULL,
                          `start_block` int NOT NULL COMMENT '开始区块',
                          `url` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '请求接口',
                          `status` tinyint(1) NOT NULL COMMENT '采集状态',
                          `first_amount` float NOT NULL COMMENT '首笔转账金额条件',
                          `second_amount` float NOT NULL COMMENT '第二笔转账金额条件',
                          `time_limit` smallint NOT NULL COMMENT '第二笔转账时间内',
                          `amount_condition` float NOT NULL COMMENT '金额条件',
                          `created_at` datetime(3) NOT NULL,
                          `updated_at` datetime(3) NOT NULL,
                          `deleted_at` datetime(3) DEFAULT NULL,
                          `tg_bot_url` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '小飞机机器人接口地址',
                          `tg_id` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '小飞机群id',
                          PRIMARY KEY (`id`),
                          KEY `idx_config_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
