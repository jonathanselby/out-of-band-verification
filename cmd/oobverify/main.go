package main

import (
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/jonathanselby/out-of-band-verification/internal/oob"
	"os"
	"time"
)

const version = "0.1.0"

type input struct {
	Operation      string       `json:"operation"`
	Person         oob.Person   `json:"person"`
	Prompt         string       `json:"prompt,omitempty"`
	Answer         string       `json:"answer,omitempty"`
	InboundContact string       `json:"inbound_contact,omitempty"`
	Session        *oob.Session `json:"session,omitempty"`
	Token          string       `json:"token,omitempty"`
	Connected      bool         `json:"connected,omitempty"`
}

func main() {
	var in string
	var v bool
	flag.StringVar(&in, "input", "", "request JSON")
	flag.BoolVar(&v, "version", false, "print version")
	flag.Parse()
	if v {
		fmt.Println("oobverify " + version)
		return
	}
	if in == "" {
		flag.Usage()
		os.Exit(2)
	}
	var x input
	b, e := os.ReadFile(in)
	if e != nil || json.Unmarshal(b, &x) != nil {
		bad(fmt.Errorf("invalid input"))
	}
	now := time.Now().UTC()
	switch x.Operation {
	case "challenge":
		r, e := oob.VerifyChallenge(x.Person, x.Prompt, x.Answer)
		if e != nil {
			bad(e)
		}
		json.NewEncoder(os.Stdout).Encode(r)
		if !r.Verified {
			os.Exit(1)
		}
	case "start-callback":
		key := sessionKey()
		s, e := oob.StartCallback(x.Person, x.InboundContact, now, 5*time.Minute, key)
		if e != nil {
			bad(e)
		}
		json.NewEncoder(os.Stdout).Encode(s)
	case "complete-callback":
		if x.Session == nil {
			bad(fmt.Errorf("session required"))
		}
		key := sessionKey()
		r, e := oob.CompleteCallback(*x.Session, x.Token, x.Person.ID, x.Connected, now, key)
		if e != nil {
			bad(e)
		}
		json.NewEncoder(os.Stdout).Encode(r)
		if !r.Verified {
			os.Exit(1)
		}
	default:
		os.Exit(2)
	}
}

func sessionKey() []byte {
	s := os.Getenv("OOB_SESSION_KEY")
	if s == "" {
		bad(fmt.Errorf("OOB_SESSION_KEY required for callback operations"))
	}
	b, err := base64.StdEncoding.DecodeString(s)
	if err != nil || len(b) < oob.MinSessionKeyBytes {
		bad(fmt.Errorf("invalid OOB_SESSION_KEY"))
	}
	return b
}

func bad(e error) {
	fmt.Fprintln(os.Stderr, e)
	os.Exit(3)
}
