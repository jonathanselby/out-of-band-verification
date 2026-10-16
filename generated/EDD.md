# Out-of-Band Verification — Recovered Engineering Design

**Evidence commit:** `caac210a85a4852d3ae090203e8401d514715c16`  
**Recovery method:** source, tests, CLI and CI inspection  
**Classification:** KNOWN / INFERRED / UNKNOWN

## 1. Observed purpose

**KNOWN.** The repository verifies a claimed person using pre-registered challenge material or a callback plan whose destination comes from the stored `Person` record.

**KNOWN.** It does not place calls, authenticate accounts, persist contacts/sessions or approve the underlying business action.

## 2. Person/challenge model

**KNOWN.** `Person` contains ID, optional callback and challenge records. `ValidatePerson` requires a non-blank ID, unique non-blank challenge prompts and valid challenge records.

**KNOWN.** Challenge records use standard-base64 salt/hash, PBKDF2-HMAC-SHA256 through Go `crypto/pbkdf2`, minimum 600,000 iterations, minimum 16-byte salt and 32-byte hash.

**KNOWN.** `VerifyChallenge` constant-time compares derived and stored hashes and returns deterministic refusal for wrong or unregistered challenges.

**UNKNOWN.** Challenge-answer entropy is not measurable from the record; the code cannot make a guessable answer high entropy.

## 3. Callback-start model

**KNOWN.** `StartCallback` validates the person, requires a non-blank pre-registered callback, non-zero time, TTL >0 and <=30 minutes, and a session key >=32 bytes.

**KNOWN.** It generates 24 random bytes using `crypto/rand` and base64url-encodes them as the session token.

**KNOWN.** `Session.Callback` is copied only from `Person.Callback`. `inboundContact` is never used as the destination; it only controls `InboundContactIgnored`.

## 4. Session integrity

**KNOWN.** A callback session contains token, person ID, callback, expiry, inbound-contact-ignored flag and HMAC seal.

**KNOWN.** `sealSession` uses HMAC-SHA256 with the deployment key over length-prefixed token/person/callback, UTC expiry Unix nanoseconds and the boolean marker.

**KNOWN.** The seal is base64url encoded and does not include itself.

**INFERRED.** The session can safely be caller-carried for integrity/authenticity as long as the deployment HMAC key remains secret.

**UNKNOWN.** The HMAC provides no confidentiality; token/session exposure protection is a deployment/application concern.

## 5. Completion order

**KNOWN.** `CompleteCallback` checks key length and non-zero completion time, then structural session fields, then session HMAC, then expiry, then person/token, then `connected`.

**KNOWN.** Completion at exactly `Expires` is refused because the code requires `now.Before(s.Expires)`.

**KNOWN.** Token and seal comparisons use constant-time comparison.

## 6. Mutation resistance

**KNOWN.** Tests mutate person ID, callback, expiry, token and the inbound-contact marker and require `callback session integrity rejected` for each.

**KNOWN.** A caller therefore cannot extend a session or reroute the callback by editing the returned JSON without invalidating the seal.

## 7. Operator trust boundary

**KNOWN.** `connected` is a caller/operator-supplied boolean. There is no telephony or identity-recognition integration.

**INFERRED.** Verification proves that a sealed callback workflow was completed according to the supplied operator assertion; it does not cryptographically prove a human voice or phone call.

## 8. CLI and secret boundary

**KNOWN.** `oobverify` supports `challenge`, `start-callback` and `complete-callback` operations.

**KNOWN.** Callback operations load `OOB_SESSION_KEY` from the environment, require standard base64 and at least 32 decoded bytes. The key is not persisted by the repository.

**KNOWN.** Exit semantics are: 0 success, 1 deterministic verification refusal, 2 invalid operation/usage, 3 malformed input/configuration/execution error.

## 9. State and dependencies

**KNOWN.** `go.mod` declares Go 1.26 and no third-party modules.

**KNOWN.** The core has no database, network client, global mutable session store or background task.

## 10. Failure model

**KNOWN.** Invalid person/challenge records, weak session key, bad callback time/TTL and malformed CLI input are errors.

**KNOWN.** Wrong answer, unknown challenge, invalid session integrity, expiry, wrong person/token and disconnected callback are structured refusals.

## 11. Test evidence

**KNOWN.** `TestChallenge` covers valid/wrong/unknown challenge results.

**KNOWN.** `TestChallengeSecurityFloorsAndValidation` covers salt, work factor and duplicate prompt rejection.

**KNOWN.** `TestCallbackIgnoresInboundContact` proves callback routing uses the stored record.

**KNOWN.** `TestCallbackSessionIntegrity` covers five session-field mutations.

**KNOWN.** `TestCallbackExpiryWrongPersonTokenAndConnection` covers expiry, person, token and connected gates.

**KNOWN.** `TestCallbackKeyAndTTLBounds` covers key length and maximum TTL.

## 12. CI evidence

**KNOWN.** CI requires clean formatting, vet, race tests, native build, exact version output, Linux ARM64 cross-build and the named security-boundary suite.

## 13. Security properties

**KNOWN.** Inbound contact cannot influence the callback destination. Caller-carried session state is authenticated before use. Challenge hashes use a current slow-KDF floor rather than a custom KDF implementation.

**UNKNOWN.** Correct callback execution, voice identity, deepfake resistance, secret-key storage and challenge-attempt throttling cannot be proven by this component.

## 14. Replay/state limit

**KNOWN.** There is no durable “session consumed” ledger. A valid unexpired session/token can therefore be presented more than once if the surrounding application permits it.

**INFERRED.** A consuming workflow that requires one-time completion must persist/mark consumption externally.

## 15. Portability

**KNOWN.** CI cross-builds Linux ARM64.

**INFERRED.** This is a portability proof, not a constrained-hardware performance benchmark.

## 16. Observed non-goals

**KNOWN.** No telephony, messaging, biometrics, deepfake detection, contact enrolment/administration, authentication session, trust scoring or underlying business approval is present.

## 17. Evidence conclusion

**KNOWN.** Source commit `caac210a85a4852d3ae090203e8401d514715c16` implements a small out-of-band verification primitive whose two strongest controls are stored-contact-only callback selection and HMAC-authenticated callback-session state.

**INFERRED.** The release remains useful precisely because it stops at verification primitives and leaves real-world calling, secret custody and business approval to the consumer.
