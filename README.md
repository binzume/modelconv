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

## Usage:

[cmd/modelconv](cmd/modelconv) : モデルデータを相互変換するサンプルプログラム．

新し目のGoがあればビルドできると思います．
[Releases](https://github.com/binzume/modelconv/releases/latest)からビルド済みのWindows用のバイナリがダウンロードできます．

```bash
go install github.com/binzume/modelconv/cmd/modelconv@latest
modelconv "path_to.pmx"
modelconv "path_to.pmx" "animation.vmd" "output.glb"
```

# License

MIT License
