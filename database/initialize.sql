SELECT 'Creating user...';
CREATE USER 'agent'@'%' IDENTIFIED BY 'SUPERSECUREUNCRACKABLEPASSWORD';

DROP DATABASE IF EXISTS dp_hashcat;
CREATE DATABASE IF NOT EXISTS dp_hashcat;
USE dp_hashcat;

DROP TABLE IF EXISTS raspberry_pi;
DROP TABLE IF EXISTS handshake;
DROP TABLE IF EXISTS client;
DROP TABLE IF EXISTS role;
DROP TABLE IF EXISTS user;

CREATE TABLE IF NOT EXISTS user (
    UUID varchar(36),
    USERNAME varchar(255) UNIQUE,
    PASSWORD varchar(255),
    PRIMARY KEY(UUID)
);

CREATE TABLE IF NOT EXISTS role (
    UUID varchar(36),
    ROLE_STRING varchar(20),
    PRIMARY KEY (UUID),
    FOREIGN KEY (`UUID`) REFERENCES `user` (`UUID`) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS raspberry_pi (
    UUID_USER varchar(36),
    UUID varchar(36),
    MACHINE_ID varchar(32) UNIQUE,
    ENCRYPTION_KEY varchar(64),
    PRIMARY KEY(UUID),
    FOREIGN KEY (`UUID_USER`) REFERENCES `user` (`UUID`) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS client (
    UUID_USER varchar(36),
    UUID varchar(36),
    NAME varchar(100),
    LATEST_IP varchar(100),
    CREATION_DATETIME DATETIME,
    LATEST_CONNECTION DATETIME,
    MACHINE_ID varchar(32) UNIQUE,
    PRIMARY KEY(UUID),
    FOREIGN KEY (`UUID_USER`) REFERENCES `user` (`UUID`) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS handshake (
    UUID_USER varchar(36),
    UUID_ASSIGNED_CLIENT varchar(36),
    UUID varchar(36),
    SSID varchar(300),
    BSSID varchar(300),
    UPLOADED_DATE DATETIME DEFAULT CURRENT_TIMESTAMP,
    STATUS varchar(20) DEFAULT 'nothing',
    CRACKED_DATE DATETIME NULL,
    HASHCAT_OPTIONS text,
    HASHCAT_LOGS LONGTEXT,
    CRACKED_HANDSHAKE varchar(1000),
    HANDSHAKE_PCAP LONGTEXT,
    PRIMARY KEY(UUID),
    FOREIGN KEY (`UUID_USER`) REFERENCES `user` (`UUID`) ON DELETE CASCADE,
    FOREIGN KEY (`UUID_ASSIGNED_CLIENT`) REFERENCES `client` (`UUID`) ON DELETE CASCADE
);

GRANT SELECT, UPDATE, INSERT, DELETE ON dp_hashcat.* TO 'agent'@'%';

-- Finalizing privileges
FLUSH PRIVILEGES;
