package handlers

import (
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"

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

func Register(db *sqlx.DB) gin.HandlerFunc {
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
