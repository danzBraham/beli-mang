CREATE EXTENSION postgis;

CREATE TABLE IF NOT EXISTS merchants (
  id VARCHAR(26) PRIMARY KEY NOT NULL,
  name VARCHAR(30) NOT NULL,
  category VARCHAR(25) NULL,
  image_url TEXT NOT NULL,
  location GEOGRAPHY(Point, 4326) NOT NULL,
  user_id VARCHAR(26) NOT NULL,
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW(),
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE NO ACTION ON UPDATE NO ACTION
);