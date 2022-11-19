CREATE TABLE `transfer_alternative`
(
    `id`             int                                                           NOT NULL AUTO_INCREMENT,
    `amount`         decimal(14, 2)                                                NOT NULL COMMENT '收款金额',
    `from_address`   varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '付款人地址',
    `to_address`     varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '收款人地址',
    `block`          int                                                           NOT NULL COMMENT '区块',
    `transaction_id` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '哈希值',
    `time`           datetime                                                      NOT NULL COMMENT '确认时间',
    `updated_at`     datetime                                                      NOT NULL COMMENT '更新时间',
    `created_at`     datetime                                                      NOT NULL COMMENT '创建时间',
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
