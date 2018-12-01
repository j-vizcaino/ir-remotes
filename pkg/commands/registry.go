package commands

import (
	"encoding/json"
	"fmt"
	"io"
)

type CommandRegistry map[string]IRCommand

func (r CommandRegistry) AddCommand(name string, irCode []byte) error {
	_, ok := r[name]
	if ok {
		return fmt.Errorf("command %s already exist in registry", name)
	}

	cmd := make([]byte, len(irCode))
	copy(cmd, irCode)
	r[name] = cmd
	return nil
}

func (r CommandRegistry) Save(out io.Writer) error {
	obj := make(map[string]string, len(r))

	for k, v := range r {
		obj[k] = v.String()
	}

	enc := json.NewEncoder(out)
	return enc.Encode(obj)
}

func (r CommandRegistry) Load(in io.Reader) error {
	dec := json.NewDecoder(in)
	obj := make(map[string]string)

	if err := dec.Decode(&obj); err != nil {
		return err
	}

	for name, strCode := range obj {
		irCode, err := IRComandFromString(strCode)
		if err != nil {
			return err
		}
		r[name] = irCode
	}
	return nil
}