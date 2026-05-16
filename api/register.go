package api

import (
	"encoding/json"
	"net/http"

	"golang.org/x/crypto/bcrypt"
	"kingultron99.com/RoleCall/core"
)

func init() {
	mapRoutes["/register"] = Route{
		Authenticated: false,
		Handler: func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			type registrationRequest struct {
				DiscordID string `json:"id"`
				Password  string `json:"pw"`
			}
			var req registrationRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, "invalid body", http.StatusBadRequest)
				return
			}

			if req.DiscordID == "" || req.Password == "" {
				http.Error(w, "missing fields", http.StatusBadRequest)
				return
			}

			hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
			if err != nil {
				http.Error(w, "server error", http.StatusInternalServerError)
				return
			}

			rows, err := core.DB.Query("SELECT discord_id FROM auth WHERE discord_id=$1", req.DiscordID)
			if err != nil && err.Error() != "sql: no rows in result set" {
				http.Error(w, "server error", http.StatusInternalServerError)
				return
			}
			defer rows.Close()

			if rows.Next() {
				http.Error(w, "user already exists", http.StatusConflict)
				return
			}

			_, err = core.DB.Exec("INSERT INTO auth (discord_id, password) VALUES ($1, $2)", req.DiscordID, string(hashedPassword))
			if err != nil {
				http.Error(w, "server error", http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusOK)
		},
	}
}
