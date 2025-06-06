-- Insérer des utilisateurs de test
INSERT INTO users (username, email, password, is_admin, created_at, updated_at) VALUES
('MC_Rhyme', 'mc.rhyme@test.com', '$2a$10$abcdefghijklmnopqrstuvwxyz', false, NOW(), NOW()),
('RapMaster', 'rap.master@test.com', '$2a$10$abcdefghijklmnopqrstuvwxyz', false, NOW(), NOW()),
('FlowKing', 'flow.king@test.com', '$2a$10$abcdefghijklmnopqrstuvwxyz', false, NOW(), NOW()),
('BeatQueen', 'beat.queen@test.com', '$2a$10$abcdefghijklmnopqrstuvwxyz', false, NOW(), NOW());

-- Insérer des battles de test
INSERT INTO battles (title, description, state, contestant1_id, contestant2_id, created_at, updated_at) VALUES
('Battle des Titans', 'Le clash du siècle entre MC_Rhyme et RapMaster', 'active', 1, 2, NOW(), NOW()),
('Reine vs Roi', 'FlowKing affronte BeatQueen dans un battle épique', 'active', 3, 4, NOW(), NOW()),
('Championnat 2024', 'Finale du tournoi de rap', 'active', 1, 3, NOW(), NOW());

-- Insérer quelques votes
INSERT INTO battle_votes (battle_id, user_id, contestant_id, created_at) VALUES
(1, 3, 1, NOW()), -- FlowKing vote pour MC_Rhyme
(1, 4, 2, NOW()), -- BeatQueen vote pour RapMaster
(2, 1, 3, NOW()), -- MC_Rhyme vote pour FlowKing
(2, 2, 4, NOW()); -- RapMaster vote pour BeatQueen 