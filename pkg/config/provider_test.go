package config

import (
	"testing"
)

func TestProviderValidation(t *testing.T) {
	tests := []struct {
		name        string
		config      Config
		expectError bool
		err         error
	}{
		{
			name: "valid-test-without-auth-token",
			config: Config{
				Provider: Provider{
					Github: Github{
						Repository: "test/repo",
						Branch:     "main",
					},
				},
			},
			expectError: false,
			err:         nil,
		},
		{
			name: "valid-test-with-auth-token",
			config: Config{
				Provider: Provider{
					Github: Github{
						Repository: "test/repo",
						Branch:     "main",
						Auth: &Auth{
							Token: "test-token-very-private",
						},
					},
				},
			},
			expectError: false,
			err:         nil,
		},
		{
			name: "invalid-test-with-auth-no-token",
			config: Config{
				Provider: Provider{
					Github: Github{
						Repository: "test/repo",
						Branch:     "main",
						Auth: &Auth{
							Token: "",
						},
					},
				},
			},
			expectError: true,
			err:         ErrInvalidPersonalAccessToken,
		},
		{
			name: "invalid-test-without-repo",
			config: Config{
				Provider: Provider{
					Github: Github{
						Repository: "",
						Branch:     "main",
					},
				},
			},
			expectError: true,
			err:         ErrInvalidRepoName,
		},
		{
			name: "invalid-test-without-branch",
			config: Config{
				Provider: Provider{
					Github: Github{
						Repository: "test/repo",
						Branch:     "",
					},
				},
			},
			expectError: true,
			err:         ErrInvalidBranchName,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.ValidateProvider()

			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error, got: %v", err.Error())
			}
			if tt.expectError && err == nil {
				t.Errorf("Expected error: %v, got no error", tt.err.Error())
			}
			if tt.err != nil && tt.err != err {
				t.Errorf("Expected error: %v, got: %v", tt.err.Error(), err.Error())
			}
		})
	}
}

func TestGetCloneURL(t *testing.T) {
	tests := []struct {
		name        string
		config      Config
		expectedURL string
	}{
		{
			name: "valid-test-without-auth-token",
			config: Config{
				Provider: Provider{
					Github: Github{
						Repository: "test/repo",
						Branch:     "main",
					},
				},
			},
			expectedURL: "https://github.com/test/repo.git",
		},
		{
			name: "valid-test-with-auth-token",
			config: Config{
				Provider: Provider{
					Github: Github{
						Repository: "test/repo",
						Branch:     "main",
						Auth: &Auth{
							Token: "test-token-very-private",
						},
					},
				},
			},
			expectedURL: "https://github.com/test/repo.git",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := tt.config.GetCloneURL()
			if url != tt.expectedURL {
				t.Errorf("Expected url: %v, got: %v", tt.expectedURL, url)
			}
		})
	}
}
