package fbx

func parseGeometry(base *Obj) *Geometry {
	mesh := &Geometry{Obj: *base}
	mesh.Vertices = base.FindChild("Vertices").Prop(0).ToVec3Array()
	normal := base.FindChild("LayerElementNormal")
	if normal.FindChild("MappingInformationType").PropString(0) == "ByPolygonVertex" {
		mesh.Normals = normal.FindChild("Normals").Prop(0).ToVec3Array()
	}
	if v := base.FindChild("PolygonVertexIndex").Prop(0).ToInt32Array(); v != nil {
		var face []int
		for _, index := range v {
			if index < 0 {
				face = append(face, int(^index))
				mesh.Faces = append(mesh.Faces, face)
				face = nil
				continue
			}
			face = append(face, int(index))
		}
	}
	return mesh
}

func parseModel(base *Obj) *Model {
	m := &Model{Obj: *base}

	m.Translation = m.GetProperty70("Lcl Translation").ToVector3(0, 0, 0)
	m.Rotation = m.GetProperty70("Lcl Rotation").ToVector3(0, 0, 0)
	m.Scaling = m.GetProperty70("Lcl Scaling").ToVector3(1, 1, 1)

	return m
}

func parseConnection(node *Node) *Connection {
	c := &Connection{
		Type: node.Prop(0).ToString(""),
		From: node.Prop(1).ToInt64(0),
		To:   node.Prop(2).ToInt64(0),
	}
	if c.Type == "OP" {
		c.Prop = node.Prop(3).ToString("")
	}
	return c
}

func BuildDocument(root *Node) (*Document, error) {
	doc := &Document{RawNode: root, Scene: &Obj{}}
	doc.Objects = map[int64]Object{0: doc.Scene}

	doc.Creator = root.FindChild("Creator").PropString(0)
	doc.CreationTime = root.FindChild("CreationTime").PropString(0)
	doc.FileId, _ = root.FindChild("FileId").PropValue(0).([]byte)

	templates := map[string]*Obj{}
	for _, node := range root.FindChild("Definitions").GetChildren() {
		if node.Name != "ObjectType" {
			continue
		}
		templates[node.PropString(0)] = &Obj{Node: node.FindChild("PropertyTemplate")}
	}
	doc.GlobalSettings = &Obj{Node: root.FindChild("GlobalSettings"), Template: templates["GlobalSettings"]}

	for _, node := range root.FindChild("Objects").GetChildren() {
		base := &Obj{Node: node, Template: templates[node.Name]}
		var obj Object = base
		switch node.Name {
		case "Geometry":
			obj = parseGeometry(base)
		case "Material":
			doc.Materials = append(doc.Materials, &Material{*base})
			obj = doc.Materials[len(doc.Materials)-1]
		case "Model":
			obj = parseModel(base)
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
