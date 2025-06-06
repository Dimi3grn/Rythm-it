-- Supprimer les votes de test
DELETE FROM battle_votes WHERE battle_id IN (1, 2, 3);

-- Supprimer les battles de test
DELETE FROM battles WHERE id IN (1, 2, 3);

-- Supprimer les utilisateurs de test
DELETE FROM users WHERE email IN (
    'mc.rhyme@test.com',
    'rap.master@test.com',
    'flow.king@test.com',
    'beat.queen@test.com'
); 