CREATE TABLE battles (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    title VARCHAR(200) NOT NULL,
    description TEXT NOT NULL,
    state VARCHAR(20) NOT NULL DEFAULT 'active',
    contestant1_id BIGINT UNSIGNED NOT NULL,
    contestant2_id BIGINT UNSIGNED NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (contestant1_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (contestant2_id) REFERENCES users(id) ON DELETE CASCADE,
    CHECK (state IN ('active', 'finished', 'cancelled')),
    CHECK (contestant1_id != contestant2_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE battle_votes (
    battle_id BIGINT UNSIGNED NOT NULL,
    user_id BIGINT UNSIGNED NOT NULL,
    contestant_id BIGINT UNSIGNED NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (battle_id, user_id),
    FOREIGN KEY (battle_id) REFERENCES battles(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (contestant_id) REFERENCES users(id) ON DELETE CASCADE,
    CHECK (contestant_id IN (
        SELECT contestant1_id FROM battles WHERE id = battle_id
        UNION
        SELECT contestant2_id FROM battles WHERE id = battle_id
    ))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci; 