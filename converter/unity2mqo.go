package converter

import (
	"io"
	"log"
	"os"
	"path"

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
	mat map[string]int
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
		mat:              map[string]int{},
	}

	// TODO
	state.dst.Materials = append(state.dst.Materials, &mqo.Material{Name: "dummy", Color: geom.Vector4{X: 1, Y: 1, Z: 1, W: 1}})

	for _, o := range secne.Objects {
		state.convertObject(o, 0)
	}

	return state.dst, nil
}

func (c *unityToMqoState) convertObject(o *unity.GameObject, d int) {
	dst := c.dst
	obj := mqo.NewObject(o.Name)
	dst.Objects = append(dst.Objects, obj)
	obj.UID = len(dst.Objects)
	obj.Depth = d

	tr := o.GetTransform()
	if tr == nil {
		return
	}

	var meshFilter *unity.MeshFilter
	var meshRenderer *unity.MeshRenderer
	if o.GetComponent(&meshFilter) && o.GetComponent(&meshRenderer) {
		mat := 0
		if len(meshRenderer.Materials) > 0 {
			matGUID := meshRenderer.Materials[0].GUID
			mat = c.mat[matGUID]
			if mat == 0 {
				mat = len(dst.Materials)
				m := &mqo.Material{Name: matGUID, Color: geom.Vector4{X: 1, Y: 1, Z: 1, W: 1}, Diffuse: 0.8}
				c.mat[matGUID] = mat
				dst.Materials = append(dst.Materials, m)
				material, err := unity.LoadMaterial(o.Scene.Assets, matGUID)
				log.Println(material)
				if err == nil {
					if c := material.GetColorProperty("_Color"); c != nil {
						m.Color = geom.Vector4{X: c.R, Y: c.G, Z: c.B, W: c.A}
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
			}
		}
		s := c.ConvertScale
		trmat := geom.NewScaleMatrix4(s, s, s).Mul(tr.GetWorldMatrix())
		if name, ok := unity.UnityMeshes[*meshFilter.Mesh]; ok {
			obj.Name += name
			if name == "Cube" {
				Cube(obj, trmat, mat)
			} else if name == "Plane" {
				Plane(obj, trmat, mat)
			} else if name == "Quad" {
				Quad(obj, trmat, mat)
			} else {
				log.Println("TODO:", name)
			}
		} else {
			asset := c.src.Assets.GetAsset(meshFilter.Mesh.GUID)
			if asset != nil {
				meta, _ := c.src.Assets.GetMetaFile(asset)
				log.Println("TODO:", asset.Path, meta.GetRecycleNameByFileID(meshFilter.Mesh.FileID))
			}
		}
	}

	for _, child := range tr.GetChildren() {
		c.convertObject(child.GetGameObject(), d+1)
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
		{X: 0.5, Y: 0.5},
		{X: -0.5, Y: 0.5},
	}
	faces := [][]int{
		{0, 1, 2, 3},
	}
	uvs := [][]geom.Vector2{
		{{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 1, Y: 1}, {X: 0, Y: 1}},
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
