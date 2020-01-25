package mqo

import (
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"
)

type BonePlugin struct {
	XMLName xml.Name `xml:"Plugin.56A31D20.71F282AB"`

	Name     string `xml:"name,attr"`
	BoneSet  BoneSet
	BoneSet2 BoneSet2
	PoseSet  PoseSet `xml:"Poses"`
	Obj      []BoneObj
}

func GetBonePlugin(mqo *MQODocument) *BonePlugin {
	for _, p := range mqo.Plugins {
		if bp, ok := p.(*BonePlugin); ok {
			return bp
		}
	}
	bp := &BonePlugin{}
	mqo.Plugins = append(mqo.Plugins, bp)
	return bp
}

func (p *BonePlugin) Bones() []*Bone {
	return p.BoneSet2.Bones
}

func (p *BonePlugin) SetBones(bones []*Bone) {
	p.BoneSet2.Bones = bones
}

func (p *BonePlugin) AddBone(b *Bone) {
	p.BoneSet2.Bones = append(p.BoneSet2.Bones, b)
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

	RotB float32 `xml:"rotB,attr"`
	RotH float32 `xml:"rotH,attr"`
	RotP float32 `xml:"rotP,attr"`

	Sc float32 `xml:"sc,attr"`

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

type BoneSet2 struct {
	Limit int     `xml:"limit,attr"`
	Bones []*Bone `xml:"Bone"`
}

type Bone struct {
	ID      int     `xml:"id,attr"`
	Name    string  `xml:"name,attr"`
	Group   int     `xml:"group,attr,omitempty"`
	Parent  int     `xml:"parent,attr,omitempty"`
	PosStr  string  `xml:"pos,attr,omitempty"`
	Pos     Vector3 `xml:"-"`
	Movable int     `xml:"movable,attr,omitempty"`
	Hide    int     `xml:"hide,attr,omitempty"`
	Dummy   int     `xml:"dummy,attr,omitempty"`
	Color   string  `xml:"color,attr,omitempty"`

	IK *BoneIK `xml:"IK,omitempty"`

	Weights []*BoneWeight2 `xml:"W"`

	weightMap map[int]*BoneWeight2
}

func (b *Bone) SetVertexWeight(objectID, vertID int, weight float32) *VertexWeight {
	w := b.weightMap[objectID]
	if w == nil {
		if b.weightMap == nil {
			b.weightMap = map[int]*BoneWeight2{}
		}
		w = &BoneWeight2{ObjectID: objectID}
		b.weightMap[objectID] = w
		b.Weights = append(b.Weights, w)
	}
	vw := &VertexWeight{vertID, weight}
	w.Vertexes = append(w.Vertexes, vw)
	return vw
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

type PoseSet struct {
	BonePoses []*BonePose `xml:"Pose"`
}

type BonePose struct {
	// oneof
	ID   int    `xml:"id,attr"`
	Name string `xml:"name,attr"`

	// Translation
	MvX float32 `xml:"mvX,attr"`
	MvY float32 `xml:"mvY,attr"`
	MvZ float32 `xml:"mvZ,attr"`

	// Rotation
	RotB float32 `xml:"rotB,attr"`
	RotH float32 `xml:"rotH,attr"`
	RotP float32 `xml:"rotP,attr"`

	// Scale
	ScB float32 `xml:"scB,attr"`
	ScH float32 `xml:"scH,attr"`
	ScP float32 `xml:"scP,attr"`
}

func (p *BonePlugin) PreSerialize(mqo *MQODocument) {
	// TODO
	p.Name = "Bone"
	for i, o := range mqo.Objects {
		if o.Depth == 0 {
			p.Obj = append(p.Obj, BoneObj{ID: i + 1})
		}
	}
	for _, b := range p.BoneSet2.Bones {
		b.PosStr = fmt.Sprintf("%v,%v,%v", b.Pos.X, b.Pos.Y, b.Pos.Z)
	}
}

func (p *BonePlugin) PostDeserialize(mqo *MQODocument) {
	for _, b := range p.BoneSet2.Bones {
		pos := strings.Split(b.PosStr, ",")
		x, _ := strconv.ParseFloat(pos[0], 32)
		y, _ := strconv.ParseFloat(pos[1], 32)
		z, _ := strconv.ParseFloat(pos[2], 32)
		b.Pos.X = float32(x)
		b.Pos.Y = float32(y)
		b.Pos.Z = float32(z)
	}
}
