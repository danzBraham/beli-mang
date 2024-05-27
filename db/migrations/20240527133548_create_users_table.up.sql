CREATE TABLE IF NOT EXISTS users (
  id VARCHAR(26) PRIMARY KEY NOT NULL,
  username VARCHAR(30) NOT NULL UNIQUE,
  password VARCHAR(60) NOT NULL,
  email VARCHAR(50) NOT NULL,
  is_admin BOOLEAN NOT NULL DEFAULT false,
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW()
);