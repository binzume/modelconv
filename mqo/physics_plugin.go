package mqo

import (
	"encoding/xml"

	"github.com/binzume/modelconv/geom"
)

// Fake plugin for keeping physics params
type PhysicsPlugin struct {
	XMLName xml.Name `xml:"Plugin.7A6E6962.43594850"`
	Name    string   `xml:"name,attr"`

	Bodies      []*PhysicsBody            `xml:"Bodies>Body"`
	Constraints []*PhysicsJointConstraint `xml:"Constraints>Joint"`
}

type Vector3XmlAttr struct {
	X geom.Element `xml:"x,attr"`
	Y geom.Element `xml:"y,attr"`
	Z geom.Element `xml:"z,attr"`
}

func (v *Vector3XmlAttr) Vec3() *Vector3 {
	return (*Vector3)(v)
}

type PhysicsBody struct {
	Shapes []*PhysicsShape `xml:"Shapes>Shape"`

	Mass           float32 `xml:"mass,attr"`
	Kinematic      bool    `xml:"kinematic,attr,omitempty"`
	CollisionGroup int
	CollisionMask  int

	// TODO: PhysicsMaterial
	Restitution    float32
	Friction       float32
	LinearDamping  float32
	AngularDamping float32

	// optional
	Name         string `xml:"name,attr,omitempty"`
	TargetBoneID int    `xml:"targetBone,attr,omitempty"`
	TargetObjID  int    `xml:"targetObj,attr,omitempty"`
}

type PhysicsShape struct {
	Type     string `xml:"type,attr"` // BOX | SPHERE | CAPSULE | CYLINDER | MESH
	Size     Vector3XmlAttr
	Position Vector3XmlAttr
	Rotation Vector3XmlAttr
}

type PhysicsJointConstraint struct {
	Type  string `xml:"type,attr,omitempty"`
	Body1 int    `xml:"body1,attr,omitempty"`
	Body2 int    `xml:"body2,attr,omitempty"`

	Name string `xml:"name,attr,omitempty"` // optional

	// Spring joint
	Position      Vector3XmlAttr
	Rotation      Vector3XmlAttr
	LinerSpring   Vector3XmlAttr
	AngulerSpring Vector3XmlAttr
}

func GetPhysicsPlugin(mqo *Document) *PhysicsPlugin {
	for _, p := range mqo.Plugins {
		if plugin, ok := p.(*PhysicsPlugin); ok {
			return plugin
		}
	}
	plugin := &PhysicsPlugin{Name: "Physics Plugin"}
	mqo.Plugins = append(mqo.Plugins, plugin)
	return plugin
}

func (p *PhysicsPlugin) PreSerialize(mqo *Document) {
}

func (p *PhysicsPlugin) PostDeserialize(mqo *Document) {
}

func (p *PhysicsPlugin) ApplyTransform(transform *Matrix4) {
	_, _, scale := transform.Decompose()
	for _, b := range p.Bodies {
		for _, s := range b.Shapes {
			pos := geom.Vector3(s.Position)
			s.Position = Vector3XmlAttr(*transform.ApplyTo(&pos))
			s.Size.X, s.Size.Y, s.Size.Z = geom.Abs(s.Size.X*scale.X), geom.Abs(s.Size.Y*scale.Y), geom.Abs(s.Size.Z*scale.Z)
		}
	}
}
