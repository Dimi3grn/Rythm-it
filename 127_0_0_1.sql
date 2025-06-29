-- phpMyAdmin SQL Dump
-- version 5.2.1
-- https://www.phpmyadmin.net/
--
-- Hôte : 127.0.0.1:3306
-- Généré le : dim. 29 juin 2025 à 16:44
-- Version du serveur : 9.1.0
-- Version de PHP : 8.3.14

SET SQL_MODE = "NO_AUTO_VALUE_ON_ZERO";
START TRANSACTION;
SET time_zone = "+00:00";


/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!40101 SET NAMES utf8mb4 */;

--
-- Base de données : `rythmit_db`
--
CREATE DATABASE IF NOT EXISTS `rythmit_db` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE `rythmit_db`;

-- --------------------------------------------------------

--
-- Structure de la table `battles`
--

DROP TABLE IF EXISTS `battles`;
CREATE TABLE IF NOT EXISTS `battles` (
  `id` int NOT NULL AUTO_INCREMENT,
  `title` varchar(200) COLLATE utf8mb4_unicode_ci NOT NULL,
  `description` text COLLATE utf8mb4_unicode_ci NOT NULL,
  `state` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'active',
  `contestant1_id` int NOT NULL,
  `contestant2_id` int NOT NULL,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `contestant1_id` (`contestant1_id`),
  KEY `contestant2_id` (`contestant2_id`),
  KEY `idx_state` (`state`)
) ENGINE=InnoDB AUTO_INCREMENT=4 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

--
-- Déchargement des données de la table `battles`
--

INSERT INTO `battles` (`id`, `title`, `description`, `state`, `contestant1_id`, `contestant2_id`, `created_at`, `updated_at`) VALUES
(1, 'Battle des Titans', 'Le clash du siècle entre MC_Rhyme et RapMaster', 'active', 2, 3, '2025-06-28 17:22:08', '2025-06-28 17:22:08'),
(2, 'Reine vs Roi', 'FlowKing affronte BeatQueen dans un battle épique', 'active', 4, 5, '2025-06-28 17:22:08', '2025-06-28 17:22:08'),
(3, 'Championnat 2024', 'Finale du tournoi de rap', 'active', 2, 4, '2025-06-28 17:22:08', '2025-06-28 17:22:08');

-- --------------------------------------------------------

--
-- Structure de la table `battle_votes`
--

DROP TABLE IF EXISTS `battle_votes`;
CREATE TABLE IF NOT EXISTS `battle_votes` (
  `battle_id` int NOT NULL,
  `user_id` int NOT NULL,
  `contestant_id` int NOT NULL,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`battle_id`,`user_id`),
  KEY `user_id` (`user_id`),
  KEY `contestant_id` (`contestant_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

--
-- Déchargement des données de la table `battle_votes`
--

INSERT INTO `battle_votes` (`battle_id`, `user_id`, `contestant_id`, `created_at`) VALUES
(1, 4, 2, '2025-06-28 17:22:08'),
(1, 5, 3, '2025-06-28 17:22:08'),
(2, 2, 4, '2025-06-28 17:22:08'),
(2, 3, 5, '2025-06-28 17:22:08');

-- --------------------------------------------------------

--
-- Structure de la table `friendships`
--

DROP TABLE IF EXISTS `friendships`;
CREATE TABLE IF NOT EXISTS `friendships` (
  `id` int NOT NULL AUTO_INCREMENT,
  `status` enum('pending','accepted','rejected') COLLATE utf8mb4_unicode_ci DEFAULT 'pending',
  `request_date` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `response_date` timestamp NULL DEFAULT NULL,
  `user_id_1` int NOT NULL,
  `user_id_2` int NOT NULL,
  `compatibility_score` int DEFAULT '0',
  PRIMARY KEY (`id`),
  UNIQUE KEY `unique_friendship` (`user_id_1`,`user_id_2`),
  KEY `user_id_2` (`user_id_2`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- --------------------------------------------------------

--
-- Structure de la table `messages`
--

DROP TABLE IF EXISTS `messages`;
CREATE TABLE IF NOT EXISTS `messages` (
  `id` int NOT NULL AUTO_INCREMENT,
  `content` text COLLATE utf8mb4_unicode_ci NOT NULL,
  `image_url` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `youtube_embed` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `spotify_embed` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `date_` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `thread_id` int NOT NULL,
  `user_id` int NOT NULL,
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_thread_id` (`thread_id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_date` (`date_`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- --------------------------------------------------------

--
-- Structure de la table `message_votes`
--

DROP TABLE IF EXISTS `message_votes`;
CREATE TABLE IF NOT EXISTS `message_votes` (
  `user_id` int NOT NULL,
  `message_id` int NOT NULL,
  `state` enum('fire','skip','neutral') COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'neutral',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`user_id`,`message_id`),
  KEY `idx_message_id` (`message_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- --------------------------------------------------------

--
-- Structure de la table `tags`
--

DROP TABLE IF EXISTS `tags`;
CREATE TABLE IF NOT EXISTS `tags` (
  `id` int NOT NULL AUTO_INCREMENT,
  `name` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL,
  `type` enum('genre','artist','album') COLLATE utf8mb4_unicode_ci DEFAULT 'genre',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `name` (`name`),
  KEY `idx_name` (`name`),
  KEY `idx_type` (`type`)
) ENGINE=InnoDB AUTO_INCREMENT=23 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

--
-- Déchargement des données de la table `tags`
--

INSERT INTO `tags` (`id`, `name`, `type`, `created_at`) VALUES
(1, 'rap', 'genre', '2025-06-28 17:22:08'),
(2, 'hip-hop', 'genre', '2025-06-28 17:22:08'),
(3, 'r&b', 'genre', '2025-06-28 17:22:08'),
(4, 'pop', 'genre', '2025-06-28 17:22:08'),
(5, 'rock', 'genre', '2025-06-28 17:22:08'),
(6, 'jazz', 'genre', '2025-06-28 17:22:08'),
(7, 'electronic', 'genre', '2025-06-28 17:22:08'),
(8, 'indie', 'genre', '2025-06-28 17:22:08'),
(9, 'metal', 'genre', '2025-06-28 17:22:08'),
(10, 'classical', 'genre', '2025-06-28 17:22:08'),
(11, 'drake', 'artist', '2025-06-28 17:22:08'),
(12, 'kendrick lamar', 'artist', '2025-06-28 17:22:08'),
(13, 'taylor swift', 'artist', '2025-06-28 17:22:08'),
(14, 'the weeknd', 'artist', '2025-06-28 17:22:08'),
(15, 'beyonce', 'artist', '2025-06-28 17:22:08'),
(16, 'kanye west', 'artist', '2025-06-28 17:22:08'),
(17, 'billie eilish', 'artist', '2025-06-28 17:22:08'),
(18, 'post malone', 'artist', '2025-06-28 17:22:08'),
(19, 'travis scott', 'artist', '2025-06-28 17:22:08'),
(20, 'dua lipa', 'artist', '2025-06-28 17:22:08'),
(21, 'général', 'artist', '2025-06-28 17:28:28'),
(22, 'discussion', 'artist', '2025-06-28 17:28:28');

-- --------------------------------------------------------

--
-- Structure de la table `threads`
--

DROP TABLE IF EXISTS `threads`;
CREATE TABLE IF NOT EXISTS `threads` (
  `id` int NOT NULL AUTO_INCREMENT,
  `title` varchar(200) COLLATE utf8mb4_unicode_ci NOT NULL,
  `desc_` text COLLATE utf8mb4_unicode_ci NOT NULL,
  `image_url` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `creation` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `state` enum('ouvert','fermé','archivé') COLLATE utf8mb4_unicode_ci DEFAULT 'ouvert',
  `visibility` enum('public','privé') COLLATE utf8mb4_unicode_ci DEFAULT 'public',
  `user_id` int NOT NULL,
  `likes_count` int NOT NULL DEFAULT '0',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_state` (`state`),
  KEY `idx_creation` (`creation`)
) ENGINE=InnoDB AUTO_INCREMENT=3 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

--
-- Déchargement des données de la table `threads`
--

INSERT INTO `threads` (`id`, `title`, `desc_`, `image_url`, `creation`, `state`, `visibility`, `user_id`, `likes_count`, `created_at`, `updated_at`) VALUES
(1, 'My new thread', 'my thread wow so cool !', NULL, '2025-06-28 17:28:11', 'ouvert', 'public', 6, 0, '2025-06-28 17:28:11', '2025-06-28 17:28:11'),
(2, 'teddteqdsdqsdqdsdddddddd', 'qsdfjiaefijaehjifezjiaefjiafjoi', '/uploads/threads/1751131702_30c479881bb7.jpg', '2025-06-28 17:28:28', 'ouvert', 'public', 6, 1, '2025-06-28 17:28:28', '2025-06-28 17:29:25');

-- --------------------------------------------------------

--
-- Structure de la table `thread_likes`
--

DROP TABLE IF EXISTS `thread_likes`;
CREATE TABLE IF NOT EXISTS `thread_likes` (
  `id` int NOT NULL AUTO_INCREMENT,
  `user_id` int NOT NULL,
  `thread_id` int NOT NULL,
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `unique_user_thread_like` (`user_id`,`thread_id`),
  KEY `idx_thread_likes_thread_id` (`thread_id`),
  KEY `idx_thread_likes_user_id` (`user_id`)
) ENGINE=InnoDB AUTO_INCREMENT=3 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

--
-- Déchargement des données de la table `thread_likes`
--

INSERT INTO `thread_likes` (`id`, `user_id`, `thread_id`, `created_at`) VALUES
(2, 6, 2, '2025-06-28 17:29:14');

-- --------------------------------------------------------

--
-- Structure de la table `thread_tags`
--

DROP TABLE IF EXISTS `thread_tags`;
CREATE TABLE IF NOT EXISTS `thread_tags` (
  `thread_id` int NOT NULL,
  `tag_id` int NOT NULL,
  PRIMARY KEY (`thread_id`,`tag_id`),
  KEY `tag_id` (`tag_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

--
-- Déchargement des données de la table `thread_tags`
--

INSERT INTO `thread_tags` (`thread_id`, `tag_id`) VALUES
(1, 11),
(1, 15),
(2, 21),
(2, 22);

-- --------------------------------------------------------

--
-- Structure de la table `users`
--

DROP TABLE IF EXISTS `users`;
CREATE TABLE IF NOT EXISTS `users` (
  `id` int NOT NULL AUTO_INCREMENT,
  `username` varchar(30) COLLATE utf8mb4_unicode_ci NOT NULL,
  `email` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `password` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `is_admin` tinyint(1) DEFAULT '0',
  `profile_pic` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `biography` text COLLATE utf8mb4_unicode_ci,
  `last_connection` timestamp NULL DEFAULT NULL,
  `message_count` int DEFAULT '0',
  `thread_count` int DEFAULT '0',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `username` (`username`),
  UNIQUE KEY `email` (`email`),
  KEY `idx_username` (`username`),
  KEY `idx_email` (`email`)
) ENGINE=InnoDB AUTO_INCREMENT=7 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

--
-- Déchargement des données de la table `users`
--

INSERT INTO `users` (`id`, `username`, `email`, `password`, `is_admin`, `profile_pic`, `biography`, `last_connection`, `message_count`, `thread_count`, `created_at`, `updated_at`) VALUES
(1, 'admin', 'admin@rythmit.com', '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewKyNiGH8IJ.2XpO', 1, NULL, NULL, NULL, 0, 0, '2025-06-28 17:22:08', '2025-06-28 17:22:08'),
(2, 'MC_Rhyme', 'mc.rhyme@test.com', '$2a$10$abcdefghijklmnopqrstuvwxyz', 0, NULL, NULL, NULL, 0, 0, '2025-06-28 17:22:08', '2025-06-28 17:22:08'),
(3, 'RapMaster', 'rap.master@test.com', '$2a$10$abcdefghijklmnopqrstuvwxyz', 0, NULL, NULL, NULL, 0, 0, '2025-06-28 17:22:08', '2025-06-28 17:22:08'),
(4, 'FlowKing', 'flow.king@test.com', '$2a$10$abcdefghijklmnopqrstuvwxyz', 0, NULL, NULL, NULL, 0, 0, '2025-06-28 17:22:08', '2025-06-28 17:22:08'),
(5, 'BeatQueen', 'beat.queen@test.com', '$2a$10$abcdefghijklmnopqrstuvwxyz', 0, NULL, NULL, NULL, 0, 0, '2025-06-28 17:22:08', '2025-06-28 17:22:08'),
(6, 'Dimitri', 'dimitri@gmail.com', '$2a$12$AiqxCPjciPnWCAsi8i0GNeXGjCJkCDZbkzRxhnFc.Av/RQQC16u92', 0, NULL, NULL, '2025-06-28 17:23:54', 0, 0, '2025-06-28 17:23:43', '2025-06-28 17:23:54');

-- --------------------------------------------------------

--
-- Structure de la table `user_music_preferences`
--

DROP TABLE IF EXISTS `user_music_preferences`;
CREATE TABLE IF NOT EXISTS `user_music_preferences` (
  `user_id` int NOT NULL,
  `tag_id` int NOT NULL,
  `score` int DEFAULT '0',
  PRIMARY KEY (`user_id`,`tag_id`),
  KEY `tag_id` (`tag_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- --------------------------------------------------------

--
-- Structure de la table `user_profiles`
--

DROP TABLE IF EXISTS `user_profiles`;
CREATE TABLE IF NOT EXISTS `user_profiles` (
  `id` int NOT NULL AUTO_INCREMENT,
  `user_id` int NOT NULL,
  `display_name` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `avatar_image` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `banner_image` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `user_id` (`user_id`),
  KEY `idx_user_id` (`user_id`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

--
-- Déchargement des données de la table `user_profiles`
--

INSERT INTO `user_profiles` (`id`, `user_id`, `display_name`, `avatar_image`, `banner_image`, `created_at`, `updated_at`) VALUES
(1, 6, NULL, '/uploads/profiles/1751133578_01cbe966ca91.jpg', '/uploads/profiles/1751133597_c34ca701624f.jpg', '2025-06-28 17:24:47', '2025-06-28 17:59:58');

--
-- Contraintes pour les tables déchargées
--

--
-- Contraintes pour la table `battles`
--
ALTER TABLE `battles`
  ADD CONSTRAINT `battles_ibfk_1` FOREIGN KEY (`contestant1_id`) REFERENCES `users` (`id`) ON DELETE CASCADE,
  ADD CONSTRAINT `battles_ibfk_2` FOREIGN KEY (`contestant2_id`) REFERENCES `users` (`id`) ON DELETE CASCADE;

--
-- Contraintes pour la table `battle_votes`
--
ALTER TABLE `battle_votes`
  ADD CONSTRAINT `battle_votes_ibfk_1` FOREIGN KEY (`battle_id`) REFERENCES `battles` (`id`) ON DELETE CASCADE,
  ADD CONSTRAINT `battle_votes_ibfk_2` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE,
  ADD CONSTRAINT `battle_votes_ibfk_3` FOREIGN KEY (`contestant_id`) REFERENCES `users` (`id`) ON DELETE CASCADE;

--
-- Contraintes pour la table `friendships`
--
ALTER TABLE `friendships`
  ADD CONSTRAINT `friendships_ibfk_1` FOREIGN KEY (`user_id_1`) REFERENCES `users` (`id`) ON DELETE CASCADE,
  ADD CONSTRAINT `friendships_ibfk_2` FOREIGN KEY (`user_id_2`) REFERENCES `users` (`id`) ON DELETE CASCADE;

--
-- Contraintes pour la table `messages`
--
ALTER TABLE `messages`
  ADD CONSTRAINT `messages_ibfk_1` FOREIGN KEY (`thread_id`) REFERENCES `threads` (`id`) ON DELETE CASCADE,
  ADD CONSTRAINT `messages_ibfk_2` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE;

--
-- Contraintes pour la table `message_votes`
--
ALTER TABLE `message_votes`
  ADD CONSTRAINT `message_votes_ibfk_1` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE,
  ADD CONSTRAINT `message_votes_ibfk_2` FOREIGN KEY (`message_id`) REFERENCES `messages` (`id`) ON DELETE CASCADE;

--
-- Contraintes pour la table `threads`
--
ALTER TABLE `threads`
  ADD CONSTRAINT `threads_ibfk_1` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE;

--
-- Contraintes pour la table `thread_likes`
--
ALTER TABLE `thread_likes`
  ADD CONSTRAINT `thread_likes_ibfk_1` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE,
  ADD CONSTRAINT `thread_likes_ibfk_2` FOREIGN KEY (`thread_id`) REFERENCES `threads` (`id`) ON DELETE CASCADE;

--
-- Contraintes pour la table `thread_tags`
--
ALTER TABLE `thread_tags`
  ADD CONSTRAINT `thread_tags_ibfk_1` FOREIGN KEY (`thread_id`) REFERENCES `threads` (`id`) ON DELETE CASCADE,
  ADD CONSTRAINT `thread_tags_ibfk_2` FOREIGN KEY (`tag_id`) REFERENCES `tags` (`id`) ON DELETE CASCADE;

--
-- Contraintes pour la table `user_music_preferences`
--
ALTER TABLE `user_music_preferences`
  ADD CONSTRAINT `user_music_preferences_ibfk_1` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE,
  ADD CONSTRAINT `user_music_preferences_ibfk_2` FOREIGN KEY (`tag_id`) REFERENCES `tags` (`id`) ON DELETE CASCADE;

--
-- Contraintes pour la table `user_profiles`
--
ALTER TABLE `user_profiles`
  ADD CONSTRAINT `user_profiles_ibfk_1` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE;
COMMIT;

/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
