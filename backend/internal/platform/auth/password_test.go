package auth

import "testing"

func TestHashAndVerifyPassword(t *testing.T) {
	hash, err := HashPassword("Admin123!")
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	ok, err := VerifyPassword(hash, "Admin123!")
	if err != nil || !ok {
		t.Fatalf("expected valid password, err=%v ok=%v", err, ok)
	}
	ok, err = VerifyPassword(hash, "wrong")
	if err != nil {
		t.Fatalf("verify error: %v", err)
	}
	if ok {
		t.Fatal("expected invalid password")
	}
}
