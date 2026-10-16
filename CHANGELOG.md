# Changelog

## 0.1.0 — 2026-10-17

- initial out-of-band person-verification component;
- validated pre-registered challenge records with Go 1.26 standard-library PBKDF2-HMAC-SHA256 and 600,000-iteration floor;
- callback planning always uses stored contact and records inbound contact as ignored;
- random 192-bit callback token and maximum 30-minute TTL;
- HMAC-SHA256 sealed callback sessions using deployment-owned key;
- refusal of fabricated/modified callback session fields before person/token/connection checks;
- explicit operator-call and shared-knowledge limitations;
- PRD, EDD, architecture, skill, CI and code-derived evidence;
- race tests and Linux ARM64 build proof.
