package mqo

import (
	"encoding/xml"
	"io"
)

type MQXDoc struct {
	XMLName    xml.Name `xml:"MetasequoiaDocument"`
	IncludedBy string
	Plugin     *BonePlugin
}

type BoneSet struct {
	Bone []*Bone
}

type BoneObj struct {
	ID int `xml:"id,attr"`
}

type BonePlugin struct {
	XMLName xml.Name `xml:"Plugin.56A31D20.71F282AB"`
	Name    string   `xml:"name,attr"`
	BoneSet BoneSet
	Obj     []BoneObj
}

type BoneWeight struct {
	ObjectID int     `xml:"oi,attr"`
	VertexID int     `xml:"vi,attr"`
	Weight   float32 `xml:"w,attr"`
}

type BoneRef struct {
	ID int `xml:"id,attr"`
}

type Bone struct {
	ID      int    `xml:"id,attr"`
	Name    string `xml:"name,attr"`
	IsDummy int    `xml:"isDummy,attr"`

	RtX float32 `xml:"rtX,attr"`
	RtY float32 `xml:"rtY,attr"`
	RtZ float32 `xml:"rtZ,attr"`
	TpX float32 `xml:"tpX,attr"`
	TpY float32 `xml:"tpY,attr"`
	TpZ float32 `xml:"tpZ,attr"`

	MvX float32 `xml:"mvX,attr"`
	MvY float32 `xml:"mvY,attr"`
	MvZ float32 `xml:"myZ,attr"`

	Sc float32 `xml:"sc,attr"`

	RotB float32 `xml:"rotB,attr"`
	RotH float32 `xml:"rotH,attr"`
	RotP float32 `xml:"rotP,attr"`

	MaxAngB float32 `xml:"maxAngB,attr"`
	MaxAngH float32 `xml:"maxAngH,attr"`
	MaxAngP float32 `xml:"maxAngP,attr"`

	MinAngB float32 `xml:"minAngB,attr"`
	MinAngH float32 `xml:"minAngH,attr"`
	MinAngP float32 `xml:"minAngP,attr"`

	Parent   BoneRef    `xml:"P"`
	Children []*BoneRef `xml:"C"`

	Weights []*BoneWeight `xml:"W"`
}

func UpdateBoneRef(mqo *MQODocument) {
	for _, bone := range mqo.Bones {
		bone.Children = nil
	}
	for idx, bone := range mqo.Bones {
		if bone.Parent.ID > 0 {
			parent := mqo.Bones[bone.Parent.ID-1]
			parent.Children = append(parent.Children, &BoneRef{ID: idx + 1})
		}
	}
}

func WriteMQX(mqo *MQODocument, w io.Writer, mqoName string) error {

	mqx := &MQXDoc{
		IncludedBy: mqoName,
		Plugin: &BonePlugin{
			Name:    "Bone",
			BoneSet: BoneSet{mqo.Bones},
		},
	}

	// TODO
	mqx.Plugin.Obj = make([]BoneObj, len(mqo.Objects))
	for i := 0; i < len(mqo.Objects); i++ {
		mqx.Plugin.Obj[i].ID = i + 1
	}

	xmlBuf, _ := xml.MarshalIndent(mqx, "", "    ")
	w.Write([]byte(xml.Header))
	w.Write(xmlBuf)
	return nil
}
