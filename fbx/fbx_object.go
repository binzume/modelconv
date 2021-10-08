package fbx

import (
	"strings"
)

type Property struct {
	AttributeList
	Type  string
	Label string
	Flag  string
}

type Connection struct {
	Type string
	To   int64
	From int64
	Prop string
}
type Object interface {
	GetNode() *Node
	NodeName() string
	ID() int64
	Name() string
	Kind() string
	GetProperty(name string) *Property
	SetProperty(name string, prop *Property) *Property
	FindRefs(name string) []Object
	AddRef(o Object)
}

type Obj struct {
	*Node
	Template   *Obj
	Refs       []Object
	properties map[string]*Property // lazty initialize
}

func newObj(typ, name, kind string, nodes []*Node) *Obj {
	children := append(nodes, &Node{Name: "Properties70"})
	obj := &Obj{Node: &Node{
		Name:       typ,
		Attributes: []*Attribute{{Value: 0}, {Value: name}, {Value: kind}},
		Children:   children,
	}}
	return obj
}

func (o *Obj) GetNode() *Node {
	return o.Node
}
func (o *Obj) NodeName() string {
	return o.Node.Name
}
func (o *Obj) ID() int64 {
	return o.Attr(0).ToInt64(0)
}
func (o *Obj) Name() string {
	return strings.ReplaceAll(o.Attr(1).ToString(), "\x00\x01", "::")
}
func (o *Obj) Kind() string {
	return o.Attr(2).ToString()
}
func (o *Obj) GetProperty(name string) *Property {
	if o.properties == nil {
		o.properties = map[string]*Property{}
		for _, node := range o.FindChild("Properties70").GetChildren() {
			o.properties[node.Attr(0).ToString()] = &Property{
				AttributeList: node.Attributes[4:],
				Type:          node.Attr(1).ToString(),
				Label:         node.Attr(2).ToString(),
				Flag:          node.Attr(3).ToString()}
		}
	}
	if p, ok := o.properties[name]; ok {
		return p
	} else if o.Template != nil {
		return o.Template.GetProperty(name)
	}
	return &Property{}
}
func (o *Obj) SetProperty(name string, prop *Property) *Property {
	if o.properties != nil {
		o.properties[name] = prop
	}
	attrs := AttributeList{
		&Attribute{Value: name},
		&Attribute{Value: prop.Type},
		&Attribute{Value: prop.Label},
		&Attribute{Value: prop.Flag},
	}
	attrs = append(attrs, prop.AttributeList...)
	properties70 := o.FindChild("Properties70")
	for _, node := range properties70.GetChildren() {
		if node.Attr(0).ToString() == name {
			node.Attributes = attrs
			return prop
		}
	}
	properties70.Children = append(properties70.Children, &Node{Name: "P", Attributes: attrs})
	return prop
}

func (o *Obj) SetIntProperty(name string, v int) *Property {
	return o.SetProperty(name, &Property{Type: "int", Label: "Integer", AttributeList: []*Attribute{{Value: int64(v)}}})
}

func (o *Obj) SetFloatProperty(name string, v float64) *Property {
	return o.SetProperty(name, &Property{Type: "double", Label: "Number", AttributeList: []*Attribute{{Value: v}}})
}

func (o *Obj) SetStringProperty(name string, v string) *Property {
	return o.SetProperty(name, &Property{Type: "KString", Label: "", AttributeList: []*Attribute{{Value: v}}})
}

func (o *Obj) SetColorProperty(name string, r, g, b float32) *Property {
	return o.SetProperty(name, &Property{Type: "ColorRGB", Label: "Color", AttributeList: []*Attribute{{Value: r}, {Value: g}, {Value: b}}})
}

func (o *Obj) FindRefs(typ string) []Object {
	var refs []Object
	for _, o := range o.Refs {
		if o.NodeName() == typ {
			refs = append(refs, o)
		}
	}
	return refs
}
func (o *Obj) AddRef(ref Object) {
	o.Refs = append(o.Refs, ref)
}

func (o *Obj) AddOrReplaceChild(node *Node) bool {
	for i, c := range o.Children {
		if c.Name == node.Name {
			o.Children[i] = node
			return false
		}
	}
	o.Children = append(o.Children, node)
	return true
}
