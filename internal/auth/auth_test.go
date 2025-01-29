package auth

import (
	"encoding/hex"
	"net/http"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func TestHashPassword(t *testing.T) {
	password := "skibidi"
	hash, err := HashPassword(password)
	if err != nil {
		t.Errorf("HashPassword fialed: %v", err)
	}
	if hash == password {
		t.Error("Hash should not be equal to password")
	}
}

func TestCheckPasswordHash(t *testing.T) {
	password := "skibidi"
	hash, _ := HashPassword(password)

	err := CheckPasswordHash(password, hash)
	if err != nil {
		t.Errorf("CheckPasswordHash failed with valid password: %v", err)
	}

	err = CheckPasswordHash("wrongpass", hash)
	if err != bcrypt.ErrMismatchedHashAndPassword {
		t.Error("CheckPasswordHash failed with invalid password")
	}
}

func TestMakeJWT(t *testing.T) {
	tests := []struct {
		name      string
		userId    uuid.UUID
		expiresIn time.Duration
		wantErr   bool
	}{
		{
			name:      "Valid token generation",
			userId:    uuid.New(),
			expiresIn: 3600 * time.Second,
			wantErr:   false,
		},
		{
			name:      "Zero duration",
			userId:    uuid.New(),
			expiresIn: 0,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := MakeJWT(tt.userId, "skibidiKey", tt.expiresIn)

			if (err != nil) != tt.wantErr {
				t.Errorf("MakeJWT() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if token != "" {
					t.Error("Expected empty token when error")
				}
				return
			}

			// Verify token
			parsedToken, err := jwt.ParseWithClaims(token, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
				return []byte("skibidiKey"), nil
			})

			if err != nil {
				t.Errorf("Failed to parse token: %v", err)
				return
			}

			if claims, ok := parsedToken.Claims.(*jwt.RegisteredClaims); ok {
				// Verify claims
				if claims.Subject != tt.userId.String() {
					t.Errorf("Wrong subject. got = %v, want = %v", claims.Subject, tt.userId)
				}

				if claims.Issuer != "chirpy" {
					t.Errorf("Wrong issuer. got = %v, want = chirpy", claims.Issuer)
				}

				expectedExpiry := time.Now().UTC().Add(tt.expiresIn)
				if claims.ExpiresAt.Time.Sub(expectedExpiry) > time.Second {
					t.Errorf("Wrong expiry. got = %v, want = %v", claims.ExpiresAt, expectedExpiry)
				}
			} else {
				t.Error("Failed to parse claims")
			}
		})
	}
}

func TestJWTValidation(t *testing.T) {
	userId := uuid.New()
	secret := "test-secret"

	// Create token
	token, err := MakeJWT(userId, secret, time.Hour)
	if err != nil {
		t.Fatalf("Failed to create token: %v", err)
	}

	// Test valid token
	extractedId, err := ValidateJWT(token, secret)
	if err != nil {
		t.Errorf("Failed to validate token: %v", err)
	}
	if extractedId != userId {
		t.Errorf("Got wrong user ID. Want %v, got %v", userId, extractedId)
	}

	// Test invalid secret
	_, err = ValidateJWT(token, "wrong-secret")
	if err == nil {
		t.Error("Expected error with wrong secret")
	}

	// Test expired token
	expiredToken, _ := MakeJWT(userId, secret, -time.Hour)
	_, err = ValidateJWT(expiredToken, secret)
	if err == nil {
		t.Error("Expected error for expired token")
	}
}

func TestGetBearerToken(t *testing.T) {
	tests := []struct {
		name          string
		headerValue   string
		expectedToken string
		expectError   bool
	}{
		{
			name:          "Valid bearer token",
			headerValue:   "Bearer abc123",
			expectedToken: "abc123",
			expectError:   false,
		},
		{
			name:          "Empty header",
			headerValue:   "",
			expectedToken: "",
			expectError:   true,
		},
		{
			name:          "Missing Bearer prefix",
			headerValue:   "abc123",
			expectedToken: "",
			expectError:   true,
		},
		{
			name:          "Bearer with no token",
			headerValue:   "Bearer ",
			expectedToken: "",
			expectError:   true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			headers := http.Header{}
			if tc.headerValue != "" {
				headers.Add("Authorization", tc.headerValue)
			}

			token, err := GetBearerToken(headers)
			if tc.expectError && err == nil {
				t.Error("Expected an error but got none")
			}
			if !tc.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}

			if token != tc.expectedToken {
				t.Errorf("Expected token %q but got %q", tc.expectedToken, token)
			}
		})
	}
}

func TestMakeRefreshToken(t *testing.T) {
    // Run multiple times to ensure consistent behavior
    for i := 0; i < 100; i++ {
        token, err := MakeRefreshToken()
        if err != nil {
            t.Errorf("error creating token: %v", err)
        }

        // Check length (32 bytes encoded as hex = 64 characters)
        expectedTokenLength := 64
        if len(token) != expectedTokenLength {
            t.Errorf("Expected token length %d, but got %d", expectedTokenLength, len(token))
        }

        // Verify it's a valid hex string
        _, err = hex.DecodeString(token)
        if err != nil {
            t.Errorf("Token is not a valid hex string: %v", err)
        }
    }
}
