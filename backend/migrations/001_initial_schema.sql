-- Création de la base de données Rythmit
CREATE DATABASE IF NOT EXISTS rythmit_db CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE rythmit_db;

-- Table User
CREATE TABLE IF NOT EXISTS users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    username VARCHAR(30) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    is_admin BOOLEAN DEFAULT FALSE,
    profile_pic VARCHAR(500),
    biography TEXT,
    last_connection TIMESTAMP NULL,
    message_count INT DEFAULT 0,
    thread_count INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_username (username),
    INDEX idx_email (email)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Table Thread
CREATE TABLE IF NOT EXISTS threads (
    id INT AUTO_INCREMENT PRIMARY KEY,
    title VARCHAR(200) NOT NULL,
    desc_ TEXT NOT NULL,
    creation TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    state ENUM('ouvert', 'fermé', 'archivé') DEFAULT 'ouvert',
    visibility ENUM('public', 'privé') DEFAULT 'public',
    user_id INT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    INDEX idx_user_id (user_id),
    INDEX idx_state (state),
    INDEX idx_creation (creation)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Table Message
CREATE TABLE IF NOT EXISTS messages (
    id INT AUTO_INCREMENT PRIMARY KEY,
    content TEXT NOT NULL,
    date_ TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    thread_id INT NOT NULL,
    user_id INT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (thread_id) REFERENCES threads(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    INDEX idx_thread_id (thread_id),
    INDEX idx_user_id (user_id),
    INDEX idx_date (date_)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Table Tag
CREATE TABLE IF NOT EXISTS tags (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(50) UNIQUE NOT NULL,
    type ENUM('genre', 'artist', 'album') DEFAULT 'genre',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_name (name),
    INDEX idx_type (type)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Table de liaison Thread-Tag (many-to-many)
CREATE TABLE IF NOT EXISTS thread_tags (
    thread_id INT NOT NULL,
    tag_id INT NOT NULL,
    PRIMARY KEY (thread_id, tag_id),
    FOREIGN KEY (thread_id) REFERENCES threads(id) ON DELETE CASCADE,
    FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Table Liked_Disliked (Fire/Skip)
CREATE TABLE IF NOT EXISTS message_votes (
    user_id INT NOT NULL,
    message_id INT NOT NULL,
    state ENUM('fire', 'skip', 'neutral') NOT NULL DEFAULT 'neutral',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, message_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (message_id) REFERENCES messages(id) ON DELETE CASCADE,
    INDEX idx_message_id (message_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Table Friendship (Musical Twins)
CREATE TABLE IF NOT EXISTS friendships (
    id INT AUTO_INCREMENT PRIMARY KEY,
    status ENUM('pending', 'accepted', 'rejected') DEFAULT 'pending',
    request_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    response_date TIMESTAMP NULL,
    user_id_1 INT NOT NULL,
    user_id_2 INT NOT NULL,
    compatibility_score INT DEFAULT 0,
    FOREIGN KEY (user_id_1) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id_2) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE KEY unique_friendship (user_id_1, user_id_2),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Table Battle
CREATE TABLE IF NOT EXISTS battles (
    id INT AUTO_INCREMENT PRIMARY KEY,
    title VARCHAR(200) NOT NULL,
    status ENUM('active', 'ended') DEFAULT 'active',
    end_date TIMESTAMP NOT NULL,
    creator_id INT NOT NULL,
    total_votes INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (creator_id) REFERENCES users(id) ON DELETE CASCADE,
    INDEX idx_status (status),
    INDEX idx_end_date (end_date)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Table Battle Options
CREATE TABLE IF NOT EXISTS battle_options (
    id INT AUTO_INCREMENT PRIMARY KEY,
    battle_id INT NOT NULL,
    name VARCHAR(100) NOT NULL,
    image VARCHAR(500),
    votes INT DEFAULT 0,
    FOREIGN KEY (battle_id) REFERENCES battles(id) ON DELETE CASCADE,
    INDEX idx_battle_id (battle_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Table Battle Votes
CREATE TABLE IF NOT EXISTS battle_votes (
    user_id INT NOT NULL,
    battle_id INT NOT NULL,
    option_id INT NOT NULL,
    voted_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, battle_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (battle_id) REFERENCES battles(id) ON DELETE CASCADE,
    FOREIGN KEY (option_id) REFERENCES battle_options(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Table User Music Preferences (pour Musical Twins)
CREATE TABLE IF NOT EXISTS user_music_preferences (
    user_id INT NOT NULL,
    tag_id INT NOT NULL,
    score INT DEFAULT 0,
    PRIMARY KEY (user_id, tag_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Insertion de tags musicaux de base
INSERT INTO tags (name, type) VALUES 
-- Genres
('rap', 'genre'),
('hip-hop', 'genre'),
('r&b', 'genre'),
('pop', 'genre'),
('rock', 'genre'),
('jazz', 'genre'),
('electronic', 'genre'),
('indie', 'genre'),
('metal', 'genre'),
('classical', 'genre'),
-- Artistes populaires
('drake', 'artist'),
('kendrick lamar', 'artist'),
('taylor swift', 'artist'),
('the weeknd', 'artist'),
('beyonce', 'artist'),
('kanye west', 'artist'),
('billie eilish', 'artist'),
('post malone', 'artist'),
('travis scott', 'artist'),
('dua lipa', 'artist');

-- Création d'un utilisateur admin par défaut
INSERT INTO users (username, email, password, is_admin) VALUES 
('admin', 'admin@rythmit.com', '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewKyNiGH8IJ.2XpO', TRUE);
-- Mot de passe par défaut: ChangeThisPassword123!