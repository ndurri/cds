package mail


import (
	"testing"
	"os"
	"github.com/stretchr/testify/assert"
	"crypto/md5"
	"io"
	"encoding/hex"
)

func TestParse(t *testing.T) {
	ins, err := os.Open("test.txt")
	if err != nil {
		t.Fatal(err)
	}
	msg, err := Parse(ins)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, "neil@durri.net", msg.From)
	assert.Equal(t,  "test 1", msg.Subject)
	assert.Equal(t, "exp", msg.TextContent)
	h := md5.New()
	io.WriteString(h, msg.XMLContent)
	assert.Equal(t, "e99260b4e14856c1a46af7d5bbcaa0e3", hex.EncodeToString(h.Sum(nil)))
}