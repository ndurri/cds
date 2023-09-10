package cds

import (
	"encoding/xml"
	"time"
	"errors"
	"regexp"
	"strings"
)

type Movement struct {
	XMLName            xml.Name `xml:"inv:inventoryLinkingMovementRequest"`
	XmlNS              string   `xml:"xmlns:inv,attr"`
	MessageCode        string   `xml:"inv:messageCode"`
	UCR                UCR      `xml:"inv:ucrBlock"`
	GoodsLocation      string   `xml:"inv:goodsLocation"`
	GoodsArrivalTime   string   `xml:"inv:goodsArrivalDateTime,omitempty"`
	GoodsDepartureTime string   `xml:"inv:goodsDepartureDateTime,omitempty"`
	MasterOpt      	   string   `xml:"inv:masterOpt,omitempty"`
	MovementReference  string   `xml:"inv:movementReference,omitempty"`
}

type UCR struct {
	UCR     string  `xml:"inv:ucr"`
	UCRType UCRType `xml:"inv:ucrType"`
}

type UCRType string

const (
	DUCR UCRType = "D"
	MUCR         = "M"
)

var (
	DUCRRegEx    = regexp.MustCompile(`^[0-9][A-Z][A-Z][0-9A-Z\(\)\-/]{6,32}$`)
	MUCRARegEx   = regexp.MustCompile(`^A:[0-9A-Z]{3}[0-9]{8}$`)
	MUCRCRegEx   = regexp.MustCompile(`^C:[A-Z]{3}[0-9A-Z]{3,30}$`)
	MUCRGB1RegEx = regexp.MustCompile(`^GB/[0-9A-Z]{3,4}-[0-9A-Z]{5,28}$`)
	MUCRGB2RegEx = regexp.MustCompile(`^GB/[0-9A-Z]{9,12}-[0-9A-Z]{1,23}$`)
)

var Undefined = errors.New("Unable to detect UCR type")

func MakeAnticipatedArrival() *Movement {
	return &Movement{
		XmlNS:       "http://gov.uk/customs/inventoryLinking/v1",
		MessageCode: "EAA",
		MasterOpt: "R",
	}
}

func MakeArrival() *Movement {
	return &Movement{
		XmlNS:       "http://gov.uk/customs/inventoryLinking/v1",
		MessageCode: "EAL",
		MasterOpt: "R",
	}
}

func MakeDeparture() *Movement {
	return &Movement{
		XmlNS:       "http://gov.uk/customs/inventoryLinking/v1",
		MessageCode: "EDL",
	}
}

func (m *Movement) SetArrivalDate(date time.Time) {
	m.GoodsArrivalTime = date.UTC().Format(time.RFC3339)
}

func (m *Movement) SetDepartureDate(date time.Time) {
	m.GoodsDepartureTime = date.UTC().Format(time.RFC3339)
}

func (m *Movement) SetDUCR(ducr string) {
	m.UCR.UCR = ducr
	m.UCR.UCRType = DUCR
}

func (m *Movement) SetMUCR(mucr string) {
	m.UCR.UCR = mucr
	m.UCR.UCRType = MUCR
}

func IsDUCR(arg string) bool {
	if len(arg) < 16 {
		return false
	}
	arg = strings.ToUpper(arg)
	return DUCRRegEx.MatchString(arg) && arg[15:16] == "-"
}

func IsMUCR(arg string) bool {
	if len(arg) < 2 {
		return false
	}
	arg = strings.ToUpper(arg)
	switch arg[0:2] {
		case "A:": return MUCRARegEx.MatchString(arg)
		case "C:": return MUCRCRegEx.MatchString(arg)
		case "GB": return MUCRGB1RegEx.MatchString(arg) || MUCRGB2RegEx.MatchString(arg)
		default: return false
	}
}

func GetUCRType(ucr string) (string, error) {
	if IsDUCR(ucr) {
		return "D", nil
	} else if IsMUCR(ucr) {
		return "M", nil
	} else {
		return "", Undefined
	}
}
