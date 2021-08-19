# Experimental 3D model converter

Goで3Dモデルファイルを読み書きするライブラリ＆変換ツールです．

| Format     | Read | Write | Comment                          |
| ---------- | ---- | ----- | -------------------------------- |
| .mqo/.mqoz |  ○  |  ○   | ボーン・モーフに対応             |
| .gltf/.glb |  ○  |  ○   | 他フォーマットへの変換は暫定実装 |
| .vrm       |  △  |  ○   | glTF 用のエクステンション        |
| .pmx       |  ○  |  ○   | Physics未対応                    |
| .pmd       |  ○  |       | Read only                        |
| .vmd       |  △  |       | 暫定実装                         |

glTFの読み書きには https://github.com/qmuntal/gltf を使っています．

データを見ながら雰囲気で実装してるので，読み込めないファイルがあるかもしれません．

# Command-line tool

package: [cmd/modelconv](cmd/modelconv)

以下の組み合わせの変換が可能です．

- (.pmd | .pmx | .mqo | .mqoz) → (.pmx | .mqo| .mqoz | .glb | .gltf | .vrm)
- (.glb | .gltf | .vrm) → (.glb | .gltf | .vrm) (※1)

※1: glTF同士の変換は特別扱いをしているため，モデルに変更を加えるオプションは未対応です．(scaleは可能)

## Install "modelconv" commant

新し目のGoがあればビルドできると思います．
[Releases](https://github.com/binzume/modelconv/releases/latest)にビルド済みのWindows用の実行ファイルを置いてあります．

```bash
go install github.com/binzume/modelconv/cmd/modelconv@latest
```

## Usage examples

### MMD to VRM

```bash
modelconv -autotpose "右腕,左腕" "model.pmx" "model.vrm"
modelconv -vrmconfig "model.vrmconfig.json" "model.pmx" "model.vrm"
```

### gltf to glb

```bash
modelconv "model.gltf" "model.glb"
modelconv -format glb "model.gltf"
```

### Scaling

```bash
modelconv -scale 1.5 "model.glb" "model_scaled.glb"
modelconv -scaleY 1.5 -scaleX 1.3 "model.mqo" "model_scaled.mqo"
```

## Flags

| Flag       | Description    | Default    |
| ---------- | -------------- | ---------- |
| -format    | Output format  |            |
| -scale     | Scale          | See `Unit` |
| -scaleX    | Scale x-axis   | 1.0        |
| -scaleY    | Scale y-axis   | 1.0        |
| -scaleZ    | Scale z-axis   | 1.0        |
| -rot180    | rotate 180 degrees around Y-axis |  |
| -hide      | hide objects (OBJ1,OBJ2,...) |  |
| -hidemat   | hide materials (MAT1,MAT2,...)  |  |
| -unlit     | unlit materials (MAT1,MAT2,...)  |  |
| -vrmconfig | Config file for VRM | "inputfile.vrmconfig.json" |
| -autotpose | Arm bone names |            |
| -chparent  | replace parent bone (BONE1:PARENT1,BONE2:PARENT2,...) |  |

### vrmconfig:

設定ファイルのjsonは [converter/vrmconfig_presets](converter/vrmconfig_presets) にあるファイルや，
[Qiitaの記事](https://qiita.com/binzume/items/d29cd21b9860809f72cf)も参考にしてください．

MMDからの変換時にはデフォルトで [mmd.json](converter/vrmconfig_presets/mmd.json) が使われます．

### hide,hidemat,unlit:

対象のオブジェクトやマテリアルの名前をカンマ区切りで指定してください．ワイルドカード(`*`)が利用可能です．

### autotpose:

腕のボーンを指定するとX軸に沿うように形状を調整します(暫定実装)

### Unit:

- MQO: 1mm
- MMD: 80mm
- glTF/VRM: 1m

例： MMD → VRM : default scale = 0.08

# API

T.B.D.

### Example: .pmx to .mqoz

```go 
	mmdModel, err :=  mmd.Load("model.pmx")
	if err != nil {
		return err
	}
	mqDoc, err := converter.NewMMDToMQOConverter(nil).Convert(mmdModel)
	if err != nil {
		return err
	}
	err = mqo.Save(mqDoc, "model.mqoz")
	if err != nil {
		return err
	}
```

# License

MIT License
