package ethclient

import (
	"errors"
	"fmt"
	"math/big"
)

type Quantity struct{ v big.Int }

func NewQuantityFromInt64(v int64) Quantity { return Quantity{v: *big.NewInt(int64(v))} }

func (s Quantity) Int64() int64 { return s.v.Int64() }

func (s Quantity) String() string {
	return fmt.Sprintf("0x%x", &s.v)
}

func (s Quantity) MarshalJSON() ([]byte, error) { return []byte(fmt.Sprintf(`"%s"`, s.String())), nil }

func (s *Quantity) UnmarshalJSON(data []byte) error {
	if len(data) == 0 {
		return errors.New("quantity should can not be empty bytes")
	}
	if len(data) < 2 || data[0] != '"' || data[len(data)-1] != '"' {
		return errors.New("quantity should be wrapped in double quotes")
	}
	// unpack string
	data = data[1 : len(data)-1]

	if len(data) < 2 {
		return errors.New("quantity should always start with 0x prefix")
	}
	// strip hex prefix
	data = data[2:]

	if len(data) == 0 {
		return errors.New("digits should always be present, zero value is 0x0")
	}

	// zero
	if len(data) == 1 && data[0] == '0' {
		s.v = *big.NewInt(0)
		return nil
	}

	// TODO: verify that no leading zeroes are present

	s.v.SetString(string(data), 16)

	return nil
}
