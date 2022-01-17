package unity

import (
	"reflect"
)

var UnityMeshes = map[Ref]string{
	{FileID: 10202, GUID: "0000000000000000e000000000000000"}: "Cube",
	{FileID: 10206, GUID: "0000000000000000e000000000000000"}: "Cylinder",
	{FileID: 10207, GUID: "0000000000000000e000000000000000"}: "Sphere",
	{FileID: 10208, GUID: "0000000000000000e000000000000000"}: "Capsule",
	{FileID: 10209, GUID: "0000000000000000e000000000000000"}: "Plane",
	{FileID: 10210, GUID: "0000000000000000e000000000000000"}: "Quad",
}

var UnityShaders = map[Ref]string{
	{FileID: 45, GUID: "0000000000000000f000000000000000"}:      "StandardSpecularSetup",
	{FileID: 46, GUID: "0000000000000000f000000000000000"}:      "Standard",
	{FileID: 47, GUID: "0000000000000000f000000000000000"}:      "AutodeskInteractive ",
	{FileID: 4800000, GUID: "933532a4fcc9baf4fa0491de14d08ed7"}: "URP/Lit ",
	{FileID: 4800000, GUID: "650dd9526735d5b46b79224bc6e94025"}: "URP/Unlit ",
	{FileID: 4800000, GUID: "8d2bb70cbf9db8d4da26e15b26e74248"}: "URP/SimpleLit ",
}

type Scene struct {
	GUID    string
	Objects []*GameObject

	Elements map[int64]Element

	Assets Assets
}

func NewScene(assets Assets, guid string) *Scene {
	return &Scene{GUID: guid, Elements: map[int64]Element{}, Assets: assets}
}

func (s *Scene) GetElement(ref *Ref) Element {
	return s.GetElement2(ref, nil)
}
func (s *Scene) GetElement2(ref *Ref, prefabInstance *Ref) Element {
	if !ref.IsValid() {
		return nil
	}
	if ref.GUID != "" && ref.GUID != s.GUID && prefabInstance.IsValid() {
		if prefab, ok := s.Elements[prefabInstance.FileID].(*PrefabInstance); ok {
			return prefab.PrefabScene.GetElement(ref)
		}
		return nil
	}
	return s.Elements[ref.FileID]
}

func (s *Scene) GetGameObject(ref *Ref) *GameObject {
	t, _ := s.GetElement(ref).(*GameObject)
	return t
}

func (s *Scene) GetTransform(ref *Ref, prefabInstance *Ref) *Transform {
	if t, ok := s.GetElement2(ref, prefabInstance).(*Transform); ok {
		// stripped.
		if !t.GameObject.IsValid() && t.CorrespondingSourceObject.IsValid() {
			if t2, ok := t.Scene.GetElement2(t.CorrespondingSourceObject, t.PrefabInstance).(*Transform); ok {
				return t2
			}
		}
		return t
	}
	return nil
}

type Ref struct {
	FileID int64  `yaml:"fileID"`
	GUID   string `yaml:"guid"`
	Type   int    `yaml:"type"`
}

type Element interface{}

func (r *Ref) IsValid() bool {
	return r != nil && r.FileID != 0
}

type GameObject struct {
	Name      string `yaml:"m_Name"`
	IsActive  int    `yaml:"m_IsActive"`
	TagString string `yaml:"m_TagString"`

	Components []*struct {
		Ref Ref `yaml:"component"`
	} `yaml:"m_Component"`

	CorrespondingSourceObject *Ref `yaml:"m_CorrespondingSourceObject"`
	PrefabInstance            *Ref `yaml:"m_PrefabInstance"`

	Scene *Scene `yaml:"-"`
}

func (o *GameObject) init(scene *Scene) {
	o.Scene = scene
}

func (o *GameObject) GetComponent(target interface{}) bool {
	typ := reflect.TypeOf(target).Elem()
	for _, c := range o.Components {
		component := o.Scene.GetElement2(&c.Ref, o.PrefabInstance)
		if reflect.TypeOf(component) == typ {
			reflect.ValueOf(target).Elem().Set(reflect.ValueOf(component))
			return true
		}
	}
	return false
}

func (o *GameObject) GetComponents(target interface{}) {
	typ := reflect.TypeOf(target).Elem().Elem()
	value := reflect.ValueOf(target).Elem()
	for _, c := range o.Components {
		component := o.Scene.GetElement2(&c.Ref, o.PrefabInstance)
		if reflect.TypeOf(component) == typ {
			value.Set(reflect.Append(value, reflect.ValueOf(component)))
		}
	}
}

func (o *GameObject) GetTransform() *Transform {
	for _, c := range o.Components {
		if t, ok := o.Scene.GetElement2(&c.Ref, o.PrefabInstance).(*Transform); ok {
			return t
		}
	}
	return nil
}

type PrefabInstance struct {
	Modification struct {
		Modifications []struct {
			Target          *Ref        `yaml:"target"`
			PropertyPath    string      `yaml:"propertyPath"`
			Value           interface{} `yaml:"value"`
			ObjectReference *Ref        `yaml:"objectReference"`
		} `yaml:"m_Modifications"`
		TransformParent   *Ref          `yaml:"m_TransformParent"`
		RemovedComponents []interface{} `yaml:"m_RemovedComponents"`
	} `yaml:"m_Modification"`

	SourcePrefab *Ref   `yaml:"m_SourcePrefab"`
	PrefabScene  *Scene `yaml:"-"`
}
