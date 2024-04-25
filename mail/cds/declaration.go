package cds

import (
	"encoding/xml"
	"strings"
)

type MetaData struct {
	XMLName        xml.Name `xml:"md:MetaData"`
	Xmlns          string   `xml:"xmlns,attr"`
	Md             string   `xml:"xmlns:md,attr"`
	Clm63055       string   `xml:"xmlns:clm63055,attr"`
	DS             string   `xml:"xmlns:ds,attr"`
	Xsi            string   `xml:"xmlns:xsi,attr"`
	SchemaLocation string   `xml:"xsi:schemaLocation,attr"`
	DMVersion      string   `xml:"md:WCODataModelVersionCode"`
	TypeName       string   `xml:"md:WCOTypeName"`
	CountryCode    string   `xml:"md:ResponsibleCountryCode"`
	AgencyName     string   `xml:"md:ResponsibleAgencyName"`
	AgencyVersion  string   `xml:"md:AgencyAssignedCustomizationVersionCode"`
	Declaration    *Declaration
}

type Declaration struct {
	P1                    string `xml:"xmlns:p1,attr"`
	Udt                   string `xml:"xmlns:udt,attr"`
	Clm5ISO42173A         string `xml:"xmlns:clm5ISO42173A,attr"`
	SchemaLocation        string `xml:"xsi:schemaLocation,attr"`
	FunctionCode          int
	FunctionalReferenceID string
	TypeCode              string
	GoodsItemQuantity     int
	TotalPackageQuantity  int
	BorderTransportMeans  TransportMeans `xml:",omitempty"`
	Consignment           Consignment    `xml:",omitempty"`
	DeclarantId           string         `xml:"Declarant>ID"`
	ExitOffice            string         `xml:"ExitOffice>ID"`
	Exporter              string         `xml:"Exporter>ID"`
	GoodsShipment         GoodsShipment  `xml:",omitempty"`
}

type TransportMeans struct {
	ID                          string
	IdentificationTypeCode      int
	RegistrationNationalityCode string `xml:",omitempty"`
	ModeCode                    int    `xml:",omitempty"`
}

type Route []Itinerary

type Consignment struct {
	CarrierId            string `xml:"Carrier>ID"`
	FreightPaymentMethod string `xml:"Freight>PaymentMethodCode"`
	Itinerary            Route  `xml:"Itinerary"`
}

type Itinerary struct {
	SequenceNumeric    int
	RoutingCountryCode string
}

type GoodsItems []GovernmentAgencyGoodsItem

type GoodsShipment struct {
	Consignee                  Consignee
	Consignment                GoodsConsignment
	Destination                string             `xml:"Destination>CountryCode"`
	ExportCountry              string             `xml:"ExportCountry>ID"`
	GovernmentAgencyGoodsItems GoodsItems         `xml:"GovernmentAgencyGoodsItem"`
	PreviousDocuments          []PreviousDocument `xml:"PreviousDocument"`
}

type Consignee struct {
	Name    string
	Address Address
}

type Address struct {
	CityName    string
	CountryCode string
	Line        string
	PostcodeID  string
}

type TransportEquipments []TransportEquipment

type GoodsConsignment struct {
	ContainerCode           int
	DepartureTransportMeans TransportMeans
	GoodsLocation           GoodsLocation
	TransportEquipment      TransportEquipments
}

type GoodsLocation struct {
	Name               string
	TypeCode           string
	AddressTypeCode    string `xml:"Address>TypeCode"`
	AddressCountryCode string `xml:"Address>CountryCode"`
}

type Seals []Seal

type TransportEquipment struct {
	SequenceNumeric int
	ID              string
	Seal            Seals
}

type Seal struct {
	SequenceNumeric int
	ID              string
}

type Packaging []Package

type GovernmentAgencyGoodsItem struct {
	SequenceNumeric        int
	StatisticalValueAmount StatisticalValueAmount
	TransactionNatureCode  int
	AdditionalInformation  AdditionalInformation
	Commodity              Commodity
	GovernmentProcedure    []GovernmentProcedure
	Packaging              Packaging
}

type StatisticalValueAmount struct {
	CurrencyID string  `xml:"currencyID,attr"`
	Amount     float64 `xml:",chardata"`
}

type AdditionalInformation struct {
	StatementCode        string
	StatementDescription string
}

type Commodity struct {
	Description    string
	Classification []Classification
	GoodsMeasure   GoodsMeasure
}

type Classification struct {
	ID                     string
	IdentificationTypeCode string
}

type GoodsMeasure struct {
	GrossMassMeasure int
	NetWeightMeasure int `xml:"NetNetWeightMeasure"`
}

type GovernmentProcedure struct {
	CurrentCode  string
	PreviousCode string `xml:",omitempty"`
}

type Package struct {
	SequenceNumeric int
	MarksNumbersID  string
	Quantity        int `xml:"QuantityQuantity"`
	TypeCode        string
}

type PreviousDocument struct {
	CategoryCode string
	ID           string
	TypeCode     string
	Line         *int `xml:"LineNumeric,omitempty"`
}

func (g GoodsItems) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	for index, gi := range g {
		gi.SequenceNumeric = index + 1
		if err := e.EncodeElement(gi, start); err != nil {
			return err
		}
	}
	return nil
}

func (r Route) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	for index, it := range r {
		it.SequenceNumeric = index + 1
		if err := e.EncodeElement(it, start); err != nil {
			return err
		}
	}
	return nil
}

func (tes TransportEquipments) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	for index, te := range tes {
		te.SequenceNumeric = index + 1
		if err := e.EncodeElement(te, start); err != nil {
			return err
		}
	}
	return nil
}

func (sls Seals) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	for index, s := range sls {
		s.SequenceNumeric = index + 1
		if err := e.EncodeElement(s, start); err != nil {
			return err
		}
	}
	return nil
}

func (pkg Packaging) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	for index, pk := range pkg {
		pk.SequenceNumeric = index + 1
		if err := e.EncodeElement(pk, start); err != nil {
			return err
		}
	}
	return nil
}

func WrapMetaData(dec *Declaration) *MetaData {
	var md = MetaData{
		Xmlns:          "urn:wco:datamodel:WCO:DEC-DMS:2",
		Md:             "urn:wco:datamodel:WCO:DocumentMetaData-DMS:2",
		Clm63055:       "urn:un:unece:uncefact:codelist:standard:UNECE:AgencyIdentificationCode:D12B",
		DS:             "urn:wco:datamodel:WCO:MetaData_DS-DMS:2",
		Xsi:            "http://www.w3.org/2001/XMLSchema-instance",
		SchemaLocation: "urn:wco:datamodel:WCO:DocumentMetaData-DMS:2 ../DocumentMetaData_2_DMS.xsd ",
		DMVersion:      "3.6",
		TypeName:       "DEC",
		CountryCode:    "GB",
		AgencyName:     "HMRC",
		AgencyVersion:  "v2.1",
		Declaration:    dec,
	}
	return &md
}

func MakePrelodgedExport() *Declaration {
	var exp = Declaration{
		P1:             "urn:wco:datamodel:WCO:Declaration_DS:DMS:2",
		Udt:            "urn:un:unece:uncefact:data:standard:UnqualifiedDataType:6",
		Clm5ISO42173A:  "urn:un:unece:uncefact:codelist:standard:ISO:ISO3AlphaCurrencyCode:2012-08-31",
		SchemaLocation: "urn:wco:datamodel:WCO:DEC-DMS:2 ../WCO_DEC_2_DMS.xsd ",
		FunctionCode:   9,
		TypeCode:       "EXD",
	}
	return &exp
}

func (d *Declaration) SetDUCR(ducr string) {
	line := 1
	doc := PreviousDocument{
		CategoryCode: "Z",
		ID:           strings.ToUpper(ducr),
		TypeCode:     "DCR",
		Line:         &line,
	}
	d.GoodsShipment.PreviousDocuments = append(d.GoodsShipment.PreviousDocuments, doc)
}

func (d *Declaration) SetMUCR(mucr string) {
	doc := PreviousDocument{
		CategoryCode: "Z",
		ID:           strings.ToUpper(mucr),
		TypeCode:     "MCR",
	}
	d.GoodsShipment.PreviousDocuments = append(d.GoodsShipment.PreviousDocuments, doc)
}

func (d *Declaration) SetGoodsDescription(description string) {
	if len(d.GoodsShipment.GovernmentAgencyGoodsItems) < 1 {
		d.GoodsShipment.GovernmentAgencyGoodsItems = append(d.GoodsShipment.GovernmentAgencyGoodsItems, GovernmentAgencyGoodsItem{})
	}
	d.GoodsShipment.GovernmentAgencyGoodsItems[0].Commodity.Description = description
}

func (d *Declaration) SetPackagingType(pktype string) {
	if len(d.GoodsShipment.GovernmentAgencyGoodsItems) < 1 {
		d.GoodsShipment.GovernmentAgencyGoodsItems = append(d.GoodsShipment.GovernmentAgencyGoodsItems, GovernmentAgencyGoodsItem{})
	}
	d.GoodsShipment.GovernmentAgencyGoodsItems[0].Packaging[0].TypeCode = strings.ToUpper(pktype)
}

func (d *Declaration) UpdateComputedTotals() {
	d.UpdateGoodsItemQuantity()
	d.UpdateTotalPackageQuantity()
}

func (d *Declaration) UpdateGoodsItemQuantity() {
	d.GoodsItemQuantity = len(d.GoodsShipment.GovernmentAgencyGoodsItems)
}

func (d *Declaration) UpdateTotalPackageQuantity() {
	ptot := 0
	for _, item := range d.GoodsShipment.GovernmentAgencyGoodsItems {
		for _, pack := range item.Packaging {
			ptot += pack.Quantity
		}
	}
	d.TotalPackageQuantity = ptot
}
