package mqo

import "encoding/xml"

type MorphPlugin struct {
	XMLName xml.Name `xml:"Plugin.56A31D20.C452C6DB"`

	Name     string `xml:"name,attr"`
	MorphSet MorphSet
}

func GetMorphPlugin(mqo *MQODocument) *MorphPlugin {
	for _, p := range mqo.Plugins {
		if bp, ok := p.(*MorphPlugin); ok {
			return bp
		}
	}
	bp := &MorphPlugin{Name: "Morph"}
	mqo.Plugins = append(mqo.Plugins, bp)
	return bp
}

type MorphSet struct {
	Targets []*MorphTargetList `xml:"TargetList"`
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
}

func (p *MorphPlugin) PostDeserialize(mqo *MQODocument) {
}
