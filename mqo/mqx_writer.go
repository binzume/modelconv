package mqo

import (
	"encoding/xml"
	"io"
)

type MQXDoc struct {
	XMLName    xml.Name `xml:"MetasequoiaDocument"`
	IncludedBy string

	Plugins []Plugin
}

func WriteMQX(mqo *Document, w io.Writer, mqoName string) error {
	mqx := &MQXDoc{IncludedBy: mqoName, Plugins: mqo.GetPlugins()}
	for _, p := range mqx.Plugins {
		p.PreSerialize(mqo)
	}

	xmlBuf, _ := xml.MarshalIndent(mqx, "", "    ")
	w.Write([]byte(xml.Header))
	w.Write(xmlBuf)
	return nil
}
