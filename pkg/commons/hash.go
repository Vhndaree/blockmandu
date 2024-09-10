package common

import "crypto/sha256"

const addressChecksumLen = 4

func HashPubKey(pubKey []byte) []byte {
	publicSHA256 := sha256.Sum256(pubKey)
	secondHasher := sha256.Sum256(publicSHA256[:])

	return secondHasher[:]
}

func Checksum(payload []byte) []byte {
	firstSHA := sha256.Sum256(payload)
	secondSHA := sha256.Sum256(firstSHA[:])

	return secondSHA[:addressChecksumLen]
}
