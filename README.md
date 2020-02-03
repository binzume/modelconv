# Experimental 3D model converter

Goで3Dモデルファイルを読み書きするライブラリ＆変換ツールです．

| Format | Comment |
| ------ | ------- |
| .pmx   | Physicsは未対応 |
| .pmd   | Read only |
| .glb   | https://github.com/qmuntal/gltf を使っています |
| .vrm   | https://github.com/qmuntal/gltf 用のエクステンションです |
| .mqo   | ボーン・モーフにも対応 |

データを見ながら雰囲気で実装してるので，おかしな挙動をするかもしれません．

## Usage:

PMX/PMDを.mqoに変換するサンプルプログラム．(上記の表の全フォーマットが扱えるわけではないです)

Pure golangなのでGoがあればビルドできると思います．
[Releases](https://github.com/binzume/modelconv/releases/latest)からビルド済みのWindows用のバイナリがダウンロードできます．

```bash
go get -u github.com/binzume/modelconv/cmd/modelconv
go build github.com/binzume/modelconv/cmd/modelconv
./modelconv "path_to.pmx"
```

# License

MIT License
