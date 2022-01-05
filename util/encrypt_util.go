package util

import "errors"

// GenerateEncryptKey Generate a random EncryptKey
func GenerateEncryptKey() (encryptKey string, err error) {
	key, nonce, err := GenerateKeyAndNonce()
	if err != nil {
		return "", err
	}

	return key + nonce, nil
}

// EncryptByEncryptKey 加密
func EncryptByEncryptKey(encryptKey string, orgStr string) (ecryptStr string, err error) {
	if len(encryptKey) != 88 {
		return "", errors.New("EncryptKey not corret")
	}
	key, nonce, err := ValidateKeyAndNonce(encryptKey[0:64], encryptKey[64:88])
	if err != nil {
		return "", err
	}
	return Encrypt(key, nonce, orgStr)
}

// DecryptByEncryptKey 解密
func DecryptByEncryptKey(encryptKey string, encryptStr string) (decryptStr string, err error) {
	if len(encryptKey) != 88 {
		return "", errors.New("EncryptKey not corret")
	}
	key, nonce, err := ValidateKeyAndNonce(encryptKey[0:64], encryptKey[64:88])
	if err != nil {
		return "", err
	}
	return Decrypt(key, nonce, encryptStr)
}
