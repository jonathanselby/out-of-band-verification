# Out-of-Band Verification — Intended vs Recovered Design

**Implementation evidence:** `caac210a85a4852d3ae090203e8401d514715c16`

| Area | Result | Recovered evidence |
|---|---|---|
| One verification primitive | MATCH | Challenge and callback are two paths to the same person-verification result; no business approval/workflow engine appears. |
| Stored-contact-only callback | MATCH | `Session.Callback` is copied from `Person.Callback`; inbound contact only sets an ignored marker. |
| Challenge record validation | MATCH | Person ID, unique prompt, salt/hash shape and KDF work factor are validated. |
| Standard-library PBKDF2 | MATCH | Go 1.26 `crypto/pbkdf2` replaces the draft custom loop. |
| Challenge KDF floor | MATCH | 600,000 iterations, >=16-byte salt and 32-byte hash. |
| Random short-lived callback token | MATCH | 24 bytes from `crypto/rand`; TTL >0 and <=30 minutes. |
| Caller-carried session integrity | MATCH / HARDENED | Release candidate HMAC-seals all security-relevant session fields; the draft did not protect caller-supplied session JSON. |
| Tamper refusal before completion checks | MATCH | Seal is verified before expiry/person/token/connected acceptance. |
| Exact expiry boundary | MATCH | `now.Before(expires)` is required; at expiry is refused. |
| Deployment-owned session key | MATCH | Callback operations require >=32-byte key; CLI reads `OOB_SESSION_KEY`. |
| Stable CLI/version/refusal split | MATCH | 0/1/2/3 semantics and 0.1.0 version are explicit. |
| Go 1.26 / ARM64 | MATCH | Module baseline and CI cross-build are explicit. |
| Telephone call placement | DELIBERATE GAP | Component returns stored callback destination but has no telephony integration. |
| Voice identity/deepfake detection | DELIBERATE GAP | Explicitly absent. |
| Contact enrolment/admin | DELIBERATE GAP | Person records are pre-registered caller inputs. |
| One-time session consumption | DELIBERATE GAP | Stateless component has no consumed-session ledger. |
| Challenge-attempt throttling | DELIBERATE GAP | Requires external state. |
| Session confidentiality | DELIBERATE GAP | HMAC authenticates session state but does not encrypt it. |
| Operator truthfulness about `connected` | UNKNOWN | Cannot be established by source code without telephony/identity integration. |
| Deployment key custody/rotation | UNKNOWN | Secret management is outside this repository. |

## Material correction made before publication

The initial draft returned callback sessions to the caller and later trusted the same caller-supplied session object. That made fabricated or edited sessions possible. The V0.1 release candidate seals token, person, callback, expiry and inbound-contact state with HMAC-SHA256 and rejects altered sessions before the remaining completion gates.

The draft also carried a hand-written PBKDF2 implementation. V0.1 uses Go 1.26 `crypto/pbkdf2`, validates record shape and raises the work-factor floor to 600,000.

These harden the original job without broadening it.

## Conclusion

**MATCH for the hardened V0.1 design.** The implementation preserves the intended out-of-band boundary and removes the draft's most important stateless-session trust flaw.
