package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
)

// LoggingMiddleware logs all HTTP requests
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a custom ResponseWriter to capture status code
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		// Log incoming request
		log.Printf("→ [%s] %s %s - Remote: %s - User-Agent: %s",
			r.Method, r.RequestURI, r.Proto, r.RemoteAddr, r.UserAgent())

		// Process request
		next.ServeHTTP(wrapped, r)

		// Log response
		duration := time.Since(start)
		log.Printf("← [%s] %s - Status: %d - Duration: %v",
			r.Method, r.RequestURI, wrapped.statusCode, duration)
	})
}

// Custom ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// NotFoundHandler handles 404 errors with logging
func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("⚠️  Route not found: [%s] %s", r.Method, r.RequestURI)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
	json.NewEncoder(w).Encode(map[string]string{
		"error":  "Route not found",
		"method": r.Method,
		"path":   r.RequestURI,
	})
}

// MethodNotAllowedHandler handles 405 errors with logging
func methodNotAllowedHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("⚠️  Method not allowed: [%s] %s", r.Method, r.RequestURI)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusMethodNotAllowed)
	json.NewEncoder(w).Encode(map[string]string{
		"error":  "Method not allowed",
		"method": r.Method,
		"path":   r.RequestURI,
	})
}

// Todo représente une tâche
type Todo struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Completed bool      `json:"completed"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Database connection
var db *sql.DB

func main() {
	// Configuration de la base de données SQLite
	dbPath := getEnv("DB_PATH", "./data/todos.db")

	// Créer le dossier data s'il n'existe pas
	os.MkdirAll("./data", 0755)

	// Connexion à SQLite
	var err error
	db, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal("Erreur lors de l'ouverture de la base de données SQLite:", err)
	}
	defer db.Close()

	// Test de connexion
	if err := db.Ping(); err != nil {
		log.Fatal("Impossible de se connecter à la base de données SQLite:", err)
	}
	log.Printf("✅ Connexion SQLite établie: %s", dbPath)

	// Création des tables
	if err := createTables(); err != nil {
		log.Fatal("Erreur lors de la création des tables:", err)
	}

	// Configuration du routeur
	r := mux.NewRouter()

	// Middleware de logging
	r.Use(loggingMiddleware)

	// Handlers pour les erreurs
	r.NotFoundHandler = http.HandlerFunc(notFoundHandler)
	r.MethodNotAllowedHandler = http.HandlerFunc(methodNotAllowedHandler)

	// Routes API
	api := r.PathPrefix("/api").Subrouter()
	api.HandleFunc("/todos", getTodos).Methods("GET")
	api.HandleFunc("/todos", createTodo).Methods("POST")
	api.HandleFunc("/todos/{id}", getTodo).Methods("GET")
	api.HandleFunc("/todos/{id}", updateTodo).Methods("PUT")
	api.HandleFunc("/todos/{id}", deleteTodo).Methods("DELETE")

	// Route de santé
	r.HandleFunc("/health", healthCheck).Methods("GET")
	// Route racine - redirige vers health pour les health checks Kubernetes
	r.HandleFunc("/", healthCheck).Methods("GET")

	// Middleware CORS
	corsHandler := handlers.CORS(
		handlers.AllowedOrigins([]string{"*"}),
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
		handlers.AllowedHeaders([]string{"Content-Type", "Authorization"}),
	)(r)

	port := getEnv("PORT", "8080")
	log.Printf("🚀 Serveur démarré sur le port %s", port)
	log.Printf("📊 Routes disponibles:")
	log.Printf("   GET    /          (health check)")
	log.Printf("   GET    /health")
	log.Printf("   GET    /api/todos")
	log.Printf("   POST   /api/todos")
	log.Printf("   GET    /api/todos/{id}")
	log.Printf("   PUT    /api/todos/{id}")
	log.Printf("   DELETE /api/todos/{id}")
	log.Fatal(http.ListenAndServe(":"+port, corsHandler))
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func createTables() error {
	query := `
	CREATE TABLE IF NOT EXISTS todos (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		completed BOOLEAN DEFAULT FALSE,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	
	CREATE TRIGGER IF NOT EXISTS update_todos_timestamp 
		AFTER UPDATE ON todos
		BEGIN
			UPDATE todos SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
		END;`

	_, err := db.Exec(query)
	if err != nil {
		return err
	}

	// Insérer quelques données de test si la table est vide
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM todos").Scan(&count)
	if err != nil {
		return err
	}

	if count == 0 {
		testData := []struct {
			title     string
			completed bool
		}{
			{"Apprendre Docker", false},
			{"Créer un chart Helm", false},
			{"Déployer sur Kubernetes", false},
			{"Tester l'application", true},
		}

		for _, todo := range testData {
			_, err := db.Exec("INSERT INTO todos (title, completed) VALUES (?, ?)", todo.title, todo.completed)
			if err != nil {
				log.Printf("⚠️  Erreur lors de l'insertion des données de test: %v", err)
			}
		}
		log.Printf("✅ Données de test créées (%d todos)", len(testData))
	}

	log.Println("✅ Tables SQLite créées avec succès")
	return nil
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Test de connexion à la base
	if err := db.Ping(); err != nil {
		log.Printf("❌ Health check failed - DB connection error: %v", err)
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "unhealthy",
			"error":  "Database connection failed",
			"time":   time.Now().Format(time.RFC3339),
		})
		return
	}

	log.Printf("✅ Health check OK")
	response := map[string]string{
		"status": "healthy",
		"time":   time.Now().Format(time.RFC3339),
	}
	json.NewEncoder(w).Encode(response)
}

func getTodos(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	log.Printf("📋 Fetching all todos")
	rows, err := db.Query("SELECT id, title, completed, created_at, updated_at FROM todos ORDER BY created_at DESC")
	if err != nil {
		log.Printf("❌ Error querying todos: %v", err)
		http.Error(w, "Erreur lors de la récupération des todos", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var todos []Todo
	for rows.Next() {
		var todo Todo
		if err := rows.Scan(&todo.ID, &todo.Title, &todo.Completed, &todo.CreatedAt, &todo.UpdatedAt); err != nil {
			log.Printf("❌ Error scanning todo row: %v", err)
			http.Error(w, "Erreur lors de la lecture des données", http.StatusInternalServerError)
			return
		}
		todos = append(todos, todo)
	}

	if todos == nil {
		todos = []Todo{}
	}

	log.Printf("✅ Successfully fetched %d todos", len(todos))
	json.NewEncoder(w).Encode(todos)
}

func createTodo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var todo Todo
	if err := json.NewDecoder(r.Body).Decode(&todo); err != nil {
		log.Printf("❌ Invalid JSON in createTodo: %v", err)
		http.Error(w, "Données JSON invalides", http.StatusBadRequest)
		return
	}

	if todo.Title == "" {
		log.Printf("⚠️  Empty title in createTodo request")
		http.Error(w, "Le titre est requis", http.StatusBadRequest)
		return
	}

	log.Printf("📝 Creating new todo: %s", todo.Title)
	result, err := db.Exec("INSERT INTO todos (title, completed) VALUES (?, ?)",
		todo.Title, todo.Completed)
	if err != nil {
		log.Printf("❌ Error inserting todo: %v", err)
		http.Error(w, "Erreur lors de la création du todo", http.StatusInternalServerError)
		return
	}

	id, err := result.LastInsertId()
	if err != nil {
		log.Printf("❌ Error getting last insert ID: %v", err)
		http.Error(w, "Erreur lors de la récupération de l'ID", http.StatusInternalServerError)
		return
	}

	// Récupérer le todo créé
	err = db.QueryRow("SELECT id, title, completed, created_at, updated_at FROM todos WHERE id = ?", id).
		Scan(&todo.ID, &todo.Title, &todo.Completed, &todo.CreatedAt, &todo.UpdatedAt)
	if err != nil {
		log.Printf("❌ Error fetching created todo: %v", err)
		http.Error(w, "Erreur lors de la récupération du todo créé", http.StatusInternalServerError)
		return
	}

	log.Printf("✅ Successfully created todo ID %d: %s", todo.ID, todo.Title)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(todo)
}

func getTodo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "ID invalide", http.StatusBadRequest)
		return
	}

	var todo Todo
	err = db.QueryRow("SELECT id, title, completed, created_at, updated_at FROM todos WHERE id = ?", id).
		Scan(&todo.ID, &todo.Title, &todo.Completed, &todo.CreatedAt, &todo.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Todo non trouvé", http.StatusNotFound)
		} else {
			http.Error(w, "Erreur lors de la récupération du todo", http.StatusInternalServerError)
			log.Println("Erreur QueryRow:", err)
		}
		return
	}

	json.NewEncoder(w).Encode(todo)
}

func updateTodo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "ID invalide", http.StatusBadRequest)
		return
	}

	var todo Todo
	if err := json.NewDecoder(r.Body).Decode(&todo); err != nil {
		http.Error(w, "Données JSON invalides", http.StatusBadRequest)
		return
	}

	if todo.Title == "" {
		http.Error(w, "Le titre est requis", http.StatusBadRequest)
		return
	}

	_, err = db.Exec("UPDATE todos SET title = ?, completed = ? WHERE id = ?",
		todo.Title, todo.Completed, id)
	if err != nil {
		http.Error(w, "Erreur lors de la mise à jour du todo", http.StatusInternalServerError)
		log.Println("Erreur Update:", err)
		return
	}

	// Récupérer le todo mis à jour
	err = db.QueryRow("SELECT id, title, completed, created_at, updated_at FROM todos WHERE id = ?", id).
		Scan(&todo.ID, &todo.Title, &todo.Completed, &todo.CreatedAt, &todo.UpdatedAt)
	if err != nil {
		http.Error(w, "Erreur lors de la récupération du todo mis à jour", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(todo)
}

func deleteTodo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		log.Printf("❌ Invalid ID in deleteTodo: %s", vars["id"])
		http.Error(w, "ID invalide", http.StatusBadRequest)
		return
	}

	log.Printf("🗑️  Deleting todo ID %d", id)
	result, err := db.Exec("DELETE FROM todos WHERE id = ?", id)
	if err != nil {
		log.Printf("❌ Error deleting todo ID %d: %v", id, err)
		http.Error(w, "Erreur lors de la suppression du todo", http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("❌ Error checking rows affected for todo ID %d: %v", id, err)
		http.Error(w, "Erreur lors de la vérification de la suppression", http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		log.Printf("⚠️  Todo ID %d not found for deletion", id)
		http.Error(w, "Todo non trouvé", http.StatusNotFound)
		return
	}

	log.Printf("✅ Successfully deleted todo ID %d", id)
	w.WriteHeader(http.StatusNoContent)
}
