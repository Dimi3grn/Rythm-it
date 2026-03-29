-- Migration pour créer la table comment_likes
-- Cette table stocke les likes sur les commentaires/messages

CREATE TABLE IF NOT EXISTS comment_likes (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL,
    message_id INT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY unique_user_message_like (user_id, message_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (message_id) REFERENCES messages(id) ON DELETE CASCADE,
    INDEX idx_message_likes_message_id (message_id),
    INDEX idx_message_likes_user_id (user_id)
);

-- Ajouter des données de test pour vérifier que ça fonctionne
-- (vous pouvez supprimer ces lignes si nécessaire)
INSERT IGNORE INTO comment_likes (user_id, message_id) VALUES 
(23, 16),  -- User 23 like le message 16
(23, 17);  -- User 23 like le message 17 