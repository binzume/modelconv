package mqo

import (
	"encoding/xml"
	"fmt"
	"log"
	"math"

	"github.com/binzume/modelconv/geom"
)

type BonePlugin struct {
	XMLName xml.Name `xml:"Plugin.56A31D20.71F282AB"`

	Name     string `xml:"name,attr"`
	BoneSet  BoneSet
	BoneSet2 BoneSet2
	PoseSet  PoseSet `xml:"Poses"`
	Obj      []BoneObj
}

func GetBonePlugin(mqo *Document) *BonePlugin {
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
	ID       int          `xml:"id,attr"`
	Name     string       `xml:"name,attr"`
	Group    int          `xml:"group,attr,omitempty"`
	Parent   int          `xml:"parent,attr,omitempty"`
	Pos      Vector3Attr  `xml:"pos,attr,omitempty"`
	Movable  int          `xml:"movable,attr,omitempty"`
	Hide     int          `xml:"hide,attr,omitempty"`
	Dummy    int          `xml:"dummy,attr,omitempty"`
	Color    string       `xml:"color,attr,omitempty"`
	UpVector *Vector3Attr `xml:"upVector,attr,omitempty"`
	Rotate   *Vector3Attr `xml:"rotate,attr,omitempty"`

	IK *BoneIK `xml:"IK,omitempty"`

	Weights []*BoneWeight2 `xml:"W"`

	weightMap map[int]*BoneWeight2

	// internal use
	RotationOffset Vector3 `xml:"-"`
}

type Vector3Attr struct{ Vector3 }

func (v *Vector3Attr) MarshalXMLAttr(name xml.Name) (xml.Attr, error) {
	value := fmt.Sprintf("%v,%v,%v", v.X, v.Y, v.Z)
	return xml.Attr{Name: name, Value: value}, nil
}

func (v *Vector3Attr) UnmarshalXMLAttr(attr xml.Attr) error {
	fmt.Sscanf(attr.Value, "%f,%f,%f", &v.X, &v.Y, &v.Z)
	return nil
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

	// MMD
	Name    string `xml:"name,attr"`
	TipName string `xml:"tipName,attr"`
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

func (p *BonePlugin) PreSerialize(mqo *Document) {
	p.Name = "Bone"

	objs := map[int]bool{}
	for _, b := range p.BoneSet2.Bones {
		for _, w := range b.Weights {
			objs[w.ObjectID] = true
		}
	}
	for _, b := range p.BoneSet.Bone {
		for _, w := range b.Weights {
			objs[w.ObjectID] = true
		}
	}

	for i, o := range mqo.Objects {
		id := o.UID
		if id <= 0 {
			id = i + 1
		}
		if objs[id] {
			p.Obj = append(p.Obj, BoneObj{ID: id})
		}
	}
}

func (p *BonePlugin) PostDeserialize(mqo *Document) {
}

func (p *BonePlugin) ApplyTransform(transform *Matrix4) {
	for _, b := range p.BoneSet2.Bones {
		b.Pos.Vector3 = *transform.ApplyTo(&b.Pos.Vector3)
	}
}

func (doc *Document) BoneTransform(baseBone *Bone, rot *geom.Quaternion, boneFn func(b *Bone)) {
	bones := GetBonePlugin(doc).Bones()
	targetBones := map[*Bone]bool{baseBone: true}
	boneByID := map[int]*Bone{}
	for _, b := range bones {
		boneByID[b.ID] = b
	}
	boneFn(baseBone)

	// broken if bone.Parent > bone.ID ...
	for _, b := range bones {
		if b.Parent > 0 && targetBones[boneByID[b.Parent]] && !targetBones[b] {
			boneFn(b)
			targetBones[b] = true
		}
	}

	// Morph target
	objectByName := map[string]int{}
	for _, obj := range doc.Objects {
		objectByName[obj.Name] = obj.UID
	}
	morphObjs := map[int][]int{} // UID => []index
	for _, m := range GetMorphPlugin(doc).Morphs() {
		if oi, ok := objectByName[m.Base]; ok {
			for _, t := range m.Target {
				morphObjs[oi] = append(morphObjs[oi], objectByName[t.Name])
			}
		}
	}

	verts := map[*Vector3]float32{}
	for b := range targetBones {
		verts[&b.Pos.Vector3] = 1
		for _, bw := range b.Weights {
			objectIDs := append(morphObjs[bw.ObjectID], bw.ObjectID)
			for _, objID := range objectIDs {
				obj := doc.GetObjectByID(objID)
				if obj == nil {
					continue
				}
				for _, vw := range bw.Vertexes {
					w := vw.Weight / 100
					v := obj.Vertexes[obj.GetVertexIndexByID(vw.VertexID)]
					verts[v] += w
				}
			}
		}
	}

	// Physics collider
	if physics := FindPhysicsPlugin(doc); physics != nil {
		for _, b := range physics.Bodies {
			if b.TargetBoneID != 0 && targetBones[boneByID[b.TargetBoneID]] {
				for _, s := range b.Shapes {
					verts[s.Position.Vec3()] = 1
					// TODO: rotation order.
					r := s.Rotation.Vec3()
					angle := geom.NewEulerFromQuaternion(rot.Mul(geom.NewEuler(r.X, r.Y, r.Z, geom.RotationOrderXYZ).ToQuaternion()), geom.RotationOrderXYZ)
					*r = angle.Vector3
				}
			}
		}
	}

	pos := &baseBone.Pos.Vector3
	for v, w := range verts {
		dv := rot.ApplyTo(v.Sub(pos))
		v.X = (dv.X+pos.X)*w + v.X*(1-w)
		v.Y = (dv.Y+pos.Y)*w + v.Y*(1-w)
		v.Z = (dv.Z+pos.Z)*w + v.Z*(1-w)
	}
}

// for T-Pose adjustment
// TODO: more generic function.
func (doc *Document) BoneAdjustX(baseBone *Bone) {
	var boneVec *Vector3
	var max float32 = 0
	for _, b := range GetBonePlugin(doc).Bones() {
		if b.Parent == baseBone.ID {
			v := b.Pos.Sub(&baseBone.Pos.Vector3)
			if v.LenSqr() > max {
				max = v.LenSqr()
				boneVec = v
			}
		}
	}
	if boneVec == nil {
		return
	}

	bd := math.Atan2(float64(boneVec.Y), float64(boneVec.X))
	log.Println("T-Pose:", baseBone.Name, boneVec, bd/math.Pi*180)
	var rot float64
	if math.Abs(bd) < math.Pi/2 {
		rot = -bd
	} else {
		rot = math.Pi - bd
	}
	q := geom.NewEuler(0, 0, float32(rot), geom.RotationOrderXYZ).ToQuaternion()
	doc.BoneTransform(baseBone, q, func(b *Bone) {
		b.RotationOffset.X += float32(rot)
	})
}
