-- Script de test pour les likes de commentaires
-- À exécuter après avoir créé la table comment_likes

-- Ajouter quelques likes de test (remplace les IDs par ceux de ta DB)
INSERT IGNORE INTO comment_likes (user_id, message_id) VALUES 
(127, 1),  -- Admin like le message 1
(127, 2),  -- Admin like le message 2
(1, 1),    -- User 1 like le message 1
(2, 1),    -- User 2 like le message 1
(3, 2);    -- User 3 like le message 2

-- Vérifier les likes
SELECT 
    cl.message_id,
    COUNT(*) as total_likes,
    GROUP_CONCAT(u.username) as liked_by
FROM comment_likes cl
LEFT JOIN users u ON cl.user_id = u.id
GROUP BY cl.message_id
ORDER BY total_likes DESC;

-- Vérifier les likes d'un utilisateur spécifique (remplace 127 par ton user_id)
SELECT 
    cl.message_id,
    m.content,
    u.username as liked_by
FROM comment_likes cl
LEFT JOIN messages m ON cl.message_id = m.id
LEFT JOIN users u ON cl.user_id = u.id
WHERE cl.user_id = 127; 