// Package token provides functionalities for encryption operations and handling of tokens.
package token

import (
	"fmt"
	"github.com/dark-enstein/vault/internal/tokenize"
)

// EnvKeyAESCipher is the environment key name for the AES cipher string.
var EnvKeyAESCipher = "CIPHER"

// EnvKeyInitializationVector is the environment key name for the initialization vector.
var EnvKeyInitializationVector = "IV"

// NewCipher generates a new cipher map containing an AES cipher string and an initialization vector.
// The sizes of the AES key and the initialization vector are specified by the parameters.
func NewCipher(aesSize, cipherSize int) map[string]string {
	return genCipherString(aesSize, cipherSize)
}

// Detokenize takes an encrypted string and decryption parameters to revert to the original string.
// It returns the decrypted string, the cipher used for the decryption, and any error encountered.
// Parameters:
// - s: the encrypted string to decrypt
// - aesSize: the size of the AES key that was used for encryption
// - cipherSize: the size of the initialization vector that was used for encryption
// - cipher: optional; an existing cipher map to use for decryption
func Detokenize(s string, aesSize, cipherSize int, cipher ...map[string]string) (string, map[string]string, error) {
	// Generate a new cipher string based on the provided AES and cipher sizes, if not provided.
	var cypher = genCipherString(aesSize, cipherSize)

	// Check if an existing cipher map was provided and use it if there's only one.
	// If none provided, use the generated cipher.
	if len(cipher) > 0 {
		if len(cipher) > 1 {
			// Return an error if more than one cipher map is passed, as it's unclear which to use.
			return "", nil, fmt.Errorf("too many cipher maps passed in")
		}
		cypher = cipher[0]
	}

	// Use the tokenize package's Detokenize function to decrypt the token.
	token, err := tokenize.Detokenize(s, cypher)
	if err != nil {
		// If there's an error during detokenization, return it.
		return "", nil, fmt.Errorf("encountered error while detokenizing: %v", err)
	}

	// Return the decrypted string and the cipher map used for decryption.
	return token, cypher, nil
}

// genCipherString is a helper function that generates a cipher map with an AES cipher string
// and an initialization vector, both of specified sizes.
func genCipherString(aesSize, cipherSize int) map[string]string {
	cipher := make(map[string]string, 2)
	// Generate an alphanumeric string of the given size for the AES cipher key.
	cipher[EnvKeyAESCipher] = GenerateAlphaNumericString(aesSize)
	// Generate an alphanumeric string of the given size for the initialization vector.
	cipher[EnvKeyInitializationVector] = GenerateAlphaNumericString(cipherSize)

	return cipher
}

// GenerateAlphaNumericString generates a random alphanumeric string of the specified size.
// It uses the tokenize package's GenAlphaNumericString function for the generation.
func GenerateAlphaNumericString(size int) string {
	return tokenize.GenAlphaNumericString(size)
}
