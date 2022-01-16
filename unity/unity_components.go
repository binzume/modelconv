package unity

import "github.com/binzume/modelconv/geom"

type MonoBehaviour struct {
	Component `yaml:",inline"`

	Enabled int  `yaml:"m_Enabled"`
	Script  *Ref `yaml:"m_Script"`

	RawData map[string]interface{} `yaml:",inline"`
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

type Rigidbody struct {
	Component `yaml:",inline"`

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
	Component `yaml:",inline"`

	Enabled   int  `yaml:"m_Enabled"`
	Mesh      *Ref `yaml:"m_Mesh"`
	IsTrigger int  `yaml:"m_IsTrigger"`
}

type BoxCollider struct {
	Component `yaml:",inline"`

	IsTrigger int          `yaml:"m_IsTrigger"`
	Center    geom.Vector3 `yaml:"m_Center"`
	Size      geom.Vector3 `yaml:"m_Size"`
}

type SphereCollider struct {
	Component `yaml:",inline"`

	IsTrigger int          `yaml:"m_IsTrigger"`
	Center    geom.Vector3 `yaml:"m_Center"`
	Radius    float32      `yaml:"m_Radius"`
}

type CapsuleCollider struct {
	Component `yaml:",inline"`

	IsTrigger int          `yaml:"m_IsTrigger"`
	Center    geom.Vector3 `yaml:"m_Center"`
	Radius    float32      `yaml:"m_Radius"`
	Direction int          `yaml:"m_Direction"`
	Height    float32      `yaml:"m_Height"`
}
