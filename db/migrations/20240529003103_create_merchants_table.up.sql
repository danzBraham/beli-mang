CREATE EXTENSION postgis;

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'merchant_categories') THEN
    CREATE TYPE merchant_categories AS ENUM (
      'SmallRestaurant',
      'MediumRestaurant',
      'LargeRestaurant',
      'MerchandiseRestaurant',
      'BoothKiosk',
      'ConvenienceStore'
    );
  END IF;
END $$;

CREATE TABLE IF NOT EXISTS merchants (
  id VARCHAR(26) PRIMARY KEY NOT NULL,
  name VARCHAR(30) NOT NULL,
  category merchant_categories NULL,
  image_url TEXT NOT NULL,
  location GEOGRAPHY(Point, 4326) NOT NULL,
  user_id VARCHAR(26) NOT NULL,
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW(),
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE NO ACTION ON UPDATE NO ACTION
);
