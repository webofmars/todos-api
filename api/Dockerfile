FROM golang:1.21-alpine AS builder

# Installation des dépendances système pour SQLite
RUN apk add --no-cache git gcc musl-dev sqlite-dev

WORKDIR /app

# Copie des fichiers de dépendances
COPY go.mod go.sum ./

# Téléchargement des dépendances
RUN go mod download

# Copie du code source
COPY . .

# Compilation de l'application avec CGO activé pour SQLite
# Use static linking to avoid runtime issues
ENV CGO_ENABLED=1
ENV GOOS=linux
ENV CGO_CFLAGS="-D_LARGEFILE64_SOURCE"
RUN go build -a -ldflags '-linkmode external -extldflags "-static"' -o main .

# Image finale
FROM alpine:latest

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
