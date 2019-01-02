package integration

import (
	"bytes"
	"cbs/api"
	"cbs/dtos"
	"cbs/models"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
)

func TestUserRouter_CreateUser(t *testing.T) {
	tests := []struct {
		name     string
		payload  dtos.ReqCreateUser
		want     interface{}
		wantstat int
	}{
		{
			name: "create user",
			payload: dtos.ReqCreateUser{
				DisplayName: "testuser1",
				Email:       "testuser1@test.com",
				Password:    "testuser1",
			},
			want: dtos.ResGetUser{
				User: &models.User{DisplayName: "testuser1", Email: "testuser1@test.com"},
			},
			wantstat: http.StatusCreated,
		},
		{
			name: "email exists",
			payload: dtos.ReqCreateUser{
				DisplayName: "testuser2",
				Email:       "testuser1@test.com",
				Password:    "testuser2",
			},
			want:     api.ErrCodeInternal,
			wantstat: http.StatusInternalServerError,
		},
		{
			name:     "no name",
			payload:  dtos.ReqCreateUser{DisplayName: "", Email: "noname@test.com", Password: "noname"},
			want:     api.ErrCodeBadBody,
			wantstat: http.StatusBadRequest,
		},
		{
			name:     "no email",
			payload:  dtos.ReqCreateUser{DisplayName: "noemail", Email: "", Password: "noemail"},
			want:     api.ErrCodeBadBody,
			wantstat: http.StatusBadRequest,
		},
		{
			name:     "no password",
			payload:  dtos.ReqCreateUser{DisplayName: "nopassword", Email: "nopass@test.com", Password: ""},
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
			_, rr, err := apiRequest("POST", "/users", bytes.NewReader(serialized))
			if err != nil {
				t.Fatal(err)
			}
			testAPIResponse(t, rr, tt.want, tt.wantstat, true)
		})
	}
}

func TestUserRouter_GetUser(t *testing.T) {
	tests := []struct {
		name     string
		id       string
		want     interface{}
		wantstat int
	}{
		{
			name: "found",
			id:   userA.ID,
			want: &dtos.ResGetUser{
				User: userA,
			},
			wantstat: http.StatusOK,
		},
		{
			name:     "not found",
			id:       "somerandomid",
			want:     api.ErrCodeNotFound,
			wantstat: http.StatusNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := fmt.Sprintf("/users/%v", tt.id)
			_, rr, err := apiRequest("GET", url, nil)
			if err != nil {
				t.Fatal("failed to create request")
			}
			testAPIResponse(t, rr, tt.want, tt.wantstat, false)
		})
	}
}
