package api

import (
	"cbs/models"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

var (
	testUserSessionID = "testusersession"
	testUser          = models.User{
		DisplayName: "testuser",
		Email:       "testuser@test.com",
	}
)

type AuthMock Service

func (m *AuthMock) User(r *http.Request) (*models.User, error) {
	c, err := r.Cookie("user_session")
	if err != nil {
		return nil, err
	}
	if c.Value == testUserSessionID {
		return &testUser, nil
	}
	return nil, nil
}

// Unused stub methods to satisfy interface implementation
func (m *AuthMock) Authenticate(string, string) (*models.User, error) {
	return nil, nil
}
func (m *AuthMock) Login(*models.User, http.ResponseWriter) error {
	return nil
}
func (m *AuthMock) Logout(http.ResponseWriter) error {
	return nil
}

func TestMwUserSession(t *testing.T) {
	services := &Services{Auth: &AuthMock{Config: nil, Providers: nil}}
	middleware := MwUserSession(services)
	tests := []struct {
		name   string
		sessid string
		want   *models.User
	}{
		{
			name:   "valid session",
			sessid: testUserSessionID,
			want:   &testUser,
		},
		{
			name:   "invalid session",
			sessid: "invalidsessionid",
			want:   nil,
		},
		{
			name:   "no session",
			sessid: "",
			want:   nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := http.NewRequest("GET", "/", nil)
			if err != nil {
				t.Fatalf("failed to create request")
			}
			rr := httptest.NewRecorder()
			r.AddCookie(&http.Cookie{Name: "user_session", Value: tt.sessid})
			handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				result, ok := r.Context().Value(UserContextKey).(*models.User)
				fmt.Println(result)
				if tt.want == nil && result != nil {
					t.Fatalf("got user %v; want nil", result)
				}
				if !ok {
					t.Fatalf("failed context type conversion")
				}
				if !reflect.DeepEqual(result, tt.want) {
					t.Fatalf("got user %v; want %v", result, tt.want)
				}
			}))
			handler.ServeHTTP(rr, r)
		})
	}
}
