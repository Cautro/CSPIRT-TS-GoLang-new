package entity

import (
	"crypto/sha256"
	"encoding/hex"
)

// BlacklistKeyPrefix namespaces revoked-access-token entries in the cache.
const BlacklistKeyPrefix = "blacklist:token:"

// BlacklistTokenKey derives the cache key for accessToken. The raw JWT is
// hashed so it is never stored in the cache in plaintext. Shared by the auth
// usecase (which writes the entry on logout) and AuthMiddleware (which reads
// it on every request), so both sides agree on the same key without either
// layer depending on the other.
func BlacklistTokenKey(accessToken string) string {
	sum := sha256.Sum256([]byte(accessToken))
	return BlacklistKeyPrefix + hex.EncodeToString(sum[:])
}
