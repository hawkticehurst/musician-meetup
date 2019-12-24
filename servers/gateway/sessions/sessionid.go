package sessions

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
)

// InvalidSessionID represents an empty, invalid session ID
const InvalidSessionID SessionID = ""

// idLength is the length of the ID portion
const idLength = 32

// signedLength is the full length of the signed session ID
// (ID portion plus signature)
const signedLength = idLength + sha256.Size

// SessionID represents a valid, digitally-signed session ID.
// This is a base64 URL encoded string created from a byte slice
// where the first `idLength` bytes are crytographically random
// bytes representing the unique session ID, and the remaining bytes
// are an HMAC hash of those ID bytes (i.e., a digital signature).
// The byte slice layout is like so:
// +-----------------------------------------------------+
// |...32 crypto random bytes...|HMAC hash of those bytes|
// +-----------------------------------------------------+
type SessionID string

// ErrInvalidID is returned when an invalid session id is passed to ValidateID()
var ErrInvalidID = errors.New("Invalid Session ID")

// NewSessionID creates and returns a new digitally-signed session ID,
// using `signingKey` as the HMAC signing key. An error is returned only
// if there was an error generating random bytes for the session ID
func NewSessionID(signingKey string) (SessionID, error) {
	if len(signingKey) == 0 {
		return InvalidSessionID, ErrInvalidID
	}

	// Create byte slice using the first 32 bytes as the cryptographically random bytes
	sID := make([]byte, idLength)
	_, err := rand.Read(sID)
	if err != nil {
		return InvalidSessionID, err
	}

	// Create a new HMAC hasher with the signingKey
	hasher := hmac.New(sha256.New, []byte(signingKey))
	// Sign the session ID data mixing a secret key and hashing the data
	hasher.Write(sID)
	// Calculate the HMAC signature
	hmacSignature := hasher.Sum(nil)
	// Combine the session ID with the HMAC signature
	s := append(sID, hmacSignature...)
	// HTTP wants text, so we encode the binary data into base 64
	encodedSessionID := SessionID(base64.URLEncoding.EncodeToString([]byte(s)))

	return encodedSessionID, nil
}

// ValidateID validates the string in the `id` parameter
// using the `signingKey` as the HMAC signing key
// and returns an error if invalid, or a SessionID if valid
func ValidateID(id string, signingKey string) (SessionID, error) {
	// Decode back into binary when token is received
	decodedSigningKey, err := base64.URLEncoding.DecodeString(id)
	if err != nil {
		return InvalidSessionID, err
	}
	// Create a new HMAC hasher with the signingKey
	hasher := hmac.New(sha256.New, []byte(signingKey))
	// Hash the first 32 bytes
	hasher.Write([]byte(decodedSigningKey[0:32]))
	// Calculate the signature
	hmacSignature1 := hasher.Sum(nil)
	// Compare the calculated hash with the hash of the last 32 bytes
	if hmac.Equal(hmacSignature1, []byte(decodedSigningKey[32:])) {
		return SessionID(id), nil
	}

	return InvalidSessionID, ErrInvalidID
}

// String returns a string representation of the sessionID
func (sid SessionID) String() string {
	return string(sid)
}
