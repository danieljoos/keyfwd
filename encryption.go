package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
)

// AES based encryption.
type Encryption struct {
	cipher cipher.Block
}

// Initialize the encryption object using the given secret.
func (t *Encryption) Initialize(secret []byte) {
	hasher := sha256.New()
	hasher.Write(secret)
	t.cipher, _ = aes.NewCipher(hasher.Sum(nil))
}

// Encrypt the given bytes using AES.
// Generates a new, random initialization vector each time called.
// This ain't very fast, but hey, we just transfer a few key strokes.
// Returns the initialization vector followed by the encrypted data
// as single byte array.
func (t *Encryption) Encrypt(data []byte) []byte {
	iv := make([]byte, t.cipher.BlockSize())
	rand.Read(iv)
	encrypter := cipher.NewCFBEncrypter(t.cipher, iv)
	ret := make([]byte, len(data))
	encrypter.XORKeyStream(ret, data)
	return append(iv, ret...)
}

// Decrypts the given byte array.
// The function expects the given data to include the initialization vector
// within the first bytes (0 to BlockSize(16)). The remaining bytes contain the actual
// data to decrypt.
// Returns the decrypted data as byte array.
func (t *Encryption) Decrypt(data []byte) []byte {
	blockSize := t.cipher.BlockSize()
	decrypter := cipher.NewCFBDecrypter(t.cipher, data[0:blockSize])
	ret := make([]byte, len(data)-blockSize)
	decrypter.XORKeyStream(ret, data[blockSize:len(data)])
	return ret
}
