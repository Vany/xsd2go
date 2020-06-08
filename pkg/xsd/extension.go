package xsd

import (
	"encoding/xml"
)

type Extension struct {
	XMLName          xml.Name    `xml:"http://www.w3.org/2001/XMLSchema extension"`
	Base             reference   `xml:"base,attr"`
	AttributesDirect []Attribute `xml:"attribute"`
	Sequence         *Sequence   `xml:"sequence"`
	typ              Type
}

func (ext *Extension) Attributes() []Attribute {
	elements := ext.Elements()
	goNames := make(map[string]struct{}, len(elements)+len(ext.AttributesDirect))
	for _, el := range ext.Elements() {
		goNames[el.GoName()] = struct{}{}
	}
	attributes := []Attribute{}
	for _, attr := range ext.AttributesDirect {
		if _, found := goNames[attr.GoName()]; found {
			if attr.DuplicateCount == 0 {
				attr.DuplicateCount += 1
			}
			attr.DuplicateCount += 1
		}
		goNames[attr.GoName()] = struct{}{}
		attributes = append(attributes, attr)
	}
	return attributes
}

func (ext *Extension) Elements() []Element {
	elements := []Element{}
	if ext.typ != nil {
		elements = append(elements, ext.typ.Elements()...)
	}
	if ext.Sequence != nil {
		elements = append(elements, ext.Sequence.Elements()...)
		if ext.typ != nil {
			elements = deduplicateElements(elements)
		}
	}
	return elements
}

func deduplicateElements(elements []Element) []Element {
	seen := make(map[string]struct{}, len(elements))
	j := 0
	for _, element := range elements {
		if _, ok := seen[element.GoName()]; ok {
			continue
		}
		seen[element.GoName()] = struct{}{}
		elements[j] = element
		j++
	}
	return elements[:j]
}

func (ext *Extension) ContainsText() bool {
	return ext.Base == "xsd:string" || (ext.typ != nil && ext.typ.ContainsText())
}

func (ext *Extension) compile(sch *Schema, parentElement *Element) {
	if ext.Sequence != nil {
		ext.Sequence.compile(sch, parentElement)
	}
	if ext.Base == "" {
		panic("Not implemented: xsd:extension/@base empty, cannot extend unknown type")
	}

	ext.typ = sch.findReferencedType(ext.Base)
	if ext.typ == nil {
		panic("Cannot build xsd:extension: unknown type: " + string(ext.Base))
	}
	ext.typ.compile(sch, parentElement)
}