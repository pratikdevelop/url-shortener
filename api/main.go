package main

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"os"
	"time"

	"github.com/didip/tollbooth/v7"
	"github.com/didip/tollbooth/v7/limiter"
	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
	"github.com/rs/cors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

var (
	dbClient  *mongo.Client
	urlColl   *mongo.Collection
	userColl  *mongo.Collection
	jwtSecret = []byte("your-very-strong-random-secret-here-CHANGE-IN-PRODUCTION-2025") // MUST CHANGE!
)

type User struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name      string             `bson:"name" json:"name"`
	Email     string             `bson:"email" json:"email"`
	Password  string             `bson:"password" json:"-"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
}

type URL struct {
	ID          primitive.ObjectID  `bson:"_id,omitempty" json:"id,omitempty"`
	ShortCode   string              `bson:"short_code" json:"short_code"`
	OriginalURL string              `bson:"original_url" json:"original_url"`
	Title       string              `bson:"title,omitempty" json:"title,omitempty"`
	CreatedAt   time.Time           `bson:"created_at" json:"created_at"`
	ExpiresAt   *time.Time          `bson:"expires_at,omitempty" json:"expires_at,omitempty"`
	UserID      *primitive.ObjectID `bson:"user_id,omitempty" json:"user_id,omitempty"`
	ClickCount  int64               `bson:"click_count" json:"click_count"`
	Custom      bool                `bson:"custom,omitempty" json:"custom,omitempty"`
}

type RegisterRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type ShortenRequest struct {
	OriginalURL string  `json:"original_url"`
	CustomAlias string  `json:"custom_alias,omitempty"`
	Title       string  `json:"title,omitempty"`
	ExpiresAt   *string `json:"expires_at,omitempty"`
}

type UpdateRequest struct {
	OriginalURL *string `json:"original_url,omitempty"`
	Title       *string `json:"title,omitempty"`
	ExpiresAt   *string `json:"expires_at,omitempty"` // ISO or empty string to remove
}

type Claims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

const (
	chars         = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	charLen       = len(chars)
	defaultLength = 7
)

var (
	globalLimiter   = tollbooth.NewLimiter(20, nil)
	strictLimiter   = tollbooth.NewLimiter(0.1, nil) // ~6 per minute
	redirectLimiter = tollbooth.NewLimiter(50, nil)
	addUrlLimiter   = tollbooth.NewLimiter(2, nil)
)

func GenerateShortCode(length int) (string, error) {
	if length == 0 {
		length = defaultLength
	}
	b := make([]byte, length)
	for i := range b {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(charLen)))
		if err != nil {
			return "", err
		}
		b[i] = chars[n.Int64()]
	}
	return string(b), nil
}

func initDB() {
	if dbClient != nil {
		return
	}

	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		log.Fatal("You must set MONGODB_URI environment variable")
	}

	clientOptions := options.Client().ApplyURI(uri)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		log.Fatal("Failed to ping MongoDB:", err)
	}

	dbClient = client
	db := dbClient.Database("url-shortner")
	urlColl = db.Collection("lists")
	userColl = db.Collection("users")

	userColl.Indexes().CreateOne(context.TODO(), mongo.IndexModel{
		Keys:    bson.D{{"email", 1}},
		Options: options.Index().SetUnique(true),
	})

	urlColl.Indexes().CreateOne(context.TODO(), mongo.IndexModel{
		Keys:    bson.D{{"short_code", 1}},
		Options: options.Index().SetUnique(true),
	})

	fmt.Println("Connected to MongoDB Atlas!")
}

func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
			http.Error(w, "Unauthorized: Missing token", http.StatusUnauthorized)
			return
		}

		tokenStr := authHeader[7:]
		claims := &Claims{}

		token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Unauthorized: Invalid token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), "user_id", claims.UserID)
		next(w, r.WithContext(ctx))
	}
}

func generateToken(userID primitive.ObjectID) (string, error) {
	expiration := time.Now().Add(30 * 24 * time.Hour)
	claims := &Claims{
		UserID: userID.Hex(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiration),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func rateLimit(lim *limiter.Limiter, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		httpErr := tollbooth.LimitByRequest(lim, w, r)
		if httpErr != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(httpErr.StatusCode)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Too Many Requests",
			})
			return
		}
		next(w, r)
	}
}

// Handlers
func register(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	if len(req.Password) < 6 {
		http.Error(w, "Password must be at least 6 characters", http.StatusBadRequest)
		return
	}

	hashed, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	user := User{Name: req.Name, Email: req.Email, Password: string(hashed), CreatedAt: time.Now()}

	initDB()
	_, err := userColl.InsertOne(context.TODO(), user)
	if err != nil {
		http.Error(w, "Email already exists", http.StatusConflict)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"message": "Registered successfully"})
}

func login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	initDB()
	var user User
	err := userColl.FindOne(context.TODO(), bson.M{"email": req.Email}).Decode(&user)
	if err != nil || bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)) != nil {
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	token, _ := generateToken(user.ID)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Login successful",
		"token":   token,
		"user":    map[string]string{"id": user.ID.Hex(), "email": user.Email, "name": user.Name},
	})
}

func addShortUrls(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	userIDHex := r.Context().Value("user_id").(string)
	userOID, _ := primitive.ObjectIDFromHex(userIDHex)

	var req ShortenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.OriginalURL == "" {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	initDB()
	shortCode := req.CustomAlias
	if shortCode == "" {
		for i := 0; i < 10; i++ {
			shortCode, _ = GenerateShortCode(defaultLength)
			if count, _ := urlColl.CountDocuments(context.TODO(), bson.M{"short_code": shortCode}); count == 0 {
				break
			}
		}
	} else if count, _ := urlColl.CountDocuments(context.TODO(), bson.M{"short_code": shortCode}); count > 0 {
		http.Error(w, "Custom alias already taken", http.StatusConflict)
		return
	}

	var expiresAt *time.Time
	if req.ExpiresAt != nil {
		if t, err := time.Parse(time.RFC3339, *req.ExpiresAt); err == nil {
			expiresAt = &t
		}
	}

	url := URL{
		ShortCode:   shortCode,
		OriginalURL: req.OriginalURL,
		Title:       req.Title,
		CreatedAt:   time.Now(),
		ExpiresAt:   expiresAt,
		UserID:      &userOID,
		ClickCount:  0,
		Custom:      shortCode == req.CustomAlias,
	}

	_, err := urlColl.InsertOne(context.TODO(), url)
	if err != nil {
		http.Error(w, "Failed to save URL", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"short_code":   shortCode,
		"short_url":    "http://localhost:8080/" + shortCode,
		"original_url": req.OriginalURL,
		"title":        req.Title,
		"created_at":   time.Now(),
	}
	if expiresAt != nil {
		response["expires_at"] = expiresAt
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func getAllUrls(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	initDB()
	userIDHex := r.Context().Value("user_id").(string)
	userOID, _ := primitive.ObjectIDFromHex(userIDHex)

	cursor, _ := urlColl.Find(context.TODO(), bson.M{"user_id": userOID})
	defer cursor.Close(context.TODO())

	var results []URL
	cursor.All(context.TODO(), &results)
	json.NewEncoder(w).Encode(results)
}

func redirect(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Path[1:]
	if code == "" || code == "api" || len(code) > 30 {
		http.NotFound(w, r)
		return
	}

	initDB()
	var url URL
	if err := urlColl.FindOne(context.TODO(), bson.M{"short_code": code}).Decode(&url); err != nil {
		http.NotFound(w, r)
		return
	}

	if url.ExpiresAt != nil && time.Now().After(*url.ExpiresAt) {
		http.Error(w, "This link has expired", http.StatusGone)
		return
	}

	// Fire-and-forget click increment with logging
	go func(id primitive.ObjectID, code string) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, err := urlColl.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$inc": bson.M{"click_count": 1}})
		if err != nil {
			log.Printf("Click increment failed for %s: %v", code, err)
		}
	}(url.ID, code)

	http.Redirect(w, r, url.OriginalURL, http.StatusMovedPermanently)
}

func statsPage(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Path[len("/s/"):]
	if code == "" || len(code) > 30 {
		http.NotFound(w, r)
		return
	}

	initDB()
	var url URL
	if err := urlColl.FindOne(context.TODO(), bson.M{"short_code": code}).Decode(&url); err != nil {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, `
<!DOCTYPE html>
<html lang="en" class="h-full bg-gray-50">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Stats for %s</title>
    <script src="https://cdn.tailwindcss.com"></script>
</head>
<body class="h-full flex items-center justify-center p-4">
    <div class="max-w-md w-full bg-white rounded-xl shadow-lg p-8 text-center">
        <h1 class="text-3xl font-bold text-gray-800 mb-6">Link Statistics</h1>
        <div class="space-y-6">
            <div><p class="text-gray-600 text-sm">Short URL</p>
                <code class="bg-gray-100 px-4 py-2 rounded-lg text-lg font-mono">http://localhost:8080/%s</code></div>
            <div><p class="text-gray-600 text-sm">Original URL</p>
                <a href="%s" target="_blank" class="text-blue-600 hover:underline break-all">%s</a></div>
            <div class="text-5xl font-bold text-blue-600 my-8">%d
                <p class="text-lg text-gray-600 mt-2">Total Clicks</p></div>
            %s
        </div>
    </div>
</body>
</html>`,
		code, code, url.OriginalURL, url.OriginalURL, url.ClickCount,
		func() string {
			if url.ExpiresAt != nil {
				return fmt.Sprintf(`<div><p class="text-gray-600 text-sm">Expires on</p><p class="text-lg font-semibold text-red-600">%s</p></div>`,
					url.ExpiresAt.Format("January 2, 2006"))
			}
			return `<div class="text-green-600 font-semibold text-xl">Never expires</div>`
		}(),
	)
}

func deleteUrl(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	userIDHex := r.Context().Value("user_id").(string)
	userOID, _ := primitive.ObjectIDFromHex(userIDHex)

	code := r.Context().Value("short_code").(string)
	initDB()

	result, err := urlColl.DeleteOne(context.TODO(), bson.M{"short_code": code, "user_id": userOID})
	if err != nil || result.DeletedCount == 0 {
		http.Error(w, "URL not found or not owned by you", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"message": "URL deleted successfully"})
}

func updateUrl(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	userIDHex := r.Context().Value("user_id").(string)
	userOID, _ := primitive.ObjectIDFromHex(userIDHex)

	code := r.Context().Value("short_code").(string)

	var req UpdateRequest
	fmt.Print(req.ExpiresAt)
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	updateFields := bson.M{}
	if req.Title != nil {
		updateFields["title"] = *req.Title
	}
	if req.OriginalURL != nil {
		if *req.OriginalURL == "" {
			http.Error(w, "original_url cannot be empty", http.StatusBadRequest)
			return
		}
		updateFields["original_url"] = *req.OriginalURL
	}
	if req.ExpiresAt != nil {
		if *req.ExpiresAt == "" {
			updateFields["expires_at"] = nil
		} else {
			t, err := time.Parse(time.RFC3339, *req.ExpiresAt)
			if err != nil {
				http.Error(w, "Invalid expires_at format", http.StatusBadRequest)
				return
			}
			updateFields["expires_at"] = t
		}
	}

	if len(updateFields) == 0 {
		http.Error(w, "No fields to update", http.StatusBadRequest)
		return
	}

	initDB()
	result, err := urlColl.UpdateOne(context.TODO(), bson.M{"short_code": code, "user_id": userOID}, bson.M{"$set": updateFields})
	if err != nil || result.MatchedCount == 0 {
		http.Error(w, "URL not found or not owned by you", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"message": "URL updated successfully"})
}

func apiStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	userIDHex := r.Context().Value("user_id").(string)
	userOID, _ := primitive.ObjectIDFromHex(userIDHex)

	code := r.Context().Value("short_code").(string)

	initDB()
	var url URL
	err := urlColl.FindOne(context.TODO(), bson.M{"short_code": code, "user_id": userOID}).Decode(&url)
	if err != nil {
		http.Error(w, "Link not found or access denied", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(map[string]any{
		"short_code":   url.ShortCode,
		"original_url": url.OriginalURL,
		"title":        url.Title,
		"click_count":  url.ClickCount,
		"created_at":   url.CreatedAt,
		"expires_at":   url.ExpiresAt,
		"short_url":    "http://localhost:8080/" + url.ShortCode,
	})
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		redirect(w, r)
		return
	}
	fmt.Fprintln(w, "URL Shortener API Running! 🚀")
}

func init() {
	_ = godotenv.Load()
}

func main() {
	initDB()

	mux := http.NewServeMux()

	mux.HandleFunc("/", rootHandler)
	mux.HandleFunc("/s/", statsPage)

	mux.HandleFunc("/api/register", rateLimit(strictLimiter, register))
	mux.HandleFunc("/api/login", rateLimit(strictLimiter, login))
	mux.HandleFunc("/api/urls", rateLimit(globalLimiter, authMiddleware(getAllUrls)))
	mux.HandleFunc("/api/add-url", rateLimit(addUrlLimiter, authMiddleware(addShortUrls)))

	// DELETE & PUT /api/url/{code}
	mux.HandleFunc("/api/url/", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Path[len("/api/url/"):]
		if code == "" || len(code) > 30 {
			http.Error(w, "Invalid short code", http.StatusBadRequest)
			return
		}
		ctx := context.WithValue(r.Context(), "short_code", code)

		switch r.Method {
		case http.MethodDelete:
			rateLimit(globalLimiter, authMiddleware(deleteUrl))(w, r.WithContext(ctx))
		case http.MethodPut:
			rateLimit(globalLimiter, authMiddleware(updateUrl))(w, r.WithContext(ctx))
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// GET /api/stats/{code}
	mux.HandleFunc("/api/stats/", rateLimit(globalLimiter, authMiddleware(func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Path[len("/api/stats/"):]
		if code == "" || len(code) > 30 {
			http.Error(w, "Invalid short code", http.StatusBadRequest)
			return
		}
		ctx := context.WithValue(r.Context(), "short_code", code)
		apiStats(w, r.WithContext(ctx))
	})))

	handler := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:4200"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowCredentials: true,
	}).Handler(mux)

	server := &http.Server{
		Addr:         ":8080",
		Handler:      tollbooth.LimitHandler(globalLimiter, handler),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	fmt.Println("Server running on http://localhost:8080")
	log.Fatal(server.ListenAndServe())
}
