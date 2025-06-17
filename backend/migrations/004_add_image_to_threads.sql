-- Migration: Add image_url field to threads table
-- This allows users to attach images to their threads
 
ALTER TABLE threads 
ADD COLUMN image_url VARCHAR(500) AFTER desc_; 