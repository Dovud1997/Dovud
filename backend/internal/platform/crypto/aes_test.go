package crypto_test

import (
	"testing"

	"github.com/Dovud1997/Dovud/backend/internal/platform/crypto"
)

func TestSecretBoxRoundTrip(t *testing.T) {
	box, err := crypto.NewSecretBox("test-secret-key-min-32-chars!!!!")
	if err != nil {
		t.Fatal(err)
	}
	enc, err := box.Seal([]byte(`{"password":"secret"}`))
	if err != nil {
		t.Fatal(err)
	}
	plain, err := box.Open(enc)
	if err != nil {
		t.Fatal(err)
	}
	if string(plain) != `{"password":"secret"}` {
		t.Fatalf("got %s", plain)
	}
}
