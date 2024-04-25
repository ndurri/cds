package cds

import (
	"encoding/xml"
)

type Query struct {
	XMLName            xml.Name `xml:"inv:inventoryLinkingQueryRequest"`
	XmlNS              string   `xml:"xmlns:inv,attr"`
	UCR                UCR      `xml:"inv:queryUCR"`
}

func MakeQuery() *Query {
	return &Query{
		XmlNS:       "http://gov.uk/customs/inventoryLinking/v1",
	}
}

func (m *Query) SetDUCR(ducr string) {
	m.UCR.UCR = ducr
	m.UCR.UCRType = DUCR
}

func (m *Query) SetMUCR(mucr string) {
	m.UCR.UCR = mucr
	m.UCR.UCRType = MUCR
}