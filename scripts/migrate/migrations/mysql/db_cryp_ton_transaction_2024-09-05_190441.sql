-- MySQL dump 10.13  Distrib 9.0.1, for macos14.4 (arm64)
--
-- Host: 127.0.0.1    Database: db_cryp_ton
-- ------------------------------------------------------
-- Server version	8.4.0

/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!50503 SET NAMES utf8mb4 */;
/*!40103 SET @OLD_TIME_ZONE=@@TIME_ZONE */;
/*!40103 SET TIME_ZONE='+00:00' */;
/*!40014 SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;

--
-- Table structure for table `transaction`
--

DROP TABLE IF EXISTS `transaction`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `transaction` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '流水號',
  `transaction_time` bigint NOT NULL DEFAULT '0' COMMENT '交易時間',
  `create_time` bigint NOT NULL DEFAULT '0' COMMENT '創建時間',
  `update_time` bigint NOT NULL DEFAULT '0' COMMENT '更新時間',
  `tx_type` tinyint NOT NULL DEFAULT '0' COMMENT '交易類型 1: 充值, 2: 提幣',
  `block_height` bigint NOT NULL DEFAULT '0' COMMENT '區塊高度',
  `transaction_index` int NOT NULL DEFAULT '0' COMMENT '區塊索引',
  `tx_hash` varchar(255) NOT NULL DEFAULT '' COMMENT '交易哈希',
  `crypto_type` varchar(50) NOT NULL DEFAULT '' COMMENT '加密貨幣類型',
  `chain_type` varchar(50) NOT NULL DEFAULT '' COMMENT '鏈類型',
  `contract_addr` varchar(255) NOT NULL DEFAULT '' COMMENT '合約地址',
  `from_address` varchar(255) NOT NULL DEFAULT '' COMMENT '來源地址',
  `to_address` varchar(255) NOT NULL DEFAULT '' COMMENT '帳面目標地址',
  `to_address_real` varchar(255) NOT NULL DEFAULT '' COMMENT '實際目標地址',
  `amount` decimal(30,18) NOT NULL DEFAULT '0.000000000000000000' COMMENT '交易數量',
  `gas` bigint NOT NULL DEFAULT '0' COMMENT '燃料',
  `gas_used` bigint NOT NULL DEFAULT '0' COMMENT '燃料使用',
  `gas_price` decimal(30,18) NOT NULL DEFAULT '0.000000000000000000' COMMENT '燃料價格',
  `fee` decimal(30,18) NOT NULL DEFAULT '0.000000000000000000' COMMENT '交易手續費',
  `fee_crypto` varchar(50) NOT NULL DEFAULT '' COMMENT '手續費使用幣別',
  `confirm` int NOT NULL DEFAULT '0' COMMENT '區塊確認數',
  `status` tinyint NOT NULL DEFAULT '0' COMMENT '狀態 0: 待確認, 1: 鏈上交易成功, 2: 鏈上交易失敗',
  `memo` varchar(255) NOT NULL DEFAULT '' COMMENT '鏈上備註',
  `notify_status` tinyint NOT NULL DEFAULT '0' COMMENT '通知狀態 0: 未處理, 1: 待通知, 2: 通知成功, 3: 通知失敗',
  `risk_control_status` tinyint NOT NULL DEFAULT '0' COMMENT '風控狀態 0: 未處理, 2: 通知成功, 3: 通知失敗',
  `workchain` int NOT NULL DEFAULT '0' COMMENT 'blockIDExt_workchain',
  `shard` bigint NOT NULL DEFAULT '0' COMMENT 'blockIDExt_shard',
  `seq_no` bigint NOT NULL DEFAULT '0' COMMENT 'blockIDExt_seq_no',
  `own_address` varchar(255) NOT NULL DEFAULT '' COMMENT 'fetchedIDs_own_address',
  `lt` bigint NOT NULL DEFAULT '0' COMMENT 'fetchedIDs_lt',
  `jetton_query_id` varchar(255) NOT NULL DEFAULT '' COMMENT 'jetton_query_id 對帳用',
  `next_message_hash` varchar(255) NOT NULL DEFAULT '' COMMENT 'next_message_hash 對帳用',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=45 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='鏈上交易紀錄';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `transaction`
--

/*!40000 ALTER TABLE `transaction` DISABLE KEYS */;
INSERT INTO `transaction` VALUES (40,1725530802,1725533418,1725533454,1,22726255,0,'a2c27b28298f8eaf992304545b4118ccccf53506c313607d12f2c7f04cacdd33','TON','TON','','EQBke4XqvImnG6LHhGUpzsoTszUJqavyA5ZWn2xrWs7hhXPd','EQD8W1_02-LN02dMYGBNVx3EP6fyezVhY5ukL0DiReusuZRf','EQD8W1_02-LN02dMYGBNVx3EP6fyezVhY5ukL0DiReusuZRf',0.290018781000000000,0,0,0.000000000000000000,0.000266669000000000,'TON',12,1,'',1,0,0,-2305843009213693952,24361754,'/Ftf9NvizdNnTGBgTVcdxD+n8ns1YWObpC9A4kXrrLk=',25558964000001,'','RqYN5kw9w766IXai1Z4p87coRBh0gwUDqgJRG5G65Lw='),(41,1725530802,1725533418,1725533455,1,22726255,0,'3c3cdcc0e45d7584c77fba5480d2f27dacd560b4545f3d4a3e204885d98fc567','TON','TON','','EQBke4XqvImnG6LHhGUpzsoTszUJqavyA5ZWn2xrWs7hhXPd','EQDjROxEHR1wVp7QihAIT167kw9HrIlcXbdken8J9NWoq-22','EQDjROxEHR1wVp7QihAIT167kw9HrIlcXbdken8J9NWoq-22',0.289981972000000000,0,0,0.000000000000000000,0.000266669000000000,'TON',12,1,'',1,0,0,-2305843009213693952,24361754,'40TsRB0dcFae0IoQCE9eu5MPR6yJXF23ZHp/CfTVqKs=',25558964000001,'','g22HSFvTCBYZ+Kw+Tb5nnlXzYl4PMHInESeMylr55GI='),(42,1725530802,1725533418,1725533455,2,22726255,0,'71509681da4b1fb0b1b9f3f95935a0c7b7612573ceb17c8e7b14b282db63979f','TON','TON','','','EQDaNL2Gcx5a33WOB2Gm2JoKdXlpBNTaTLfTSR6vltz6Rr3v','EQBVgfn2XwrknMJpWwpwTtFm1OTChtq9w8S2356fxO-oF47Z',0.000000000000000000,0,0,0.000000000000000000,0.002063032000000000,'TON',12,1,'',1,0,0,-2305843009213693952,24361754,'2jS9hnMeWt91jgdhptiaCnV5aQTU2ky300ker5bc+kY=',25558964000001,'','lqKW0iTyhcZ77pPDD4owkVfw2qNdxbh+QQt4YwoJz8c='),(43,1725530802,1725533418,1725533456,1,22726255,0,'97791b32803faeb403dbc280c6cf77b26f6793ecc6d380b29c6cbd53e3e1b0b0','TON','TON','','EQBke4XqvImnG6LHhGUpzsoTszUJqavyA5ZWn2xrWs7hhXPd','EQDZuvsdxrbKWp2412aeIxoDhFBJcnZZIBJ86dm_OlYgzvH5','EQDZuvsdxrbKWp2412aeIxoDhFBJcnZZIBJ86dm_OlYgzvH5',0.289636372000000000,0,0,0.000000000000000000,0.000266669000000000,'TON',12,1,'',1,0,0,-2305843009213693952,24361754,'2br7Hca2ylqduNdmniMaA4RQSXJ2WSASfOnZvzpWIM4=',25558964000001,'','619dWIZDtU92XY6YyZZ8wGPshwNAMMpY9qn9JAnLupw='),(44,1725530810,1725533443,1725533443,2,22726258,0,'347671fcf3c49bda92fa7fb13797134269795134b3088e47aa20f2bbf9caa180','TON','TON','','EQDaNL2Gcx5a33WOB2Gm2JoKdXlpBNTaTLfTSR6vltz6Rr3v','EQBVgfn2XwrknMJpWwpwTtFm1OTChtq9w8S2356fxO-oF47Z','EQBVgfn2XwrknMJpWwpwTtFm1OTChtq9w8S2356fxO-oF47Z',0.100000000000000000,0,0,0.000000000000000000,0.000266669000000000,'TON',0,0,'',0,0,0,6917529027641081856,24393148,'VYH59l8K5JzCaVsKcE7RZtTkwobavcPEtt+en8TvqBc=',25558968000001,'','');
/*!40000 ALTER TABLE `transaction` ENABLE KEYS */;
/*!40103 SET TIME_ZONE=@OLD_TIME_ZONE */;

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;

-- Dump completed on 2024-09-05 19:04:56
