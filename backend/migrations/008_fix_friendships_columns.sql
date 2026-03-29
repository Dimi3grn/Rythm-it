-- Migration: Correction des colonnes de la table friendships
-- Date: 2025-11-27
-- Description: Renommer les colonnes pour correspondre au code Go

-- Supprimer l'index unique avant de renommer les colonnes
ALTER TABLE friendships DROP INDEX unique_friendship;

-- Renommer les colonnes (les contraintes FK seront automatiquement mises à jour)
ALTER TABLE friendships CHANGE COLUMN user_id_1 requester_id INT NOT NULL;
ALTER TABLE friendships CHANGE COLUMN user_id_2 addressee_id INT NOT NULL;
ALTER TABLE friendships CHANGE COLUMN request_date created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE friendships CHANGE COLUMN response_date updated_at TIMESTAMP NULL;

-- Recréer l'index unique avec les nouvelles colonnes
ALTER TABLE friendships ADD UNIQUE KEY unique_friendship (requester_id, addressee_id);

-- Ajouter des index pour améliorer les performances
ALTER TABLE friendships ADD INDEX idx_requester (requester_id);
ALTER TABLE friendships ADD INDEX idx_addressee (addressee_id);
ALTER TABLE friendships ADD INDEX idx_status_created (status, created_at);
