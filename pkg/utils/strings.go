package utils

import (
	crand "crypto/rand"
)

func RandStringN(n int) string {
	return crand.Text()[:n]
}
