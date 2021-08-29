package fbx

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
