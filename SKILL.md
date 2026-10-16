# Out-of-Band Verification — Skill

## Use this when

Use this component to verify a claimed person using verification material registered before the inbound instruction: a stored challenge or callback contact.

## Core rule

> **Never verify an instruction using contact details supplied with that instruction.**

## Change discipline

1. Keep inbound contact data out of callback selection.
2. Validate person/challenge records before verification.
3. Use Go standard-library PBKDF2, not a custom KDF loop.
4. Never accept caller-returned callback session state without verifying its seal first.
5. Keep session key custody outside this repository.
6. Treat `connected` as an operator/application assertion, not cryptographic proof of voice identity.
7. Do not add telephony, biometrics, deepfake detection, contact administration or business approval here.
8. Keep deterministic refusals separate from malformed/configuration errors.
9. Run format, vet, race, native and ARM64 gates after changes.
10. Regenerate code-derived evidence before release.
