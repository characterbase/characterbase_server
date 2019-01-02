package integration

import (
	"bytes"
	"cbs/api"
	"cbs/dtos"
	"cbs/models"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func findCookie(name string, cookies []*http.Cookie) *http.Cookie {
	for _, c := range cookies {
		if c.Name == name {
			return c
		}
	}
	return nil
}

func TestAuthRouter_LogIn(t *testing.T) {
	tests := []struct {
		name     string
		payload  dtos.ReqLogIn
		want     interface{}
		wantstat int
	}{
		{
			name:     "valid",
			payload:  dtos.ReqLogIn{Email: userA.Email, Password: userAPassword},
			want:     &dtos.ResGetUser{User: userA},
			wantstat: http.StatusOK,
		},
		{
			name:     "invalid",
			payload:  dtos.ReqLogIn{Email: userA.Email, Password: "wrongpassword"},
			want:     api.ErrCodeBadAuth,
			wantstat: http.StatusUnauthorized,
		},
		{
			name:     "no email",
			payload:  dtos.ReqLogIn{Email: "", Password: "elliptical"},
			want:     api.ErrCodeBadBody,
			wantstat: http.StatusBadRequest,
		},
		{
			name:     "no password",
			payload:  dtos.ReqLogIn{Email: userA.Email, Password: ""},
			want:     api.ErrCodeBadBody,
			wantstat: http.StatusBadRequest,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serialized, err := json.Marshal(tt.payload)
			if err != nil {
				t.Fatal("failed to marshal payload")
			}
			_, rr, err := apiRequest("POST", "/login", bytes.NewReader(serialized))
			if err != nil {
				t.Fatal(err)
			}
			testAPIResponse(t, rr, tt.want, tt.wantstat, false)
			if fmt.Sprintf("%T", tt.want) == "*dtos.ResGetUser" {
				sesscookie := findCookie("user_session", rr.Result().Cookies())
				if sesscookie == nil {
					t.Fatalf("response did not get session cookie")
				}
				sess, err := server.Providers.Redis.Get(fmt.Sprintf("session:%v", sesscookie.Value)).Result()
				if err != nil {
					t.Fatalf("failed to get session from redis store")
				}
				wantsess, err := json.Marshal(tt.want)
				if err != nil {
					t.Fatalf("failed to marshal wanted session")
				}
				if sess != string(wantsess) {
					t.Errorf("got session %v; want %v", sess, string(wantsess))
				}
			}
		})
	}
}

func TestAuthRouter_Me(t *testing.T) {
	tests := []struct {
		name     string
		user     *models.User
		sessid   string
		want     interface{}
		wantstat int
	}{
		{
			name:     "authenticated",
			user:     userA,
			want:     &dtos.ResGetUser{User: userA},
			wantstat: http.StatusOK,
		},
		{
			name:     "unauthenticated",
			user:     nil,
			want:     api.ErrCodeBadAuth,
			wantstat: http.StatusUnauthorized,
		},
		{
			name:     "invalid session",
			user:     nil,
			sessid:   "somerandomid",
			want:     api.ErrCodeBadAuth,
			wantstat: http.StatusUnauthorized,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := http.NewRequest("GET", "/me", nil)
			if err != nil {
				t.Fatalf("failed to create request")
			}
			if tt.user != nil {
				if _, err := loginUser(r, tt.user); err != nil {
					t.Fatalf("failed to login user")
				}
			}
			if tt.sessid != "" {
				r.AddCookie(&http.Cookie{Name: "user_session", Value: tt.sessid})
			}
			rr := httptest.NewRecorder()
			server.ServeHTTP(rr, r)
			testAPIResponse(t, rr, tt.want, tt.wantstat, false)
		})
	}
}

func TestAuthRouter_LogOut(t *testing.T) {
	tests := []struct {
		name     string
		user     *models.User
		want     interface{}
		wantstat int
	}{
		{
			name:     "logged in",
			user:     userA,
			want:     "",
			wantstat: http.StatusNoContent,
		},
		{
			name:     "not logged in",
			user:     nil,
			want:     api.ErrCodeBadAuth,
			wantstat: http.StatusUnauthorized,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := http.NewRequest("GET", "/logout", nil)
			if err != nil {
				t.Fatalf("failed to create request")
			}
			if tt.user != nil {
				if _, err := loginUser(r, tt.user); err != nil {
					t.Fatalf("failed to login user")
				}
			}
			rr := httptest.NewRecorder()
			server.ServeHTTP(rr, r)
			testAPIResponse(t, rr, tt.want, tt.wantstat, false)
		})
	}
}
