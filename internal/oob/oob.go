package oob

import (
	"crypto/hmac"
	"crypto/pbkdf2"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"hash"
	"strings"
	"time"
)

const (
	MinChallengeIterations = 600000
	MinChallengeSaltBytes  = 16
	ChallengeHashBytes     = 32
	MinSessionKeyBytes     = 32
	MaxCallbackTTL         = 30 * time.Minute
)

type Challenge struct {
	Prompt     string `json:"prompt"`
	Salt       string `json:"salt"`
	Iterations int    `json:"iterations"`
	Hash       string `json:"hash"`
}

type Person struct {
	ID         string      `json:"id"`
	Callback   string      `json:"callback,omitempty"`
	Challenges []Challenge `json:"challenges,omitempty"`
}

type Session struct {
	Token                 string    `json:"token"`
	PersonID              string    `json:"person_id"`
	Callback              string    `json:"callback"`
	Expires               time.Time `json:"expires"`
	InboundContactIgnored bool      `json:"inbound_contact_ignored"`
	Seal                  string    `json:"seal"`
}

type Result struct {
	Verified bool   `json:"verified"`
	PersonID string `json:"person_id"`
	Method   string `json:"method"`
	Reason   string `json:"reason,omitempty"`
}

func ValidatePerson(p Person) error {
	if strings.TrimSpace(p.ID) == "" {
		return errors.New("person id is required")
	}
	seen := map[string]bool{}
	for _, c := range p.Challenges {
		if strings.TrimSpace(c.Prompt) == "" || seen[c.Prompt] {
			return errors.New("invalid or duplicate challenge prompt")
		}
		seen[c.Prompt] = true
		if err := validateChallenge(c); err != nil {
			return err
		}
	}
	return nil
}

func VerifyChallenge(p Person, prompt, answer string) (Result, error) {
	if err := ValidatePerson(p); err != nil {
		return Result{}, err
	}
	if strings.TrimSpace(prompt) == "" {
		return Result{}, errors.New("challenge prompt is required")
	}
	for _, c := range p.Challenges {
		if c.Prompt == prompt {
			salt, _ := base64.StdEncoding.DecodeString(c.Salt)
			want, _ := base64.StdEncoding.DecodeString(c.Hash)
			got, err := pbkdf2.Key(sha256.New, answer, salt, c.Iterations, len(want))
			if err != nil {
				return Result{}, err
			}
			ok := subtle.ConstantTimeCompare(got, want) == 1
			return Result{Verified: ok, PersonID: p.ID, Method: "challenge", Reason: reason(ok)}, nil
		}
	}
	return Result{PersonID: p.ID, Method: "challenge", Reason: "challenge not registered"}, nil
}

func NewChallenge(prompt, answer string, salt []byte, iterations int) (Challenge, error) {
	if strings.TrimSpace(prompt) == "" || answer == "" || len(salt) < MinChallengeSaltBytes || iterations < MinChallengeIterations {
		return Challenge{}, errors.New("challenge parameters below minimum")
	}
	h, err := pbkdf2.Key(sha256.New, answer, salt, iterations, ChallengeHashBytes)
	if err != nil {
		return Challenge{}, err
	}
	return Challenge{
		Prompt:     prompt,
		Salt:       base64.StdEncoding.EncodeToString(salt),
		Iterations: iterations,
		Hash:       base64.StdEncoding.EncodeToString(h),
	}, nil
}

func validateChallenge(c Challenge) error {
	if c.Iterations < MinChallengeIterations {
		return errors.New("challenge work factor below minimum")
	}
	salt, err := base64.StdEncoding.DecodeString(c.Salt)
	if err != nil || len(salt) < MinChallengeSaltBytes {
		return errors.New("invalid challenge salt")
	}
	h, err := base64.StdEncoding.DecodeString(c.Hash)
	if err != nil || len(h) != ChallengeHashBytes {
		return errors.New("invalid challenge hash")
	}
	return nil
}

func StartCallback(p Person, inboundContact string, now time.Time, ttl time.Duration, sessionKey []byte) (Session, error) {
	if err := ValidatePerson(p); err != nil {
		return Session{}, err
	}
	if strings.TrimSpace(p.Callback) == "" {
		return Session{}, errors.New("pre-registered callback required")
	}
	if now.IsZero() || ttl <= 0 || ttl > MaxCallbackTTL {
		return Session{}, errors.New("invalid callback time or ttl")
	}
	if len(sessionKey) < MinSessionKeyBytes {
		return Session{}, errors.New("session key below minimum")
	}
	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		return Session{}, err
	}
	s := Session{
		Token:                 base64.RawURLEncoding.EncodeToString(b),
		PersonID:              p.ID,
		Callback:              p.Callback,
		Expires:               now.Add(ttl).UTC(),
		InboundContactIgnored: strings.TrimSpace(inboundContact) != "",
	}
	s.Seal = sealSession(s, sessionKey)
	return s, nil
}

func CompleteCallback(s Session, token, personID string, connected bool, now time.Time, sessionKey []byte) (Result, error) {
	r := Result{PersonID: personID, Method: "callback"}
	if len(sessionKey) < MinSessionKeyBytes {
		return Result{}, errors.New("session key below minimum")
	}
	if now.IsZero() {
		return Result{}, errors.New("callback completion time required")
	}
	if strings.TrimSpace(s.Token) == "" || strings.TrimSpace(s.PersonID) == "" || strings.TrimSpace(s.Callback) == "" || s.Expires.IsZero() || strings.TrimSpace(s.Seal) == "" {
		r.Reason = "invalid callback session"
		return r, nil
	}
	wantSeal := sealSession(s, sessionKey)
	if subtle.ConstantTimeCompare([]byte(wantSeal), []byte(s.Seal)) != 1 {
		r.Reason = "callback session integrity rejected"
		return r, nil
	}
	if !now.Before(s.Expires) {
		r.Reason = "callback session expired"
		return r, nil
	}
	if personID != s.PersonID || subtle.ConstantTimeCompare([]byte(token), []byte(s.Token)) != 1 {
		r.Reason = "session token or person mismatch"
		return r, nil
	}
	if !connected {
		r.Reason = "callback did not connect to registered person"
		return r, nil
	}
	r.Verified = true
	return r, nil
}

func sealSession(s Session, key []byte) string {
	m := hmac.New(sha256.New, key)
	writeField(m, s.Token)
	writeField(m, s.PersonID)
	writeField(m, s.Callback)
	var b [8]byte
	binary.BigEndian.PutUint64(b[:], uint64(s.Expires.UTC().UnixNano()))
	m.Write(b[:])
	if s.InboundContactIgnored {
		m.Write([]byte{1})
	} else {
		m.Write([]byte{0})
	}
	return base64.RawURLEncoding.EncodeToString(m.Sum(nil))
}

func writeField(m hash.Hash, s string) {
	var b [4]byte
	binary.BigEndian.PutUint32(b[:], uint32(len(s)))
	m.Write(b[:])
	m.Write([]byte(s))
}

func reason(ok bool) string {
	if ok {
		return ""
	}
	return "challenge response rejected"
}
