package api

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/lib/pq"
	"kingultron99.com/RoleCall/core"
	"kingultron99.com/RoleCall/middleware"
	"kingultron99.com/RoleCall/utils"
)

func init() {
	mapRoutes["/auth"] = Route{
		Authenticated: false,
		Handler: func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}

			key, err := utils.LoadPrivateKey()
			if err != nil {
				http.Error(w, "server error", http.StatusInternalServerError)
				return
			}

			type credentials struct {
				DiscordID string `json:"id"`
				Password  string `json:"pw"`
			}
			var creds credentials
			err = json.NewDecoder(r.Body).Decode(&creds)
			if err != nil {
				log.Printf("failed to decode auth request body: %v", err)
				http.Error(w, "invalid request body", http.StatusBadRequest)
				return
			}

			var (
				hash       string
				userScopes []string
			)
			err = core.DB.QueryRow("SELECT password, scopes FROM auth WHERE discord_id=$1", creds.DiscordID).Scan(&hash, (*pq.StringArray)(&userScopes))
			if err != nil {
				log.Printf("failed to query auth table: %v", err)
				http.Error(w, "invalid credentials", http.StatusUnauthorized)
				return
			}
			if utils.VerifyPassword(hash, creds.Password) {
				log.Printf("failed to authenticate user: %v", err)
				http.Error(w, "invalid credentials", http.StatusUnauthorized)
				return
			}

			var now = time.Now()
			claims := middleware.Claims{
				Scopes: userScopes,
				RegisteredClaims: jwt.RegisteredClaims{
					Issuer:    "RoleCall-API",
					Audience:  []string{"RoleCall-Clients"},
					Subject:   creds.DiscordID,
					IssuedAt:  jwt.NewNumericDate(now),
					ExpiresAt: jwt.NewNumericDate(now.Add(15 * time.Minute)),
					NotBefore: jwt.NewNumericDate(now),
				},
			}

			token := jwt.NewWithClaims(jwt.SigningMethodES512, claims)
			signed, err := token.SignedString(key)
			if err != nil {
				log.Printf("failed to sign token: %v", err)
				http.Error(w, "Failed to sign token", http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{
				"token": signed,
			})
		},
	}
}
