-- Migration: Add image_url field to messages table
-- This allows users to attach images to their comments
 
ALTER TABLE messages 
ADD COLUMN image_url VARCHAR(500) AFTER content; 