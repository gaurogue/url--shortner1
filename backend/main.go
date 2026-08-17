package main

import (
    "crypto/md5"
    "encoding/hex"
    "encoding/json"
    "errors"
    "fmt"
    "net/http"
    "os"
    "strings"
    "time"

    "github.com/golang-jwt/jwt/v5"
    "golang.org/x/crypto/bcrypt"
)

type URL struct {
	ID           string    `json:"id"`
	OriginalURL  string    `json:"original_url"`
	ShortURL     string    `json:"short_url"`
	CreationDate time.Time `json:"creation_date"`
}

type User struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

var urlDB = make(map[string]URL)
var userURLs = make(map[string][]URL)
var userDB = make(map[string]User)

var jwtKey = []byte("my_secret_key")

func enableCors(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
}

func generateShortURL(originalURL string) string {
	hasher := md5.New()
	hasher.Write([]byte(originalURL))

	hash := hex.EncodeToString(hasher.Sum(nil))

	return hash[:8]
}

func createURL(originalURL string) string {
	shortURL := generateShortURL(originalURL)

	urlDB[shortURL] = URL{
		ID:           shortURL,
		OriginalURL:  originalURL,
		ShortURL:     shortURL,
		CreationDate: time.Now(),
	}

	return shortURL
}

func getURL(id string) (URL, error) {
	url, ok := urlDB[id]

	if !ok {
		return URL{}, errors.New("URL not found")
	}

	return url, nil
}

// SHORTEN API

func ShortURLHandler(w http.ResponseWriter, r *http.Request) {

	enableCors(w)

	if r.Method == http.MethodOptions {
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	tokenStr := r.Header.Get("Authorization")

	if tokenStr == "" {
		http.Error(w, "Missing token", http.StatusUnauthorized)
		return
	}

	if !strings.HasPrefix(tokenStr, "Bearer ") {
		http.Error(w, "Invalid authorization header", http.StatusUnauthorized)
		return
	}

	tokenStr = strings.TrimPrefix(tokenStr, "Bearer ")

	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil || !token.Valid {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)

	if !ok {
		http.Error(w, "Invalid token claims", http.StatusUnauthorized)
		return
	}

	email, ok := claims["email"].(string)

	if !ok {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	var data struct {
		URL string `json:"url"`
	}

	err = json.NewDecoder(r.Body).Decode(&data)

	if err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if data.URL == "" {
		http.Error(w, "URL is required", http.StatusBadRequest)
		return
	}

	shortURL := createURL(data.URL)

	newURL := URL{
		ID:           shortURL,
		OriginalURL:  data.URL,
		ShortURL:     shortURL,
		CreationDate: time.Now(),
	}

	userURLs[email] = append(userURLs[email], newURL)

	// IMPORTANT:
	// Use the deployed domain automatically.
	scheme := "http"

if r.TLS != nil {
    scheme = "https"
}

if r.Header.Get("X-Forwarded-Proto") == "https" {
    scheme = "https"
}

response := map[string]string{
    "short_url": fmt.Sprintf(
        "%s://%s/redirect/%s",
        scheme,
        r.Host,
        shortURL,
    ),
}

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(response)
}

// REDIRECT

func redirectURLHandler(w http.ResponseWriter, r *http.Request) {

	id := strings.TrimPrefix(r.URL.Path, "/redirect/")

	url, err := getURL(id)

	if err != nil {
		http.Error(w, "URL not found", http.StatusNotFound)
		return
	}

	http.Redirect(w, r, url.OriginalURL, http.StatusFound)
}

// REGISTER

func RegisterHandler(w http.ResponseWriter, r *http.Request) {

	enableCors(w)

	if r.Method == http.MethodOptions {
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var user User

	err := json.NewDecoder(r.Body).Decode(&user)

	if err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if _, exists := userDB[user.Email]; exists {
		http.Error(w, "User already exists", http.StatusBadRequest)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword(
		[]byte(user.Password),
		bcrypt.DefaultCost,
	)

	if err != nil {
		http.Error(w, "Error hashing password", http.StatusInternalServerError)
		return
	}

	user.Password = string(hashedPassword)

	userDB[user.Email] = user

	w.Write([]byte("User registered successfully"))
}

// LOGIN

func LoginHandler(w http.ResponseWriter, r *http.Request) {

	enableCors(w)

	if r.Method == http.MethodOptions {
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var user User

	err := json.NewDecoder(r.Body).Decode(&user)

	if err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	storedUser, exists := userDB[user.Email]

	if !exists {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	err = bcrypt.CompareHashAndPassword(
		[]byte(storedUser.Password),
		[]byte(user.Password),
	)

	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		jwt.MapClaims{
			"email": user.Email,
			"exp":   time.Now().Add(time.Hour * 1).Unix(),
		},
	)

	tokenString, err := token.SignedString(jwtKey)

	if err != nil {
		http.Error(w, "Could not create token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(
		map[string]string{
			"token": tokenString,
		},
	)
}

func main() {

	// Serve frontend files
	frontendFS := http.FileServer(http.Dir("frontend"))

	http.Handle("/", frontendFS)

	// API routes
	http.HandleFunc("/shorten", ShortURLHandler)
	http.HandleFunc("/redirect/", redirectURLHandler)
	http.HandleFunc("/register", RegisterHandler)
	http.HandleFunc("/login", LoginHandler)

	port := os.Getenv("PORT")

	if port == "" {
		port = "3000"
	}

	fmt.Println("🚀 Server running on port", port)

	err := http.ListenAndServe("0.0.0.0:"+port, nil)

	if err != nil {
		fmt.Println("Error starting server:", err)
	}
}