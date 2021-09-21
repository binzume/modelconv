package converter

import (
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"strings"

	"github.com/binzume/modelconv/fbx"
	"github.com/binzume/modelconv/geom"
	"github.com/binzume/modelconv/mqo"
	"github.com/binzume/modelconv/unity"
)

type UnityToMQOOption struct {
	SaveTexrure  bool
	ConvertScale float32
}

type UnityToMQOConverter struct {
	options *UnityToMQOOption
}

type unityToMqoState struct {
	UnityToMQOOption
	src *unity.Scene
	dst *mqo.Document
	mat map[string]struct {
		index   int
		uvScale *geom.Vector2
	}
}

func NewUnityToMQOConverter(options *UnityToMQOOption) *UnityToMQOConverter {
	if options == nil {
		options = &UnityToMQOOption{SaveTexrure: true}
	}
	if options.ConvertScale == 0 {
		options.ConvertScale = 1000
	}
	return &UnityToMQOConverter{
		options: options,
	}
}

func (conv *UnityToMQOConverter) Convert(secne *unity.Scene) (*mqo.Document, error) {
	state := unityToMqoState{
		UnityToMQOOption: *conv.options,
		src:              secne,
		dst:              mqo.NewDocument(),
		mat: map[string]struct {
			index   int
			uvScale *geom.Vector2
		}{},
	}

	for _, o := range secne.Objects {
		state.convertObject(o, 0, true)
	}

	if len(state.dst.Materials) == 0 {
		state.dst.Materials = append(state.dst.Materials, &mqo.Material{Name: "dummy", Color: geom.Vector4{X: 1, Y: 1, Z: 1, W: 1}})
	}

	return state.dst, nil
}

func (c *unityToMqoState) convertObject(o *unity.GameObject, d int, active bool) {
	dst := c.dst
	obj := mqo.NewObject(o.Name)
	active = active && o.IsActive != 0
	dst.Objects = append(dst.Objects, obj)
	obj.UID = len(dst.Objects)
	obj.Depth = d
	obj.Visible = active

	tr := o.GetTransform()
	if tr == nil {
		return
	}
	obj.Translation = tr.LocalPosition.Scale(c.ConvertScale)
	obj.Scale = &tr.LocalScale

	var meshFilter *unity.MeshFilter
	var meshRenderer *unity.MeshRenderer
	if o.GetComponent(&meshFilter) && o.GetComponent(&meshRenderer) && meshFilter.Mesh.IsValid() {
		var materials []int
		for _, materialRef := range meshRenderer.Materials {
			matGUID := materialRef.GUID
			if m, ok := c.mat[matGUID]; ok {
				materials = append(materials, m.index)
				continue
			}
			mat := struct {
				index   int
				uvScale *geom.Vector2
			}{len(dst.Materials), nil}
			materials = append(materials, mat.index)
			m := &mqo.Material{Name: matGUID, Color: geom.Vector4{X: 1, Y: 1, Z: 1, W: 1}, Diffuse: 0.8}
			material, err := unity.LoadMaterial(o.Scene.Assets, matGUID)
			if err == nil {
				if c := material.GetColorProperty("_Color"); c != nil {
					m.Color = geom.Vector4{X: c.R, Y: c.G, Z: c.B, W: c.A}
				}
				if c := material.GetColorProperty("_EmissionColor"); c != nil {
					emmision := &geom.Vector3{X: c.R, Y: c.G, Z: c.B}
					if emmision.Len() > 0 {
						m.EmissionColor = emmision
					}
				}
				m.Name = matGUID + "_" + material.Name

				if t := material.GetTextureProperty("_MainTex"); t != nil && t.Texture.IsValid() {
					texAsset := o.Scene.Assets.GetAsset(t.Texture.GUID)
					if c.SaveTexrure && texAsset != nil {
						m.Texture, err = c.saveTexrure(texAsset)
						if err != nil {
							log.Println(err)
						}
					}
					mat.uvScale = &t.Scale
				}
				if t := material.GetTextureProperty("_BumpMap"); t != nil && t.Texture.IsValid() {
					texAsset := o.Scene.Assets.GetAsset(t.Texture.GUID)
					if c.SaveTexrure && texAsset != nil {
						m.BumpTexture, err = c.saveTexrure(texAsset)
						if err != nil {
							log.Println(err)
						}
					}
				}
			}
			dst.Materials = append(dst.Materials, m)
			c.mat[matGUID] = mat
		}

		s := c.ConvertScale
		trmat := geom.NewScaleMatrix4(s, s, s).Mul(tr.GetWorldMatrix())
		if name, ok := unity.UnityMeshes[*meshFilter.Mesh]; ok {
			mat := struct {
				index   int
				uvScale *geom.Vector2
			}{}
			if len(meshRenderer.Materials) > 0 {
				mat = c.mat[meshRenderer.Materials[0].GUID]
			}
			obj.Name += name
			if name == "Cube" {
				Cube(obj, trmat, mat.index, mat.uvScale)
			} else if name == "Plane" {
				Plane(obj, trmat, mat.index, mat.uvScale)
			} else if name == "Quad" {
				Quad(obj, trmat, mat.index, mat.uvScale)
			} else {
				log.Println("TODO:", name)
			}
		} else {
			err := c.importMesh(meshFilter.Mesh, obj, materials, trmat)
			if err != nil {
				log.Println("Can not import mesh: ", obj.Name, err)
			}
		}
	}

	for _, child := range tr.GetChildren() {
		childObj := child.GetGameObject()
		if childObj == nil {
			log.Println("GameObject not found", child.GameObject)
			continue
		}
		c.convertObject(childObj, d+1, active)
	}
}

func (c *unityToMqoState) saveTexrure(texAsset *unity.Asset) (string, error) {
	fileName := "textures/" + texAsset.GUID + "_" + path.Base(texAsset.Path)
	_ = os.Mkdir("textures/", 0755)
	if _, err := os.Stat(fileName); err == nil {
		return fileName, nil
	}
	r, err := c.src.Assets.Open(texAsset.Path)
	if err != nil {
		return "", err
	}
	defer r.Close()
	w, err := os.Create(fileName)
	if err != nil {
		return "", err
	}
	defer w.Close()
	_, err = io.Copy(w, r)
	return fileName, err
}

func (c *unityToMqoState) importMesh(mesh *unity.Ref, obj *mqo.Object, materials []int, transform *geom.Matrix4) error {
	asset := c.src.Assets.GetAsset(mesh.GUID)
	if asset == nil {
		return fmt.Errorf("asset not found %s", mesh.GUID)
	}
	if !strings.HasSuffix(asset.Path, ".fbx") {
		return fmt.Errorf("not supported: %s", asset.Path)
	}

	obj.Name += "(FBX)"
	meta, err := c.src.Assets.GetMetaFile(asset)
	if err != nil {
		return err
	}
	log.Println("Import mesh:", asset, meta.GetRecycleNameByFileID(mesh.FileID))

	r, err := c.src.Assets.Open(asset.Path)
	if err != nil {
		return err
	}
	defer r.Close()
	doc, err := fbx.Parse(r)
	if err != nil {
		return err
	}
	scale := doc.GlobalSettings.GetProperty70("UnitScaleFactor").ToFloat32(1) * 0.01
	_, err = NewFBXToMQOConverter(&FBXToMQOOption{
		ObjectDepth:      obj.Depth + 1,
		TargetModelName:  meta.GetRecycleNameByFileID(mesh.FileID),
		MaterialOverride: materials,
		RootTransform:    transform.Mul(geom.NewScaleMatrix4(-scale, scale, -scale)),
	}).ConvertTo(c.dst, doc)
	return err
}

func AddGeometry(o *mqo.Object, tr *geom.Matrix4, mat int, vs []*geom.Vector3, faces [][]int, uvs [][]geom.Vector2, uvScale *geom.Vector2) {
	voffset := len(o.Vertexes)
	for _, v := range vs {
		o.Vertexes = append(o.Vertexes, tr.ApplyTo(v))
	}
	for n, f := range faces {
		for i := range f {
			f[i] += voffset
		}
		face := &mqo.Face{Material: mat, Verts: f}
		if len(uvs) > n {
			face.UVs = uvs[n]
			if uvScale != nil {
				for i := range face.UVs {
					face.UVs[i].X *= uvScale.X
					face.UVs[i].Y *= uvScale.Y
				}
			}
		}
		o.Faces = append(o.Faces, face)
	}
}

func Cube(o *mqo.Object, tr *geom.Matrix4, mat int, uvScale *geom.Vector2) {
	vs := []*geom.Vector3{
		{X: -0.5, Y: -0.5, Z: -0.5},
		{X: 0.5, Y: -0.5, Z: -0.5},
		{X: 0.5, Y: 0.5, Z: -0.5},
		{X: -0.5, Y: 0.5, Z: -0.5},
		{X: -0.5, Y: -0.5, Z: 0.5},
		{X: 0.5, Y: -0.5, Z: 0.5},
		{X: 0.5, Y: 0.5, Z: 0.5},
		{X: -0.5, Y: 0.5, Z: 0.5},
	}
	faces := [][]int{
		{0, 1, 2, 3}, {7, 6, 5, 4},
		{4, 5, 1, 0}, {3, 2, 6, 7},
		{2, 1, 5, 6}, {0, 3, 7, 4},
	}
	uvs := [][]geom.Vector2{
		{{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 1, Y: 1}, {X: 0, Y: 1}},
		{{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 1, Y: 1}, {X: 0, Y: 1}},
		{{X: 0, Y: 1}, {X: 1, Y: 1}, {X: 1, Y: 0}, {X: 0, Y: 0}},
		{{X: 0, Y: 1}, {X: 1, Y: 1}, {X: 1, Y: 0}, {X: 0, Y: 0}},
		{{X: 1, Y: 1}, {X: 1, Y: 0}, {X: 0, Y: 0}, {X: 0, Y: 1}},
		{{X: 1, Y: 1}, {X: 1, Y: 0}, {X: 0, Y: 0}, {X: 0, Y: 1}},
	}
	AddGeometry(o, tr, mat, vs, faces, uvs, uvScale)
}

func Quad(o *mqo.Object, tr *geom.Matrix4, mat int, uvScale *geom.Vector2) {
	vs := []*geom.Vector3{
		{X: -0.5, Y: -0.5},
		{X: 0.5, Y: -0.5},
		{X: -0.5, Y: 0.5},
		{X: 0.5, Y: 0.5},
	}
	faces := [][]int{
		{0, 1, 3, 2},
	}
	uvs := [][]geom.Vector2{
		{{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 0, Y: 1}, {X: 1, Y: 1}},
	}
	AddGeometry(o, tr, mat, vs, faces, uvs, uvScale)
}

func Plane(o *mqo.Object, tr *geom.Matrix4, mat int, uvScale *geom.Vector2) {
	vs := []*geom.Vector3{
		{X: -5, Y: 0, Z: -5},
		{X: 5, Y: 0, Z: -5},
		{X: 5, Y: 0, Z: 5},
		{X: -5, Y: 0, Z: 5},
	}
	faces := [][]int{
		{0, 1, 2, 3},
	}
	uvs := [][]geom.Vector2{
		{{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 1, Y: 1}, {X: 0, Y: 1}},
	}
	AddGeometry(o, tr, mat, vs, faces, uvs, uvScale)
}
