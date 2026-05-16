package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"kingultron99.com/RoleCall/utils"
)

type Claims struct {
	Scopes []string `json:"scopes"`
	jwt.RegisteredClaims
}

var claimsKey = "RoleCall_Claims"

func (c *Claims) HasScope(scope string) bool {
	for _, s := range c.Scopes {
		if s == scope {
			return true
		}
	}
	return false
}

func ValidateToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwt.SigningMethodES512 {
			return nil, jwt.ErrSignatureInvalid
		}
		return utils.LoadPublicKey()
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

func JWTAuth(scope string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")

		if header == "" {
			http.Error(w, "missing auth header", http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(header, "Bearer ")

		claims, err := ValidateToken(tokenString)
		if err != nil {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		if !claims.ExpiresAt.Time.After(time.Now()) {
			http.Error(w, "Token has expired", http.StatusForbidden)
			return
		}

		if !claims.HasScope(scope) {
			http.Error(w, "insufficient scope", http.StatusForbidden)
			return
		}

		ctx := context.WithValue(
			r.Context(),
			claimsKey,
			claims,
		)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetClaims(ctx context.Context) *Claims {
	claims, ok := ctx.Value(claimsKey).(*Claims)
	if !ok {
		return nil
	}

	return claims
}
