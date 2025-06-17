-- Migration: Add avatar_image field to user_profiles table
-- This allows users to upload and store custom avatar images
 
ALTER TABLE user_profiles 
ADD COLUMN avatar_image VARCHAR(500) AFTER display_name; 