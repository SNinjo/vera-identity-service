package test

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
)

type IDTokenClaims struct {
	Subject string `json:"sub"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	Picture string `json:"picture"`
}
type OAuthAPI struct {
	URL               string
	AuthorizationCode string
	IDTokenClaims     *IDTokenClaims
}

func SetupOAuthAPI() *OAuthAPI {
	expectedCode := "mock-code"
	claims := &IDTokenClaims{
		Subject: "mock-subject",
		Name:    "Jo Liao",
		Email:   "user@example.com",
		Picture: "https://example.com/picture.png",
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(fmt.Sprintf(`{"error": "failed to parse form", "message": "%v"}`, err)))
			return
		}
		actualCode := r.FormValue("code")
		if actualCode != expectedCode {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(fmt.Sprintf(`{"error": "unexpected code value", "message": "got %s, want %s"}`, actualCode, expectedCode)))
			return
		}

		claimsBytes, _ := json.Marshal(claims)
		header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"none"}`))
		payload := base64.RawURLEncoding.EncodeToString(claimsBytes)
		mockIDToken := header + "." + payload + "."
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf(`{
			"id_token": "%s",
			"access_token": "mock_access_token",
			"token_type": "Bearer",
			"expires_in": 3600,
			"refresh_token": "mock_refresh_token"
		}`, mockIDToken)))
	})

	server := httptest.NewServer(mux)
	return &OAuthAPI{
		URL:               server.URL,
		AuthorizationCode: expectedCode,
		IDTokenClaims:     claims,
	}
}
