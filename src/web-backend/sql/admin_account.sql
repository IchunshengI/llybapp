-- Admin account table for /admin/register and /admin/login
CREATE TABLE IF NOT EXISTS admin_account (
  id BIGINT NOT NULL AUTO_INCREMENT,
  username VARCHAR(64) NOT NULL,
  password_hash CHAR(32) NOT NULL,
  password_salt CHAR(32) NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_username (username)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
