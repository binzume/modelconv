package mqo

import (
	"encoding/xml"
	"io"
)

func ReadMQX(r io.Reader) (*MQXDoc, error) {
	var data struct {
		MQXDoc
		BonePlugin  *BonePlugin
		MorphPlugin *MorphPlugin
	}
	err := xml.NewDecoder(r).Decode(&data)
	doc := data.MQXDoc
	if data.BonePlugin != nil {
		doc.Plugins = append(doc.Plugins, data.BonePlugin)
	}
	if data.MorphPlugin != nil {
		doc.Plugins = append(doc.Plugins, data.MorphPlugin)
	}
	for _, p := range doc.Plugins {
		p.PostDeserialize(nil)
	}
	return &doc, err
}
