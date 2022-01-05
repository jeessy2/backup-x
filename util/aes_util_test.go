package util

import (
	"testing"
)

func TestAesGcm(t *testing.T) {
	keyHexStr, nonceHexStr, err := GenerateKeyAndNonce()
	if err != nil {
		t.Error(err)
	}

	key, nonce, err := ValidateKeyAndNonce(keyHexStr, nonceHexStr)
	if err != nil {
		t.Error(err)
	}

	plainText := "abc123"

	cipherText, err := Encrypt(key, nonce, plainText)
	if err != nil {
		t.Error(err)
	}

	plainText2, err := Decrypt(key, nonce, cipherText)
	if err != nil {
		t.Error(err)
	}

	if plainText2 != plainText {
		t.Error("aes gcm error")
	}

}
