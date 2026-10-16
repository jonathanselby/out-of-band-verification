# Out-of-Band Verification — Product Requirements Document

**Version:** 0.1

## 1. Problem

An attacker who controls an instruction can often also supply a convenient phone number, email address or other contact route and then “verify” the same instruction through that attacker-controlled route. Out-of-band verification must break that loop by using verification material registered before the suspicious instruction arrived.

A stateless implementation has a second problem: callback-session JSON supplied back to the verifier must not be forgeable or editable by the caller.

## 2. Product objective

Provide a small Go verification primitive that can verify a claimed person through pre-registered challenge material or a callback plan using only stored contact details, with tamper-evident short-lived callback sessions.

> **Never verify an instruction using contact details supplied with that instruction.**

## 3. In scope

- pre-registered person IDs;
- optional pre-registered callback contact;
- slow-hashed challenge records;
- challenge verification;
- callback planning that ignores inbound contact details;
- random short-lived callback-session token;
- HMAC-SHA256 sealing of callback session state;
- callback completion using seal, expiry, person, token and connected status;
- structured results and stable CLI exits;
- Go 1.26 and Linux ARM64 build proof.

## 4. Out of scope

Placing calls, sending messages, voice/face biometrics, deepfake detection, person/contact enrolment, secret-store implementation, challenge-answer recovery, trust scoring, user monitoring, workflow/approval of the underlying action and durable verification-session storage.

## 5. Functional requirements

### FR1 — Person/challenge validation
Person ID is mandatory. Challenge prompts are non-blank and unique. Challenge records require valid base64 salt/hash, salt >=16 bytes, hash 32 bytes and PBKDF2-HMAC-SHA256 >=600,000 iterations.

### FR2 — Challenge result
A registered prompt with correct answer verifies; wrong answer or unknown prompt returns deterministic refusal without exposing the expected answer.

### FR3 — Stored callback only
`StartCallback` requires a pre-registered callback on the person and never substitutes the inbound contact. Presence of inbound contact is recorded as ignored.

### FR4 — Bounded callback session
Callback TTL is >0 and <=30 minutes. A cryptographically random 24-byte token is generated.

### FR5 — Session integrity
All security-relevant session fields are sealed using HMAC-SHA256 with a deployment-owned key of at least 32 bytes. Any field modification causes refusal.

### FR6 — Callback completion
Completion requires a valid seal, time before expiry, matching person, matching token and `connected=true`.

### FR7 — Secret boundary
The session HMAC key is supplied by deployment and never persisted in the repository. CLI callback operations obtain it from `OOB_SESSION_KEY` as standard base64.

## 6. Acceptance criteria

Automated evidence must show:

1. valid challenge success and bad/unknown challenge refusal;
2. weak challenge salt/work factor and duplicate prompt rejection;
3. inbound contact cannot alter stored callback destination;
4. a normal sealed callback session completes;
5. mutation of person, callback, expiry, token or inbound-contact marker invalidates the session;
6. wrong person/token, disconnected callback and exact-expiry completion are refused;
7. short session key and excessive TTL are rejected;
8. format, vet, race, native build, exact version and Linux ARM64 build pass.

## 7. Definition of done

All criteria have executable evidence; the operator-assertion and shared-knowledge limitations are explicit; no caller-provided contact can become the callback route; and generated design evidence is recovered from the exact release-candidate source commit.
