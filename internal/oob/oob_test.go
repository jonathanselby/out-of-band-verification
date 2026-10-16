package oob

import (
	"testing"
	"time"
)

var testSessionKey = []byte("0123456789abcdef0123456789abcdef")

func challengePerson(t *testing.T) Person {
	t.Helper()
	c, err := NewChallenge("first school", "Bafflewick", []byte("1234567890123456"), MinChallengeIterations)
	if err != nil {
		t.Fatal(err)
	}
	return Person{ID: "ed", Callback: "+61-2-5550-0100", Challenges: []Challenge{c}}
}

func TestChallenge(t *testing.T) {
	p := challengePerson(t)
	r, err := VerifyChallenge(p, "first school", "Bafflewick")
	if err != nil || !r.Verified {
		t.Fatalf("valid challenge: %+v %v", r, err)
	}
	r, err = VerifyChallenge(p, "first school", "wrong")
	if err != nil || r.Verified || r.Reason != "challenge response rejected" {
		t.Fatalf("bad answer: %+v %v", r, err)
	}
	r, err = VerifyChallenge(p, "unregistered", "anything")
	if err != nil || r.Verified || r.Reason != "challenge not registered" {
		t.Fatalf("unregistered challenge: %+v %v", r, err)
	}
}

func TestChallengeSecurityFloorsAndValidation(t *testing.T) {
	if _, err := NewChallenge("q", "answer", []byte("short"), MinChallengeIterations); err == nil {
		t.Fatal("short salt accepted")
	}
	if _, err := NewChallenge("q", "answer", []byte("1234567890123456"), MinChallengeIterations-1); err == nil {
		t.Fatal("weak work factor accepted")
	}
	p := challengePerson(t)
	p.Challenges = append(p.Challenges, p.Challenges[0])
	if err := ValidatePerson(p); err == nil {
		t.Fatal("duplicate prompt accepted")
	}
}

func TestCallbackIgnoresInboundContact(t *testing.T) {
	now := time.Date(2026, 10, 17, 0, 0, 0, 0, time.UTC)
	s, err := StartCallback(challengePerson(t), "+61-400-EVIL", now, time.Minute, testSessionKey)
	if err != nil {
		t.Fatal(err)
	}
	if s.Callback != "+61-2-5550-0100" || !s.InboundContactIgnored {
		t.Fatalf("inbound contact affected callback: %+v", s)
	}
	r, err := CompleteCallback(s, s.Token, "ed", true, now, testSessionKey)
	if err != nil || !r.Verified {
		t.Fatalf("callback: %+v %v", r, err)
	}
}

func TestCallbackSessionIntegrity(t *testing.T) {
	now := time.Date(2026, 10, 17, 0, 0, 0, 0, time.UTC)
	s, err := StartCallback(challengePerson(t), "", now, time.Minute, testSessionKey)
	if err != nil {
		t.Fatal(err)
	}

	mutations := []func(*Session){
		func(x *Session) { x.PersonID = "mallory" },
		func(x *Session) { x.Callback = "+61-400-EVIL" },
		func(x *Session) { x.Expires = x.Expires.Add(time.Hour) },
		func(x *Session) { x.Token = "attacker-token" },
		func(x *Session) { x.InboundContactIgnored = !x.InboundContactIgnored },
	}
	for i, mutate := range mutations {
		x := s
		mutate(&x)
		r, err := CompleteCallback(x, x.Token, x.PersonID, true, now, testSessionKey)
		if err != nil || r.Verified || r.Reason != "callback session integrity rejected" {
			t.Fatalf("mutation %d accepted: %+v %v", i, r, err)
		}
	}
}

func TestCallbackExpiryWrongPersonTokenAndConnection(t *testing.T) {
	now := time.Date(2026, 10, 17, 0, 0, 0, 0, time.UTC)
	s, err := StartCallback(challengePerson(t), "", now, time.Minute, testSessionKey)
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name, token, person string
		connected           bool
		at                  time.Time
		reason              string
	}{
		{"wrong person", s.Token, "mallory", true, now, "session token or person mismatch"},
		{"wrong token", "wrong", "ed", true, now, "session token or person mismatch"},
		{"not connected", s.Token, "ed", false, now, "callback did not connect to registered person"},
		{"at expiry", s.Token, "ed", true, s.Expires, "callback session expired"},
	}
	for _, tc := range cases {
		r, err := CompleteCallback(s, tc.token, tc.person, tc.connected, tc.at, testSessionKey)
		if err != nil || r.Verified || r.Reason != tc.reason {
			t.Fatalf("%s: %+v %v", tc.name, r, err)
		}
	}
}

func TestCallbackKeyAndTTLBounds(t *testing.T) {
	now := time.Date(2026, 10, 17, 0, 0, 0, 0, time.UTC)
	p := challengePerson(t)
	if _, err := StartCallback(p, "", now, time.Minute, []byte("short")); err == nil {
		t.Fatal("short session key accepted")
	}
	if _, err := StartCallback(p, "", now, MaxCallbackTTL+time.Second, testSessionKey); err == nil {
		t.Fatal("excessive ttl accepted")
	}
	s, err := StartCallback(p, "", now, time.Minute, testSessionKey)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := CompleteCallback(s, s.Token, "ed", true, now, []byte("short")); err == nil {
		t.Fatal("completion accepted short session key")
	}
}
