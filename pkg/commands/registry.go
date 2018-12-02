package commands

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
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

func (r CommandRegistry) Commands() []string {
	out := make([]string, 0, len(r))
	for k := range r {
		out = append(out, k)
	}
	return out
}

func (r CommandRegistry) Save(out io.Writer) error {
	obj := make(map[string]string, len(r))

	for k, v := range r {
		obj[k] = v.String()
	}

	enc := json.NewEncoder(out)
	return enc.Encode(obj)
}

func (r CommandRegistry) SaveToFile(filename string) error {
	fd, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer fd.Close()

	return r.Save(fd)
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

func (r CommandRegistry) LoadFromFile(filename string) error {
	fd, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer fd.Close()
	return r.Load(fd)
}
