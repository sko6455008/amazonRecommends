create table IF not exists `user`
(
 `id`               INT(20) AUTO_INCREMENT,
 `name`             VARCHAR(20) NOT NULL,
 `age`              INT(20),
 `created_at`       Datetime DEFAULT NULL,
 `updated_at`       Datetime DEFAULT NULL,
    PRIMARY KEY (`id`)
) DEFAULT CHARSET=utf8 COLLATE=utf8_bin;
