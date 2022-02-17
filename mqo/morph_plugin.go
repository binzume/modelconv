package mqo

import (
	"encoding/xml"
	"strings"
)

type MorphPlugin struct {
	XMLName xml.Name `xml:"Plugin.56A31D20.C452C6DB"`

	Name     string `xml:"name,attr"`
	MorphSet MorphSet
}

func (p *MorphPlugin) Morphs() []*MorphTargetList {
	return p.MorphSet.Targets
}

func GetMorphPlugin(mqo *Document) *MorphPlugin {
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

func (p *MorphPlugin) PreSerialize(mqo *Document) {
}

func (p *MorphPlugin) PostDeserialize(mqo *Document) {
}

func (p *MorphPlugin) Apply(doc *Document, name string, value float32) (updated bool) {
	for _, m := range p.Morphs() {
		for i, t := range m.Target {
			if t.Name == name {
				updated = m.Apply(doc, i, float32(value)) || updated
			}
		}
	}
	p.ApplyMaterialMorph(doc, name, value)
	return
}

func (p *MorphPlugin) ApplyMaterialMorph(doc *Document, name string, value float32) {
	matByID := map[string]*Material{}
	for _, m := range doc.Materials {
		matByID[m.Name] = m
	}
	morphPrefix := "$MORPH:" + name + ":"
	for _, m := range doc.Materials {
		if strings.HasPrefix(m.Name, morphPrefix) {
			t := matByID[m.Name[len(morphPrefix):]]
			if m == nil {
				continue
			}
			t.Color = *t.Color.Scale(1 - value).Add(m.Color.Scale(value))
		}

	}
}

func (m *MorphTargetList) Apply(doc *Document, target int, value float32) bool {
	objByID := map[string]*Object{}
	for _, o := range doc.Objects {
		objByID[o.Name] = o
	}
	bobj, tobj := objByID[m.Base], objByID[m.Target[target].Name]
	if bobj == nil || tobj == nil || len(bobj.Vertexes) != len(tobj.Vertexes) {
		return false
	}

	diff := make([]*Vector3, len(tobj.Vertexes))
	for i := range bobj.Vertexes {
		diff[i] = tobj.Vertexes[i].Sub(bobj.Vertexes[i]).Scale(value)
	}
	// Apply
	for i, t := range m.Target {
		o := objByID[t.Name]
		if i == target {
			o = bobj
		}
		if o == nil || len(o.Vertexes) != len(tobj.Vertexes) {
			continue
		}
		for i := range o.Vertexes {
			o.Vertexes[i] = o.Vertexes[i].Add(diff[i])
		}
	}
	return true
}
