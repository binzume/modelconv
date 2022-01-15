package unity

import "github.com/binzume/modelconv/geom"

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
