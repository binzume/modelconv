package unity

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"
)

type Assets interface {
	GetSourcePath() string
	GetAsset(guid string) *Asset
	GetAssetByPath(assetPath string) *Asset
	GetAllAssets() []*Asset
	GetMetaFile(asset *Asset) (*MetaFile, error)
	Open(assetPath string) (fs.File, error)
	Close() error
}
type Asset struct {
	GUID string
	Path string
}

type MetaFile struct {
	FileFormatVersion    int                    `yaml:"fileFormatVersion"`
	GUID                 string                 `yaml:"guid"`
	NativeFormatImporter map[string]interface{} `yaml:"NativeFormatImporter"`
	TextureImporter      map[string]interface{} `yaml:"TextureImporter"`
	ModelImporter        *struct {
		FileIDToRecycleName map[int64]string `yaml:"fileIDToRecycleName"`
	} `yaml:"ModelImporter"`
	RawData map[string]interface{} `yaml:",inline"`
}

func (m *MetaFile) GetRecycleNameByFileID(fileID int64) string {
	if m.ModelImporter == nil || m.ModelImporter.FileIDToRecycleName == nil {
		return ""
	}
	return m.ModelImporter.FileIDToRecycleName[fileID]
}

// OpenProject opens Unity project.
func OpenProject(projectDir string) (Assets, error) {
	return scanAssets(filepath.Join(projectDir, "Assets"))
}

// OpenAssets opens Assets dir.
func OpenAssets(assetsDir string) (Assets, error) {
	return scanAssets(assetsDir)
}

// OpenPackage opens .unitypackage file or dir.
func OpenPackage(packagePath string) (Assets, error) {
	stat, err := os.Stat(packagePath)
	if err != nil {
		return nil, err
	}
	if stat.IsDir() {
		return scanPackage(packagePath, false, packagePath)
	}
	tmpDir, err := ioutil.TempDir("", "modelconv_assets_")
	if err != nil {
		return nil, err
	}
	err = extractPackage(packagePath, tmpDir)
	if err != nil {
		return nil, err
	}
	return scanPackage(tmpDir, true, packagePath)
}

type assets struct {
	Assets       map[string]*Asset
	AssetsByPath map[string]*Asset
}

func (a *assets) GetAsset(guid string) *Asset {
	return a.Assets[guid]
}

func (a *assets) GetAssetByPath(path string) *Asset {
	return a.AssetsByPath[filepath.ToSlash(path)]
}

func (a *assets) GetAllAssets() []*Asset {
	var assets []*Asset
	for _, a := range a.Assets {
		assets = append(assets, a)
	}
	return assets
}

type packageFs struct {
	assets
	OriginalPackagePath string
	PackageDir          string
	Temp                bool
	HideMetaFile        bool
}

func (a *packageFs) GetSourcePath() string {
	return a.OriginalPackagePath
}

func (a *packageFs) Open(path string) (fs.File, error) {
	asset := a.AssetsByPath[path]
	if asset == nil {
		return nil, fs.ErrNotExist
	}
	return os.Open(filepath.Join(a.PackageDir, asset.GUID, "asset"))
}

func (a *packageFs) GetMetaFile(asset *Asset) (*MetaFile, error) {
	r, err := os.Open(filepath.Join(a.PackageDir, asset.GUID, "asset.meta"))
	if err != nil {
		return nil, err
	}
	defer r.Close()

	var meta MetaFile
	err = yaml.NewDecoder(r).Decode(&meta)
	return &meta, err
}

func (a *packageFs) Close() error {
	if a.Temp {
		return os.RemoveAll(a.PackageDir)
	}
	return nil
}

func scanPackage(packageDir string, tmp bool, originalPath string) (*packageFs, error) {
	ent, err := os.ReadDir(packageDir)
	if err != nil {
		return nil, err
	}

	pkg := &packageFs{
		assets: assets{
			Assets:       map[string]*Asset{},
			AssetsByPath: map[string]*Asset{},
		},
		Temp:                tmp,
		PackageDir:          packageDir,
		OriginalPackagePath: originalPath,
	}
	for _, f := range ent {
		if !f.IsDir() {
			continue
		}
		pathname := filepath.Join(packageDir, f.Name(), "pathname")
		if _, err = os.Stat(pathname); err != nil {
			continue
		}
		r, err := os.Open(pathname)
		if err != nil {
			return nil, err
		}
		defer r.Close()
		b, err := ioutil.ReadAll(r)
		if err != nil {
			return nil, err
		}
		path := string(b)
		guid := f.Name()
		asset := &Asset{
			GUID: guid,
			Path: path,
		}
		pkg.Assets[guid] = asset
		pkg.AssetsByPath[path] = asset
	}

	return pkg, nil
}

func extractPackage(pacakage, dst string) error {
	r, err := os.Open(pacakage)
	if err != nil {
		return err
	}
	defer r.Close()
	// TODO: always tar.gz?
	gzr, err := gzip.NewReader(r)
	if err != nil {
		return err
	}
	defer gzr.Close()
	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()
		switch {

		case err == io.EOF:
			return nil

		case err != nil:
			return err

		case header == nil:
			continue
		}

		name := filepath.Join(dst, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			if _, err := os.Stat(name); err != nil {
				if err := os.MkdirAll(name, 0755); err != nil {
					return err
				}
			}
		case tar.TypeReg:
			f, err := os.Create(name)
			if err != nil {
				return err
			}
			if _, err := io.Copy(f, tr); err != nil {
				f.Close()
				return err
			}
			f.Close()
		}
	}
}

type assetsFs struct {
	assets
	ProjectDir   string
	AssetsDir    string
	HideMetaFile bool
}

func (a *assetsFs) GetSourcePath() string {
	return a.AssetsDir
}

func (a *assetsFs) Open(path string) (fs.File, error) {
	return os.Open(filepath.Join(a.ProjectDir, path))
}

func (a *assetsFs) GetMetaFile(asset *Asset) (*MetaFile, error) {
	return a.loadMetaFile(asset.Path)
}

func (a *assetsFs) loadMetaFile(assetPath string) (*MetaFile, error) {
	r, err := os.Open(filepath.Join(a.ProjectDir, assetPath+".meta"))
	if err != nil {
		return nil, err
	}
	defer r.Close()

	var meta MetaFile
	err = yaml.NewDecoder(r).Decode(&meta)
	return &meta, err
}

func (a *assetsFs) Close() error {
	return nil
}

func scanAssets(assetsDir string) (*assetsFs, error) {
	assets := &assetsFs{
		assets: assets{
			Assets:       map[string]*Asset{},
			AssetsByPath: map[string]*Asset{},
		},
		ProjectDir: filepath.Dir(assetsDir),
		AssetsDir:  assetsDir,
	}
	err := filepath.Walk(assetsDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !strings.HasSuffix(path, ".meta") {
			return nil
		}
		path, _ = filepath.Rel(assets.ProjectDir, strings.TrimSuffix(path, ".meta"))
		meta, err := assets.loadMetaFile(path)
		if err != nil {
			log.Println(path, err)
		} else {
			guid := meta.GUID
			asset := &Asset{
				GUID: guid,
				Path: path,
			}
			assets.Assets[guid] = asset
			assets.AssetsByPath[filepath.ToSlash(path)] = asset
		}
		return nil
	})
	return assets, err
}
