# Out-of-Band Verification

[![test](https://github.com/jonathanselby/out-of-band-verification/actions/workflows/test.yml/badge.svg)](https://github.com/jonathanselby/out-of-band-verification/actions/workflows/test.yml)

**Never verify an instruction using contact details supplied with that instruction.**

`out-of-band-verification` is a small Go component for confirming a claimed person using pre-registered verification material. It supports a slow-hashed shared-knowledge challenge and a callback plan that always uses the stored contact record rather than contact data supplied with the inbound instruction.

V0.1 is independently usable; CRAB is one consumer, not part of the component contract.

## What V0.1 proves

- challenge records are validated and PBKDF2-HMAC-SHA256 uses a 600,000-iteration floor;
- duplicate/blank challenge prompts and malformed records are rejected;
- callback destination always comes from the pre-registered person record;
- inbound contact details are ignored and explicitly recorded as ignored;
- callback sessions contain a 192-bit random token and a bounded expiry;
- the complete callback session is sealed with HMAC-SHA256 using a deployment-owned key;
- fabricated or modified person, callback, expiry, token or inbound-contact state is refused;
- callback completion checks the seal, expiry, person, token and operator-recorded connection result.

## Install

Requires Go 1.26+.

```bash
go install github.com/jonathanselby/out-of-band-verification/cmd/oobverify@main
oobverify -version
```

Expected:

```text
oobverify 0.1.0
```

## Callback session key

Callback operations require `OOB_SESSION_KEY`, standard-base64 encoded and at least 32 decoded bytes.

```bash
export OOB_SESSION_KEY='<deployment-owned-base64-key>'
```

The key is not stored by this repository. Rotation and secret custody belong to the deployment.

## CLI

```bash
oobverify -input request.json
```

Operations are `challenge`, `start-callback` and `complete-callback`.

## Exit codes

| Code | Meaning |
|---:|---|
| `0` | verification/start succeeded |
| `1` | deterministic verification refusal |
| `2` | invalid CLI operation/usage |
| `3` | malformed input/configuration or execution error |

## Security boundary

The callback component **does not place a call**. It returns the pre-registered callback destination and later accepts an operator/application assertion that the call connected to the registered person. It cannot independently prove voice identity, detect deepfakes or prove that an operator actually dialled the number.

Shared-knowledge challenge answers are low-entropy by nature. PBKDF2 protects stored answers against inexpensive offline guessing, but challenge verification should be treated as a weaker fallback signal rather than a substitute for a strong authentication factor for high-risk actions.

The component does not approve the underlying transaction. It only returns a person-verification result.

## Repository map

```text
out-of-band-verification/
├── .github/workflows/test.yml
├── PRD.md
├── EDD.md
├── README.md
├── SKILL.md
├── RELEASE.md
├── VERSION
├── CHANGELOG.md
├── LICENSE
├── architecture.txt
├── go.mod
├── cmd/oobverify/main.go
├── internal/oob/
├── examples/bafflewick/
└── generated/
    ├── EDD.md
    └── COMPARISON.md
```

## Design rule

> **Never verify an instruction using contact details supplied with that instruction.**

## Licence

MIT. See [`LICENSE`](LICENSE).
