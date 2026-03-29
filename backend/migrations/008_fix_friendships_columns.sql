-- Migration 008: Rebuild friendships table with correct column names
-- Drops and recreates the table so foreign key / index conflicts are avoided.
-- Safe because this runs on a fresh database with no friendship data yet.

DROP TABLE IF EXISTS friendships;

CREATE TABLE IF NOT EXISTS friendships (
    id           INT AUTO_INCREMENT PRIMARY KEY,
    requester_id INT NOT NULL,
    addressee_id INT NOT NULL,
    status       ENUM('pending', 'accepted', 'rejected', 'blocked') DEFAULT 'pending',
    created_at   TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at   TIMESTAMP NULL ON UPDATE CURRENT_TIMESTAMP,
    compatibility_score INT DEFAULT 0,
    FOREIGN KEY (requester_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (addressee_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE KEY unique_friendship (requester_id, addressee_id),
    INDEX idx_requester          (requester_id),
    INDEX idx_addressee          (addressee_id),
    INDEX idx_status             (status),
    INDEX idx_status_created     (status, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
