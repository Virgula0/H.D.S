CREATE USER 'agent' IDENTIFIED BY 'SUPERSECUREUNCRACKABLEPASSWORD';

DROP DATABASE IF EXISTS dp_hashcat;
CREATE DATABASE IF NOT EXISTS dp_hashcat;

USE dp_hashcat;

DROP TABLE IF EXISTS raspberry_pi;
DROP TABLE IF EXISTS handshake;
DROP TABLE IF EXISTS client;
DROP TABLE IF EXISTS role;
DROP TABLE IF EXISTS user;

CREATE TABLE IF NOT EXISTS user
(
        UUID varchar(36), -- uuid format saved a string
        USERNAME varchar(255) UNIQUE,
        PASSWORD varchar(255),

        PRIMARY KEY(UUID)
);

CREATE TABLE IF NOT EXISTS role
(
        UUID varchar(36),
        ROLE_STRING varchar(20),

        PRIMARY KEY (UUID),
        FOREIGN KEY (`UUID`) REFERENCES `user` (`UUID`) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS raspberry_pi
(
    UUID_USER varchar(36),
    UUID varchar(36),
    MACHINE_ID varchar(32) UNIQUE, -- md5 format
    ENCRYPTION_KEY varchar(64), -- double encryption key for sharing cap files

    PRIMARY KEY(UUID),
    FOREIGN KEY (`UUID_USER`) REFERENCES `user` (`UUID`) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS client
(
        UUID_USER varchar(36), -- uuid format saved a string
        UUID varchar(36),
        NAME varchar(100),
        LATEST_IP varchar(100),
        CREATION_DATETIME DATETIME, -- YYYY-MM-DD HH:MM:SS
        LATEST_CONNECTION DATETIME, -- YYYY-MM-DD HH:MM:SS
        MACHINE_ID varchar(32), -- md5 format

        PRIMARY KEY(UUID),
        FOREIGN KEY (`UUID_USER`) REFERENCES `user` (`UUID`) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS handshake
(   
    UUID_USER varchar(36),
    UUID_ASSIGNED_CLIENT varchar(36),
    UUID_ASSIGNED_RASPBERRY_PI varchar(36),
    UUID varchar(36),
    SSID varchar(300) UNIQUE,
    BSSID varchar(300),
    UPLOADED_DATE DATETIME DEFAULT CURRENT_TIMESTAMP,
    STATUS varchar(20) DEFAULT 'pending',
    CRACKED_DATE DATETIME NULL,
    HASHCAT_OPTIONS text,
    HASHCAT_LOGS text,
    CRACKED_HANDSHAKE varchar(1000),

    PRIMARY KEY(UUID),
    FOREIGN KEY (`UUID_USER`) REFERENCES `user` (`UUID`) ON DELETE CASCADE,
    FOREIGN KEY (`UUID_ASSIGNED_CLIENT`) REFERENCES `client` (`UUID`),
    FOREIGN KEY (`UUID_ASSIGNED_RASPBERRY_PI`) REFERENCES `raspberry_pi` (`UUID`) 
);

GRANT SELECT, UPDATE, INSERT, DELETE ON dp_hashcat.* TO 'agent'@'%';
FLUSH PRIVILEGES;
