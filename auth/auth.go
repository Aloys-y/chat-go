package auth

import (
	"errors"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/jinzhu/gorm"
	"github.com/yourusername/chat-go/config"
	"github.com/yourusername/chat-go/db"
	"github.com/yourusername/chat-go/models"
	"golang.org/x/crypto/bcrypt"
)

type Claims struct {
	UserID uint   `json:"user_id"`
	Email  string `json:"email"`
	jwt.StandardClaims
}

// HashPassword generates a bcrypt hash of the password
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// CheckPassword compares a bcrypt hashed password with its plaintext version
func CheckPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

// GenerateToken generates a JWT token for the user
func GenerateToken(user *models.User) (string, error) {
	cfg := config.AppConfig.Auth
	expiry, err := time.ParseDuration(cfg.TokenExpiry)
	if err != nil {
		expiry = 24 * time.Hour // Default to 24 hours
	}

	claims := &Claims{
		UserID: user.ID,
		Email:  user.Email,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(expiry).Unix(),
			IssuedAt:  time.Now().Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(cfg.SecretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ValidateToken validates a JWT token and returns the claims
func ValidateToken(tokenString string) (*Claims, error) {
	cfg := config.AppConfig.Auth
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(cfg.SecretKey), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

// RegisterUser registers a new user
func RegisterUser(username, email, password, displayName string) (*models.User, error) {
	// Check if user already exists
	var existingUser models.User
	if err := db.DB.Where("username = ? OR email = ?", username, email).First(&existingUser).Error; err == nil {
		return nil, errors.New("user already exists with this username or email")
	}

	// Hash password
	hashedPassword, err := HashPassword(password)
	if err != nil {
		return nil, err
	}

	// Create new user
	user := &models.User{
		Username:     username,
		Email:        email,
		PasswordHash: hashedPassword,
		DisplayName:  displayName,
		LastLogin:    time.Now(),
		IsOnline:     true,
	}

	if err := db.DB.Create(user).Error; err != nil {
		return nil, err
	}

	return user, nil
}

// LoginUser authenticates a user and returns a token
func LoginUser(email, password string) (*models.User, string, error) {
	var user models.User
	if err := db.DB.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, "", errors.New("invalid email or password")
	}

	if err := CheckPassword(user.PasswordHash, password); err != nil {
		return nil, "", errors.New("invalid email or password")
	}

	// Update last login and online status
	user.LastLogin = time.Now()
	user.IsOnline = true
	if err := db.DB.Save(&user).Error; err != nil {
		return nil, "", err
	}

	// Generate token
	token, err := GenerateToken(&user)
	if err != nil {
		return nil, "", err
	}

	return &user, token, nil
}

// GetUserByID retrieves a user by ID
func GetUserByID(userID uint) (*models.User, error) {
	var user models.User
	if err := db.DB.First(&user, userID).Error; err != nil {
		return nil, err
	}
	return &user, nil
}