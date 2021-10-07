package converter

import (
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
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
	matToGUID map[int]string

	// TODL: LRU cache
	lastFbx   *fbx.Document
	lastFbxID string
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
		matToGUID: map[int]string{},
	}

	s := state.ConvertScale
	transform := geom.NewScaleMatrix4(s, s, s)
	for _, o := range secne.Objects {
		state.convertObject(o, 0, transform, true)
	}

	if len(state.dst.Materials) == 0 {
		state.dst.Materials = append(state.dst.Materials, &mqo.Material{Name: "dummy", Color: geom.Vector4{X: 1, Y: 1, Z: 1, W: 1}})
	}

	return state.dst, nil
}

func (c *unityToMqoState) convertObject(o *unity.GameObject, d int, parentTransform *geom.Matrix4, active bool) {
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
	transform := parentTransform.Mul(tr.GetMatrix())

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
			c.matToGUID[len(dst.Materials)] = matGUID
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
				m.Name = material.Name

				if t := material.GetTextureProperty("_MainTex"); t != nil && t.Texture.IsValid() {
					texAsset := c.src.Assets.GetAsset(t.Texture.GUID)
					if c.SaveTexrure && texAsset != nil {
						m.Texture, err = c.saveTexrure(c.src.Assets, texAsset)
						if err != nil {
							log.Println(err)
						}
					}
					mat.uvScale = &t.Scale
				}
				if t := material.GetTextureProperty("_BumpMap"); t != nil && t.Texture.IsValid() {
					texAsset := c.src.Assets.GetAsset(t.Texture.GUID)
					if c.SaveTexrure && texAsset != nil {
						m.BumpTexture, err = c.saveTexrure(c.src.Assets, texAsset)
						if err != nil {
							log.Println(err)
						}
					}
				}
			}
			dst.Materials = append(dst.Materials, m)
			c.mat[matGUID] = mat
		}

		meshObjectIndex := len(dst.Objects)
		if name, ok := unity.UnityMeshes[*meshFilter.Mesh]; ok {
			meshObjectIndex--
			mat := 0
			if len(materials) > 0 {
				mat = materials[0]
			}
			obj.Name += "(" + name + ")"
			if name == "Cube" {
				Cube(obj, transform, mat)
			} else if name == "Plane" {
				Plane(obj, transform, mat)
			} else if name == "Quad" {
				Quad(obj, transform, mat)
			} else {
				log.Println("TODO:", name)
			}
		} else {
			err := c.importMesh(meshFilter.Mesh, obj, materials, transform)
			if err != nil {
				log.Println("Can not import mesh: ", obj.Name, err)
			}
		}
		for meshObjectIndex < len(dst.Objects) {
			obj := dst.Objects[meshObjectIndex]
			obj.Visible = active && meshRenderer.Enabled != 0
			for _, face := range obj.Faces {
				scale := c.mat[c.matToGUID[face.Material]].uvScale
				if scale != nil {
					for i := range face.UVs {
						face.UVs[i].X *= scale.X
						face.UVs[i].Y *= scale.Y
					}
				}
			}
			meshObjectIndex++
		}
	}

	for _, child := range tr.GetChildren() {
		childObj := child.GetGameObject()
		if childObj == nil {
			log.Println("GameObject not found", child.GameObject, tr.CorrespondingSourceObject)
			continue
		}
		c.convertObject(childObj, d+1, transform, active)
	}
}

func (c *unityToMqoState) saveTexrure(assets unity.Assets, texAsset *unity.Asset) (string, error) {
	texDir := filepath.Join(filepath.Dir(assets.GetSourcePath()), "saved_textures")
	texPath := filepath.Join(texDir, texAsset.GUID+"_"+path.Base(texAsset.Path))
	texName := "saved_textures/" + texAsset.GUID + "_" + path.Base(texAsset.Path)
	_ = os.Mkdir(texDir, 0755)
	if _, err := os.Stat(texPath); err == nil {
		return texName, nil
	}
	r, err := c.src.Assets.Open(texAsset.Path)
	if err != nil {
		return "", err
	}
	defer r.Close()
	w, err := os.Create(texPath)
	if err != nil {
		return "", err
	}
	defer w.Close()
	_, err = io.Copy(w, r)
	return texName, err
}

func (c *unityToMqoState) importMesh(mesh *unity.Ref, obj *mqo.Object, materials []int, transform *geom.Matrix4) error {
	asset := c.src.Assets.GetAsset(mesh.GUID)
	if asset == nil {
		return fmt.Errorf("asset not found %s", mesh.GUID)
	}
	meta, err := c.src.Assets.GetMetaFile(asset)
	if err != nil {
		return err
	}
	log.Println("Import mesh:", asset, meta.GetRecycleNameByFileID(mesh.FileID))

	if c.lastFbxID != mesh.GUID { // TODO
		if !strings.HasSuffix(asset.Path, ".fbx") {
			return fmt.Errorf("not supported: %s", asset.Path)
		}

		r, err := c.src.Assets.Open(asset.Path)
		if err != nil {
			return err
		}
		defer r.Close()
		doc, err := fbx.Parse(r)
		if err != nil {
			return err
		}
		c.lastFbx = doc
		c.lastFbxID = mesh.GUID
	}
	doc := c.lastFbx

	obj.Name += "(FBX)"
	objectIdx := len(c.dst.Objects)
	scale := doc.GlobalSettings.GetProperty("UnitScaleFactor").ToFloat32(1) * 0.01
	_, err = NewFBXToMQOConverter(&FBXToMQOOption{
		ObjectDepth:      obj.Depth + 1,
		TargetModelName:  meta.GetRecycleNameByFileID(mesh.FileID),
		MaterialOverride: materials,
		RootTransform:    transform.Mul(geom.NewScaleMatrix4(-scale, scale, -scale)),
	}).ConvertTo(c.dst, doc)
	if len(c.dst.Objects) == objectIdx+1 {
		c.dst.Objects[objectIdx].SharedGeometryHint = &mqo.SharedGeometryHint{
			Key:       mesh.GUID + fmt.Sprint(mesh.FileID),
			Transform: transform.Mul(geom.NewScaleMatrix4(-scale, scale, -scale)),
		}
	}
	return err
}

func AddGeometry(o *mqo.Object, tr *geom.Matrix4, mat int, vs []*geom.Vector3, faces [][]int, uvs [][]geom.Vector2) {
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
		}
		o.Faces = append(o.Faces, face)
	}
}

func Cube(o *mqo.Object, tr *geom.Matrix4, mat int) {
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
	AddGeometry(o, tr, mat, vs, faces, uvs)
}

func Quad(o *mqo.Object, tr *geom.Matrix4, mat int) {
	vs := []*geom.Vector3{
		{X: -0.5, Y: -0.5},
		{X: 0.5, Y: -0.5},
		{X: -0.5, Y: 0.5},
		{X: 0.5, Y: 0.5},
	}
	faces := [][]int{
		{1, 0, 2, 3},
	}
	uvs := [][]geom.Vector2{
		{{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 0, Y: 1}, {X: 1, Y: 1}},
	}
	AddGeometry(o, tr, mat, vs, faces, uvs)
}

func Plane(o *mqo.Object, tr *geom.Matrix4, mat int) {
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
	AddGeometry(o, tr, mat, vs, faces, uvs)
}
