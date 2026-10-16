# Release checklist

## Source/version
- [ ] `VERSION`, CLI version and release tag agree.
- [ ] Go 1.26 baseline and changelog are current.

## Security gate
- [ ] challenge security floors and record validation are unchanged or deliberately revised.
- [ ] inbound contact cannot become callback destination.
- [ ] session integrity covers token, person, callback, expiry and inbound-contact marker.
- [ ] seal verification occurs before expiry/person/token/connection acceptance.
- [ ] callback session key is deployment-owned and >=32 bytes.
- [ ] operator-assertion and shared-knowledge limitations remain explicit.

## Engineering gate
- [ ] `gofmt` clean.
- [ ] `go vet ./...` passes.
- [ ] `go test -race -count=1 ./...` passes.
- [ ] native/version and Linux ARM64 builds pass.
- [ ] Bafflewick security-boundary acceptance passes.

## Evidence
- [ ] all PRD criteria map to executable evidence.
- [ ] generated EDD is pinned to exact source commit.
- [ ] generated comparison has no unexplained mismatch.
