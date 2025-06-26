# Todo Application

Cette application est composée de deux microservices avec une base de données SQLite locale :

## 🎯 Architecture

- **Frontend React** : Interface utilisateur pour gérer les tâches quotidiennes
- **API Golang** : API RESTful pour la gestion des todos
- **Base de données SQLite** : Stockage persistent local (zéro configuration requise !)

## ✨ Avantages de SQLite

- **Zéro configuration** : Pas de serveur de base de données à configurer
- **Fichier unique** : Toutes les données dans un seul fichier
- **Démarrage instantané** : Plus d'attente pour l'initialisation de la DB
- **Léger** : Parfait pour le développement et les petites applications
- **Portable** : Le fichier de données peut être facilement sauvegardé/partagé
- **Multiplateforme** : Forcé en AMD64 pour compatibilité ARM64/Intel

## 🚀 Démarrage rapide

### Prérequis
- Docker et Docker Compose installés
- Ports 3000 et 8080 disponibles

### Démarrage avec Docker Compose

```bash
# Se placer dans le dossier src
cd src

# Démarrer tous les services
docker-compose up -d

# Voir les logs
docker-compose logs -f

# Arrêter les services
docker-compose down
```

### Accès aux services

- **Frontend** : <http://localhost:3000>
- **API** : <http://localhost:8080>
- **Health check API** : <http://localhost:8080/health>

## 📁 Structure du projet

```text
src/
├── frontend/              # Application React
│   ├── src/
│   │   ├── App.js        # Composant principal
│   │   ├── index.js      # Point d'entrée
│   │   └── index.css     # Styles
│   ├── public/
│   ├── package.json
│   ├── Dockerfile
│   └── nginx.conf
├── api/                   # API Golang
│   ├── main.go           # Serveur API avec SQLite
│   ├── go.mod
│   ├── Dockerfile
│   └── .gitignore
├── docker-compose.yml     # Configuration Docker Compose
└── README.md
```

## 🔧 API Endpoints

- `GET /api/todos` - Récupérer tous les todos
- `POST /api/todos` - Créer un nouveau todo
- `GET /api/todos/{id}` - Récupérer un todo spécifique
- `PUT /api/todos/{id}` - Mettre à jour un todo
- `DELETE /api/todos/{id}` - Supprimer un todo
- `GET /health` - Vérification de santé de l'API

## 📊 Variables d'environnement

### API (Golang)
```env
DB_PATH=/root/data/todos.db
PORT=8080
```

### Frontend (React)
```env
REACT_APP_API_URL=http://localhost:8080/api
```

## 🏗️ Build des images Docker

### Frontend
```bash
cd frontend
docker build -t todo-frontend .
```

### API
```bash
cd api
docker build -t todo-api .
```

## 🧪 Test de l'API

```bash
# Créer un todo
curl -X POST http://localhost:8080/api/todos \
  -H "Content-Type: application/json" \
  -d '{"title": "Test todo", "completed": false}'

# Récupérer tous les todos
curl http://localhost:8080/api/todos

# Health check
curl http://localhost:8080/health
```

## 💾 Persistance des données

Les données SQLite sont stockées dans un volume Docker nommé `api_data`. Elles persistent entre les redémarrages des conteneurs.

```bash
# Voir les volumes
docker volume ls

# Inspecter le volume des données
docker volume inspect src_api_data

# Sauvegarder les données
docker run --rm -v src_api_data:/data -v $(pwd):/backup alpine tar czf /backup/todos-backup.tar.gz -C /data .

# Restaurer les données
docker run --rm -v src_api_data:/data -v $(pwd):/backup alpine tar xzf /backup/todos-backup.tar.gz -C /data
```

## 🐛 Dépannage

### Problèmes courants

1. **Port déjà utilisé**
   - Modifier les ports dans `docker-compose.yml`
   - Vérifier avec `netstat -an | grep :3000`

2. **Données corrompues**
   - Supprimer le volume : `docker volume rm src_api_data`
   - Redémarrer : `docker-compose up -d`

3. **Problème CORS**
   - L'API est configurée pour accepter toutes les origines
   - Vérifier la configuration nginx pour le frontend

4. **Problème de compilation SQLite sur ARM64/Apple Silicon**
   - Les conteneurs sont forcés en AMD64 pour éviter les problèmes CGO/SQLite
   - Si vous avez des erreurs de build, essayez : `docker-compose build --no-cache`

### Logs utiles

```bash
# Logs de tous les services
docker-compose logs

# Logs d'un service spécifique
docker-compose logs api
docker-compose logs frontend

# Logs en temps réel
docker-compose logs -f api
```

## 🔄 Redémarrage des services

```bash
# Redémarrer un service spécifique
docker-compose restart api

# Reconstruire et redémarrer
docker-compose up --build -d

# Nettoyer et redémarrer (supprime les données !)
docker-compose down -v
docker-compose up -d
```
