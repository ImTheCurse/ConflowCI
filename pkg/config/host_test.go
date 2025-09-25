package config

import (
	"os/user"
	"reflect"
	"testing"
)

func TestIsValidHost(t *testing.T) {
	tests := []struct {
		name     string
		endpoint EndpointInfo
		wantErr  error
	}{
		{
			name:     "valid-host",
			endpoint: EndpointInfo{User: "user", Host: "host", Port: 22},
			wantErr:  nil,
		},
		{
			name:     "invalid-user",
			endpoint: EndpointInfo{User: "", Host: "host", Port: 22},
			wantErr:  ErrInvalidUser,
		},
		{
			name:     "invalid-host",
			endpoint: EndpointInfo{User: "user", Host: "", Port: 22},
			wantErr:  ErrInvalidHost,
		},
		{
			name:     "invalid-port",
			endpoint: EndpointInfo{User: "user", Host: "host", Port: 0},
			wantErr:  ErrInvalidPortNum,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEndpoint(tt.endpoint)
			if err != tt.wantErr {
				t.Errorf("Expected error: %v, got: %v", err, tt.wantErr)
			}
		})
	}
}

func TestParseHost(t *testing.T) {
	currentUser, err := user.Current()
	if err != nil {
		t.Fatalf("Failed to get current user: %v", err)
	}
	defaultUser := currentUser.Username

	// Invalid tests are tested in ValidateEndpoint
	tests := []struct {
		name     string
		endpoint string
		want     EndpointInfo
		wantErr  error
	}{
		{
			name:     "valid-endpoint",
			endpoint: "user@host:22",
			want:     EndpointInfo{User: "user", Host: "host", Port: 22},
			wantErr:  nil,
		},
		{
			name:     "valid-endpoint-without-user",
			endpoint: "host:2222",
			want:     EndpointInfo{User: defaultUser, Host: "host", Port: 2222},
			wantErr:  nil,
		},
		{
			name:     "valid-endpoint-without-port",
			endpoint: "user@host",
			want:     EndpointInfo{User: "user", Host: "host", Port: 22},
			wantErr:  nil,
		},
		{
			name:     "valid-endpoint-without-user-and-port",
			endpoint: "host",
			want:     EndpointInfo{User: defaultUser, Host: "host", Port: 22},
			wantErr:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseHost(tt.endpoint)
			if err != tt.wantErr {
				t.Errorf("Expected error: %v, got: %v", tt.wantErr, err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Expected endpoint: %v, got: %v", tt.want, got)
			}
		})
	}
}
