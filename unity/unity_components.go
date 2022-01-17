package unity

import "github.com/binzume/modelconv/geom"

type componentDesc struct {
	Transform     *Transform     `yaml:"Transform" typeid:"unity3d.com,2011:4"`
	MonoBehaviour *MonoBehaviour `yaml:"MonoBehaviour" typeid:"unity3d.com,2011:114"`

	// Mesh
	MeshRenderer *MeshRenderer `yaml:"MeshRenderer" typeid:"unity3d.com,2011:23"`
	MeshFilter   *MeshFilter   `yaml:"MeshFilter" typeid:"unity3d.com,2011:33"`

	// Physics
	Rigidbody       *Rigidbody       `yaml:"Rigidbody" typeid:"unity3d.com,2011:54"`
	MeshCollider    *MeshCollider    `yaml:"MeshCollider" typeid:"unity3d.com,2011:64"`
	BoxCollider     *BoxCollider     `yaml:"BoxCollider" typeid:"unity3d.com,2011:65"`
	SphereCollider  *SphereCollider  `yaml:"SphereCollider" typeid:"unity3d.com,2011:135"`
	CapsuleCollider *CapsuleCollider `yaml:"CapsuleCollider" typeid:"unity3d.com,2011:136"`
}

type Component interface {
	GetGameObject() *GameObject
	init(*Scene)
}

type BaseComponent struct {
	Scene *Scene `yaml:"-"`

	ObjectHideFlags int `yaml:"m_ObjectHideFlags"`
	PrefabInternal  Ref `yaml:"m_PrefabInternal"`
	GameObject      Ref `yaml:"m_GameObject"`

	CorrespondingSourceObject *Ref `yaml:"m_CorrespondingSourceObject"`
	PrefabInstance            *Ref `yaml:"m_PrefabInstance"`
}

func (c *BaseComponent) init(scene *Scene) {
	c.Scene = scene
}

func (c *BaseComponent) GetGameObject() *GameObject {
	obj, _ := c.Scene.GetElement2(&c.GameObject, c.PrefabInstance).(*GameObject)
	return obj
}

type Transform struct {
	BaseComponent `yaml:",inline"`
	Father        Ref    `yaml:"m_Father"`
	Children      []*Ref `yaml:"m_Children"`

	LocalRotation geom.Vector4 `yaml:"m_LocalRotation"`
	LocalPosition geom.Vector3 `yaml:"m_LocalPosition"`
	LocalScale    geom.Vector3 `yaml:"m_LocalScale"`

	RootOrder int `yaml:"m_RootOrder"`

	children []*Transform
}

func (tr *Transform) GetMatrix() *geom.Matrix4 {
	pos := geom.NewTranslateMatrix4(tr.LocalPosition.X, tr.LocalPosition.Y, -tr.LocalPosition.Z)
	rot := geom.NewRotationMatrix4FromQuaternion(geom.NewQuaternion(-tr.LocalRotation.X, -tr.LocalRotation.Y, tr.LocalRotation.Z, tr.LocalRotation.W))
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
	if tr.children != nil {
		return tr.children
	}
	var children []*Transform
	for _, c := range tr.Children {
		if t := tr.Scene.GetTransform(c, tr.PrefabInstance); t != nil {
			children = append(children, t)
		}
	}
	tr.children = children
	return children
}

func (tr *Transform) AddChild(child *Transform) bool {
	if tr.children == nil {
		tr.GetChildren()
	}
	for _, c := range tr.children {
		if child == c {
			return false
		}
	}
	tr.children = append(tr.children, child)
	return true
}

func (tr *Transform) GetParent() *Transform {
	// TODO: prefab parent
	t, _ := tr.Scene.GetElement(&tr.Father).(*Transform)
	return t
}

type MonoBehaviour struct {
	BaseComponent `yaml:",inline"`

	Enabled int  `yaml:"m_Enabled"`
	Script  *Ref `yaml:"m_Script"`

	RawData map[string]interface{} `yaml:",inline"`
}

type MeshFilter struct {
	BaseComponent `yaml:",inline"`
	Mesh          *Ref `yaml:"m_Mesh"`
}

type MeshRenderer struct {
	BaseComponent `yaml:",inline"`
	Enabled       int `yaml:"m_Enabled"`

	CastShadows          int `yaml:"m_CastShadows"`
	ReceiveShadows       int `yaml:"m_ReceiveShadows"`
	DynamicOccludee      int `yaml:"m_DynamicOccludee"`
	MotionVectors        int `yaml:"m_MotionVectors"`
	LightProbeUsage      int `yaml:"m_LightProbeUsage"`
	ReflectionProbeUsage int `yaml:"m_ReflectionProbeUsage"`

	Materials []*Ref `yaml:"m_Materials"`
}

type Rigidbody struct {
	BaseComponent `yaml:",inline"`

	Mass               float32 `yaml:"m_Mass"`
	Drag               float32 `yaml:"m_Drag"`
	AngularDrag        float32 `yaml:"m_AngularDrag"`
	UseGravity         int     `yaml:"m_UseGravity"`
	IsKinematic        int     `yaml:"m_IsKinematic"`
	Interpolate        int     `yaml:"m_Interpolate"`
	Constraints        int     `yaml:"m_Constraints"`
	CollisionDetection int     `yaml:"m_CollisionDetection"`
}

type MeshCollider struct {
	BaseComponent `yaml:",inline"`

	Enabled   int  `yaml:"m_Enabled"`
	Mesh      *Ref `yaml:"m_Mesh"`
	IsTrigger int  `yaml:"m_IsTrigger"`
}

type BoxCollider struct {
	BaseComponent `yaml:",inline"`

	IsTrigger int          `yaml:"m_IsTrigger"`
	Center    geom.Vector3 `yaml:"m_Center"`
	Size      geom.Vector3 `yaml:"m_Size"`
}

type SphereCollider struct {
	BaseComponent `yaml:",inline"`

	IsTrigger int          `yaml:"m_IsTrigger"`
	Center    geom.Vector3 `yaml:"m_Center"`
	Radius    float32      `yaml:"m_Radius"`
}

type CapsuleCollider struct {
	BaseComponent `yaml:",inline"`

	IsTrigger int          `yaml:"m_IsTrigger"`
	Center    geom.Vector3 `yaml:"m_Center"`
	Radius    float32      `yaml:"m_Radius"`
	Direction int          `yaml:"m_Direction"`
	Height    float32      `yaml:"m_Height"`
}
