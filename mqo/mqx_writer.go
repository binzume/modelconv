package mqo

import (
	"encoding/xml"
	"fmt"
	"io"
)

type MQXDoc struct {
	XMLName     xml.Name `xml:"MetasequoiaDocument"`
	IncludedBy  string
	BonePlugin  *BonePlugin  `xml:"Plugin.56A31D20.71F282AB"`
	MorphPlugin *MorphPlugin `xml:"Plugin.56A31D20.C452C6DB"`
}

type BonePlugin struct {
	Name     string `xml:"name,attr"`
	BoneSet  BoneSet
	BoneSet2 BoneSet2
	Obj      []BoneObj
}

type BoneSet struct {
	Bone []*BoneOld
}

type BoneObj struct {
	ID int `xml:"id,attr"`
}

type BoneWeight struct {
	ObjectID int     `xml:"oi,attr"`
	VertexID int     `xml:"vi,attr"`
	Weight   float32 `xml:"w,attr"`
}

type BoneRef struct {
	ID int `xml:"id,attr"`
}

type BoneOld struct {
	ID      int    `xml:"id,attr"`
	Name    string `xml:"name,attr"`
	Group   int    `xml:"group,attr"`
	IsDummy int    `xml:"isDummy,attr"`

	RtX float32 `xml:"rtX,attr"`
	RtY float32 `xml:"rtY,attr"`
	RtZ float32 `xml:"rtZ,attr"`
	TpX float32 `xml:"tpX,attr"`
	TpY float32 `xml:"tpY,attr"`
	TpZ float32 `xml:"tpZ,attr"`

	MvX float32 `xml:"mvX,attr"`
	MvY float32 `xml:"mvY,attr"`
	MvZ float32 `xml:"mvZ,attr"`

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

type BoneSet2 struct {
	Bone []*Bone
}

type Bone struct {
	ID      int     `xml:"id,attr"`
	Name    string  `xml:"name,attr"`
	Group   int     `xml:"group,attr,omitempty"`
	Parent  int     `xml:"parent,attr,omitempty"`
	PosStr  string  `xml:"pos,attr,omitempty"`
	Pos     Vector3 `xml:"-"`
	Movable int     `xml:"movable,attr"`
	Color   string  `xml:"color,attr,omitempty"`

	IK *BoneIK `xml:"IK,omitempty"`

	Weights []*BoneWeight2 `xml:"W"`

	weightMap map[int]*BoneWeight2
}

func (b *Bone) SetVertexWeight(objectID, vertID int, weight float32) {
	w := b.weightMap[objectID]
	if w == nil {
		if b.weightMap == nil {
			b.weightMap = map[int]*BoneWeight2{}
		}
		w = &BoneWeight2{ObjectID: objectID}
		b.weightMap[objectID] = w
		b.Weights = append(b.Weights, w)
	}
	w.Vertexes = append(w.Vertexes, &VertexWeight{vertID, weight})
}

type BoneWeight2 struct {
	ObjectID int             `xml:"obj,attr"`
	Vertexes []*VertexWeight `xml:"V"`
}

type VertexWeight struct {
	VertexID int     `xml:"v,attr"`
	Weight   float32 `xml:"w,attr"`
}

type BoneIK struct {
	ChainCount int `xml:"chain,attr"`
}

type MorphPlugin struct {
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

func UpdateBoneRef(bones []*BoneOld) {
	for _, bone := range bones {
		bone.Children = nil
	}
	for idx, bone := range bones {
		if bone.Parent.ID > 0 {
			parent := bones[bone.Parent.ID-1]
			parent.Children = append(parent.Children, &BoneRef{ID: idx + 1})
		}
	}
}

func WriteMQX(mqo *MQODocument, w io.Writer, mqoName string) error {

	mqx := &MQXDoc{IncludedBy: mqoName}

	if len(mqo.Bones) > 0 {
		mqx.BonePlugin = &BonePlugin{
			Name:     "Bone",
			BoneSet2: BoneSet2{mqo.Bones},
		}
		// TODO
		mqx.BonePlugin.Obj = make([]BoneObj, len(mqo.Objects))
		for i := 0; i < len(mqo.Objects); i++ {
			mqx.BonePlugin.Obj[i].ID = i + 1
		}
		for _, b := range mqo.Bones {
			b.PosStr = fmt.Sprintf("%v,%v,%v", b.Pos.X, b.Pos.Y, b.Pos.Z)
		}
	}

	if len(mqo.Morphs) > 0 {
		mqx.MorphPlugin = &MorphPlugin{
			Name:     "Morph",
			MorphSet: MorphSet{mqo.Morphs},
		}
	}

	xmlBuf, _ := xml.MarshalIndent(mqx, "", "    ")
	w.Write([]byte(xml.Header))
	w.Write(xmlBuf)
	return nil
}
