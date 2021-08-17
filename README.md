# Experimental 3D model converter

Goで3Dモデルファイルを読み書きするライブラリ＆変換ツールです．

| Format | Read | Write | Comment |
| ------ | ---- | ----- | ------- |
| .mqo   |  ○  |  ○  | ボーン・モーフにも対応 |
| .vrm   |      |  ○  | https://github.com/qmuntal/gltf 用のエクステンションです |
| .glb   |  △  |  ○  | https://github.com/qmuntal/gltf を使っています |
| .pmx   |  ○  |  ○  | Physicsは未対応 |
| .pmd   |  ○  |      | Read only |
| .vmd   |  △  |      |  |

データを見ながら雰囲気で実装してるので，おかしな挙動をするかもしれません．

# Command-line tool

package: [cmd/modelconv](cmd/modelconv) : モデルデータを相互変換するサンプルプログラム．

以下の組み合わせの変換ができます．

- (.pmd | .pmx | .mqo) → (.mqo | .pmx | .glb | .gltf | .vrm)
- (.glb | .gltf | .vrm) → (.glb | .gltf | .vrm) (※1)

(※1: glTF同士の変換は特別扱いをしているため，モデルに変更を加えるオプションは動きません．scaleは可能です)

新し目のGoがあればビルドできると思います．
[Releases](https://github.com/binzume/modelconv/releases/latest)からビルド済みのWindows用のバイナリがダウンロードできます．

```bash
go install github.com/binzume/modelconv/cmd/modelconv@latest
```

### Example: MMD to VRM

```bash
modelconv "model.pmx" "model.vrm"
modelconv "model.pmx" -format vrm
modelconv -vrmconfig "model.vrmconfig.json" "model.pmx" "model.vrm"
```

### Example: gltf to glb

```bash
modelconv "model.gltf" "model.glb"
modelconv "model.gltf" -format glb
```

### Flags

| Flag       | Description   | Default |
| ---------- | ------------- | ------- |
| -format    | output format | |
| -scale     | Scale         | 0:auto, see `Unit` |
| -scaleX    | Scale x-axis  | 1.0 |
| -scaleY    | Scale y-axis  | 1.0 |
| -scaleZ    | Scale z-axis  | 1.0 |
| -rot180    | rotate 180 degrees around Y-axis |  |
| -hide      | hide objects (OBJ1,OBJ2,...) |  |
| -hidemat   | hide materials (MAT1,MAT2,...)  |  |
| -unlit     | unlit materials (MAT1,MAT2,...)  |  |
| -vrmconfig | Config file for VRM | "inputfile.vrmconfig.json" |
| -autotpose | Arm bone names |  |
| -chparent  | replace parent bone (BONE1:PARENT1,BONE2:PARENT2,...) |  |

Unit:

- MQO: 1mm
- MMD: 80mm
- glTF/VRM: 1m

MQO to GLTG : default scale = 0.001

# License

MIT License
