# Rythm'it - Contexte Claude

## 🎵 Vue d'ensemble du projet

**Nom** : Rythm'it  
**Type** : Forum musical interactif avec système de battles d'artistes  
**Équipe** : Dimitri (Backend Go) + Romain (Frontend JS/HTML/CSS)  
**Deadline** : 16/06/2025 23h59  
**Repo** : [URL à compléter]

## 🎯 Concept unique

Forum musical où les utilisateurs peuvent :
- Créer des discussions sur des artistes/albums/genres
- Lancer des battles musicales (ex: Drake vs Kendrick)
- Voter avec Fire 🔥 (like) ou Skip ⏭️ (dislike)  
- Découvrir leur compatibilité musicale avec d'autres users
- Intégrer des embeds YouTube/Spotify dans les messages

## 📋 Spécifications techniques obligatoires

### Contraintes imposées
- **Backend** : Go obligatoire (architecture MVC)
- **Frontend** : HTML/CSS/JavaScript (pas de framework imposé)
- **Auth** : JWT + hash SHA512 minimum pour mots de passe
- **BDD** : Persistance obligatoire (MySQL recommandé)
- **Refresh** : JavaScript pour éviter les rechargements de page

### Fonctionnalités obligatoires (FT-1 à FT-12)
1. **FT-1** : Inscription (username unique, email unique, mdp 12+ chars + majuscule + spécial)
2. **FT-2** : Connexion (username/email + mdp → JWT)
3. **FT-3** : Création threads (titre, desc, tags, état: ouvert/fermé/archivé)
4. **FT-4** : Consultation threads (même non-auth)
5. **FT-5** : Messages dans threads (auth uniquement)
6. **FT-6** : Fire/Skip sur messages (pas les deux simultanément)
7. **FT-7** : CRUD propriétaire + admin peut tout supprimer
8. **FT-8** : Tri messages (chronologique OU popularité via Fire/Skip)
9. **FT-9** : Pagination (10/20/30/tout, défaut: 10)
10. **FT-10** : Filtrage par tags
11. **FT-11** : Recherche (titre OU tags)
12. **FT-12** : Dashboard admin (modérer, bannir, changer états)

## 🗄️ Structure Base de Données

### Tables principales
```sql
User (User_Id, username, email, password, is_Admin, profile_pic, Biographie, last_connection, message_count, threat_count)

thread (Thread_Id, title, desc_, creation, state, visibility, User_Id)

Message (Message_Id, content, date_, Thread_Id, User_Id)

tag (tag_id, name)

have_tag (Thread_Id, tag_Id) -- relation N:N

liked_disliked (User_Id, Message_Id, state) -- Fire/Skip

friendship (Friendship_id, status, request_date, response_date, User_Id, User_Id_1) -- Musical Twins
```

### États des threads
- **ouvert** : consultation + nouveaux messages OK
- **fermé** : consultation OK, nouveaux messages NON
- **archivé** : plus visible ni accessible

### États Fire/Skip
- **Fire** : +1 au score popularité
- **Skip** : -1 au score popularité
- **Neutre** : aucun vote

## 🔗 API Endpoints (Dimitri → Romain)

### Authentification
```
POST /api/register
POST /api/login
GET /api/profile
```

### Threads
```
GET /api/threads?page=1&limit=10&tag=&search=
POST /api/threads
GET /api/threads/:id
PUT /api/threads/:id (propriétaire ou admin)
DELETE /api/threads/:id (propriétaire ou admin)
```

### Messages
```
GET /api/threads/:id/messages?page=1&limit=10&sort=date|popularity
POST /api/threads/:id/messages
POST /api/messages/:id/fire
POST /api/messages/:id/skip
DELETE /api/messages/:id (propriétaire ou admin)
```

### Battles (spécifique Rythm'it)
```
POST /api/battles
GET /api/battles/:id
POST /api/battles/:id/vote
GET /api/battles/active
```

### Musique (intégrations)
```
GET /api/search/artists?q=
GET /api/search/albums?q=
GET /api/compatibility/:userId
```

### Admin
```
GET /api/admin/dashboard
PUT /api/admin/threads/:id/state
DELETE /api/admin/threads/:id
DELETE /api/admin/messages/:id
POST /api/admin/users/:id/ban
```

## 📊 Formats JSON critiques

### Thread Response
```json
{
  "id": 1,
  "title": "Drake vs Kendrick : qui est le GOAT ?",
  "description": "Battle épique entre les deux kings...",
  "tags": ["rap", "drake", "kendrick"],
  "creation_date": "2025-05-23T14:30:00Z",
  "author": {
    "id": 1,
    "username": "MusicLover"
  },
  "state": "ouvert",
  "message_count": 42,
  "fire_count": 15,
  "skip_count": 3
}
```

### Message Response
```json
{
  "id": 1,
  "content": "Kendrick > Drake fight me 🔥",
  "date": "2025-05-23T14:35:00Z",
  "author": {
    "id": 2,
    "username": "KendrickFan"
  },
  "popularity_score": 8,
  "user_vote": "fire", // null, "fire", ou "skip"
  "embeds": {
    "youtube": "https://youtube.com/watch?v=...",
    "spotify": "https://open.spotify.com/track/..."
  }
}
```

### Battle Response
```json
{
  "id": 1,
  "title": "Drake vs Kendrick Lamar",
  "options": [
    {
      "name": "Drake",
      "image": "https://...",
      "votes": 142
    },
    {
      "name": "Kendrick Lamar", 
      "image": "https://...",
      "votes": 189
    }
  ],
  "total_votes": 331,
  "user_vote": "kendrick", // null si pas voté
  "status": "active", // active, ended
  "end_date": "2025-05-30T23:59:59Z"
}
```

## 🎨 Spécificités UI/UX (Romain)

### Thème musical
- **Couleurs** : Dégradés sombres avec accents dorés/violets
- **Icônes** : Vinyles, casques, ondes sonores, feu/skip
- **Animations** : Equalizer, rotation vinyle, ondes lors des votes
- **Fonts** : Modern, légèrement "street" pour le côté musical

### Éléments clés
- **Cards threads** : Style vinyle/cassette/CD selon le genre
- **Fire/Skip buttons** : Animations au hover et clic
- **Battle interface** : Barres de progression animées
- **Embeds** : Intégration native YouTube/Spotify
- **Loading states** : Equalizer animé

## 🔄 Features spécifiques Rythm'it

### Battle System
- **Création** : Drag & drop de 2 artistes/albums
- **Vote** : Animation de disque qui penche
- **Résultats** : Effets visuels pour le gagnant
- **Historique** : Battles gagnées/perdues par user

### Musical Twins
- **Compatibilité** : Calcul basé sur votes similaires
- **Affichage** : Pourcentage + graphique radar des goûts
- **Discovery** : Page "Find Your Musical Twin"

### Tags intelligents
- **Hiérarchie** : Genre > Artiste > Album
- **Auto-complétion** : Suggestions basées sur APIs externes
- **Visual** : Badges colorés par genre musical

## 🚀 Workflow de développement

### Communication
- **Daily** : 15min chaque matin
- **Sync points** : Validation formats JSON, endpoints
- **Urgences** : Discord pour questions rapides

### Git
- **Branches** : `feature/dimitri-backend` et `feature/romain-frontend`
- **Main** : Merge après review
- **Structure** :
  ```
  repo/
  ├── backend/ (Go, Dimitri)
  ├── frontend/ (HTML/CSS/JS, Romain)
  ├── docs/ (ce fichier + specs)
  └── database/ (scripts SQL)
  ```

### Phases critiques
1. **Phase 0-1** : Auth + architecture (priorité absolue)
2. **Phase 2-3** : Threads + messages (cœur fonctionnel)  
3. **Phase 4** : Battle system (différenciation)
4. **Phase 5+** : Features bonus si temps

## ⚠️ Points d'attention

### Sécurité
- **Mots de passe** : bcrypt recommandé (plus sûr que SHA512)
- **JWT** : Expiration + refresh token
- **Validation** : Côté client ET serveur
- **XSS** : Sanitisation des messages utilisateur

### Performance  
- **Pagination** : Obligatoire pour scalabilité
- **Cache** : APIs externes musicales (limites de rate)
- **Optimisation** : Index sur colonnes recherchées fréquemment

### Intégrations externes
- **YouTube API** : Quota limitations
- **Spotify Web API** : Client credentials flow
- **Embeds** : Vérification des URLs avant intégration

## 📝 Documentation requise

### README.md
- Installation et lancement
- Liste des routes (vues VS API)
- Composition équipe

### Rapport de projet
- Décomposition en phases
- Répartition des tâches  
- Gestion du temps et priorités
- Stratégie de documentation

### Soutenance (10min + 5min questions)
- Pitch du projet
- Architecture technique
- Démonstration live
- Organisation équipe
- Difficultés et solutions

---

**Dernière mise à jour** : [À compléter à chaque modification importante]  
**Prochaine sync** : [Date du prochain point équipe]