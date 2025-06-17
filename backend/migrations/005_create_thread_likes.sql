-- Migration pour ajouter le système de likes des threads
-- Crée la table thread_likes pour stocker les likes des utilisateurs

CREATE TABLE IF NOT EXISTS thread_likes (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL,
    thread_id INT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY unique_user_thread_like (user_id, thread_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (thread_id) REFERENCES threads(id) ON DELETE CASCADE,
    INDEX idx_thread_likes_thread_id (thread_id),
    INDEX idx_thread_likes_user_id (user_id)
);

-- Ajouter une colonne likes_count aux threads pour optimiser les performances
ALTER TABLE threads ADD COLUMN likes_count INT NOT NULL DEFAULT 0; 