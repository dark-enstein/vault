package tokenize

import (
	"context"
	"fmt"
	"github.com/dark-enstein/vault/internal/model"
	"github.com/dark-enstein/vault/pkg/store"
	"github.com/dark-enstein/vault/pkg/vlog"
	"github.com/joho/godotenv"
	"github.com/pkg/errors"
	"math/rand"
	"os"
	"strings"
	"time"
	"unsafe"
)

var (
	ErrKeyAlreadyExists = errors.New("key already exists. not overriding")
	ErrKeyDoesNotExists = "key %s does not exist"
	ErrDuplicateKeys    = errors.New("key already exists in request. accepted only the first one")
)

var (
	DefaultCipherLoc           = "./.cipher"
	EnvKeyAESCipher            = "CIPHER"
	EnvKeyInitializationVector = "IV"
	KeyDelimiter               = "__"
)

type Manager struct {
	store     store.Store
	cipher    map[string]string
	cipherLoc string
	log       *vlog.Logger
}

// NewManager creates a new instance of Manager. It manages token operations (retrieval, storage, servicing) throughout the lifetime of the server.
func NewManager(ctx context.Context, logger *vlog.Logger, opts ...Options) *Manager {
	log := logger.Logger()
	var manager = &Manager{}
	manager.log = logger
	manager.cipherLoc = DefaultCipherLoc
	manager.cipher = map[string]string{}
	if len(opts) == 0 {
		manager.store = store.NewSyncMap(ctx, manager.log)
	} else {
		log.Debug().Msg("a separate store option was passed in")
		for i := 0; i < len(opts); i++ {
			opts[i](manager)
		}
	}

	b, err := manager.store.Connect(ctx)
	if err != nil || !b {
		log.Debug().Msgf("connection failed to store: %s\n", err.Error())
	}

	// if cipher file doesn't exist
	if _, err := os.Stat(manager.cipherLoc); err != nil {
		manager.log.Logger().Debug().Msgf("cipher file %s not found, generating it", manager.cipherLoc)
		err = manager.GenerateCipher()
		if err != nil {
			manager.log.Logger().Error().Msgf("error encountered while writing cipher to file: %s\n", err.Error())
		}
	}

	// if cipher file already exists, emvMap is empty, so read from file
	if len(manager.cipher) == 0 {
		manager.cipher, err = godotenv.Read(manager.cipherLoc)
		if err != nil {
			manager.log.Logger().Error().Msgf("error encountered while reading cipher from file %s: %s\n", manager.cipherLoc, err.Error())
		}
	}

	return manager
}

// GenerateCipher generates a new AES cipher and Initialization Vector pais, and persists it to disk
func (m *Manager) GenerateCipher() error {
	// generate 32 digit key
	m.cipher[EnvKeyAESCipher] = GenAlphaNumericString(32)
	// generate 16 digit initialization vector
	m.cipher[EnvKeyInitializationVector] = GenAlphaNumericString(16)
	// write to file
	return godotenv.Write(m.cipher, m.cipherLoc)
}

// GetTokenByID returns the token owned by a specific ID/Key
func (m *Manager) GetTokenByID(ctx context.Context, id string) (*model.Tokenize, error) {
	log := m.log.Logger()
	var tokenStr string

	if val, err := m.store.Retrieve(ctx, id); err != nil {
		return nil, fmt.Errorf(ErrKeyDoesNotExists, id)
	} else {
		tokenStr = fmt.Sprint(val)
	}
	log.Debug().Msg("successfully ranged over store data")

	ss := strings.Split(id, KeyDelimiter)

	log.Debug().Msg("found token in store")
	return &model.Tokenize{
		ID: ss[0],
		Data: []model.Child{
			{
				Key:   ss[1],
				Value: tokenStr,
			},
		},
	}, nil
}

// GetAllTokens returns all tokens currently in the store
func (m *Manager) GetAllTokens(ctx context.Context) ([]*model.Tokenize, error) {
	log := m.log.Logger()
	allTokens := map[string]*model.Tokenize{}

	// pass a range func over the contents of the store and get the contents
	allTokenMap, err := m.store.RetrieveAll(ctx)
	if err != nil {
		log.Error().Msgf("error while retrieving all keys: %s\n", err.Error())
	}

	// parse all tokens into a slice of model.Tokenize
	for k, v := range allTokenMap {
		ss := strings.Split(k, KeyDelimiter)
		if val, ok := allTokens[k]; ok {
			if len(val.ID) == 0 {
				log.Debug().Msgf("token with id %s is already stored. continuing.", val.ID)
			}
			val.ID = ss[0]
			val.Data = append(val.Data, model.Child{
				Key:   ss[1],
				Value: v,
			})
			continue
		}
		allTokens[k] = &model.Tokenize{
			ID: ss[0],
		}
		allTokens[k].Data = append(allTokens[k].Data, model.Child{
			Key:   ss[1],
			Value: v,
		})
	}

	respTokens := []*model.Tokenize{}
	for _, v := range allTokens {
		respTokens = append(respTokens, v)
	}

	log.Debug().Msg("successfully parsed all tokens into a token array")
	return respTokens, nil
}

// ValidateResponse holds the error response from validation and the associated key.
type ValidateResponse struct {
	Key string
	Err error
}

// Validate is the high level api for validating all the user provided data
func (m *Manager) Validate(ctx context.Context, token *model.Tokenize, patch bool) ([]*ValidateResponse, bool) {
	keysValidationResp, ok := m.ValidateKeys(ctx, token)
	if !ok && !patch {
		m.log.Logger().Error().Msgf("error while validating keys")
		return keysValidationResp, ok
	}

	//valValidationResp, ok := m.ValidateValues(token)
	return keysValidationResp, true
}

// ValidateKeys validates the Keys used in the request, ensuring it doesn't already exist, and that it conforms with the standards.
func (m *Manager) ValidateKeys(ctx context.Context, token *model.Tokenize) ([]*ValidateResponse, bool) {
	tempMap := make(map[string]bool, len(token.Data))
	valResp := []*ValidateResponse{}
	var verdict = true
	parentKey := token.ID
	for i := 0; i < len(token.Data); i++ {
		childKey := token.Data[i].Key
		combinedKeyName := GetCombinedKey(parentKey, childKey)
		// check that key doesn't already exist
		var err error
		if err = keysIsPresent(ctx, combinedKeyName, tempMap, m.store); err != nil {
			verdict = false
			valResp = append(valResp, &ValidateResponse{combinedKeyName, fmt.Errorf("error validating keys: %s\n", err.Error())})
		}
	}
	return valResp, verdict
}

func keysIsPresent(ctx context.Context, key string, tempStore map[string]bool, store store.Store) error {
	if _, ok := tempStore[key]; ok {
		return ErrDuplicateKeys
	}
	if _, err := store.Retrieve(ctx, key); err == nil {
		return ErrKeyAlreadyExists
	}

	tempStore[key] = true
	return nil
}

// Tokenize manages the tokenization, and stores generated tokens in an internal store, for easy retrieval
func (m *Manager) Tokenize(ctx context.Context, key, val string) (string, error) {

	// Tokenize
	token, err := Tokenize(val, m.cipher)
	if err != nil {
		m.log.Logger().Error().Msgf("error occurred while generating token: %s\n", err.Error())
		return "", err
	}

	// proceed to store generated token
	err = m.store.Store(ctx, key, token.token)
	if err != nil {
		m.log.Logger().Error().Msgf("error occurred while storing token: %s\n", err.Error())
		return "", err
	}
	return token.token, nil
}

// Detokenize retrieves the value represented by a particular token, identified by the particular key
func (m *Manager) Detokenize(ctx context.Context, key, token string) (bool, string, error) {

	// ensure that token matches what is in store
	storedToken, err := m.store.Retrieve(ctx, key)
	if err != nil {
		m.log.Logger().Error().Msgf("error while confirming token key: %s\n", err.Error())
		return false, "", err
	}

	// check if the stored token match the provided token. abort if no match
	if storedToken != token {
		m.log.Logger().Error().Msgf("provided token does not match stored token. provided token: %s\n", store.Redact(token))
		return false, "", fmt.Errorf("provided token does not match stored token. provided token: %s\n", store.Redact(token))
	}

	// Detokenize
	decryptedStr, err := Detokenize(token, m.cipher)
	if err != nil {
		m.log.Logger().Error().Msgf("error occurred while decrypting token: %s\n", err.Error())
		return false, "", err
	}

	return true, decryptedStr, nil
}

// gotten from https://stackoverflow.com/questions/22892120/how-to-generate-a-random-string-of-a-fixed-length-in-go#:~:text=%22Mimicing%22%20strings.Builder%20with%20package%20unsafe
const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

var src = rand.NewSource(time.Now().UnixNano())

// GenAlphaNumericString mimics strings.Builder with package unsafe. According to the author it is one of the pastest implementation of strings builder.
func GenAlphaNumericString(n int) string {
	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return *(*string)(unsafe.Pointer(&b))
}

// DeleteTokenByID deletes the token from the store identified by ID
func (m *Manager) DeleteTokenByID(ctx context.Context, id string) (bool, error) {
	log := m.log.Logger()

	if b, err := m.store.Delete(ctx, id); err != nil || !b {
		return false, fmt.Errorf(ErrKeyDoesNotExists, id)
	}

	log.Debug().Msg("successfully deleted ID from store")

	log.Debug().Msg("found token in store")
	return true, nil
}

// PatchTokenByID updates a token in the store identified by ID
func (m *Manager) PatchTokenByID(ctx context.Context, key, val string) (string, error) {
	log := m.log.Logger()

	// Tokenize
	token, err := Tokenize(val, m.cipher)
	if err != nil {
		m.log.Logger().Error().Msgf("error occurred while generating token: %s\n", err.Error())
		return "", err
	}

	// patch token entry
	if b, err := m.store.Patch(ctx, key, token.token); err != nil || !b {
		return "", fmt.Errorf("error patching token: %s\n", err.Error())
	}

	log.Debug().Msg("successfully patched ID from store")
	return token.String(), nil
}

// IsErrKeyAlreadyExist enables easy checking of error
func IsErrKeyAlreadyExist(err error) bool {
	if err == ErrKeyAlreadyExists {
		return true
	}
	return false
}

// GetCombinedKey creates a key string unique to every value in the request object. This key string is a concatenation of all the parent keys that constitute the request data
func GetCombinedKey(s ...string) (cs string) {
	for i := 0; i < len(s); i++ {
		delimiter := ""
		if i > 0 {
			delimiter = KeyDelimiter
		}
		cs += delimiter + s[i]
	}
	return
}
