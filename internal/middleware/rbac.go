package middleware

import (
	"net/http"

	"gorm.io/gorm"
	"briers-api/internal/models"
)

func RequireRole(db *gorm.DB, roles ...string) func(http.Handler) http.Handler {
	allowed := make(map[string]bool, len(roles))
	for _, r := range roles {
		allowed[r] = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := GetClaims(r)
			if !ok {
				writeError(w, http.StatusUnauthorized, "unauthorized")
				return
			}

			var profile models.Profile
			if err := db.Where("id = ?", claims.Sub).First(&profile).Error; err != nil {
				writeError(w, http.StatusUnauthorized, "profile not found")
				return
			}

			if !allowed[profile.Role] {
				writeError(w, http.StatusForbidden, "forbidden: insufficient role")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
