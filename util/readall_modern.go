// +build !go1.12 && !go1.13 && !go1.14 && !go1.15

package util

import (
	"io"
)

func ReadAll(r io.Reader) ([]byte, error) {
	return io.ReadAll(r)
}
