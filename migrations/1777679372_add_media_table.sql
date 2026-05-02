CREATE TABLE media (
  id INTEGER PRIMARY KEY,
  filename TEXT,
  file_path TEXT, -- e.g., "/var/www/app/uploads/audio_123.mp3"
  content_type TEXT,
  user_id INTEGER,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TRIGGER IF NOT EXISTS update_media_updated_at
AFTER
UPDATE ON media BEGIN
UPDATE media
SET updated_at = CURRENT_TIMESTAMP
WHERE id = NEW.id;
END;