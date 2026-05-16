package api

import (
	"encoding/json"
	"net/http"

	"kingultron99.com/RoleCall/services"
)

type BroadcastRequest struct {
	Message []services.BroadCastMessage `json:"message"`
}

func init() {
	mapRoutes["/broadcast"] = Route{
		Authenticated: true,
		Scope:         "broadcast:write",
		Handler: func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}

			var req BroadcastRequest

			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, "invalid body", http.StatusBadRequest)
				return
			}

			results := services.BroadcastMessage(r.Context(), req.Message)

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(results)
		},
	}
}
