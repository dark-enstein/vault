package token

import (
	"fmt"
	"github.com/dark-enstein/vault/internal/tokenize"
)

// Tokenize takes a string and encryption parameters to generate an encrypted token.
// It returns the encrypted token, the cipher used for the encryption, and any error encountered.
// Parameters:
// - s: the original string to encrypt
// - aesSize: the size of the AES key to generate
// - cipherSize: the size of the initialization vector to generate
// - cipher: optional; an existing cipher map to use for encryption
func Tokenize(s string, aesSize, cipherSize int, cipher ...map[string]string) (string, map[string]string, error) {
	// Generate a new cipher string based on the provided AES and cipher sizes.
	var cypher = genCipherString(aesSize, cipherSize)

	// If a cipher map is provided, use that instead of the generated one.
	if len(cypher) == 0 {
		// If more than one cipher map is passed, return an error as it's not clear which to use.
		if len(cipher) > 1 {
			return "", nil, fmt.Errorf("too many cipher maps passed in")
		}
		cypher = cipher[0]
	}

	// Use the tokenize package's Tokenize function to generate the token.
	token, err := tokenize.Tokenize(s, cypher)
	if err != nil {
		// If there's an error during tokenization, return it.
		return "", nil, fmt.Errorf("encountered error while tokenizing: %v", err)
	}

	// Return the encrypted token and the cipher map used for encryption.
	return token.String(), cypher, nil
}
