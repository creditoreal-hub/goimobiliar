package types

import "strconv"

type Float64 float64

func (f Float64) MarshalJSON() ([]byte, error) {
	s := strconv.FormatFloat(float64(f), 'f', 2, 64)
	return []byte(s), nil
}
