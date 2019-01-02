package integration

import (
	"cbs/api"
	"cbs/models"
	"testing"
)

func TestMwUserSession(t *testing.T) {
	tests := []struct {
		name   string
		sessid string
		user   *models.User
		want   interface{}
	}{
		{
			name: "valid",
			user: userA,
			want: nil,
		},
		{
			name:   "invalid",
			sessid: "somesession",
			user:   nil,
			want:   api.ErrCodeBadAuth,
		},
		{
			name: "no session",
			user: nil,
			want: api.ErrCodeBadAuth,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

		})
	}
}
