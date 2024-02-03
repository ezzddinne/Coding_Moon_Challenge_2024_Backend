package middleware

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

type Session struct {
	UserID   uint
	RoleName string
	SquadID  uint
}

// Generate token
func GenerateToken(id, squad uint, role string) string {

	duration, _ := strconv.Atoi(os.Getenv("TOKEN_DURATION"))

	claims := jwt.MapClaims{
		"exp":       time.Now().Add(time.Hour * time.Duration(duration)).Unix(),
		"iat":       time.Now().Unix(),
		"user_id":   id,
		"role_name": role,
		"squad_id":  squad,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	auth, _ := token.SignedString([]byte(os.Getenv("TOKEN_SECRET")))

	return auth

}

// Extract the token value
func extractToken(ctx *gin.Context) string {

	bearerToken := strings.Fields(ctx.Request.Header["Authorization"][0])[1]

	if len(bearerToken) == 0 {
		return ""
	} else {
		return bearerToken
	}
}

// extract values from token
func ExtractTokenValues(ctx *gin.Context) Session {

	//init vars
	session := Session{}

	tokenString := extractToken(ctx)
	token, _ := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("TOKEN_SECRET")), nil
	})

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {

		session.UserID = uint(claims["user_id"].(float64))
		session.SquadID = uint(claims["squad_id"].(float64))
		session.RoleName, _ = claims["role_name"].(string)
		return session
	}
	return Session{}
}

// validate the given token
func validateToken(token string) (*jwt.Token, error) {
	//2nd arg function return secret key after checking if the signing method is HMAC and returned key is used by 'Parse' to decode the token)
	return jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			//nil secret key
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("TOKEN_SECRET")), nil
	})
}

// AuthorizeJWT -> to authorize JWT Token
func AuthorizeJWT() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		const BearerSchema string = "Bearer "
		authHeader := ctx.GetHeader("Authorization")
		if authHeader == "" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"Error": "No Authorization header found"})
		}
		tokenString := authHeader[len(BearerSchema):]
		if token, err := validateToken(tokenString); err != nil {
			fmt.Println("token", tokenString, err.Error())
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"Error": "Not Valid Token"})
		} else {
			if claims, ok := token.Claims.(jwt.MapClaims); !ok {
				ctx.AbortWithStatus(http.StatusUnauthorized)
			} else {
				if token.Valid {
					ctx.Set("user_id", claims["user_id"])
					ctx.Set("squad_id", claims["squad_id"])
					ctx.Set("role_name", claims["role_name"])
				} else {
					ctx.AbortWithStatus(http.StatusUnauthorized)
				}
			}
		}
	}
}

func DeleteSession(db *gorm.DB, squad_id uint) error {
	return db.Where("squad_id = ?").Delete(&Session{}).Error
}
