package mail


import (
	"testing"
	"github.com/stretchr/testify/assert"
)

const testEmail =
`From: <neil@durri.net>
To: <beta@cdsdec.uk>
Subject: test 1
Content-Type: text/plain

exp
`

func TestParse(t *testing.T) {
	msg, err := Parse(testEmail)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, "neil@durri.net", msg.From)
	assert.Equal(t,  "test 1", msg.Subject)
	assert.Equal(t, "exp", msg.TextContent)
}