package utils

import "testing"

func TestHashAndCheckPassword(t *testing.T) {
	hash, err := HashPassword("s3cret-pass")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	if hash == "" || hash == "s3cret-pass" {
		t.Fatalf("hash looks wrong: %q", hash)
	}
	if !CheckPassword(hash, "s3cret-pass") {
		t.Fatal("CheckPassword: expected match for correct password")
	}
	if CheckPassword(hash, "wrong-pass") {
		t.Fatal("CheckPassword: expected mismatch for wrong password")
	}
}

func TestHashPasswordIsSalted(t *testing.T) {
	a, _ := HashPassword("same-input")
	b, _ := HashPassword("same-input")
	if a == b {
		t.Fatal("bcrypt should produce different hashes for the same input (random salt)")
	}
}
