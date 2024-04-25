package xmlfactory

import (
	"encoding/xml"
	"reflect"
)

var typeRegistry = make(map[string]reflect.Type)

type Envelope struct {
	Content interface{}
	ContentType string
}

func ptr(pType reflect.Type, value reflect.Value) interface{} {
	pv := reflect.New(pType).Elem()
	pv.Set(value)
	return pv.Interface()
}

func (e *Envelope) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	v := reflect.New(typeRegistry[start.Name.Local].Elem()).Elem()
	pv := ptr(typeRegistry[start.Name.Local], v.Addr())
	if err := d.DecodeElement(pv, &start); err != nil {
		return err
	}
	e.Content = pv
	e.ContentType = reflect.TypeOf(pv).String()
	return nil
}

func Register(xmlTag string, handler interface{}) {
	typeRegistry[xmlTag] = reflect.TypeOf(handler)
}