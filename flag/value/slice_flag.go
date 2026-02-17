package value

import (
	"flag"
	"fmt"
)

type SliceFlag []string

// Set implements [flag.Value].
func (s *SliceFlag) Set(str string) error {
	*s = append(*s, str)
	return nil
}

// String implements [flag.Value].
func (s *SliceFlag) String() string {
	return fmt.Sprintf("%v", *s)
}

var _ flag.Value = (*SliceFlag)(nil)
