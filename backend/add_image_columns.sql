-- Script pour ajouter les colonnes image_url aux tables threads et messages
USE rythmit_db;

-- Ajouter image_url à la table threads
ALTER TABLE threads ADD COLUMN image_url VARCHAR(500) AFTER desc_;

-- Ajouter image_url à la table messages  
ALTER TABLE messages ADD COLUMN image_url VARCHAR(500) AFTER content;

-- Vérifier que les colonnes ont été ajoutées
SHOW COLUMNS FROM threads;
SHOW COLUMNS FROM messages; 