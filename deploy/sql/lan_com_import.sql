-- Baota import for database lan_com
-- Select DB lan_com in panel, then import this file
-- Do NOT create database lanping

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS=0;


/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
 SET NAMES utf8mb4 ;
/*!40103 SET @OLD_TIME_ZONE=@@TIME_ZONE */;
/*!40103 SET TIME_ZONE='+00:00' */;
/*!40014 SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;
DROP TABLE IF EXISTS `goose_db_version`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
 SET character_set_client = utf8mb4 ;
CREATE TABLE `goose_db_version` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `version_id` bigint(20) NOT NULL,
  `is_applied` tinyint(1) NOT NULL,
  `tstamp` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`)
) ENGINE=MyISAM AUTO_INCREMENT=8 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

LOCK TABLES `goose_db_version` WRITE;
/*!40000 ALTER TABLE `goose_db_version` DISABLE KEYS */;
INSERT INTO `goose_db_version` VALUES (1,0,1,'2026-11-09 09:06:54'),(2,1,1,'2026-11-09 09:07:04'),(3,2,1,'2026-11-09 09:07:05'),(4,3,1,'2026-11-09 09:33:06'),(5,4,1,'2026-11-09 09:47:59'),(6,5,1,'2026-11-10 11:18:01'),(7,6,1,'2026-11-10 11:50:09');
/*!40000 ALTER TABLE `goose_db_version` ENABLE KEYS */;
UNLOCK TABLES;
DROP TABLE IF EXISTS `lp_appointment`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
 SET character_set_client = utf8mb4 ;
CREATE TABLE `lp_appointment` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `customer_id` bigint(20) NOT NULL,
  `project_id` bigint(20) NOT NULL,
  `store_id` bigint(20) NOT NULL,
  `staff_id` bigint(20) DEFAULT NULL,
  `appoint_date` date NOT NULL,
  `appoint_time` varchar(8) NOT NULL,
  `status` varchar(16) NOT NULL DEFAULT '待确认',
  `source` varchar(16) NOT NULL DEFAULT '门店',
  `note` varchar(255) NOT NULL DEFAULT '',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_store_date` (`store_id`,`appoint_date`)
) ENGINE=InnoDB AUTO_INCREMENT=6 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

LOCK TABLES `lp_appointment` WRITE;
/*!40000 ALTER TABLE `lp_appointment` DISABLE KEYS */;
INSERT INTO `lp_appointment` VALUES (1,1,1,2,2,'2026-11-10','19:18','已取消','门店','smoke','2026-11-10 19:18:05','2026-11-10 19:30:44'),(2,16,6,1,4,'2026-11-10','19:28','已取消','门店','','2026-11-10 19:28:55','2026-11-10 19:29:05'),(3,16,6,1,5,'2026-11-10','19:29','已取消','门店','','2026-11-10 19:29:27','2026-11-10 19:30:38'),(4,16,6,1,4,'2026-11-10','19:30','已确认','门店','','2026-11-10 19:30:52','2026-11-10 19:30:55'),(5,16,6,1,8,'2026-11-10','19:53','待确认','门店','','2026-11-10 19:54:04','2026-11-10 19:54:04');
/*!40000 ALTER TABLE `lp_appointment` ENABLE KEYS */;
UNLOCK TABLES;
DROP TABLE IF EXISTS `lp_assessment`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
 SET character_set_client = utf8mb4 ;
CREATE TABLE `lp_assessment` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `user_id` bigint(20) NOT NULL,
  `store_id` bigint(20) NOT NULL,
  `month_key` varchar(7) NOT NULL,
  `name` varchar(64) NOT NULL,
  `weight` decimal(8,2) NOT NULL DEFAULT '1.00',
  `target_value` decimal(12,2) NOT NULL DEFAULT '0.00',
  `done_value` decimal(12,2) NOT NULL DEFAULT '0.00',
  `passed` tinyint(4) NOT NULL DEFAULT '0',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_month_store` (`month_key`,`store_id`)
) ENGINE=InnoDB AUTO_INCREMENT=5 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

LOCK TABLES `lp_assessment` WRITE;
/*!40000 ALTER TABLE `lp_assessment` DISABLE KEYS */;
INSERT INTO `lp_assessment` VALUES (1,4,1,'2026-11','月度业绩（元）',1.00,8000.00,9200.00,1,'2026-11-09 17:07:05','2026-11-09 17:07:05'),(2,5,1,'2026-11','月度业绩（元）',1.00,8000.00,6500.00,0,'2026-11-09 17:07:05','2026-11-09 17:07:05'),(3,6,2,'2026-11','月度业绩（元）',1.00,8000.00,8800.00,1,'2026-11-09 17:07:05','2026-11-09 17:07:05'),(4,2,2,'2026-11','月度业绩（元）',1.00,20000.00,21000.00,1,'2026-11-09 17:07:05','2026-11-09 17:07:05');
/*!40000 ALTER TABLE `lp_assessment` ENABLE KEYS */;
UNLOCK TABLES;
DROP TABLE IF EXISTS `lp_attendance`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
 SET character_set_client = utf8mb4 ;
CREATE TABLE `lp_attendance` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `user_id` bigint(20) NOT NULL,
  `store_id` bigint(20) NOT NULL,
  `work_date` date NOT NULL,
  `check_in` datetime DEFAULT NULL,
  `check_out` datetime DEFAULT NULL,
  `method` varchar(16) NOT NULL DEFAULT '定位',
  `status` varchar(16) NOT NULL DEFAULT '正常',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_date` (`user_id`,`work_date`)
) ENGINE=InnoDB AUTO_INCREMENT=3 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

LOCK TABLES `lp_attendance` WRITE;
/*!40000 ALTER TABLE `lp_attendance` DISABLE KEYS */;
INSERT INTO `lp_attendance` VALUES (1,2,2,'2026-11-09','2026-11-09 18:09:58','2026-11-09 18:10:01','定位','正常','2026-11-09 18:09:58'),(2,6,2,'2026-11-10','2026-11-10 19:07:28',NULL,'定位','正常','2026-11-10 19:07:28');
/*!40000 ALTER TABLE `lp_attendance` ENABLE KEYS */;
UNLOCK TABLES;
DROP TABLE IF EXISTS `lp_audit_apply`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
 SET character_set_client = utf8mb4 ;
CREATE TABLE `lp_audit_apply` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `name` varchar(32) NOT NULL,
  `phone` varchar(20) NOT NULL,
  `store_id` bigint(20) NOT NULL,
  `role` varchar(16) NOT NULL DEFAULT 'staff',
  `status` varchar(16) NOT NULL DEFAULT 'pending',
  `apply_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `reviewed_at` datetime DEFAULT NULL,
  `reviewer_id` bigint(20) DEFAULT NULL,
  `remark` varchar(255) NOT NULL DEFAULT '',
  PRIMARY KEY (`id`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB AUTO_INCREMENT=5 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

LOCK TABLES `lp_audit_apply` WRITE;
/*!40000 ALTER TABLE `lp_audit_apply` DISABLE KEYS */;
INSERT INTO `lp_audit_apply` VALUES (1,'何佳','13800009901',2,'staff','待审核','2026-11-09 17:07:05',NULL,NULL,''),(2,'林晓','13800009902',1,'staff','待审核','2026-11-09 17:07:05',NULL,NULL,''),(3,'测试注册','13970001111',2,'staff','待审核','2026-11-10 19:21:29',NULL,NULL,''),(4,'刘菲菲','18874550001',1,'staff','已通过','2026-11-10 19:40:28','2026-11-10 19:42:02',1,'');
/*!40000 ALTER TABLE `lp_audit_apply` ENABLE KEYS */;
UNLOCK TABLES;
DROP TABLE IF EXISTS `lp_card_flow`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
 SET character_set_client = utf8mb4 ;
CREATE TABLE `lp_card_flow` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `card_id` bigint(20) NOT NULL,
  `customer_id` bigint(20) NOT NULL,
  `change_times` int(11) NOT NULL DEFAULT '0',
  `change_amount` decimal(12,2) NOT NULL DEFAULT '0.00',
  `biz_type` varchar(32) NOT NULL,
  `biz_no` varchar(64) NOT NULL DEFAULT '',
  `remark` varchar(255) NOT NULL DEFAULT '',
  `operator_id` bigint(20) DEFAULT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_card` (`card_id`)
) ENGINE=InnoDB AUTO_INCREMENT=7 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

LOCK TABLES `lp_card_flow` WRITE;
/*!40000 ALTER TABLE `lp_card_flow` DISABLE KEYS */;
INSERT INTO `lp_card_flow` VALUES (3,9,16,8,3664.00,'open','CD3857456900','开卡',1,'2026-11-10 18:56:23'),(4,9,16,-1,458.00,'verify','HX202611101252','核销确认',NULL,'2026-11-10 19:47:54'),(5,10,16,5,500.00,'open','CD3262833100','开卡',1,'2026-11-10 19:50:23'),(6,11,16,0,1000.00,'open','CD3471788300','充值开卡',1,'2026-11-10 19:50:23');
/*!40000 ALTER TABLE `lp_card_flow` ENABLE KEYS */;
UNLOCK TABLES;
DROP TABLE IF EXISTS `lp_customer`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
 SET character_set_client = utf8mb4 ;
CREATE TABLE `lp_customer` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `customer_no` varchar(32) NOT NULL,
  `name` varchar(32) NOT NULL,
  `mobile` varchar(20) NOT NULL,
  `gender` tinyint(4) NOT NULL DEFAULT '0',
  `birthday` date DEFAULT NULL,
  `level_name` varchar(32) NOT NULL DEFAULT '银卡会员',
  `source` varchar(32) NOT NULL DEFAULT '到店',
  `store_id` bigint(20) NOT NULL,
  `employee_id` bigint(20) DEFAULT NULL,
  `total_amount` decimal(12,2) NOT NULL DEFAULT '0.00',
  `last_visit_at` datetime DEFAULT NULL,
  `status` tinyint(4) NOT NULL DEFAULT '1',
  `remark` varchar(500) NOT NULL DEFAULT '',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_customer_no` (`customer_no`),
  UNIQUE KEY `uk_mobile` (`mobile`),
  KEY `idx_store` (`store_id`)
) ENGINE=InnoDB AUTO_INCREMENT=17 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

LOCK TABLES `lp_customer` WRITE;
/*!40000 ALTER TABLE `lp_customer` DISABLE KEYS */;
INSERT INTO `lp_customer` VALUES (1,'LP2025000001','王小美','13900000001',2,'1995-04-28','钻石会员','到店',1,4,12680.00,'2026-11-08 17:07:05',1,'','2026-11-09 17:07:05','2026-11-09 17:07:05',NULL),(2,'LP2025000002','李婷婷','13900000002',2,'1993-04-29','金卡会员','转介绍',1,5,6420.00,'2026-11-03 17:07:05',1,'','2026-11-09 17:07:05','2026-11-09 17:07:05',NULL),(3,'LP2025000003','陈静','13900000003',2,'1998-05-01','银卡会员','线上',2,6,2180.00,'2026-10-01 17:07:05',1,'','2026-11-09 17:07:05','2026-11-09 17:07:05',NULL),(4,'LP2025000004','周雪','13900000004',2,'1992-07-12','钻石会员','到店',2,6,18960.00,'2026-11-07 17:07:05',1,'','2026-11-09 17:07:05','2026-11-09 17:07:05',NULL),(5,'LP2025000005','赵梦琪','13900000005',2,'1996-03-15','金卡会员','活动',3,3,8600.00,'2026-10-28 17:07:05',1,'','2026-11-09 17:07:05','2026-11-09 17:07:05',NULL),(6,'LP2025000006','孙悦','13900000006',2,'1994-09-08','银卡会员','到店',1,4,3200.00,'2026-11-04 17:07:05',1,'','2026-11-09 17:07:05','2026-11-09 17:07:05',NULL),(14,'LP770196274800','刘','15578871111',0,NULL,'银卡会员','到店',2,NULL,0.00,NULL,1,'','2026-11-09 18:56:10','2026-11-09 18:56:10',NULL),(15,'LP129820998000','1','2',0,NULL,'银卡会员','到店',2,NULL,0.00,NULL,1,'3','2026-11-10 18:55:29','2026-11-10 18:55:32','2026-11-10 18:55:32'),(16,'LP145121110800','陈佳佳','18874580001',0,NULL,'银卡会员','到店',1,NULL,5164.00,'2026-11-10 19:47:54',1,'','2026-11-10 18:55:45','2026-11-10 19:50:23',NULL);
/*!40000 ALTER TABLE `lp_customer` ENABLE KEYS */;
UNLOCK TABLES;
DROP TABLE IF EXISTS `lp_customer_card`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
 SET character_set_client = utf8mb4 ;
CREATE TABLE `lp_customer_card` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `card_no` varchar(32) NOT NULL,
  `customer_id` bigint(20) NOT NULL,
  `project_id` bigint(20) DEFAULT NULL,
  `store_id` bigint(20) NOT NULL,
  `employee_id` bigint(20) DEFAULT NULL,
  `total_times` int(11) NOT NULL DEFAULT '0',
  `used_times` int(11) NOT NULL DEFAULT '0',
  `remain_times` int(11) NOT NULL DEFAULT '0',
  `frozen_times` int(11) NOT NULL DEFAULT '0',
  `total_amount` decimal(12,2) NOT NULL DEFAULT '0.00',
  `pay_method` varchar(16) NOT NULL DEFAULT '微信',
  `card_kind` varchar(16) NOT NULL DEFAULT 'times' COMMENT 'times次卡/amount储值',
  `status` varchar(16) NOT NULL DEFAULT '使用中',
  `version` int(11) NOT NULL DEFAULT '0',
  `valid_from` date DEFAULT NULL,
  `valid_to` date DEFAULT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_card_no` (`card_no`),
  KEY `idx_customer` (`customer_id`),
  KEY `idx_store` (`store_id`)
) ENGINE=InnoDB AUTO_INCREMENT=12 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

LOCK TABLES `lp_customer_card` WRITE;
/*!40000 ALTER TABLE `lp_customer_card` DISABLE KEYS */;
INSERT INTO `lp_customer_card` VALUES (1,'CD20250001',1,1,1,4,10,2,8,0,3980.00,'微信','times','使用中',0,'2026-10-10','2027-10-10','2026-11-09 17:07:05','2026-11-09 17:07:05'),(2,'CD20250002',2,2,1,5,10,3,7,0,6980.00,'刷卡','times','使用中',0,'2026-09-30','2027-09-30','2026-11-09 17:07:05','2026-11-09 17:07:05'),(3,'CD20250003',3,3,2,6,5,1,4,0,9900.00,'微信','times','使用中',0,'2026-10-20','2027-10-20','2026-11-09 17:07:05','2026-11-09 17:07:05'),(4,'CD20250004',4,2,2,6,10,5,5,0,6980.00,'微信','times','使用中',0,'2026-09-10','2027-09-10','2026-11-09 17:07:05','2026-11-09 17:07:05'),(5,'CD20250005',5,4,3,3,10,2,8,0,2980.00,'刷卡','times','使用中',0,'2026-10-25','2027-10-25','2026-11-09 17:07:05','2026-11-09 17:07:05'),(6,'CD20250006',6,1,1,4,10,0,10,0,3980.00,'微信','times','使用中',0,'2026-11-09','2027-11-09','2026-11-09 17:07:05','2026-11-09 17:07:05'),(9,'CD3857456900',16,5,1,1,8,1,7,0,3664.00,'微信','times','使用中',2,'2026-11-10','2027-11-10','2026-11-10 18:56:23','2026-11-10 19:47:54'),(10,'CD3262833100',16,1,1,1,5,0,5,1,500.00,'??','times','使用中',1,'2026-11-10','2027-11-10','2026-11-10 19:50:23','2026-11-11 13:02:32'),(11,'CD3471788300',16,NULL,1,1,0,0,0,0,1000.00,'??','amount','使用中',0,'2026-11-10','2027-11-10','2026-11-10 19:50:23','2026-11-10 19:50:23');
/*!40000 ALTER TABLE `lp_customer_card` ENABLE KEYS */;
UNLOCK TABLES;
DROP TABLE IF EXISTS `lp_message`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
 SET character_set_client = utf8mb4 ;
CREATE TABLE `lp_message` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `user_id` bigint(20) NOT NULL,
  `title` varchar(128) NOT NULL,
  `content` varchar(500) NOT NULL DEFAULT '',
  `is_read` tinyint(4) NOT NULL DEFAULT '0',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_user` (`user_id`)
) ENGINE=InnoDB AUTO_INCREMENT=3 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

LOCK TABLES `lp_message` WRITE;
/*!40000 ALTER TABLE `lp_message` DISABLE KEYS */;
INSERT INTO `lp_message` VALUES (1,4,'欢迎使用兰瓶门店端','今日待办与核销入口已就绪',0,'2026-11-09 17:07:05'),(2,2,'库存提醒','万象城店水杨酸精华库存偏低',0,'2026-11-09 17:07:05');
/*!40000 ALTER TABLE `lp_message` ENABLE KEYS */;
UNLOCK TABLES;
DROP TABLE IF EXISTS `lp_operate_log`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
 SET character_set_client = utf8mb4 ;
CREATE TABLE `lp_operate_log` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `user_id` bigint(20) DEFAULT NULL,
  `module` varchar(64) NOT NULL,
  `action` varchar(64) NOT NULL,
  `detail` varchar(1000) NOT NULL DEFAULT '',
  `ip` varchar(64) NOT NULL DEFAULT '',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_time` (`created_at`)
) ENGINE=InnoDB AUTO_INCREMENT=54 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

LOCK TABLES `lp_operate_log` WRITE;
/*!40000 ALTER TABLE `lp_operate_log` DISABLE KEYS */;
INSERT INTO `lp_operate_log` VALUES (1,2,'auth','login','李静','127.0.0.1','2026-11-09 17:13:14'),(2,2,'customer','create','????','127.0.0.1','2026-11-09 17:13:14'),(3,2,'auth','login','李静','127.0.0.1','2026-11-09 17:13:29'),(4,2,'card','open','CD9621619400','127.0.0.1','2026-11-09 17:13:29'),(5,1,'auth','login','张美丽','127.0.0.1','2026-11-09 17:13:29'),(6,2,'auth','login','李静','127.0.0.1','2026-11-09 17:14:03'),(7,2,'auth','login','李静','127.0.0.1','2026-11-09 17:14:29'),(8,4,'auth','login','陈晓','127.0.0.1','2026-11-09 17:14:30'),(9,1,'auth','login','张美丽','127.0.0.1','2026-11-09 17:14:30'),(10,2,'customer','create','????','127.0.0.1','2026-11-09 17:14:30'),(11,2,'card','open','CD692177100','127.0.0.1','2026-11-09 17:14:30'),(12,4,'customer','create','staff?','127.0.0.1','2026-11-09 17:14:31'),(13,2,'auth','login','李静','127.0.0.1','2026-11-09 17:20:59'),(14,1,'auth','login','张美丽','127.0.0.1','2026-11-09 17:33:09'),(15,1,'auth','login','张美丽','127.0.0.1','2026-11-09 17:48:00'),(16,1,'auth','login','张美丽','127.0.0.1','2026-11-09 18:00:58'),(17,1,'auth','login','张美丽','127.0.0.1','2026-11-09 18:03:42'),(18,1,'auth','login','张美丽','127.0.0.1','2026-11-09 18:04:10'),(19,1,'auth','login','张美丽','127.0.0.1','2026-11-09 18:08:05'),(20,1,'auth','login','张美丽','127.0.0.1','2026-11-09 18:08:35'),(21,1,'auth','login','张美丽','127.0.0.1','2026-11-09 18:09:03'),(22,1,'auth','login','张美丽','127.0.0.1','2026-11-09 18:10:36'),(23,2,'auth','login','李静','127.0.0.1','2026-11-09 18:11:19'),(24,2,'customer','create','?????','127.0.0.1','2026-11-09 18:11:19'),(25,2,'auth','login','李静','127.0.0.1','2026-11-09 18:12:06'),(26,2,'customer','create','测试客户乙API','127.0.0.1','2026-11-09 18:12:07'),(27,2,'customer','create','刘','127.0.0.1','2026-11-09 18:56:10'),(28,1,'auth','login','张美丽','127.0.0.1','2026-11-10 18:52:05'),(29,1,'auth','login','张美丽','127.0.0.1','2026-11-10 18:53:05'),(30,1,'customer','create','1','127.0.0.1','2026-11-10 18:55:29'),(31,1,'customer','create','陈佳佳','127.0.0.1','2026-11-10 18:55:45'),(32,1,'card','open','CD3857456900','127.0.0.1','2026-11-10 18:56:24'),(33,2,'auth','login','李静','127.0.0.1','2026-11-10 18:58:53'),(34,6,'auth','login','赵敏','127.0.0.1','2026-11-10 18:59:05'),(35,2,'auth','login','李静','127.0.0.1','2026-11-10 19:11:34'),(36,2,'auth','login','李静','127.0.0.1','2026-11-10 19:18:05'),(37,1,'auth','login','张美丽','127.0.0.1','2026-11-10 19:21:29'),(38,1,'user','create','新店长/manager','127.0.0.1','2026-11-10 19:21:30'),(39,5,'auth','login','刘婷','127.0.0.1','2026-11-10 19:26:43'),(40,1,'auth','login','张美丽','127.0.0.1','2026-11-10 19:40:41'),(41,2,'auth','login','李静','127.0.0.1','2026-11-10 19:40:41'),(42,8,'auth','login','刘菲菲','127.0.0.1','2026-11-10 19:42:08'),(43,1,'auth','login','张美丽','127.0.0.1','2026-11-10 19:50:22'),(44,1,'card','open','CD3262833100','127.0.0.1','2026-11-10 19:50:23'),(45,1,'card','open','CD3471788300','127.0.0.1','2026-11-10 19:50:23'),(46,2,'auth','login','李静','127.0.0.1','2026-11-10 20:03:16'),(47,2,'auth','login','李静','127.0.0.1','2026-11-10 20:09:14'),(48,2,'inventory','create_item','??????','127.0.0.1','2026-11-10 20:09:15'),(49,2,'inventory','update_item','?????????','127.0.0.1','2026-11-10 20:09:15'),(50,2,'inventory','order','??','127.0.0.1','2026-11-10 20:09:15'),(51,1,'inventory','create_item','是非得失','127.0.0.1','2026-11-10 20:11:33'),(52,1,'inventory','create_item','玻尿酸补水面膜','127.0.0.1','2026-11-10 20:14:00'),(53,1,'auth','login','张美丽','127.0.0.1','2026-11-11 12:57:21');
/*!40000 ALTER TABLE `lp_operate_log` ENABLE KEYS */;
UNLOCK TABLES;
DROP TABLE IF EXISTS `lp_project`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
 SET character_set_client = utf8mb4 ;
CREATE TABLE `lp_project` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `name` varchar(64) NOT NULL,
  `category` varchar(32) NOT NULL DEFAULT '',
  `price` decimal(12,2) NOT NULL DEFAULT '0.00',
  `member_price` decimal(12,2) NOT NULL DEFAULT '0.00',
  `total_times` int(11) NOT NULL DEFAULT '1',
  `valid_days` int(11) NOT NULL DEFAULT '365',
  `commission_rate` decimal(8,4) NOT NULL DEFAULT '0.0200',
  `duration` int(11) NOT NULL DEFAULT '60',
  `status` varchar(16) NOT NULL DEFAULT '上架',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=7 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

LOCK TABLES `lp_project` WRITE;
/*!40000 ALTER TABLE `lp_project` DISABLE KEYS */;
INSERT INTO `lp_project` VALUES (1,'小气泡深层清洁','面部清洁',398.00,358.00,10,365,0.0200,60,'上架','2026-11-09 17:07:05','2026-11-09 17:07:05'),(2,'水光补水护理','面部护理',698.00,628.00,10,365,0.0200,75,'上架','2026-11-09 17:07:05','2026-11-09 17:07:05'),(3,'光子嫩肤','光电项目',1980.00,1780.00,5,365,0.0300,45,'上架','2026-11-09 17:07:05','2026-11-09 17:07:05'),(4,'肩颈舒缓','身体护理',298.00,268.00,10,365,0.0200,45,'上架','2026-11-09 17:07:05','2026-11-09 17:07:05'),(5,'背部精油SPA','身体护理',458.00,418.00,8,365,0.0200,60,'上架','2026-11-09 17:07:05','2026-11-09 17:07:05'),(6,'热玛吉紧致','光电项目',6800.00,6200.00,3,540,0.0400,90,'上架','2026-11-09 17:07:05','2026-11-09 17:07:05');
/*!40000 ALTER TABLE `lp_project` ENABLE KEYS */;
UNLOCK TABLES;
DROP TABLE IF EXISTS `lp_salary_detail`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
 SET character_set_client = utf8mb4 ;
CREATE TABLE `lp_salary_detail` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `month_key` varchar(7) NOT NULL,
  `user_id` bigint(20) NOT NULL,
  `store_id` bigint(20) NOT NULL,
  `base_salary` decimal(12,2) NOT NULL DEFAULT '0.00',
  `commission` decimal(12,2) NOT NULL DEFAULT '0.00',
  `assess_bonus` decimal(12,2) NOT NULL DEFAULT '0.00',
  `attendance_deduct` decimal(12,2) NOT NULL DEFAULT '0.00',
  `total_payable` decimal(12,2) NOT NULL DEFAULT '0.00',
  `status` varchar(16) NOT NULL DEFAULT '已计算',
  `calc_snapshot` json DEFAULT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_month_user` (`month_key`,`user_id`)
) ENGINE=InnoDB AUTO_INCREMENT=18 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

LOCK TABLES `lp_salary_detail` WRITE;
/*!40000 ALTER TABLE `lp_salary_detail` DISABLE KEYS */;
INSERT INTO `lp_salary_detail` VALUES (1,'2026-11',2,2,3500.00,0.00,0.00,135.04,3364.96,'已计算','{\"note\": \"店长：固定3500 + 开卡阶梯提成 + 核销5%；含业绩上下限\", \"sales_rate\": 0.02, \"store_sales\": 16880, \"verify_rate\": 0.05, \"fixed_salary\": 3500, \"sales_amount\": 0, \"total_payable\": 3364.96, \"verify_amount\": 0, \"manager_penalty\": 135.04, \"sales_cap_ratio\": 0.6, \"sales_commission\": 0, \"sales_amount_used\": 0, \"verify_commission\": 0, \"working_headcount\": 0}','2026-11-09 17:33:09','2026-11-11 12:59:40'),(2,'2026-11',3,3,3500.00,35.76,0.00,0.00,3535.76,'已计算','{\"note\": \"店长：固定3500 + 开卡阶梯提成 + 核销5%；含业绩上下限\", \"sales_rate\": 0.02, \"store_sales\": 2980, \"verify_rate\": 0.05, \"fixed_salary\": 3500, \"sales_amount\": 2980, \"total_payable\": 3535.76, \"verify_amount\": 0, \"manager_penalty\": 0, \"sales_cap_ratio\": 0.6, \"sales_commission\": 35.76, \"sales_amount_used\": 1788, \"verify_commission\": 0, \"working_headcount\": 0}','2026-11-09 17:33:09','2026-11-09 18:01:45'),(3,'2026-11',4,1,3000.00,159.20,0.00,0.00,3159.20,'已计算','{\"note\": \"店员：固定3000 + 开卡阶梯提成 + 核销5%\", \"sales_rate\": 0.02, \"store_sales\": 20104, \"verify_rate\": 0.05, \"fixed_salary\": 3000, \"sales_amount\": 7960, \"total_payable\": 3159.2, \"verify_amount\": 0, \"manager_penalty\": 0, \"sales_cap_ratio\": 0, \"sales_commission\": 159.2, \"sales_amount_used\": 7960, \"verify_commission\": 0, \"working_headcount\": 0}','2026-11-09 17:33:09','2026-11-11 12:59:41'),(4,'2026-11',5,1,3000.00,139.60,0.00,0.00,3139.60,'已计算','{\"note\": \"店员：固定3000 + 开卡阶梯提成 + 核销5%\", \"sales_rate\": 0.02, \"store_sales\": 20104, \"verify_rate\": 0.05, \"fixed_salary\": 3000, \"sales_amount\": 6980, \"total_payable\": 3139.6, \"verify_amount\": 0, \"manager_penalty\": 0, \"sales_cap_ratio\": 0, \"sales_commission\": 139.6, \"sales_amount_used\": 6980, \"verify_commission\": 0, \"working_headcount\": 0}','2026-11-09 17:33:09','2026-11-11 12:59:41'),(5,'2026-11',6,2,3000.00,1350.40,0.00,0.00,4350.40,'已计算','{\"note\": \"店员：固定3000 + 开卡阶梯提成 + 核销5%\", \"sales_rate\": 0.08, \"store_sales\": 16880, \"verify_rate\": 0.05, \"fixed_salary\": 3000, \"sales_amount\": 16880, \"total_payable\": 4350.4, \"verify_amount\": 0, \"manager_penalty\": 0, \"sales_cap_ratio\": 0, \"sales_commission\": 1350.4, \"sales_amount_used\": 16880, \"verify_commission\": 0, \"working_headcount\": 0}','2026-11-09 17:33:09','2026-11-11 12:59:41'),(16,'2026-11',7,2,3500.00,0.00,0.00,135.04,3364.96,'已计算','{\"note\": \"店长：固定3500 + 开卡阶梯提成 + 核销5%；含业绩上下限\", \"sales_rate\": 0.02, \"store_sales\": 16880, \"verify_rate\": 0.05, \"fixed_salary\": 3500, \"sales_amount\": 0, \"total_payable\": 3364.96, \"verify_amount\": 0, \"manager_penalty\": 135.04, \"sales_cap_ratio\": 0.6, \"sales_commission\": 0, \"sales_amount_used\": 0, \"verify_commission\": 0, \"working_headcount\": 0}','2026-11-11 12:59:41','2026-11-11 12:59:41'),(17,'2026-11',8,1,3000.00,22.90,0.00,0.00,3022.90,'已计算','{\"note\": \"店员：固定3000 + 开卡阶梯提成 + 核销5%\", \"sales_rate\": 0.02, \"store_sales\": 20104, \"verify_rate\": 0.05, \"fixed_salary\": 3000, \"sales_amount\": 0, \"total_payable\": 3022.9, \"verify_amount\": 458, \"manager_penalty\": 0, \"sales_cap_ratio\": 0, \"sales_commission\": 0, \"sales_amount_used\": 0, \"verify_commission\": 22.9, \"working_headcount\": 0}','2026-11-11 12:59:41','2026-11-11 12:59:41');
/*!40000 ALTER TABLE `lp_salary_detail` ENABLE KEYS */;
UNLOCK TABLES;
DROP TABLE IF EXISTS `lp_salary_rule`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
 SET character_set_client = utf8mb4 ;
CREATE TABLE `lp_salary_rule` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `name` varchar(64) NOT NULL DEFAULT '默认规则',
  `base_ratio` decimal(8,4) NOT NULL DEFAULT '1.0000',
  `commission_rate` decimal(8,4) NOT NULL DEFAULT '0.0200',
  `pass_bonus` decimal(12,2) NOT NULL DEFAULT '500.00',
  `absent_deduct` decimal(12,2) NOT NULL DEFAULT '150.00',
  `full_attendance_bonus` decimal(12,2) NOT NULL DEFAULT '200.00',
  `late_deduct` decimal(12,2) NOT NULL DEFAULT '50.00',
  `config_json` longtext COMMENT '完整薪资规则JSON',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

LOCK TABLES `lp_salary_rule` WRITE;
/*!40000 ALTER TABLE `lp_salary_rule` DISABLE KEYS */;
INSERT INTO `lp_salary_rule` VALUES (1,'兰瓶美业固定薪酬+开卡/核销提成',1.0000,0.0500,0.00,0.00,0.00,0.00,'{\"manager\":{\"fixed_items\":{\"base_salary\":1600,\"non_compete\":100,\"overtime\":700,\"social_subsidy\":1100},\"sales_tiers\":[{\"label\":\"1万以下\",\"max\":10000,\"min\":0,\"rate\":0.02},{\"label\":\"1万含-1.5万\",\"max\":15000,\"min\":10000,\"rate\":0.05},{\"label\":\"1.5万含-2万\",\"max\":20000,\"min\":15000,\"rate\":0.08},{\"label\":\"2万含-2.5万\",\"max\":25000,\"min\":20000,\"rate\":0.11},{\"label\":\"2.5万含-3万\",\"max\":30000,\"min\":25000,\"rate\":0.12},{\"label\":\"3万及以上\",\"max\":0,\"min\":30000,\"rate\":0.13}],\"verify_rate\":0.05},\"manager_extra\":{\"attend_days_gt\":20,\"cap_by_headcount\":[{\"cap_ratio\":0.6,\"max_people\":2},{\"cap_ratio\":0.4,\"people\":3},{\"cap_ratio\":0.3,\"min_people\":4}],\"min_shortfall_penalty\":0.08,\"min_store_share\":0.1},\"name\":\"兰瓶美业固定薪酬+开卡/核销提成\",\"note\":\"销售提成=开卡提成；消耗提成=核销提成\",\"staff\":{\"fixed_items\":{\"base_salary\":1200,\"non_compete\":0,\"overtime\":700,\"social_subsidy\":1100},\"sales_tiers\":[{\"label\":\"1万以下\",\"max\":10000,\"min\":0,\"rate\":0.02},{\"label\":\"1万含-1.5万\",\"max\":15000,\"min\":10000,\"rate\":0.05},{\"label\":\"1.5万含-2万\",\"max\":20000,\"min\":15000,\"rate\":0.08},{\"label\":\"2万含-2.5万\",\"max\":25000,\"min\":20000,\"rate\":0.11},{\"label\":\"2.5万含-3万\",\"max\":30000,\"min\":25000,\"rate\":0.12},{\"label\":\"3万及以上\",\"max\":0,\"min\":30000,\"rate\":0.13}],\"verify_rate\":0.05}}','2026-11-09 18:09:48');
/*!40000 ALTER TABLE `lp_salary_rule` ENABLE KEYS */;
UNLOCK TABLES;
DROP TABLE IF EXISTS `lp_settings`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
 SET character_set_client = utf8mb4 ;
CREATE TABLE `lp_settings` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `config_key` varchar(64) NOT NULL,
  `config_value` text NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_key` (`config_key`)
) ENGINE=InnoDB AUTO_INCREMENT=97 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

LOCK TABLES `lp_settings` WRITE;
/*!40000 ALTER TABLE `lp_settings` DISABLE KEYS */;
INSERT INTO `lp_settings` VALUES (1,'brand_name','兰瓶美业'),(2,'birthday_remind','true'),(3,'birthday_days','3'),(4,'sms_enabled','true'),(5,'sms_sign','兰瓶美业'),(6,'clockin_wifi','LanPing-Office'),(7,'clockin_range','200'),(8,'clockin_face','true'),(25,'verify_mode','real'),(26,'customer_share','true');
/*!40000 ALTER TABLE `lp_settings` ENABLE KEYS */;
UNLOCK TABLES;
DROP TABLE IF EXISTS `lp_sms_log`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
 SET character_set_client = utf8mb4 ;
CREATE TABLE `lp_sms_log` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `mobile` varchar(20) NOT NULL,
  `template_code` varchar(32) NOT NULL DEFAULT '',
  `content` varchar(500) NOT NULL DEFAULT '',
  `status` varchar(16) NOT NULL DEFAULT 'sent',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_mobile_time` (`mobile`,`created_at`)
) ENGINE=InnoDB AUTO_INCREMENT=3 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

LOCK TABLES `lp_sms_log` WRITE;
/*!40000 ALTER TABLE `lp_sms_log` DISABLE KEYS */;
INSERT INTO `lp_sms_log` VALUES (1,'18874580001','SMS_VERIFY','【模拟】核销码已生成，未真实发送短信 token=85c4fbbf743e99c9549ea467274c88a9','模拟未发','2026-11-10 19:47:32'),(2,'18874580001','SMS_VERIFY','您的核销确认链接 token=c66ad94409c841d1bc876cb49852965e','已发送','2026-11-11 13:02:32');
/*!40000 ALTER TABLE `lp_sms_log` ENABLE KEYS */;
UNLOCK TABLES;
DROP TABLE IF EXISTS `lp_stock`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
 SET character_set_client = utf8mb4 ;
CREATE TABLE `lp_stock` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `store_id` bigint(20) NOT NULL,
  `item_id` bigint(20) NOT NULL,
  `qty` int(11) NOT NULL DEFAULT '0',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_store_item` (`store_id`,`item_id`)
) ENGINE=InnoDB AUTO_INCREMENT=10 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

LOCK TABLES `lp_stock` WRITE;
/*!40000 ALTER TABLE `lp_stock` DISABLE KEYS */;
INSERT INTO `lp_stock` VALUES (1,1,1,120,'2026-11-09 17:07:05'),(2,1,2,40,'2026-11-09 17:07:05'),(3,2,1,80,'2026-11-09 17:07:05'),(4,2,2,60,'2026-11-09 17:07:05'),(5,3,3,40,'2026-11-09 17:07:05'),(6,1,4,38,'2026-11-10 20:09:15'),(8,1,5,1,'2026-11-10 20:11:33'),(9,1,6,0,'2026-11-10 20:14:00');
/*!40000 ALTER TABLE `lp_stock` ENABLE KEYS */;
UNLOCK TABLES;
DROP TABLE IF EXISTS `lp_stock_flow`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
 SET character_set_client = utf8mb4 ;
CREATE TABLE `lp_stock_flow` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `store_id` bigint(20) NOT NULL,
  `item_id` bigint(20) NOT NULL,
  `flow_type` varchar(32) NOT NULL,
  `change_qty` int(11) NOT NULL,
  `before_qty` int(11) NOT NULL,
  `after_qty` int(11) NOT NULL,
  `biz_no` varchar(32) NOT NULL DEFAULT '',
  `operator_id` bigint(20) DEFAULT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_store_time` (`store_id`,`created_at`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

LOCK TABLES `lp_stock_flow` WRITE;
/*!40000 ALTER TABLE `lp_stock_flow` DISABLE KEYS */;
INSERT INTO `lp_stock_flow` VALUES (1,1,4,'??',2,36,38,'IO555434462100',2,'2026-11-10 20:09:15');
/*!40000 ALTER TABLE `lp_stock_flow` ENABLE KEYS */;
UNLOCK TABLES;
DROP TABLE IF EXISTS `lp_stock_item`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
 SET character_set_client = utf8mb4 ;
CREATE TABLE `lp_stock_item` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `name` varchar(64) NOT NULL,
  `sku` varchar(32) NOT NULL,
  `unit` varchar(16) NOT NULL DEFAULT '件',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_sku` (`sku`)
) ENGINE=InnoDB AUTO_INCREMENT=7 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

LOCK TABLES `lp_stock_item` WRITE;
/*!40000 ALTER TABLE `lp_stock_item` DISABLE KEYS */;
INSERT INTO `lp_stock_item` VALUES (1,'玻尿酸补水面膜','M-001','片','2026-11-09 17:07:05'),(2,'水杨酸精华','S-014','瓶','2026-11-09 17:07:05'),(3,'舒缓修护乳','C-007','瓶','2026-11-09 17:07:05'),(4,'?????????','M-ROSE-01','?','2026-11-10 20:09:14'),(5,'是非得失','100','瓶','2026-11-10 20:11:33'),(6,'玻尿酸补水面膜','P143739600','瓶','2026-11-10 20:14:00');
/*!40000 ALTER TABLE `lp_stock_item` ENABLE KEYS */;
UNLOCK TABLES;
DROP TABLE IF EXISTS `lp_stock_order`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
 SET character_set_client = utf8mb4 ;
CREATE TABLE `lp_stock_order` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `order_no` varchar(32) NOT NULL,
  `store_id` bigint(20) NOT NULL,
  `item_id` bigint(20) NOT NULL,
  `qty` int(11) NOT NULL,
  `direction` varchar(8) NOT NULL,
  `channel` varchar(32) NOT NULL DEFAULT '门店调度',
  `status` varchar(16) NOT NULL DEFAULT '待发货',
  `operator_id` bigint(20) DEFAULT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_order_no` (`order_no`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

LOCK TABLES `lp_stock_order` WRITE;
/*!40000 ALTER TABLE `lp_stock_order` DISABLE KEYS */;
INSERT INTO `lp_stock_order` VALUES (1,'IO555434462100',1,4,2,'??','????','已完成',2,'2026-11-10 20:09:15','2026-11-10 20:09:15');
/*!40000 ALTER TABLE `lp_stock_order` ENABLE KEYS */;
UNLOCK TABLES;
DROP TABLE IF EXISTS `lp_store`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
 SET character_set_client = utf8mb4 ;
CREATE TABLE `lp_store` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `name` varchar(64) NOT NULL,
  `address` varchar(255) NOT NULL DEFAULT '',
  `phone` varchar(32) NOT NULL DEFAULT '',
  `manager_name` varchar(32) NOT NULL DEFAULT '',
  `status` varchar(16) NOT NULL DEFAULT '营业中',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=5 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

LOCK TABLES `lp_store` WRITE;
/*!40000 ALTER TABLE `lp_store` DISABLE KEYS */;
INSERT INTO `lp_store` VALUES (1,'兰瓶总店','怀化市鹤城区人民路 88 号','0745-2288001','张美丽','营业中','2026-11-09 17:07:05','2026-11-09 17:07:05'),(2,'万象城店','怀化市鹤城区万象大酒店3F','0745-2288002','李静','营业中','2026-11-09 17:07:05','2026-11-10 19:22:40'),(3,'天河店','怀化市鹤城区天河北路 12 号','0745-2288003','王芳','营业中','2026-11-09 17:07:05','2026-11-09 17:07:05'),(4,'测试门店','测试路1号','0745-111','小测','营业中','2026-11-10 19:21:29','2026-11-10 19:21:29');
/*!40000 ALTER TABLE `lp_store` ENABLE KEYS */;
UNLOCK TABLES;
DROP TABLE IF EXISTS `lp_user`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
 SET character_set_client = utf8mb4 ;
CREATE TABLE `lp_user` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `name` varchar(32) NOT NULL,
  `phone` varchar(20) NOT NULL,
  `password_hash` varchar(255) NOT NULL,
  `role` varchar(16) NOT NULL,
  `role_name` varchar(32) NOT NULL DEFAULT '',
  `store_id` bigint(20) DEFAULT NULL,
  `position` varchar(32) NOT NULL DEFAULT '',
  `hire_date` date DEFAULT NULL,
  `status` varchar(16) NOT NULL DEFAULT 'active',
  `level_code` varchar(16) NOT NULL DEFAULT '',
  `work_start` varchar(8) NOT NULL DEFAULT '09:00',
  `work_end` varchar(8) NOT NULL DEFAULT '18:00',
  `base_salary` decimal(12,2) NOT NULL DEFAULT '0.00',
  `avatar` varchar(255) NOT NULL DEFAULT '',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_phone` (`phone`),
  KEY `idx_store_role` (`store_id`,`role`)
) ENGINE=InnoDB AUTO_INCREMENT=9 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

LOCK TABLES `lp_user` WRITE;
/*!40000 ALTER TABLE `lp_user` DISABLE KEYS */;
INSERT INTO `lp_user` VALUES (1,'张美丽','13800000001','$2a$10$juEexCLUMeMRwMcx97U33uQ7Ic.Mszpnkm4t.KcoiB2xJmNUU1jL.','admin','超级管理员',1,'总经理','2021-03-01','active','P7','09:00','18:00',5200.00,'','2026-11-09 17:07:05','2026-11-09 17:07:05'),(2,'李静','13800000002','$2a$10$juEexCLUMeMRwMcx97U33uQ7Ic.Mszpnkm4t.KcoiB2xJmNUU1jL.','manager','店长',2,'店长','2022-05-12','active','P5','09:00','18:00',3500.00,'','2026-11-09 17:07:05','2026-11-09 17:48:01'),(3,'王芳','13800000003','$2a$10$juEexCLUMeMRwMcx97U33uQ7Ic.Mszpnkm4t.KcoiB2xJmNUU1jL.','manager','店长',3,'店长','2022-08-20','active','P5','09:00','18:00',3500.00,'','2026-11-09 17:07:05','2026-11-09 17:48:01'),(4,'陈晓','13800000004','$2a$10$juEexCLUMeMRwMcx97U33uQ7Ic.Mszpnkm4t.KcoiB2xJmNUU1jL.','staff','美业员工',1,'高级美容师','2023-02-06','active','P3','09:00','18:00',3000.00,'','2026-11-09 17:07:05','2026-11-09 17:48:01'),(5,'刘婷','13800000005','$2a$10$juEexCLUMeMRwMcx97U33uQ7Ic.Mszpnkm4t.KcoiB2xJmNUU1jL.','staff','美业员工',1,'美容师','2023-06-15','active','P2','09:00','18:00',3000.00,'','2026-11-09 17:07:05','2026-11-09 17:48:01'),(6,'赵敏','13800000006','$2a$10$juEexCLUMeMRwMcx97U33uQ7Ic.Mszpnkm4t.KcoiB2xJmNUU1jL.','staff','美业员工',2,'美容师','2024-01-08','active','P2','09:00','18:00',3000.00,'','2026-11-09 17:07:05','2026-11-09 17:48:01'),(7,'新店长','13970002222','$2a$10$7ZCabY3D.NoRBeFCjIP2k.ub0IED9BrP/BJ6xuBtyXqq.RRQM8BNK','manager','店长',2,'店长',NULL,'active','','09:00','18:00',3500.00,'','2026-11-10 19:21:30','2026-11-10 19:21:30'),(8,'刘菲菲','18874550001','$2a$10$mpjNRiZd0tZNNkzPbFY8GuQ6oTXu7AWRPICfE1ecWa6bbU3FxHtbC','staff','美业员工',1,'美业员工',NULL,'active','','09:00','18:00',3000.00,'','2026-11-10 19:42:02','2026-11-10 19:42:02');
/*!40000 ALTER TABLE `lp_user` ENABLE KEYS */;
UNLOCK TABLES;
DROP TABLE IF EXISTS `lp_verify_item`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
 SET character_set_client = utf8mb4 ;
CREATE TABLE `lp_verify_item` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `verify_id` bigint(20) NOT NULL,
  `project_id` bigint(20) NOT NULL,
  `project_name` varchar(64) NOT NULL DEFAULT '',
  `times` int(11) NOT NULL DEFAULT '1',
  `amount` decimal(12,2) NOT NULL DEFAULT '0.00',
  PRIMARY KEY (`id`),
  KEY `idx_verify` (`verify_id`)
) ENGINE=InnoDB AUTO_INCREMENT=3 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

LOCK TABLES `lp_verify_item` WRITE;
/*!40000 ALTER TABLE `lp_verify_item` DISABLE KEYS */;
INSERT INTO `lp_verify_item` VALUES (1,1,5,'背部精油SPA',1,458.00),(2,2,1,'小气泡深层清洁',1,100.00);
/*!40000 ALTER TABLE `lp_verify_item` ENABLE KEYS */;
UNLOCK TABLES;
DROP TABLE IF EXISTS `lp_verify_order`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
 SET character_set_client = utf8mb4 ;
CREATE TABLE `lp_verify_order` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `verify_no` varchar(32) NOT NULL,
  `customer_id` bigint(20) NOT NULL,
  `card_id` bigint(20) NOT NULL,
  `project_id` bigint(20) NOT NULL,
  `store_id` bigint(20) NOT NULL,
  `employee_id` bigint(20) NOT NULL,
  `total_amount` decimal(12,2) NOT NULL DEFAULT '0.00',
  `times` int(11) NOT NULL DEFAULT '1',
  `status` varchar(16) NOT NULL DEFAULT '待确认',
  `token` varchar(64) NOT NULL DEFAULT '',
  `token_expire_at` datetime DEFAULT NULL,
  `scanned_at` datetime DEFAULT NULL,
  `remark` varchar(255) NOT NULL DEFAULT '',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_verify_no` (`verify_no`),
  KEY `idx_token` (`token`),
  KEY `idx_store_time` (`store_id`,`created_at`)
) ENGINE=InnoDB AUTO_INCREMENT=3 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

LOCK TABLES `lp_verify_order` WRITE;
/*!40000 ALTER TABLE `lp_verify_order` DISABLE KEYS */;
INSERT INTO `lp_verify_order` VALUES (1,'HX202611101252',16,9,5,1,8,458.00,1,'已完成','85c4fbbf743e99c9549ea467274c88a9','2026-11-10 20:17:33','2026-11-10 19:47:54','','2026-11-10 19:47:32','2026-11-10 19:47:54'),(2,'HX202611113352',16,10,1,1,4,100.00,1,'待确认','c66ad94409c841d1bc876cb49852965e','2026-11-11 13:32:32',NULL,'','2026-11-11 13:02:32','2026-11-11 13:02:32');
/*!40000 ALTER TABLE `lp_verify_order` ENABLE KEYS */;
UNLOCK TABLES;
DROP TABLE IF EXISTS `lp_welfare`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
 SET character_set_client = utf8mb4 ;
CREATE TABLE `lp_welfare` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `name` varchar(64) NOT NULL,
  `amount` decimal(12,2) NOT NULL DEFAULT '0.00',
  `description` varchar(255) NOT NULL DEFAULT '',
  `status` varchar(16) NOT NULL DEFAULT '启用',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=6 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

LOCK TABLES `lp_welfare` WRITE;
/*!40000 ALTER TABLE `lp_welfare` DISABLE KEYS */;
INSERT INTO `lp_welfare` VALUES (1,'生日礼金',200.00,'员工生日当月发放','启用','2026-11-09 17:07:05'),(2,'全勤奖',200.00,'当月无迟到早退缺勤','启用','2026-11-09 17:07:05'),(3,'节日福利',500.00,'春节/中秋统一发放','启用','2026-11-09 17:07:05'),(4,'测试福利',88.00,'自动测试','启用','2026-11-10 19:40:41'),(5,'全员生日卡',1000.00,'生日员工福利','启用','2026-11-10 19:59:39');
/*!40000 ALTER TABLE `lp_welfare` ENABLE KEYS */;
UNLOCK TABLES;
DROP TABLE IF EXISTS `lp_welfare_issued`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
 SET character_set_client = utf8mb4 ;
CREATE TABLE `lp_welfare_issued` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `welfare_id` bigint(20) NOT NULL,
  `user_id` bigint(20) NOT NULL,
  `store_id` bigint(20) NOT NULL,
  `amount` decimal(12,2) NOT NULL DEFAULT '0.00',
  `issued_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

LOCK TABLES `lp_welfare_issued` WRITE;
/*!40000 ALTER TABLE `lp_welfare_issued` DISABLE KEYS */;
/*!40000 ALTER TABLE `lp_welfare_issued` ENABLE KEYS */;
UNLOCK TABLES;
/*!40103 SET TIME_ZONE=@OLD_TIME_ZONE */;

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;


SET FOREIGN_KEY_CHECKS=1;
