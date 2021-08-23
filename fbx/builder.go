package fbx

import "github.com/binzume/modelconv/geom"

func buildMesh(node *Node) *Mesh {
	mesh := &Mesh{Obj: Obj{Node: node}}
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

func parseModel(node *Node) *Model {
	m := &Model{Obj: Obj{Node: node}, Scaling: geom.Vector3{X: 1, Y: 1, Z: 1}}
	for _, p := range node.ChildOrEmpty("Properties70").Children {
		if p.PropString(0) == "Lcl Translation" {
			m.Translation.X = p.PropFloat(4)
			m.Translation.Y = p.PropFloat(5)
			m.Translation.Z = p.PropFloat(6)
		} else if p.PropString(0) == "Lcl Rotation" {
			m.Rotation.X = p.PropFloat(4)
			m.Rotation.Y = p.PropFloat(5)
			m.Rotation.Z = p.PropFloat(6)
		} else if p.PropString(0) == "Lcl Scaling" {
			m.Scaling.X = p.PropFloat(4)
			m.Scaling.Y = p.PropFloat(5)
			m.Scaling.Z = p.PropFloat(6)
		}
	}
	return m
}

func parseConnection(node *Node) *Connection {
	c := &Connection{
		Type: node.PropString(0),
		From: node.PropInt(1),
		To:   node.PropInt(2),
	}
	if c.Type == "OP" {
		c.Prop = node.PropString(3)
	}
	return c
}

func BuildDocument(root *Node) (*Document, error) {
	doc := &Document{RawNode: root, Scene: &Obj{}}
	doc.Objects = map[int]Object{0: doc.Scene}

	doc.Creator = root.Child("Creator").PropString(0)
	doc.CreationTime = root.Child("CreationTime").PropString(0)
	doc.FileId, _ = root.Child("FileId").PropValue(0).([]byte)

	for _, node := range root.ChildOrEmpty("Objects").Children {
		var obj Object
		switch node.Name {
		case "Geometry":
			obj = buildMesh(node)
		case "Material":
			doc.Materials = append(doc.Materials, &Material{Obj{Node: node}})
			obj = doc.Materials[len(doc.Materials)-1]
		case "Model":
			obj = parseModel(node)
		default:
			obj = &Obj{Node: node}
		}
		doc.Objects[obj.ID()] = obj
	}

	for _, node := range root.ChildOrEmpty("Connections").Children {
		if node.Name == "C" {
			c := parseConnection(node)
			if c.Type == "OO" || c.Type == "OP" {
				from := doc.Objects[c.From]
				to := doc.Objects[c.To]
				if to != nil && from != nil {
					to.AddRef(from)
				}
			}
		}
	}

	return doc, nil
}
