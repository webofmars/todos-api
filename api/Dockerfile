FROM --platform=linux/amd64 golang:1.21-alpine AS builder

# Installation des dépendances système pour SQLite
RUN apk add --no-cache git gcc musl-dev sqlite-dev

WORKDIR /app

# Copie des fichiers de dépendances
COPY go.mod go.sum ./

# Téléchargement des dépendances
RUN go mod download

# Copie du code source
COPY . .

# Compilation de l'application avec CGO activé pour SQLite (force AMD64)
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -a -ldflags '-linkmode external -extldflags "-static"' -o main .

# Image finale (force AMD64)
FROM --platform=linux/amd64 alpine:latest

RUN apk --no-cache add ca-certificates sqlite
WORKDIR /root/

# Création du dossier pour la base de données
RUN mkdir -p /root/data

# Copie de l'exécutable depuis l'étape de build
COPY --from=builder /app/main .

# Volume pour persister les données SQLite
VOLUME ["/root/data"]

# Exposition du port
EXPOSE 8080

# Variables d'environnement
ENV DB_PATH=/root/data/todos.db

# Commande par défaut
CMD ["./main"]
