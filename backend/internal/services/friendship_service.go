package services

import (
	"fmt"
	"rythmitbackend/internal/models"
	"rythmitbackend/internal/repositories"
)

// FriendshipService interface pour la logique métier des amitiés
type FriendshipService interface {
	// Gestion des demandes d'amitié
	SendFriendRequest(requesterID, addresseeID uint) error
	AcceptFriendRequest(requesterID, addresseeID uint) error
	RejectFriendRequest(requesterID, addresseeID uint) error
	CancelFriendRequest(requesterID, addresseeID uint) error

	// Gestion des amis
	RemoveFriend(userID1, userID2 uint) error
	BlockUser(blockerID, blockedID uint) error
	UnblockUser(blockerID, blockedID uint) error

	// Récupération des données
	GetFriends(userID uint) ([]*models.Friend, error)
	GetFriendRequests(userID uint) ([]*models.FriendRequest, error)
	GetSentRequests(userID uint) ([]*models.FriendRequest, error)
	GetMutualFriends(userID1, userID2 uint) ([]*models.Friend, error)

	// Recherche et découverte
	SearchUsers(query string, currentUserID uint, limit int) ([]*models.UserSearchResult, error)
	GetSuggestedFriends(userID uint, limit int) ([]*models.UserSearchResult, error)

	// Statistiques et informations
	GetFriendshipStats(userID uint) (*FriendshipStats, error)
	GetFriendshipStatus(userID1, userID2 uint) (*string, error)

	// Vérifications
	AreFriends(userID1, userID2 uint) (bool, error)
	CanSendRequest(requesterID, addresseeID uint) (bool, string, error)
}

// FriendshipStats représente les statistiques d'amitié d'un utilisateur
type FriendshipStats struct {
	FriendsCount         int `json:"friends_count"`
	PendingRequestsCount int `json:"pending_requests_count"`
	SentRequestsCount    int `json:"sent_requests_count"`
}

// friendshipService implémentation concrète
type friendshipService struct {
	friendshipRepo repositories.FriendshipRepository
	userRepo       repositories.UserRepository
}

// NewFriendshipService crée une nouvelle instance du service
func NewFriendshipService(friendshipRepo repositories.FriendshipRepository, userRepo repositories.UserRepository) FriendshipService {
	return &friendshipService{
		friendshipRepo: friendshipRepo,
		userRepo:       userRepo,
	}
}

// SendFriendRequest envoie une demande d'amitié avec validations
func (s *friendshipService) SendFriendRequest(requesterID, addresseeID uint) error {
	// Vérifications préliminaires
	canSend, reason, err := s.CanSendRequest(requesterID, addresseeID)
	if err != nil {
		return fmt.Errorf("erreur vérification envoi demande: %w", err)
	}

	if !canSend {
		return fmt.Errorf("impossible d'envoyer la demande: %s", reason)
	}

	// Vérifier que l'utilisateur destinataire existe
	_, err = s.userRepo.FindByID(addresseeID)
	if err != nil {
		return fmt.Errorf("utilisateur destinataire non trouvé")
	}

	// Envoyer la demande
	err = s.friendshipRepo.SendFriendRequest(requesterID, addresseeID)
	if err != nil {
		return fmt.Errorf("erreur envoi demande d'amitié: %w", err)
	}

	return nil
}

// AcceptFriendRequest accepte une demande d'amitié
func (s *friendshipService) AcceptFriendRequest(requesterID, addresseeID uint) error {
	// Vérifier que la demande existe et est en attente
	hasPending, err := s.friendshipRepo.HasPendingRequest(requesterID, addresseeID)
	if err != nil {
		return fmt.Errorf("erreur vérification demande en attente: %w", err)
	}

	if !hasPending {
		return fmt.Errorf("aucune demande d'amitié en attente trouvée")
	}

	// Accepter la demande
	err = s.friendshipRepo.AcceptFriendRequest(requesterID, addresseeID)
	if err != nil {
		return fmt.Errorf("erreur acceptation demande: %w", err)
	}

	return nil
}

// RejectFriendRequest rejette une demande d'amitié
func (s *friendshipService) RejectFriendRequest(requesterID, addresseeID uint) error {
	// Vérifier que la demande existe et est en attente
	hasPending, err := s.friendshipRepo.HasPendingRequest(requesterID, addresseeID)
	if err != nil {
		return fmt.Errorf("erreur vérification demande en attente: %w", err)
	}

	if !hasPending {
		return fmt.Errorf("aucune demande d'amitié en attente trouvée")
	}

	// Rejeter la demande
	err = s.friendshipRepo.RejectFriendRequest(requesterID, addresseeID)
	if err != nil {
		return fmt.Errorf("erreur rejet demande: %w", err)
	}

	return nil
}

// CancelFriendRequest annule une demande d'amitié envoyée
func (s *friendshipService) CancelFriendRequest(requesterID, addresseeID uint) error {
	// Vérifier que la demande existe et est en attente
	hasPending, err := s.friendshipRepo.HasPendingRequest(requesterID, addresseeID)
	if err != nil {
		return fmt.Errorf("erreur vérification demande en attente: %w", err)
	}

	if !hasPending {
		return fmt.Errorf("aucune demande d'amitié en attente trouvée")
	}

	// Annuler la demande (même logique que rejeter)
	err = s.friendshipRepo.RejectFriendRequest(requesterID, addresseeID)
	if err != nil {
		return fmt.Errorf("erreur annulation demande: %w", err)
	}

	return nil
}

// RemoveFriend supprime une amitié
func (s *friendshipService) RemoveFriend(userID1, userID2 uint) error {
	// Vérifier que les utilisateurs sont amis
	areFriends, err := s.friendshipRepo.AreFriends(userID1, userID2)
	if err != nil {
		return fmt.Errorf("erreur vérification amitié: %w", err)
	}

	if !areFriends {
		return fmt.Errorf("les utilisateurs ne sont pas amis")
	}

	// Supprimer l'amitié
	err = s.friendshipRepo.RemoveFriend(userID1, userID2)
	if err != nil {
		return fmt.Errorf("erreur suppression amitié: %w", err)
	}

	return nil
}

// BlockUser bloque un utilisateur
func (s *friendshipService) BlockUser(blockerID, blockedID uint) error {
	if blockerID == blockedID {
		return fmt.Errorf("impossible de se bloquer soi-même")
	}

	// Vérifier que l'utilisateur à bloquer existe
	_, err := s.userRepo.FindByID(blockedID)
	if err != nil {
		return fmt.Errorf("utilisateur à bloquer non trouvé")
	}

	// Bloquer l'utilisateur
	err = s.friendshipRepo.BlockUser(blockerID, blockedID)
	if err != nil {
		return fmt.Errorf("erreur blocage utilisateur: %w", err)
	}

	return nil
}

// UnblockUser débloque un utilisateur
func (s *friendshipService) UnblockUser(blockerID, blockedID uint) error {
	// Vérifier que l'utilisateur est bloqué
	isBlocked, err := s.friendshipRepo.IsBlocked(blockerID, blockedID)
	if err != nil {
		return fmt.Errorf("erreur vérification blocage: %w", err)
	}

	if !isBlocked {
		return fmt.Errorf("l'utilisateur n'est pas bloqué")
	}

	// Débloquer l'utilisateur
	err = s.friendshipRepo.UnblockUser(blockerID, blockedID)
	if err != nil {
		return fmt.Errorf("erreur déblocage utilisateur: %w", err)
	}

	return nil
}

// GetFriends récupère la liste des amis avec enrichissement des données
func (s *friendshipService) GetFriends(userID uint) ([]*models.Friend, error) {
	friends, err := s.friendshipRepo.GetFriends(userID)
	if err != nil {
		return nil, fmt.Errorf("erreur récupération amis: %w", err)
	}

	// Enrichir les données des amis (activité, etc.)
	for _, friend := range friends {
		// Ici on pourrait ajouter la logique pour récupérer l'activité actuelle
		// Par exemple, la dernière piste écoutée, le statut personnalisé, etc.
		s.enrichFriendData(friend)
	}

	return friends, nil
}

// GetFriendRequests récupère les demandes d'amitié reçues
func (s *friendshipService) GetFriendRequests(userID uint) ([]*models.FriendRequest, error) {
	requests, err := s.friendshipRepo.GetFriendRequests(userID)
	if err != nil {
		return nil, fmt.Errorf("erreur récupération demandes d'amitié: %w", err)
	}

	return requests, nil
}

// GetSentRequests récupère les demandes d'amitié envoyées
func (s *friendshipService) GetSentRequests(userID uint) ([]*models.FriendRequest, error) {
	requests, err := s.friendshipRepo.GetSentRequests(userID)
	if err != nil {
		return nil, fmt.Errorf("erreur récupération demandes envoyées: %w", err)
	}

	return requests, nil
}

// GetMutualFriends récupère les amis mutuels entre deux utilisateurs
func (s *friendshipService) GetMutualFriends(userID1, userID2 uint) ([]*models.Friend, error) {
	friends, err := s.friendshipRepo.GetMutualFriends(userID1, userID2)
	if err != nil {
		return nil, fmt.Errorf("erreur récupération amis mutuels: %w", err)
	}

	return friends, nil
}

// SearchUsers recherche des utilisateurs avec informations d'amitié
func (s *friendshipService) SearchUsers(query string, currentUserID uint, limit int) ([]*models.UserSearchResult, error) {
	if len(query) < 2 {
		return nil, fmt.Errorf("la recherche doit contenir au moins 2 caractères")
	}

	if limit <= 0 || limit > 50 {
		limit = 20 // Limite par défaut
	}

	users, err := s.friendshipRepo.SearchUsers(query, currentUserID, limit)
	if err != nil {
		return nil, fmt.Errorf("erreur recherche utilisateurs: %w", err)
	}

	return users, nil
}

// GetSuggestedFriends récupère des suggestions d'amis basées sur les amis mutuels
func (s *friendshipService) GetSuggestedFriends(userID uint, limit int) ([]*models.UserSearchResult, error) {
	if limit <= 0 || limit > 20 {
		limit = 10 // Limite par défaut pour les suggestions
	}

	// Pour l'instant, une implémentation simple qui récupère des utilisateurs aléatoirement
	// Dans une vraie application, on utiliserait des algorithmes plus sophistiqués
	users, err := s.friendshipRepo.SearchUsers("", userID, limit*3) // Récupérer plus pour filtrer
	if err != nil {
		return nil, fmt.Errorf("erreur récupération suggestions: %w", err)
	}

	// Filtrer pour ne garder que ceux qui ne sont pas déjà amis ou en demande
	var suggestions []*models.UserSearchResult
	for _, user := range users {
		if user.FriendshipStatus == nil && len(suggestions) < limit {
			suggestions = append(suggestions, user)
		}
	}

	return suggestions, nil
}

// GetFriendshipStats récupère les statistiques d'amitié d'un utilisateur
func (s *friendshipService) GetFriendshipStats(userID uint) (*FriendshipStats, error) {
	stats := &FriendshipStats{}

	// Compter les amis
	friendsCount, err := s.friendshipRepo.GetFriendsCount(userID)
	if err != nil {
		return nil, fmt.Errorf("erreur comptage amis: %w", err)
	}
	stats.FriendsCount = friendsCount

	// Compter les demandes reçues en attente
	pendingCount, err := s.friendshipRepo.GetPendingRequestsCount(userID)
	if err != nil {
		return nil, fmt.Errorf("erreur comptage demandes en attente: %w", err)
	}
	stats.PendingRequestsCount = pendingCount

	// Compter les demandes envoyées
	sentRequests, err := s.GetSentRequests(userID)
	if err != nil {
		return nil, fmt.Errorf("erreur comptage demandes envoyées: %w", err)
	}
	stats.SentRequestsCount = len(sentRequests)

	return stats, nil
}

// GetFriendshipStatus récupère le statut d'amitié entre deux utilisateurs
func (s *friendshipService) GetFriendshipStatus(userID1, userID2 uint) (*string, error) {
	status, err := s.friendshipRepo.GetFriendshipStatus(userID1, userID2)
	if err != nil {
		return nil, fmt.Errorf("erreur récupération statut amitié: %w", err)
	}

	return status, nil
}

// AreFriends vérifie si deux utilisateurs sont amis
func (s *friendshipService) AreFriends(userID1, userID2 uint) (bool, error) {
	areFriends, err := s.friendshipRepo.AreFriends(userID1, userID2)
	if err != nil {
		return false, fmt.Errorf("erreur vérification amitié: %w", err)
	}

	return areFriends, nil
}

// CanSendRequest vérifie si un utilisateur peut envoyer une demande d'amitié
func (s *friendshipService) CanSendRequest(requesterID, addresseeID uint) (bool, string, error) {
	// Vérifier qu'on n'essaie pas de s'ajouter soi-même
	if requesterID == addresseeID {
		return false, "impossible de s'ajouter soi-même comme ami", nil
	}

	// Vérifier s'ils sont déjà amis
	areFriends, err := s.friendshipRepo.AreFriends(requesterID, addresseeID)
	if err != nil {
		return false, "", fmt.Errorf("erreur vérification amitié: %w", err)
	}

	if areFriends {
		return false, "vous êtes déjà amis", nil
	}

	// Vérifier s'il y a déjà une demande en attente (dans les deux sens)
	hasPendingRequest, err := s.friendshipRepo.HasPendingRequest(requesterID, addresseeID)
	if err != nil {
		return false, "", fmt.Errorf("erreur vérification demande en attente: %w", err)
	}

	if hasPendingRequest {
		return false, "demande d'amitié déjà envoyée", nil
	}

	// Vérifier s'il y a une demande dans l'autre sens
	hasIncomingRequest, err := s.friendshipRepo.HasPendingRequest(addresseeID, requesterID)
	if err != nil {
		return false, "", fmt.Errorf("erreur vérification demande entrante: %w", err)
	}

	if hasIncomingRequest {
		return false, "cet utilisateur vous a déjà envoyé une demande d'amitié", nil
	}

	// Vérifier si l'utilisateur est bloqué
	isBlocked, err := s.friendshipRepo.IsBlocked(addresseeID, requesterID)
	if err != nil {
		return false, "", fmt.Errorf("erreur vérification blocage: %w", err)
	}

	if isBlocked {
		return false, "impossible d'envoyer une demande à cet utilisateur", nil
	}

	return true, "", nil
}

// enrichFriendData enrichit les données d'un ami avec des informations supplémentaires
func (s *friendshipService) enrichFriendData(friend *models.Friend) {
	// Ici on pourrait ajouter la logique pour récupérer:
	// - L'activité musicale actuelle
	// - Le statut personnalisé
	// - Les playlists récentes
	// - etc.

	// Pour l'instant, on simule une activité basée sur le statut en ligne
	if friend.OnlineStatus == "online" {
		activities := []string{
			"🎵 Écoute: \"Chill Vibes Mix\"",
			"🎧 Découvre de nouveaux artistes",
			"🎼 Crée une playlist",
			"🎤 Partage ses découvertes",
		}

		// Utiliser l'ID de l'ami pour avoir une activité "stable"
		activityIndex := int(friend.ID) % len(activities)
		activity := activities[activityIndex]
		friend.Activity = &activity
	}
}
