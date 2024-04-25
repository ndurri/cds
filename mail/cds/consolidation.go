package cds

import (
	"encoding/xml"
)

type Consolidation struct {
	XMLName     xml.Name `xml:"inv:inventoryLinkingConsolidationRequest"`
	Xmlns       string   `xml:"xmlns:inv,attr"`
	MessageCode string   `xml:"inv:messageCode"`
	MasterUCR   string   `xml:"inv:masterUCR,omitempty"`
	ChildUCR    *UCR     `xml:"inv:ucrBlock,omitempty"`
}

func MakeShut() *Consolidation {
	return &Consolidation{
		Xmlns:       "http://gov.uk/customs/inventoryLinking/v1",
		MessageCode: "CST",
	}
}

func MakeConsolidation() *Consolidation {
	return &Consolidation{
		Xmlns:       "http://gov.uk/customs/inventoryLinking/v1",
		MessageCode: "EAC",
	}
}

func (c *Consolidation) SetChild(ucr string, ucrtype UCRType) {
	c.ChildUCR = &UCR{
		UCR:     ucr,
		UCRType: ucrtype,
	}
}
