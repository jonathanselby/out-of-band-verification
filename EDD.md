# Out-of-Band Verification — Engineering Design Document

**Version:** 0.1

## 1. Purpose

A small stateless verification primitive for pre-registered shared knowledge and callback contact.

> **Never verify an instruction using contact details supplied with that instruction.**

## 2. Runtime model

```text
PRE-REGISTERED PERSON RECORD
 ID / callback / challenge hashes
             |
             +-----------------------+
             |                       |
             v                       v
       VerifyChallenge          StartCallback
             |                  ignore inbound contact
             |                  random token + expiry
             |                  HMAC seal entire session
             |                       |
             |                       v
             |                operator/application calls
             |                stored callback destination
             |                       |
             |                       v
             |                CompleteCallback
             |                seal / expiry / person /
             |                token / connected
             |                       |
             +-----------+-----------+
                         v
                  VERIFIED / REFUSED
```

## 3. Person and challenge model

`Person` holds `ID`, optional stored `Callback` and zero or more `Challenge` records.

`ValidatePerson` requires a non-blank ID, unique non-blank prompts and valid challenge records.

Challenges use Go 1.26 `crypto/pbkdf2` with SHA-256. Security floor: 600,000 iterations, 16-byte salt, 32-byte hash. Stored salt/hash values are standard base64. Answer comparison is constant time.

## 4. Callback planning

`StartCallback` requires a valid person, non-blank stored callback, non-zero time, TTL >0 and <=30 minutes, and a session key >=32 bytes.

It generates 24 random bytes for the session token and always copies `Person.Callback` into the session. `inboundContact` can only set `InboundContactIgnored=true`; it can never become the callback destination.

## 5. Stateless session integrity

Because the session is returned to the caller and supplied again during completion, V0.1 seals it with HMAC-SHA256.

The MAC covers length-prefixed token, person ID and callback, UTC expiry nanoseconds and the inbound-contact-ignored flag. `Seal` itself is excluded.

The session key is deployment-owned and must contain at least 32 bytes. The CLI receives it through `OOB_SESSION_KEY` as standard base64.

## 6. Callback completion order

`CompleteCallback`:

1. validates session-key and completion time inputs;
2. rejects structurally incomplete sessions;
3. recomputes and constant-time compares the HMAC seal;
4. refuses at or after expiry;
5. constant-time compares the token and checks person ID;
6. requires `connected=true`;
7. returns `Verified=true` only after all gates pass.

A changed expiry or callback therefore cannot extend/re-route a valid session without invalidating its seal.

## 7. Trust boundary

The component does not make the callback. It trusts a consuming operator/application to dial `Session.Callback` and accurately set `connected` after reaching the registered person.

The HMAC proves session state came from a holder of the deployment key; it does not prove the phone conversation itself occurred.

## 8. Shared-knowledge boundary

Challenge answers are inherently lower entropy than cryptographic credentials. PBKDF2 raises offline-guess cost but cannot make a guessable answer high entropy. Challenge verification is therefore a fallback verification signal, not a recommended sole factor for high-risk business approval.

## 9. CLI

`oobverify -input request.json` supports `challenge`, `start-callback` and `complete-callback`. Callback operations require `OOB_SESSION_KEY`.

Exit codes:

```text
0 verification/start succeeded
1 deterministic verification refusal
2 invalid operation/usage
3 malformed input/configuration or execution error
```

## 10. State and persistence

No persistent session database exists. Integrity comes from the session seal; confidentiality is not provided by the seal. Callers must protect returned session/token data appropriately.

## 11. Cryptography

- challenge KDF: PBKDF2-HMAC-SHA256 via Go `crypto/pbkdf2`;
- session token: 24 bytes from `crypto/rand`;
- session seal: HMAC-SHA256;
- comparisons: constant-time where secret/token/MAC comparison applies.

## 12. Failure model

Malformed person/challenge records, weak session key, invalid time/TTL and malformed CLI input are execution/configuration errors. Wrong challenge, unknown prompt, invalid session seal, expiry, wrong person/token or disconnected callback are deterministic refusals.

## 13. Test strategy

Named tests cover challenge success/refusal, challenge security floors/duplicate prompts, inbound-contact substitution, session field tampering, expiry/person/token/connected gates and key/TTL bounds.

CI adds formatting, vet, race tests, native/version, Linux ARM64 and named Bafflewick security-boundary acceptance.

## 14. Operational boundary

No network library, telephony API, biometric engine, deepfake detector, contact registry, secret store or durable audit store is included.

## 15. Dependencies

Go 1.26 standard library only. No third-party modules.

## 16. Known limits

- operator/application must actually perform the callback;
- HMAC provides integrity/authenticity, not confidentiality;
- deployment must securely provision/rotate the HMAC key;
- challenge answers may be guessable despite slow hashing;
- no replay ledger exists after a session has successfully completed;
- no challenge-attempt rate limiting;
- ARM64 is build proof, not runtime benchmark.

## 17. Acceptance-to-evidence map

| Requirement | Evidence |
|---|---|
| challenge behaviour | `TestChallenge` |
| challenge floors/duplicates | `TestChallengeSecurityFloorsAndValidation` |
| stored callback only | `TestCallbackIgnoresInboundContact` |
| session tamper refusal | `TestCallbackSessionIntegrity` |
| expiry/person/token/connected | `TestCallbackExpiryWrongPersonTokenAndConnection` |
| key + TTL bounds | `TestCallbackKeyAndTTLBounds` |
| format/vet/race/build/version/ARM64 | `.github/workflows/test.yml` |
