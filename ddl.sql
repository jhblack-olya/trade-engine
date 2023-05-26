CREATE TABLE g_product (
   id varchar(255) NOT NULL,
   created_at timestamp NULL DEFAULT NULL,
   updated_at timestamp NULL DEFAULT NULL,
   base_currency varchar(255) NOT NULL,
   quote_currency varchar(255) NOT NULL,
   base_min_size decimal(32,16) NOT NULL,
   base_max_size decimal(32,16) NOT NULL,
   base_scale int(11) NOT NULL,
   quote_scale int(11) NOT NULL,
   quote_increment double NOT NULL,
   quote_min_size decimal(32,16) NOT NULL,
   quote_max_size decimal(32,16) NOT NULL,
   PRIMARY KEY (`id`)
 );

CREATE TABLE OrderBooks (
   id bigint(20) NOT NULL AUTO_INCREMENT,
   product_id varchar(255) DEFAULT NULL,
   user bigint(20) DEFAULT NULL,
   size decimal(32,16) DEFAULT NULL,
   funds decimal(32,16) DEFAULT NULL,
   filled_size decimal(32,16) DEFAULT NULL,
   executed_value decimal(32,16) DEFAULT NULL,
   price decimal(32,16) DEFAULT NULL,
   orderType bigint(11) DEFAULT NULL,
   side varchar(255) DEFAULT NULL,
   status varchar(255) DEFAULT NULL,
   expires_in bigint(20) DEFAULT NULL,
   cancelledAt timestamp NULL DEFAULT NULL,
   executedAt timestamp NULL DEFAULT NULL,
   deletedAt timestamp NULL DEFAULT NULL,
   created_at timestamp NULL DEFAULT NULL,
   updated_at timestamp NULL DEFAULT NULL,
   commission decimal(32,16) DEFAULT NULL,
   settled tinyint(1) DEFAULT NULL,
   PRIMARY KEY (`id`)
 );

CREATE TABLE g_bill (
id bigint(20) NOT NULL AUTO_INCREMENT,
created_at timestamp NULL DEFAULT NULL,
updated_at timestamp NULL DEFAULT NULL,
user_id bigint(20) NOT NULL,
currency varchar(255) NOT NULL,
available decimal(32,16) NOT NULL DEFAULT 0.0000000000000000,
hold decimal(32,16) NOT NULL DEFAULT 0.0000000000000000,
type varchar(255) NOT NULL,
settled tinyint(1) NOT NULL DEFAULT 0,
notes varchar(255) DEFAULT NULL,
PRIMARY KEY (id),
KEY idx_gsoci (user_id,currency,settled,id),
KEY idx_s (settled)
);

CREATE TABLE g_fill (
   id bigint(20) NOT NULL AUTO_INCREMENT,
   created_at timestamp NULL DEFAULT NULL,
   updated_at timestamp NULL DEFAULT NULL,
   trade_id bigint(20) NOT NULL DEFAULT 0,
   order_id bigint(20) NOT NULL DEFAULT 0,
   product_id varchar(255) NOT NULL,
   size decimal(32,16) NOT NULL,
   price decimal(32,16) NOT NULL,
   funds decimal(32,16) NOT NULL DEFAULT 0.0000000000000000,
   fee decimal(32,16) NOT NULL DEFAULT 0.0000000000000000,
   liquidity varchar(255) NOT NULL,
   settled tinyint(1) NOT NULL DEFAULT 0,
   side varchar(255) NOT NULL,
   done tinyint(1) NOT NULL DEFAULT 0,
   done_reason varchar(255) NOT NULL,
   message_seq bigint(20) NOT NULL,
   log_offset bigint(20) NOT NULL DEFAULT 0,
   log_seq bigint(20) NOT NULL DEFAULT 0,
   client_oid varchar(36) NOT NULL DEFAULT '',
   art bigint(20) DEFAULT NULL,
   expires_in bigint(20) DEFAULT 0,
   cancelled_at varchar(255) DEFAULT '',
   executed_at varchar(255) DEFAULT '',
   PRIMARY KEY (id),
   UNIQUE KEY o_m (order_id,message_seq),
   KEY idx_gsoi (order_id,settled,id),
   KEY idx_si (settled,id)
 );

CREATE TABLE g_trade (
   `id` bigint(20) NOT NULL AUTO_INCREMENT,
   `created_at` timestamp NULL DEFAULT NULL,
   `updated_at` timestamp NULL DEFAULT NULL,
   `product_id` varchar(255) NOT NULL,
   `taker_order_id` bigint(20) NOT NULL,
   `maker_order_id` bigint(20) NOT NULL,
   `price` decimal(32,16) NOT NULL,
   `size` decimal(32,16) NOT NULL,
   `side` varchar(255) NOT NULL,
   `time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
   `log_offset` bigint(20) NOT NULL DEFAULT '0',
   `log_seq` bigint(20) NOT NULL DEFAULT '0',
   PRIMARY KEY (`id`)
 ); 


insert into g_product(
id,base_currency,quote_currency,base_min_size,base_max_size,base_scale,quote_scale,quote_increment,quote_min_size,quote_max_size)values('BTC-XUS', 'BTC', 'XUS', '0.0000100000000000', '10000000.0000000000000000', '4', '2', '0.01', '0.0000000000000000', '0.0000000000000000'),
('DBX-XUS', 'DBX', 'XUS', '0.0010000000000000', '1000.0000000000000000', '4', '2', '0.01', '0.0000000000000000', '0.0000000000000000'),
('ETH-XUS', 'ETH', 'XUS', '0.0001000000000000', '10000.0000000000000000', '4', '2', '0.01', '0.0000000000000000', '0.0000000000000000');