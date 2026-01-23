package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/Nerzal/gocloak/v13"
	"github.com/anoop-dryad/bridgehead/internal/service"
)

type AuthMiddleware struct {
	client       *gocloak.GoCloak
	cache        *service.CacheService
	clientID     string
	clientSecret string
	realm        string
}

func NewAuthMiddleware(client *gocloak.GoCloak, clientID, clientSecret, realm string, cache *service.CacheService) *AuthMiddleware {
	return &AuthMiddleware{
		client:       client,
		cache:        cache,
		clientID:     clientID,
		clientSecret: clientSecret,
		realm:        realm,
	}
}

func (m *AuthMiddleware) ValidateToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 1. Extract Bearer Token
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "Authorization header missing", http.StatusUnauthorized)
			return
		}
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// 2. Validate with Keycloak (Introspection)
		rptResult, err := m.client.RetrospectToken(context.Background(), tokenString, m.clientID, m.clientSecret, m.realm)
		if err != nil || !*rptResult.Active {
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		// 2. Fetch UserInfo to get the 'sub' (User ID)
		userInfo, err := m.client.GetUserInfo(r.Context(), tokenString, m.realm)
		if err != nil {
			http.Error(w, "Unauthorized: Could not fetch user details", http.StatusUnauthorized)
			return
		}

		// 3. Extract the ID and Email for your "Bridge" logic
		keycloakID := *userInfo.Sub
		// Check the dedicated User bucket in the CacheService
		if user, ok := m.cache.UserCache.Get(keycloakID); ok {
			// Fast path: User found in memory
			ctx := context.WithValue(r.Context(), "user", user)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		// Slow path: Cache miss, proceed to DB lookup...

		/* email := ""
		if userInfo.Email != nil {
			email = *userInfo.Email
		} */

		// 4. SYNC LOGIC: Check your Go Database
		// user, err := yourDB.SyncUser(keycloakID, email)

		// 4. Pass the Subject ID into the context
		ctx := context.WithValue(r.Context(), "user_id", keycloakID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})

}
