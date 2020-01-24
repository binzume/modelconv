package mqo

import "encoding/xml"

type MorphPlugin struct {
	XMLName xml.Name `xml:"Plugin.56A31D20.C452C6DB"`

	Name     string `xml:"name,attr"`
	MorphSet MorphSet
	Obj      []BoneObj
}

type MorphSet struct {
	TargetList []*MorphTargetList
}

type MorphTargetList struct {
	Base   string `xml:"base,attr"`
	Target []*MorphTarget
}

type MorphTarget struct {
	Name  string `xml:"name,attr"`
	Param int    `xml:"param,attr"`
}

func (p *MorphPlugin) PreSerialize(mqo *MQODocument) {
	p.Name = "Morph"
}
