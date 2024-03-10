package store

import (
	"fmt"
	"github.com/rs/zerolog"
	"os"
	"path/filepath"
	"strings"
)

const (
	ErrTokenTypeNotString   = "token of type string required"
	DefaultUnredactedLength = 4
	// DefaultRedactedLength ensures infomation about the actual length of the encoded token isn't exposed to an unauthenticated client
	DefaultRedactedLength = 10
	DefaultRedactedToken  = "*"
)

// InterfaceIsString checks if an interface underlying type is a string
func InterfaceIsString(a any) (bool, string) {
	switch v := a.(type) {
	case string:
		return true, v
	default:
		return false, ""
	}
}

// Redact censors sensitive token before printing them in logs or as a response.
func Redact(s string) string {
	redacted := ""
	redacted += s[:DefaultUnredactedLength]

	redacted += strings.Repeat(DefaultRedactedToken, DefaultRedactedLength-len(redacted))
	return redacted
}

// IsValidFile checks if a given file address is valid
func IsValidFile(loc string, log *zerolog.Logger) error {
	var abs string
	var err error
	// some sanity check
	if loc[0] != '/' {
		abs, err = filepath.Abs(loc)
		if err != nil {
			log.Error().Msgf("cannot obtain absolute path to referenced path: %s\n", err.Error())
			return fmt.Errorf("cannot obtain absolute path to referenced path: %s\n", err.Error())
		}
	}

	// confirm dir of path exists
	dir := filepath.Dir(abs)
	if _, err = os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			log.Info().Msgf("base dir %s does not exist, creating it\n", dir)
			err := os.MkdirAll(dir, 0777)
			if err != nil {
				log.Info().Msgf("error while creating base dir %s: %s\n", dir, err.Error())
				return err
			}
		} else {
			return err
		}
	}
	return nil
}
