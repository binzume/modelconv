package unity

import (
	"github.com/binzume/modelconv/geom"
)

type Scene struct {
	GUID    string
	Objects []*GameObject

	Elements        map[int64]interface{}
	PrefabInstances map[string]*Scene
}

func (s *Scene) GetElement(ref *Ref) interface{} {
	if ref == nil {
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

func (s *Scene) GetTransform(ref *Ref) *Transform {
	t, _ := s.GetElement(ref).(*Transform)
	return t
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

func (o *GameObject) GetComponent(ref *Ref) interface{} {
	return o.Scene.GetElement(ref)
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

func (tr *Transform) GetChildren() []*Transform {
	var children []*Transform
	for _, c := range tr.Children {
		if t, ok := tr.Scene.GetElement(c).(*Transform); ok {
			children = append(children, t)
		}
	}
	return children
}

type MeshFilter struct {
	Component `yaml:",inline"`
	Mesh      *Ref `yaml:"m_Mesh"`
}

type MeshRenderer struct {
	Component `yaml:",inline"`

	Enabled int `yaml:"m_Enabled"`

	CastShadows          int `yaml:"m_CastShadows"`
	ReceiveShadows       int `yaml:"m_ReceiveShadows"`
	DynamicOccludee      int `yaml:"m_DynamicOccludee"`
	MotionVectors        int `yaml:"m_MotionVectors"`
	LightProbeUsage      int `yaml:"m_LightProbeUsage"`
	ReflectionProbeUsage int `yaml:"m_ReflectionProbeUsage"`

	Materials []*Ref `yaml:"m_Materials"`
}

func (r *MeshRenderer) GetMeshFilter() *MeshFilter {
	o := r.GetGameObject()
	if o == nil {
		return nil
	}
	for _, c := range o.Components {
		if t, ok := o.Scene.GetElement(&c.Ref).(*MeshFilter); ok {
			return t
		}
	}
	return nil
}

type PrefabInstance struct {
	Modification struct {
		Modifications     []interface{} `yaml:"m_Modifications"`
		TransformParent   *Ref          `yaml:"m_TransformParent"`
		RemovedComponents []interface{} `yaml:"m_RemovedComponents"`
	} `yaml:"m_Modification"`

	SourcePrefab *Ref `yaml:"m_SourcePrefab"`
}

type MetaFile struct {
	FileFormatVersion int         `yaml:"fileFormatVersion"`
	GUID              string      `yaml:"guid"`
	ModelImporter     interface{} `yaml:"ModelImporter"`
}
