package formatutil

import (
	"encoding/hex"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSha1SumWithSalt(t *testing.T) {
	result := Sha1SumWithSalt("/broadcaster/1ki2qjg0i135l8fr", "UC8dnxqCS8U26Ax6HR4xWLVN")

	assert.Equal(t, "d4af93af41fab8f4c10999c76c00a0a61dd5bc68", result)
}

func TestAes128Encrypt(t *testing.T) {
	testedData := "0020e26f71144c8bada2ba1aa274449b-1448934753"
	log.Printf("data: %s", testedData)

	encrypted, _ := Aes128Encrypt("livehouse1234567", "abcdef1234567890", testedData, hex.EncodeToString)
	log.Printf("encrypted: %s", encrypted)

	// decode fails
	decrypted, err := Aes128Decrypt("livehouse1234567", "abcdef1234567890", "RANDOM", hex.DecodeString)
	assert.NotNil(t, err)

	decrypted, _ = Aes128Decrypt("livehouse1234567", "abcdef1234567890", encrypted, hex.DecodeString)
	log.Printf("decrypted: %s", decrypted)

	assert.Equal(t, testedData, decrypted)
}

func TestIsBase32Encoded(t *testing.T) {
	testedData := "0020e26f71144c8bada2ba1aa274449b"
	assert.False(t, IsBase32Encoded(testedData))

	testedData = "RW7A5VWF2KK5OZFFZQFTRSWBS3Y2EY673WBAJSNHWUEMCMLQMW24ZF5TLOUOI4ZWHLQB6==="
	assert.True(t, IsBase32Encoded(testedData))
}
