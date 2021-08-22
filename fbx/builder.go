package fbx

func buildMesh(node *Node) *Mesh {
	mesh := &Mesh{Object: Object{node}}
	mesh.Vertices = node.Child("Vertices").Prop(0).ToVec3Array()
	normal := node.Child("LayerElementNormal")
	if normal.Child("MappingInformationType").PropString(0) == "ByPolygonVertex" {
		mesh.Normals = normal.Child("Normals").Prop(0).ToVec3Array()
	}
	if v, ok := node.Child("PolygonVertexIndex").PropValue(0).([]int32); ok {
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

func BuildDocument(root *Node) (*Document, error) {
	doc := &Document{RawNode: root}
	doc.Objects = map[int]*Object{}

	doc.Creator = root.Child("Creator").PropString(0)
	doc.CreationTime = root.Child("CreationTime").PropString(0)
	doc.FileId, _ = root.Child("FileId").PropValue(0).([]byte)

	for _, node := range root.ChildOrEmpty("Objects").Children {
		var obj *Object
		switch node.Name {
		case "Geometry":
			doc.Meshes = append(doc.Meshes, buildMesh(node))
			obj = &doc.Meshes[len(doc.Meshes)-1].Object
		case "Material":
			doc.Materials = append(doc.Materials, &Material{Object{node}})
			obj = &doc.Materials[len(doc.Materials)-1].Object
		default:
			obj = &Object{node}
		}
		doc.Objects[obj.ID()] = obj
	}

	for _, node := range root.ChildOrEmpty("Connections").Children {
		if node.Name == "C" {
			doc.Connections = append(doc.Connections, &Connection{
				Type: node.PropString(0),
				From: node.PropInt(1),
				To:   node.PropInt(2),
			})
		}
	}

	return doc, nil
}
