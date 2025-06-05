-- Ajout des colonnes pour les embeds dans la table messages
ALTER TABLE messages
ADD COLUMN youtube_embed VARCHAR(500) NULL,
ADD COLUMN spotify_embed VARCHAR(500) NULL; 