package handlers

import (
	"crypto/rand"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jmoiron/sqlx"
	"github.com/srvsurya/system-monitor/internal/notify"
	"golang.org/x/crypto/bcrypt"

	"encoding/hex"
	"regexp"

	"strings"

	"github.com/srvsurya/system-monitor/internal/models"
)

const tokenExpiry = 72 * time.Hour // token valid for 3 days

// register func

type RegisterRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func GenerateToken() (string, error) {
	byteString := make([]byte, 32)
	if _, err := rand.Read(byteString); err != nil {
		return "", err
	}
	return hex.EncodeToString(byteString), nil
}

func Register(db *sqlx.DB, mailer *notify.Mailer) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req RegisterRequest
		var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		// email valid?
		if !emailRegex.MatchString(req.Email) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Not a Valid Email"})
			return
		}
		// email already exists check
		var count int
		db.Get(&count, `SELECT COUNT(*) FROM users WHERE email = $1`, req.Email)
		if count > 0 {
			c.JSON(http.StatusConflict, gin.H{"error": "email already registered"})
			return
		}
		// name empty?
		if req.Name == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Name field empty"})
			return
		}
		// password data validation
		if req.Password == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Nothing entered in the Password field"})
			return
		}
		if len(req.Password) < 8 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Password needs atleast 8 characters"})
			return
		}
		specialChars := "!@#$%^&*()-_=+[]{}|;:'\",.<>?/`~"
		if !strings.ContainsAny(req.Password, specialChars) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Password should atleast contain one special character"})
			return
		}
		// Hash password
		hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
			return
		}

		// Insert user
		var userID int
		err = db.QueryRow(`
			INSERT INTO users (name, email, hashed_password)
			VALUES ($1, $2, $3)
			RETURNING id`,
			req.Name, req.Email, string(hashed),
		).Scan(&userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message": "account created",
			"user_id": userID,
		})
		// Email verification sent to user's email
		token, err := GenerateToken()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
			return
		}

		_, err = db.Exec(`
			INSERT INTO verification_tokens (user_id, token, expires_at)
			VALUES ($1, $2, NOW() + INTERVAL '24 hours')`,
			userID, token,
		)

		verifyURL := fmt.Sprintf("http://localhost:8080/auth/verify?token=%s", token)
		mailer.SendVerification(req.Email, verifyURL)
	}
}

// login func

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func Login(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		// email OR password empty?
		if req.Email == "" || req.Password == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Email or Password fields empty"})
			return
		}

		// Fetch user by email
		var user models.User
		err := db.Get(&user, `SELECT * FROM users WHERE email = $1`, req.Email)
		if err != nil {
			// Don't reveal whether email exists or password is wrong
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}

		// Compare password
		err = bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(req.Password))
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}
		if !user.Verified {
			c.JSON(http.StatusForbidden, gin.H{"error": "User email not verified"})
			return
		}

		// Generate JWT
		expiresAt := time.Now().Add(tokenExpiry)
		claims := jwt.MapClaims{
			"user_id": user.ID,
			"email":   user.Email,
			"exp":     expiresAt.Unix(),
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenStr, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
			return
		}

		// Store session in DB
		_, err = db.Exec(`
			INSERT INTO sessions (user_id, token, expires_at)
			VALUES ($1, $2, $3)`,
			user.ID, tokenStr, expiresAt,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create session"})
			return
		}

		// Update last_logged
		db.Exec(`UPDATE users SET last_logged = NOW() WHERE id = $1`, user.ID)

		c.JSON(http.StatusOK, gin.H{
			"token":      tokenStr,
			"expires_at": expiresAt,
		})
	}
}

// logout func

func Logout(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Token already validated by middleware — just pull it out
		token := c.GetString("token")

		_, err := db.Exec(`DELETE FROM sessions WHERE token = $1`, token)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to logout"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "logged out"})
	}
}

// email verification + update and delete from token tables
func VerifyEmail(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Query("token")
		if token == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "token required"})
			return
		}

		var userID int
		var expiresAt time.Time
		err := db.QueryRow(`
            SELECT user_id, expires_at FROM verification_tokens
            WHERE token = $1`, token,
		).Scan(&userID, &expiresAt)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid token"})
			return
		}
		if time.Now().After(expiresAt) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "token expired"})
			return
		}

		db.Exec(`UPDATE users SET verified = true WHERE id = $1`, userID)
		db.Exec(`DELETE FROM verification_tokens WHERE token = $1`, token)

		c.JSON(http.StatusOK, gin.H{"message": "email verified successfully"})
	}
}

// user data retrieval for session

func Me(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetInt("user_id") // set by middleware

		var user models.User
		err := db.Get(&user, `SELECT * FROM users WHERE id = $1`, userID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}

		c.JSON(http.StatusOK, user)
	}
}
