package utils

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"time"

	"golang.org/x/crypto/chacha20poly1305"
)

type Claims struct {
	Id        string `json:"id"`
	ExpiresAt int64  `json:"expiresAt"`
	IssuedAt  int64  `json:"issuedAt"`
}

func GenerateAccessToken(userId string, hexKey string) (string, error) {
	claims := Claims{
		Id:        userId,
		ExpiresAt: time.Now().Add(15 * time.Minute).Unix(),
		IssuedAt:  time.Now().Unix(),
	}

	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		panic(err)
	}

	claimsBytes := []byte(string(claimsJSON))

	accessToken, err := EncryptData(claimsBytes, hexKey)

	if err != nil {
		return "", err
	}

	return accessToken, nil
}

func GenerateRefreshToken(userId string, hexKey string) (string, error) {
	claims := Claims{
		Id:        userId,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour).Unix(),
		IssuedAt:  time.Now().Unix(),
	}

	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		panic(err)
	}

	claimsBytes := []byte(string(claimsJSON))

	refreshToken, err := EncryptData(claimsBytes, hexKey)

	if err != nil {
		return "", err
	}

	return refreshToken, nil
}

func ValidateToken(tokenStr string, hexKey string) (*Claims, error) {
	plaintext, err := DecryptData(tokenStr, hexKey)

	if err != nil {
		return nil, err
	}

	claims := &Claims{}
	err = json.Unmarshal([]byte(plaintext), claims)

	if err != nil {
		return nil, err
	}

	if claims.ExpiresAt < time.Now().Unix() {
		return nil, fmt.Errorf("token expired")
	}

	return claims, nil
}

func EncryptData(plaintext []byte, hexKey string) (string, error) {
	key, err := hex.DecodeString(hexKey)
	if err != nil {
		panic(err)
	}

	aead, err := chacha20poly1305.New(key)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := aead.Seal(nil, nonce, plaintext, nil)

	return hex.EncodeToString(append(nonce, ciphertext...)), nil
}

func DecryptData(ciphertextHex string, hexKey string) (string, error) {

	key, err := hex.DecodeString(hexKey)
	if err != nil {
		log.Println("Error decoding key:", err)
		panic(err)
	}
	ciphertext, err := hex.DecodeString(ciphertextHex)
	if err != nil {
		return "", err
	}

	aead, err := chacha20poly1305.New(key)
	if err != nil {
		return "", err
	}

	nonceSize := aead.NonceSize()
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	plaintext, err := aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}
