package commands

import (
	. "github.com/onsi/gomega"
	"testing"
)

func TestIRComand(t *testing.T) {
	g := NewGomegaWithT(t)

	ir := IRCommand{1,10,16,32,42}
	s := ir.String()
	g.Expect(s).To(Equal("010a10202a"))

	ir2, err := IRComandFromString(s)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(ir2).To(Equal(ir))

	_, err = IRComandFromString("12320")
	g.Expect(err).To(HaveOccurred())
}