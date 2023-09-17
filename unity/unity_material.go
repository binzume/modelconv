package unity

import (
	"fmt"
	"io/ioutil"

	"github.com/binzume/modelconv/geom"
)

type Material struct {
	Name           string `yaml:"m_Name"`
	Shader         *Ref   `yaml:"m_Shader"`
	ShaderKeywords string `yaml:"m_ShaderKeywords"`

	LightmapFlags            int               `yaml:"m_LightmapFlags"`
	EnableInstancingVariants int               `yaml:"m_EnableInstancingVariants"`
	DoubleSidedGI            int               `yaml:"m_DoubleSidedGI"`
	CustomRenderQueue        int               `yaml:"m_CustomRenderQueue"`
	StringTagMap             map[string]string `yaml:"stringTagMap"`
	DisabledShaderPasses     []string          `yaml:"disabledShaderPasses"`

	SavedProperties struct {
		TexEnvs []map[string]*TextureEnv `yaml:"m_TexEnvs"`
		Floats  []map[string]float32     `yaml:"m_Floats"`
		Colors  []map[string]*Color      `yaml:"m_Colors"`
	} `yaml:"m_SavedProperties"`
}

type Color struct {
	R float32
	G float32
	B float32
	A float32
}

type TextureEnv struct {
	Texture *Ref         `yaml:"m_Texture"`
	Scale   geom.Vector2 `yaml:"m_Scale"`
	Offset  geom.Vector2 `yaml:"m_Offset"`
}

func (m *Material) GetTextureProperty(name string) *TextureEnv {
	for _, t := range m.SavedProperties.TexEnvs {
		if tex, ok := t[name]; ok {
			return tex
		}
	}
	return nil
}

func (m *Material) GetColorProperty(name string) *Color {
	for _, t := range m.SavedProperties.Colors {
		if col, ok := t[name]; ok {
			return col
		}
	}
	return nil
}

func (m *Material) GetFloatProperty(name string) (float32, bool) {
	for _, t := range m.SavedProperties.Floats {
		if col, ok := t[name]; ok {
			return col, true
		}
	}
	return 0, false
}

func LoadMaterial(assets Assets, guid string) (*Material, error) {
	asset := assets.GetAsset(guid)
	if asset == nil {
		return nil, fmt.Errorf("Material not found: %s", guid)
	}
	r, err := assets.Open(asset.Path)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	b, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	for _, doc := range ParseYamlDocuments(b) {
		if doc.Tag != "" && doc.Tag != "tag:unity3d.com,2011:21" {
			continue
		}
		var mat struct {
			Material Material `yaml:"Material"`
		}
		err = doc.Decode(&mat)
		if err == nil {
			return &mat.Material, err
		}
	}
	return nil, fmt.Errorf("Material not found: %s", asset.Path)
}
