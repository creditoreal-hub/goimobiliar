package types

import "encoding/json"

type CpfCnpj string

func (c *CpfCnpj) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil
	}

	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		*c = CpfCnpj(s)
		return nil
	}

	var n json.Number
	if err := json.Unmarshal(data, &n); err != nil {
		return err
	}

	*c = CpfCnpj(n.String())
	return nil
}
