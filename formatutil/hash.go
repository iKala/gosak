package formatutil

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha1"
	"encoding/base32"
	"fmt"
	"log"
)

type binaryToText func(data []byte) string
type textToBinary func(data string) ([]byte, error)

// Sha1SumWithSalt caculates SHA1 hash with salt
func Sha1SumWithSalt(seed, salt string) string {
	input := seed + salt

	result := fmt.Sprintf("%x", sha1.Sum([]byte(input)))

	log.Printf("Sha1Sum: seed[%s], salt[%s], result[%s]", seed, salt, result)

	return result
}

// Aes128Encrypt encrypts by AES128
func Aes128Encrypt(key, cipherString, src string, encode binaryToText) (string, error) {

	block, _ := aes.NewCipher([]byte(key))
	ciphertext := []byte(cipherString)
	str := []byte(src)

	iv := ciphertext[:aes.BlockSize]
	encrypter := cipher.NewCFBEncrypter(block, iv)

	encrypted := make([]byte, len(str))
	encrypter.XORKeyStream(encrypted, str)

	result := encode(encrypted)

	return result, nil
}

// Aes128Decrypt decrypts by AES128
func Aes128Decrypt(key, cipherString, encodedSrc string, decode textToBinary) (string, error) {
	encrypted, err := decode(encodedSrc)
	if err != nil {
		log.Printf("Fail decode: err[%s]", err.Error())
		return "", err
	}

	block, _ := aes.NewCipher([]byte(key))
	ciphertext := []byte(cipherString)

	iv := ciphertext[:aes.BlockSize]
	decrypter := cipher.NewCFBDecrypter(block, iv)

	decrypted := make([]byte, len(encrypted))
	decrypter.XORKeyStream(decrypted, encrypted)

	return string(decrypted[:]), nil
}

// IsBase32Encoded checks if data is Base32 encoded
func IsBase32Encoded(data string) bool {
	_, err := base32.StdEncoding.DecodeString(data)

	return err == nil
}
