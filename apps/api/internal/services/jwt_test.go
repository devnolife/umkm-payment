package services

import (
	"testing"

	"github.com/devnolife/umkm-api/internal/models"
)

func TestSignAndParseToken(t *testing.T) {
	tok, err := SignToken("user-123", models.RoleBuyer)
	if err != nil {
		t.Fatalf("SignToken: %v", err)
	}
	if tok == "" {
		t.Fatal("expected non-empty token")
	}

	claims, err := ParseToken(tok)
	if err != nil {
		t.Fatalf("ParseToken: %v", err)
	}
	if claims.UserID != "user-123" {
		t.Errorf("UserID = %q, want %q", claims.UserID, "user-123")
	}
	if claims.Role != models.RoleBuyer {
		t.Errorf("Role = %q, want %q", claims.Role, models.RoleBuyer)
	}
	if claims.Subject != "user-123" {
		t.Errorf("Subject = %q, want %q", claims.Subject, "user-123")
	}
}

func TestParseTokenRejectsGarbage(t *testing.T) {
	if _, err := ParseToken("not-a-jwt"); err == nil {
		t.Fatal("expected error for garbage token")
	}
}

func TestParseTokenRejectsWrongSecret(t *testing.T) {
	tok, err := SignToken("u", models.RoleAdmin)
	if err != nil {
		t.Fatal(err)
	}
	// Tamper: flip the last char of signature.
	if len(tok) < 2 {
		t.Fatal("token too short")
	}
	last := tok[len(tok)-1]
	flipped := byte('A')
	if last == 'A' {
		flipped = 'B'
	}
	tampered := tok[:len(tok)-1] + string(flipped)
	if _, err := ParseToken(tampered); err == nil {
		t.Fatal("expected error for tampered token")
	}
}
