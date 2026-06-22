package session

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

const CookieName = "session_id"

// GenerateID returns a cryptographically random 64-character hex string.
func GenerateID() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate session id: %w", err)
	}
	return hex.EncodeToString(b), nil
}

// SetCookie writes the session cookie to the response.
func SetCookie(c *gin.Context, sessionID string) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     CookieName,
		Value:    sessionID,
		Path:     "/",
		MaxAge:   604800, // 7 days
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   false, // set to true in production (HTTPS)
	})
}

// GetFromCookie reads the session ID from the request cookie.
// Returns an empty string if the cookie is absent.
func GetFromCookie(c *gin.Context) string {
	cookie, err := c.Request.Cookie(CookieName)
	if err != nil {
		return ""
	}
	return cookie.Value
}

// ClearCookie instructs the browser to delete the session cookie.
func ClearCookie(c *gin.Context) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     CookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}
