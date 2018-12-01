package commands

import (
	"bytes"
	. "github.com/onsi/gomega"
	"testing"
)

func TestRegistry(t *testing.T) {
	g := NewGomegaWithT(t)

	reg := CommandRegistry{}

	g.Expect(reg.AddCommand("foo", IRCommand{1,2,3})).To(Succeed())
	g.Expect(reg.AddCommand("bar", IRCommand{16,17,42,12})).To(Succeed())
	// Already exist
	g.Expect(reg.AddCommand("foo", IRCommand{2,3})).NotTo(Succeed())

	buf := bytes.Buffer{}
	g.Expect(reg.Save(&buf)).To(Succeed())

	reg2 := CommandRegistry{}
	g.Expect(reg2.Load(&buf)).To(Succeed())
	g.Expect(reg2).To(Equal(reg))
}
