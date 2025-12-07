package ulid

import (
	"crypto/rand"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/oklog/ulid/v2"
)

const Length = ulid.EncodedSize

var (
	once     sync.Once
	instance *generator
)

type generator struct {
	timestamp uint64
	entropy   *ulid.MonotonicEntropy
}

func New() string {
	u := initGenerator()

	return strings.ToLower(ulid.MustNew(u.timestamp, u.entropy).String())
}

func Verify(s string) error {
	_, err := ulid.Parse(s)
	if err != nil {
		return fmt.Errorf("id %s is not allowed, must satisfy the ulid format", s)
	}
	return nil
}

func initGenerator() *generator {
	once.Do(func() {
		instance = &generator{
			timestamp: ulid.Timestamp(time.Now()),
			entropy:   ulid.Monotonic(rand.Reader, 0),
		}
	})

	return instance
}
