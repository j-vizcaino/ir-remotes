package remotes

import (
	"encoding/json"
	. "github.com/onsi/gomega"
	"testing"
)

func TestIRComand(t *testing.T) {
	g := NewGomegaWithT(t)

	ir := IRCommand{1, 10, 16, 32, 42}
	raw, err := json.Marshal(&ir)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(string(raw)).To(Equal(`"010a10202a"`))

	ir2 := IRCommand{}
	err = json.Unmarshal(raw, &ir2)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(ir2).To(Equal(ir))

	err = json.Unmarshal([]byte("12320"), &ir2)
	g.Expect(err).To(HaveOccurred())
}
