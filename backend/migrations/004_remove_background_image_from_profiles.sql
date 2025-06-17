-- Migration 004: Supprimer la colonne background_image redondante
-- Date: 2025-01-06
-- Description: Supprime la colonne background_image de user_profiles car elle fait doublon avec banner_image
 
ALTER TABLE user_profiles DROP COLUMN background_image; 