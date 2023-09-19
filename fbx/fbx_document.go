package fbx

import "github.com/binzume/modelconv/geom"

type Document struct {
	FileId       []byte
	Creator      string
	CreationTime string

	GlobalSettings Object
	Scene          *Model

	Materials []*Material

	ObjectByID   map[int64]Object
	RawNode      *Node
	NextObjectID int64
}

func NewDocument() *Document {
	globalSettings := &Obj{Node: &Node{
		Name: "GlobalSettings",
		Children: []*Node{
			NewNode("Version", 1000),
			{Name: "Properties70"},
		},
	}}
	globalSettings.SetIntProperty("UpAxis", 1)
	globalSettings.SetIntProperty("UpAxisSign", 1)
	globalSettings.SetIntProperty("FrontAxis", 2)
	globalSettings.SetIntProperty("FrontAxisSign", 1)
	globalSettings.SetIntProperty("CoordAxis", 0)
	globalSettings.SetIntProperty("CoordAxisSign", 1)
	globalSettings.SetFloatProperty("UnitScaleFactor", 1.0)
	root := &Node{Children: []*Node{
		{Name: "FBXHeaderExtension", Children: []*Node{
			NewNode("FBXHeaderVersion", 1003),
			NewNode("FBXVersion", 7500),
			NewNode("Creator", "modelconv"),
		}},
		globalSettings.Node,
		{Name: "Definitions", Children: []*Node{
			NewNode("Version", 100),
			NewNode("ObjectType", "GlobalSettings"),
			NewNode("ObjectType", "Model"),
			NewNode("ObjectType", "Geometry"),
			NewNode("ObjectType", "Material"),
		}},
		{Name: "Objects"},
		{Name: "Connections"},
	}}

	doc, _ := BuildDocument(root)
	return doc
}

func (doc *Document) newID() int64 {
	for doc.ObjectByID[doc.NextObjectID] != nil {
		doc.NextObjectID++
	}
	doc.NextObjectID++
	return doc.NextObjectID - 1
}

func BuildDocument(root *Node) (*Document, error) {
	doc := &Document{RawNode: root, Scene: &Model{Obj: Obj{Node: NewNode("Scene\x00\x01Model")}}}
	doc.ObjectByID = map[int64]Object{0: doc.Scene}

	doc.Creator = root.FindChild("Creator").GetString()
	doc.CreationTime = root.FindChild("CreationTime").GetString()
	if fileId := root.FindChild("FileId").Attr(0); fileId != nil {
		doc.FileId, _ = fileId.Value.([]byte)
	}

	fbxVersion := root.FindChild("FBXHeaderExtension").FindChild("FBXVersion").Attr(0).ToInt(0)

	templates := map[string]*Obj{}
	for _, node := range root.FindChild("Definitions").GetChildren() {
		if node.Name != "ObjectType" {
			continue
		}
		templates[node.GetString()] = &Obj{Node: node.FindChild("PropertyTemplate")}
	}
	doc.GlobalSettings = &Obj{Node: root.FindChild("GlobalSettings"), Template: templates["GlobalSettings"]}
	objectByName := map[string]Object{"Scene\x00\x01Model": doc.Scene}
	for _, node := range root.FindChild("Objects").GetChildren() {
		if fbxVersion == 6100 {
			// TODO: clone
			node2 := *node
			node = &node2
			node.Attributes = append([]*Attribute{{Value: doc.newID()}}, node.Attributes...)
		}
		base := &Obj{Node: node, Template: templates[node.Name]}
		var obj Object = base
		switch node.Name {
		case "Deformer":
			obj = &Deformer{*base}
		case "Geometry":
			geometry := &Geometry{Obj: *base}
			geometry.Init()
			obj = geometry
		case "Material":
			doc.Materials = append(doc.Materials, &Material{*base})
			obj = doc.Materials[len(doc.Materials)-1]
		case "Model":
			obj = &Model{Obj: *base}
			if fbxVersion == 6100 && base.FindChild("Vertices") != nil {
				// TODO
				g := &Geometry{Obj: *base}
				g.Init()
				obj.AddRef(g)
			}
		}
		doc.ObjectByID[obj.ID()] = obj
		objectByName[obj.GetNode().Attr(1).ToString()] = obj
	}

	for _, node := range root.FindChild("Connections").GetChildren() {
		ctype := node.Attr(0).ToString()
		if (ctype == "OO" || ctype == "OP") && len(node.Attributes) >= 3 {
			from := doc.ObjectByID[node.Attr(1).ToInt64(0)]
			to := doc.ObjectByID[node.Attr(2).ToInt64(0)]
			if fbxVersion == 6100 {
				from = objectByName[node.Attr(1).ToString()]
				to = objectByName[node.Attr(2).ToString()]
			}
			if to != nil && from != nil {
				to.AddRef(from)

				// build model tree
				if toModel, ok := to.(*Model); ok {
					if fromModel, ok := from.(*Model); ok {
						fromModel.Parent = toModel
					}
				}
			}
		}
	}

	return doc, nil
}

func (doc *Document) AddObject(obj Object) int64 {
	existing := doc.ObjectByID[obj.ID()]
	if existing == obj {
		return obj.ID()
	}
	if obj.ID() == 0 || existing != nil {
		obj.GetNode().Attributes[0].Value = doc.newID()
	}
	doc.ObjectByID[obj.ID()] = obj
	objs := doc.RawNode.FindChild("Objects")
	objs.Children = append(objs.Children, obj.GetNode())
	return obj.ID()
}

func (doc *Document) AddConnection(parent, child Object) {
	parent.AddRef(child)
	conns := doc.RawNode.FindChild("Connections")
	conns.Children = append(conns.Children, NewNode("C", "OO", child.ID(), parent.ID()))
}

func (doc *Document) AddPropConnection(parent, child Object, prop string) {
	parent.AddRef(child)
	conns := doc.RawNode.FindChild("Connections")
	conns.Children = append(conns.Children, NewNode("C", "OP", child.ID(), parent.ID(), prop))
}

func (doc *Document) CoordMatrix() *geom.Matrix4 {
	gs := doc.GlobalSettings
	mat := geom.Matrix4{
		0, 0, 0, 0,
		0, 0, 0, 0,
		0, 0, 0, 0,
		0, 0, 0, 1,
	}
	mat[gs.GetProperty("CoordAxis").ToInt(0)*4] = gs.GetProperty("CoordAxisSign").ToFloat32(1)
	mat[gs.GetProperty("UpAxis").ToInt(1)*4+1] = gs.GetProperty("UpAxisSign").ToFloat32(1)
	mat[gs.GetProperty("FrontAxis").ToInt(2)*4+2] = gs.GetProperty("FrontAxisSign").ToFloat32(1)
	return &mat
}
