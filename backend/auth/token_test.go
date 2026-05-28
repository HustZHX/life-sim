package auth

import (
	"testing"
	"time"
)

func TestTokenSignVerify(t *testing.T) {
	issuer := NewTokenIssuer("test-secret")
	token, err := issuer.Sign(TokenKindGate, "", "", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	claims, err := issuer.Verify(token, TokenKindGate)
	if err != nil {
		t.Fatal(err)
	}
	if claims.Kind != TokenKindGate {
		t.Fatalf("kind = %s", claims.Kind)
	}
}

func TestTokenWrongKind(t *testing.T) {
	issuer := NewTokenIssuer("test-secret")
	token, _ := issuer.Sign(TokenKindGate, "", "", time.Hour)
	if _, err := issuer.Verify(token, TokenKindAccess); err == nil {
		t.Fatal("expected error")
	}
}
