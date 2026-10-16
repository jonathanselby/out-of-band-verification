# Bafflewick acceptance example

Bafflewick is fictional. The package tests generate no live contact or secret material.

```bash
go test ./internal/oob -count=1 -v
```

The acceptance suite demonstrates that inbound contact cannot replace the stored callback and that a callback session cannot be modified to change its person, callback, expiry, token or inbound-contact marker without invalidating its HMAC seal.

The component does not place a call. A consuming operator/application must use the callback destination returned by `StartCallback` and accurately record whether it connected to the registered person.
