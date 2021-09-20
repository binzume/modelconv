package unity

import (
	"reflect"

	"github.com/binzume/modelconv/geom"
)

var UnityMeshes = map[Ref]string{
	{FileID: 10202, GUID: "0000000000000000e000000000000000"}: "Cube",
	{FileID: 10209, GUID: "0000000000000000e000000000000000"}: "Plane",
	{FileID: 10210, GUID: "0000000000000000e000000000000000"}: "Quad",
}

type Scene struct {
	GUID    string
	Objects []*GameObject

	Elements        map[int64]Element
	PrefabInstances map[string]*Scene

	Assets Assets
}

func NewScene(assets Assets, guid string) *Scene {
	return &Scene{GUID: guid, Elements: map[int64]Element{}, PrefabInstances: map[string]*Scene{}, Assets: assets}
}

func (s *Scene) GetElement(ref *Ref) Element {
	if !ref.IsValid() {
		return nil
	}
	if ref.GUID != "" && ref.GUID != s.GUID {
		if p, ok := s.PrefabInstances[ref.GUID]; ok {
			return p.GetElement(ref)
		}
		return nil
	}
	return s.Elements[ref.FileID]
}

func (s *Scene) GetGameObject(ref *Ref) *GameObject {
	t, _ := s.GetElement(ref).(*GameObject)
	return t
}

type Ref struct {
	FileID int64  `yaml:"fileID"`
	GUID   string `yaml:"guid"`
	Type   int    `yaml:"type"`
}

type Element interface{}

func (r *Ref) IsValid() bool {
	return r != nil && r.FileID != 0
}

type GameObject struct {
	Name      string `yaml:"m_Name"`
	IsActive  int    `yaml:"m_IsActive"`
	TagString string `yaml:"m_TagString"`

	Components []*struct {
		Ref Ref `yaml:"component"`
	} `yaml:"m_Component"`

	CorrespondingSourceObject *Ref `yaml:"m_CorrespondingSourceObject"`
	PrefabInstance            *Ref `yaml:"m_PrefabInstance"`

	Scene *Scene `yaml:"-"`
}

func (o *GameObject) GetComponent(target interface{}) bool {
	typ := reflect.TypeOf(target).Elem()
	for _, c := range o.Components {
		component := o.Scene.GetElement(&c.Ref)
		if reflect.TypeOf(component) == typ {
			reflect.ValueOf(target).Elem().Set(reflect.ValueOf(component))
			return true
		}
	}
	return false
}

func (o *GameObject) GetTransform() *Transform {
	for _, c := range o.Components {
		if t, ok := o.Scene.GetElement(&c.Ref).(*Transform); ok {
			return t
		}
	}
	return nil
}

type Component struct {
	Scene *Scene `yaml:"-"`

	ObjectHideFlags int `yaml:"m_ObjectHideFlags"`
	PrefabInternal  Ref `yaml:"m_PrefabInternal"`
	GameObject      Ref `yaml:"m_GameObject"`
}

func (c *Component) GetGameObject() *GameObject {
	obj, _ := c.Scene.GetElement(&c.GameObject).(*GameObject)
	return obj
}

type Transform struct {
	Component `yaml:",inline"`
	Father    Ref    `yaml:"m_Father"`
	Children  []*Ref `yaml:"m_Children"`

	LocalRotation geom.Vector4 `yaml:"m_LocalRotation"`
	LocalPosition geom.Vector3 `yaml:"m_LocalPosition"`
	LocalScale    geom.Vector3 `yaml:"m_LocalScale"`

	RootOrder int `yaml:"m_RootOrder"`
}

func (tr *Transform) GetMatrix() *geom.Matrix4 {
	pos := geom.NewTranslateMatrix4(tr.LocalPosition.X, tr.LocalPosition.Y, -tr.LocalPosition.Z)
	rot := geom.NewRotationMatrix4FromQuaternion(&tr.LocalRotation)
	sacle := geom.NewScaleMatrix4(tr.LocalScale.X, tr.LocalScale.Y, tr.LocalScale.Z)
	return pos.Mul(rot).Mul(sacle)
}

func (tr *Transform) GetWorldMatrix() *geom.Matrix4 {
	parent := tr.GetParent()
	if parent == nil {
		return tr.GetMatrix()
	}
	return parent.GetWorldMatrix().Mul(tr.GetMatrix())
}

func (tr *Transform) GetChildren() []*Transform {
	var children []*Transform
	for _, c := range tr.Children {
		if t, ok := tr.Scene.GetElement(c).(*Transform); ok {
			children = append(children, t)
		}
	}
	return children
}

func (tr *Transform) GetParent() *Transform {
	t, _ := tr.Scene.GetElement(&tr.Father).(*Transform)
	return t
}

type MeshFilter struct {
	Component `yaml:",inline"`
	Mesh      *Ref `yaml:"m_Mesh"`
}

type MeshRenderer struct {
	Component `yaml:",inline"`
	Enabled   int `yaml:"m_Enabled"`

	CastShadows          int `yaml:"m_CastShadows"`
	ReceiveShadows       int `yaml:"m_ReceiveShadows"`
	DynamicOccludee      int `yaml:"m_DynamicOccludee"`
	MotionVectors        int `yaml:"m_MotionVectors"`
	LightProbeUsage      int `yaml:"m_LightProbeUsage"`
	ReflectionProbeUsage int `yaml:"m_ReflectionProbeUsage"`

	Materials []*Ref `yaml:"m_Materials"`
}

type MonoBehaviour struct {
	Component `yaml:",inline"`
	Enabled   int  `yaml:"m_Enabled"`
	Script    *Ref `yaml:"m_Script"`

	RawData map[string]interface{} `yaml:",inline"`
}

type PrefabInstance struct {
	Modification struct {
		Modifications     []interface{} `yaml:"m_Modifications"`
		TransformParent   *Ref          `yaml:"m_TransformParent"`
		RemovedComponents []interface{} `yaml:"m_RemovedComponents"`
	} `yaml:"m_Modification"`

	SourcePrefab *Ref `yaml:"m_SourcePrefab"`
}
