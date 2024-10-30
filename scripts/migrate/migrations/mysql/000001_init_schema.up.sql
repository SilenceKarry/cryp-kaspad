CREATE TABLE `address` (
    `id` bigint(20) AUTO_INCREMENT NOT NULL COMMENT '流水號',
    `create_time` bigint DEFAULT 0 NOT NULL COMMENT '創建時間',
    `update_time` bigint DEFAULT 0 NOT NULL COMMENT '更新時間',
    `merchant_type` tinyint DEFAULT 0 NOT NULL COMMENT '商戶類型',
    `address` varchar(255) DEFAULT '' NOT NULL COMMENT '鏈上地址',
    `chain_type` varchar(50) DEFAULT '' NOT NULL COMMENT '鏈類型',
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='鏈上錢包地址';

create index idx_address on `address` (`address`);

CREATE TABLE `tokens` (
    `id` bigint(20) AUTO_INCREMENT NOT NULL COMMENT '流水號',
    `create_time` bigint DEFAULT 0 NOT NULL COMMENT '創建時間',
    `update_time` bigint DEFAULT 0 NOT NULL COMMENT '更新時間',
    `crypto_type` varchar(50) DEFAULT '' NOT NULL COMMENT '加密貨幣類型',
    `chain_type` varchar(50) DEFAULT '' NOT NULL COMMENT '鏈類型',
    `contract_addr` varchar(255) DEFAULT '' NOT NULL COMMENT '合約地址',
    `decimals` int DEFAULT 0 NOT NULL COMMENT '精準位數',
    `gas_limit` bigint DEFAULT 0 NOT NULL COMMENT '燃料限制',
    `gas_price` decimal(30, 18) DEFAULT 0 NOT NULL COMMENT '燃料價格',
    `contract_abi` text COMMENT '合約 ABI',
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='有支援的 crypto_type';

create index idx_crypto_type on `tokens` (`crypto_type`);

CREATE TABLE `block_height` (
   `id` bigint(20) AUTO_INCREMENT NOT NULL COMMENT '流水號',
   `create_time` bigint DEFAULT 0 NOT NULL COMMENT '創建時間',
   `update_time` bigint DEFAULT 0 NOT NULL COMMENT '更新時間',
   `block_height` bigint DEFAULT 0 NOT NULL COMMENT '區塊高度',
   PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='區塊高度';

CREATE TABLE `transaction` (
    `id` bigint(20) AUTO_INCREMENT NOT NULL COMMENT '流水號',
    `transaction_time` bigint DEFAULT 0 NOT NULL COMMENT '交易時間',
    `create_time` bigint DEFAULT 0 NOT NULL COMMENT '創建時間',
    `update_time` bigint DEFAULT 0 NOT NULL COMMENT '更新時間',
    `tx_type` tinyint DEFAULT 0 NOT NULL COMMENT '交易類型 1: 充值, 2: 提幣',
    `block_height` bigint DEFAULT 0 NOT NULL COMMENT '區塊高度',
    `transaction_index` int DEFAULT 0 NOT NULL COMMENT '區塊索引',
    `tx_hash` varchar(255) DEFAULT '' NOT NULL COMMENT '交易哈希',
    `crypto_type` varchar(50) DEFAULT '' NOT NULL COMMENT '加密貨幣類型',
    `chain_type` varchar(50) DEFAULT '' NOT NULL COMMENT '鏈類型',
    `contract_addr` varchar(255) DEFAULT '' NOT NULL COMMENT '合約地址',
    `from_address` varchar(255) DEFAULT '' NOT NULL COMMENT '來源地址',
    `to_address` varchar(255) DEFAULT '' NOT NULL COMMENT '目標地址',
    `amount` decimal(30, 18) DEFAULT 0 NOT NULL COMMENT '交易數量',
    `gas` bigint DEFAULT 0 NOT NULL COMMENT '燃料',
    `gas_used` bigint DEFAULT 0 NOT NULL COMMENT '燃料使用',
    `gas_price` decimal(30, 18) DEFAULT 0 NOT NULL COMMENT '燃料價格',
    `fee` decimal(30, 18) DEFAULT 0 NOT NULL COMMENT '交易手續費',
    `fee_crypto` varchar(50) DEFAULT '' NOT NULL COMMENT '手續費使用幣別',
    `cpu_usage` int DEFAULT 0 NOT NULL COMMENT 'CPU 使用',
    `net_usage_words` int DEFAULT 0 NOT NULL COMMENT 'NET word 使用',
    `confirm` int DEFAULT 0 NOT NULL COMMENT '區塊確認數',
    `status` tinyint DEFAULT 0 NOT NULL COMMENT '狀態 0: 待確認, 1: 鏈上交易成功, 2: 鏈上交易失敗',
    `memo` varchar(255) DEFAULT '' NOT NULL COMMENT '鏈上備註',
    `notify_status` tinyint DEFAULT 0 NOT NULL COMMENT '通知狀態 0: 未處理, 1: 待通知, 2: 通知成功, 3: 通知失敗',
    `risk_control_status` tinyint(4) DEFAULT '0' NOT NULL COMMENT '風控狀態 0: 未處理, 2: 通知成功, 3: 通知失敗',
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='鏈上交易紀錄';

create index idx_tx_hash on `transaction` (`tx_hash`);
create index idx_status_block_height on `transaction` (`status`, `block_height`);
create index idx_notify_status on `transaction` (`notify_status`);

CREATE TABLE `withdraw` (
    `id` bigint(20) AUTO_INCREMENT NOT NULL COMMENT '流水號',
    `create_time` bigint DEFAULT 0 NOT NULL COMMENT '創建時間',
    `update_time` bigint DEFAULT 0 NOT NULL COMMENT '更新時間',
    `merchant_type` tinyint DEFAULT 0 NOT NULL COMMENT '商戶類型',
    `tx_hash` varchar(255) DEFAULT '' NOT NULL COMMENT '交易哈希',
    `crypto_type` varchar(50) DEFAULT '' NOT NULL COMMENT '加密貨幣類型',
    `chain_type` varchar(50) DEFAULT '' NOT NULL COMMENT '鏈類型',
    `from_address` varchar(255) DEFAULT '' NOT NULL COMMENT '來源地址',
    `to_address` varchar(255) DEFAULT '' NOT NULL COMMENT '目標地址',
    `amount` decimal(30, 18) DEFAULT 0 NOT NULL COMMENT '提幣數量',
    `memo` varchar(255) DEFAULT '' NOT NULL COMMENT '鏈上備註',
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='提幣請求紀錄';

create index idx_tx_hash on `withdraw` (`tx_hash`);

CREATE TABLE `config` (
    `id` bigint(20) AUTO_INCREMENT NOT NULL COMMENT '流水號',
    `create_time` bigint DEFAULT 0 NOT NULL COMMENT '創建時間',
    `update_time` bigint DEFAULT 0 NOT NULL COMMENT '更新時間',
    `key` varchar(255) DEFAULT '' NOT NULL COMMENT 'viperkey',
    `value` varchar(255) DEFAULT '' NOT NULL COMMENT 'vipervalue',
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='可變動設定檔';

ALTER TABLE `db_cryp_kaspad`.`withdraw` ADD COLUMN `nonce` BIGINT NOT NULL DEFAULT 0 AFTER `to_address`;
ALTER TABLE `db_cryp_kaspad`.`tokens` ADD COLUMN `transaction_fee` decimal(30, 18) default 0.000000000000000000 not null comment '交易手續費';
create index from_nonce_index on `db_cryp_kaspad`.`withdraw` (`from_address`, `nonce`);

ALTER TABLE `db_cryp_kaspad`.withdraw ADD COLUMN `has_retried` tinyint  default 0    not null comment '是否已經重新提幣';
ALTER TABLE `db_cryp_kaspad`.withdraw ADD COLUMN `has_chain` tinyint  default 0    not null comment '是否上鏈';

-- dev 測試資料
-- insert into `db_cryp_kaspad`.tokens(create_time, update_time, crypto_type, chain_type, contract_addr, decimals, gas_limit, gas_price, contract_abi)
-- values (UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 'EOS', 'EOS', '', 18, 21000, 10000000000, ''),
--        (UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 'USDC', 'EOS', '0x07865c6E87B9F70255377e024ace6630C1Eaa37F', 6, 100000, 20000000000, '[{"constant":true,"inputs":[],"name":"mintingFinished","outputs":[{"name":"","type":"bool"}],"payable":false,"type":"function"},{"constant":true,"inputs":[],"name":"name","outputs":[{"name":"","type":"string"}],"payable":false,"type":"function"},{"constant":false,"inputs":[{"name":"_spender","type":"address"},{"name":"_value","type":"uint256"}],"name":"approve","outputs":[],"payable":false,"type":"function"},{"constant":true,"inputs":[],"name":"totalSupply","outputs":[{"name":"","type":"uint256"}],"payable":false,"type":"function"},{"constant":false,"inputs":[{"name":"_from","type":"address"},{"name":"_to","type":"address"},{"name":"_value","type":"uint256"}],"name":"transferFrom","outputs":[],"payable":false,"type":"function"},{"constant":true,"inputs":[],"name":"decimals","outputs":[{"name":"","type":"uint256"}],"payable":false,"type":"function"},{"constant":false,"inputs":[],"name":"unpause","outputs":[{"name":"","type":"bool"}],"payable":false,"type":"function"},{"constant":false,"inputs":[{"name":"_to","type":"address"},{"name":"_amount","type":"uint256"}],"name":"mint","outputs":[{"name":"","type":"bool"}],"payable":false,"type":"function"},{"constant":true,"inputs":[],"name":"paused","outputs":[{"name":"","type":"bool"}],"payable":false,"type":"function"},{"constant":true,"inputs":[{"name":"_owner","type":"address"}],"name":"balanceOf","outputs":[{"name":"balance","type":"uint256"}],"payable":false,"type":"function"},{"constant":false,"inputs":[],"name":"finishMinting","outputs":[{"name":"","type":"bool"}],"payable":false,"type":"function"},{"constant":false,"inputs":[],"name":"pause","outputs":[{"name":"","type":"bool"}],"payable":false,"type":"function"},{"constant":true,"inputs":[],"name":"owner","outputs":[{"name":"","type":"address"}],"payable":false,"type":"function"},{"constant":true,"inputs":[],"name":"symbol","outputs":[{"name":"","type":"string"}],"payable":false,"type":"function"},{"constant":false,"inputs":[{"name":"_to","type":"address"},{"name":"_value","type":"uint256"}],"name":"transfer","outputs":[],"payable":false,"type":"function"},{"constant":false,"inputs":[{"name":"_to","type":"address"},{"name":"_amount","type":"uint256"},{"name":"_releaseTime","type":"uint256"}],"name":"mintTimelocked","outputs":[{"name":"","type":"address"}],"payable":false,"type":"function"},{"constant":true,"inputs":[{"name":"_owner","type":"address"},{"name":"_spender","type":"address"}],"name":"allowance","outputs":[{"name":"remaining","type":"uint256"}],"payable":false,"type":"function"},{"constant":false,"inputs":[{"name":"newOwner","type":"address"}],"name":"transferOwnership","outputs":[],"payable":false,"type":"function"},{"anonymous":false,"inputs":[{"indexed":true,"name":"to","type":"address"},{"indexed":false,"name":"value","type":"uint256"}],"name":"Mint","type":"event"},{"anonymous":false,"inputs":[],"name":"MintFinished","type":"event"},{"anonymous":false,"inputs":[],"name":"Pause","type":"event"},{"anonymous":false,"inputs":[],"name":"Unpause","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"name":"owner","type":"address"},{"indexed":true,"name":"spender","type":"address"},{"indexed":false,"name":"value","type":"uint256"}],"name":"Approval","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"name":"from","type":"address"},{"indexed":true,"name":"to","type":"address"},{"indexed":false,"name":"value","type":"uint256"}],"name":"Transfer","type":"event"}]')
-- ;

-- prod/online 資料, online merchant_type 要記得改成 2
insert into `db_cryp_kaspad`.tokens(create_time, update_time, crypto_type, chain_type, contract_addr, decimals, gas_limit, gas_price, contract_abi)
values (UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 'EOS', 'EOS', 'eosio.token', 4, 21000, 800000000000, ''),
(UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 'YOZI', 'EOS', 'bonkbonk1234', 7, 21000, 800000000000, '');

insert into `db_cryp_kaspad`.address(create_time, update_time, merchant_type, address, chain_type)
values (UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 1, 'shikanokopay', 'EOS');

insert into `db_cryp_kaspad`.address(create_time, update_time, merchant_type, address, chain_type)
values (UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 1, 'eosexdeposit', 'EOS');

insert into `db_cryp_kaspad`.address(create_time, update_time, merchant_type, address, chain_type)
values (UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 1, 'eosexpayment', 'EOS');

insert into `db_cryp_kaspad`.address(create_time, update_time, merchant_type, address, chain_type)
values (UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 1, 'eosexpool111', 'EOS');