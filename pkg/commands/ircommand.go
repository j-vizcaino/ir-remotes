package commands

import (
	"fmt"
)

type IRCommand []byte

func (i IRCommand) String() string {
	return fmt.Sprintf("%x", []byte(i))
}

func IRComandFromString(s string) (IRCommand, error) {
	var out []byte
	matched, err := fmt.Sscanf(s, "%x", &out)
	if err != nil {
		return nil, err
	}

	if matched != 1 {
		return nil, fmt.Errorf("invalid string format")
	}
	return out, nil
}
