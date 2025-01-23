package auth

import (
	"testing"

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
