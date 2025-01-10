package linep

import (
	"crypto/rand"
	"fmt"
	"math"
	"math/big"
	"os"
)

func randInt() uint64 {
	n, _ := rand.Int(rand.Reader, big.NewInt(math.MaxInt64))
	return n.Uint64()
}

func tempDirPattern(dir, pattern string) string {
	return fmt.Sprintf("%s/%s%020d",
		dir, pattern, randInt(),
	)
}

func MkdirTemp(dir, pattern string) (string, error) {
	d := tempDirPattern(dir, pattern)
	if err := os.MkdirAll(d, 0755); err != nil {
		return "", err
	}
	return d, nil
}
