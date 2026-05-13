package services

import (
	"crypto/sha512"
	"encoding/hex"
	"os"
	"testing"

	"github.com/devnolife/umkm-api/internal/config"
)

// resetConfig is needed because config.cfg is a package-level singleton.
// Tests using env-dependent helpers must reset between cases.
func resetConfig() {
	// Force re-load by clearing the cached pointer via Load() returning fresh.
	// We can't access the unexported var, so we rely on Load only being called
	// once; tests set env BEFORE the first call.
	_ = config.Get()
}

func TestVerifyNotificationSignature(t *testing.T) {
	serverKey := "test-server-key"
	t.Setenv("MIDTRANS_SERVER_KEY", serverKey)
	t.Setenv("JWT_SECRET", "x")
	t.Setenv("DATABASE_URL", "postgres://x")
	resetConfig()

	orderID := "ORD-20250101-ABC"
	statusCode := "200"
	grossAmount := "15000.00"

	raw := orderID + statusCode + grossAmount + serverKey
	sum := sha512.Sum512([]byte(raw))
	want := hex.EncodeToString(sum[:])

	if !VerifyNotificationSignature(orderID, statusCode, grossAmount, want) {
		t.Fatalf("expected valid signature to verify")
	}
	if VerifyNotificationSignature(orderID, statusCode, grossAmount, "deadbeef") {
		t.Fatalf("expected invalid signature to fail")
	}
	if VerifyNotificationSignature(orderID, statusCode, "0", want) {
		t.Fatalf("expected mismatched grossAmount to fail")
	}
}

func TestMain(m *testing.M) {
	// Ensure required env exists before config.Load() is invoked anywhere.
	if os.Getenv("JWT_SECRET") == "" {
		os.Setenv("JWT_SECRET", "test-jwt-secret")
	}
	if os.Getenv("DATABASE_URL") == "" {
		os.Setenv("DATABASE_URL", "postgres://test")
	}
	if os.Getenv("MIDTRANS_SERVER_KEY") == "" {
		os.Setenv("MIDTRANS_SERVER_KEY", "test-server-key")
	}
	os.Exit(m.Run())
}
