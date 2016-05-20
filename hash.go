package gosak

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha1"
	"encoding/base32"
	"encoding/base64"
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

	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		log.Printf("Fail NewCipher: err[%s]", err.Error())

		return "", err
	}

	str := []byte(src)

	ciphertext := []byte(cipherString)
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
		log.Printf("Fail base32 decode: err[%s]", err.Error())

		return "", err
	}

	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		log.Printf("Fail NewCipher: err[%s]", err.Error())

		return "", err
	}

	ciphertext := []byte(cipherString)
	iv := ciphertext[:aes.BlockSize]
	decrypter := cipher.NewCFBDecrypter(block, iv)

	decrypted := make([]byte, len(encrypted))
	decrypter.XORKeyStream(decrypted, encrypted)

	return string(decrypted[:]), nil
}

// Base64CustomeAlphabetEncode encodes wht custome alphabet by Base64
func Base64CustomeAlphabetEncode(alphabet string, src []byte) string {
	encoding := base64.NewEncoding(alphabet)

	return encoding.EncodeToString(src)
}

// Base64CustomeAlphabetDecode decodes wht custome alphabet by Base64
func Base64CustomeAlphabetDecode(alphabet string, base64Src string) ([]byte, error) {
	encoding := base64.NewEncoding(alphabet)

	return encoding.DecodeString(base64Src)
}

// IsBase32Encoded checks if data is Base32 encoded
func IsBase32Encoded(data string) bool {
	_, err := base32.StdEncoding.DecodeString(data)

	return err == nil
}
