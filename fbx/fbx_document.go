package fbx

type Document struct {
	FileId       []byte
	Creator      string
	CreationTime string

	GlobalSettings Object
	Objects        map[int64]Object
	Scene          *Model

	Materials []*Material

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

func parseGeometry(base *Obj) *Geometry {
	geometry := &Geometry{Obj: *base}
	geometry.Vertices = geometry.GetVertices()
	polygonVertices := base.FindChild("PolygonVertexIndex").GetInt32Array()
	geometry.PolygonVertexCount = len(polygonVertices)
	if v := polygonVertices; v != nil {
		var face []int
		for _, index := range v {
			if index < 0 {
				face = append(face, int(^index))
				geometry.Polygons = append(geometry.Polygons, face)
				face = nil
				continue
			}
			face = append(face, int(index))
		}
	}
	return geometry
}

func parseConnection(node *Node) *Connection {
	c := &Connection{
		Type: node.Attr(0).ToString(),
		From: node.Attr(1).ToInt64(0),
		To:   node.Attr(2).ToInt64(0),
	}
	if c.Type == "OP" {
		c.Prop = node.Attr(3).ToString()
	}
	return c
}

func BuildDocument(root *Node) (*Document, error) {
	doc := &Document{RawNode: root, Scene: &Model{Obj: Obj{}}}
	doc.Objects = map[int64]Object{0: doc.Scene}

	doc.Creator = root.FindChild("Creator").GetString()
	doc.CreationTime = root.FindChild("CreationTime").GetString()
	if fileId := root.FindChild("FileId").Attr(0); fileId != nil {
		doc.FileId, _ = fileId.Value.([]byte)
	}

	templates := map[string]*Obj{}
	for _, node := range root.FindChild("Definitions").GetChildren() {
		if node.Name != "ObjectType" {
			continue
		}
		templates[node.GetString()] = &Obj{Node: node.FindChild("PropertyTemplate")}
	}
	doc.GlobalSettings = &Obj{Node: root.FindChild("GlobalSettings"), Template: templates["GlobalSettings"]}

	for _, node := range root.FindChild("Objects").GetChildren() {
		base := &Obj{Node: node, Template: templates[node.Name]}
		var obj Object = base
		switch node.Name {
		case "Deformer":
			obj = &Deformer{*base}
		case "Geometry":
			obj = parseGeometry(base)
		case "Material":
			doc.Materials = append(doc.Materials, &Material{*base})
			obj = doc.Materials[len(doc.Materials)-1]
		case "Model":
			obj = &Model{Obj: *base}
		}
		doc.Objects[obj.ID()] = obj
	}

	for _, node := range root.FindChild("Connections").GetChildren() {
		if node.Name == "C" {
			c := parseConnection(node)
			if c.Type == "OO" || c.Type == "OP" {
				from := doc.Objects[c.From]
				to := doc.Objects[c.To]
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
	}

	return doc, nil
}

func (doc *Document) AddObject(obj Object) int64 {
	existing := doc.Objects[obj.ID()]
	if existing == obj {
		return obj.ID()
	}
	if obj.ID() == 0 || existing != nil {
		for doc.Objects[doc.NextObjectID] != nil {
			doc.NextObjectID++
		}
		obj.GetNode().Attributes[0].Value = doc.NextObjectID
		doc.NextObjectID++
	}
	doc.Objects[obj.ID()] = obj
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
