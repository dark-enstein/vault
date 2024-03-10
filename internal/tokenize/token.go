package tokenize

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"github.com/pkg/errors"
)

var (
	ErrCipherToken404AES                 = errors.New("aes cipher not found in cipher map")
	ErrCipherToken404IV                  = errors.New("initialization vector (IV) not found in cipher map")
	ErrTokenInvalidPadding               = errors.New("invalid token string: padded bytes larger than aes block size: 16")
	ErrTokenInvalidPaddingNotHomogeneous = errors.New("invalid token string: padded bytes are not all the same")
	ErrTokenInvalidBlockSize             = errors.New("invalid token string: decrypted bytes size is not a multiple of the block size")
)

type Token struct {
	token string
}

func (t *Token) String() string {
	return t.token
}

func Tokenize(s string, cypher map[string]string) (*Token, error) {
	// resolve aes cipher and initialization vector
	var aesKey, iv string
	var ok bool
	if aesKey, ok = cypher[EnvKeyAESCipher]; !ok {
		return nil, ErrCipherToken404AES
	}

	if iv, ok = cypher[EnvKeyInitializationVector]; !ok {
		return nil, ErrCipherToken404IV
	}

	// get request string padded bytes. Padding is done following PKCS #7: https://en.wikipedia.org/wiki/PKCS_7
	padded := getPaddedBlock(s)

	block, err := aes.NewCipher([]byte(aesKey))
	if err != nil {
		return nil, err
	}

	// generate cipher mode using cipher block
	mode := cipher.NewCBCEncrypter(block, []byte(iv))

	// create a byte block to hold the encrypted bytes. It will be the length of the padded request string
	encryptedBytes := make([]byte, len(padded))
	mode.CryptBlocks(encryptedBytes, padded)

	// encode encryptedBytes using base64 encoding
	token := base64.StdEncoding.EncodeToString(encryptedBytes)
	return &Token{token: token}, nil
}

func Detokenize(token string, cypher map[string]string) (string, error) {
	// resolve aes cipher and initialization vector
	var aesKey, iv string
	var ok bool
	if aesKey, ok = cypher[EnvKeyAESCipher]; !ok {
		return "", ErrCipherToken404AES
	}

	if iv, ok = cypher[EnvKeyInitializationVector]; !ok {
		return "", ErrCipherToken404IV
	}

	// base64 decode token string
	encryptedBytes, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return "", err
	}

	// begin decryption process
	block, err := aes.NewCipher([]byte(aesKey))
	if err != nil {
		return "", err
	}

	// check if the ciphertext length is a multiple of the block size
	if len(encryptedBytes)%aes.BlockSize != 0 {
		return "", ErrTokenInvalidBlockSize
	}

	// decryption core
	decrypted := make([]byte, len(encryptedBytes))
	mode := cipher.NewCBCDecrypter(block, []byte(iv))
	mode.CryptBlocks(decrypted, encryptedBytes)

	// extract padding metadata
	padding := int(decrypted[len(decrypted)-1])

	// test decrypted padding byte integrity
	if padding > aes.BlockSize || decrypted[len(decrypted)-1] == 0 {
		return "", ErrTokenInvalidPadding
	}

	// check that all the padded strings are the same
	for _, v := range decrypted[len(decrypted)-padding:] {
		if padding != int(v) {
			return "", ErrTokenInvalidPaddingNotHomogeneous
		}
	}

	// return decrypted without padding
	return string(decrypted[:len(decrypted)-padding]), nil
}

// getPaddedBlock returns a properly padded bytes block such that it is works with AES encrypting requirements. The padding is done following PKCS #7: https://en.wikipedia.org/wiki/PKCS_7
func getPaddedBlock(s string) []byte {
	// get bytes representation and length of string. this is needed for block checking and AES encryption.
	sBytes := []byte(s)
	length := len(sBytes)

	// calculate the mod 16 of the bytes length, to determing how much is required to make the block a multiple of 16. AES standard. https://en.wikipedia.org/wiki/Advanced_Encryption_Standard
	paddingRequired := 16 - length%16

	// following PKCS #7, the padding to be added (if needed) will be a repetition of the byte representation of the reminder
	// create a mod16 block
	sPaddedBlock := make([]byte, length)
	// first copy the source bytes into the new mod16 block
	copy(sPaddedBlock, sBytes)

	padding := bytes.Repeat([]byte{byte(paddingRequired)}, paddingRequired)
	// then add the padded bytes to the same mod16 block
	sPaddedBlock = append(sPaddedBlock, padding...)
	return sPaddedBlock
}
