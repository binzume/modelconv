# Experimental 3D model converter

Goで3Dモデルファイルを読み書きするライブラリ＆変換ツールです．

| Format | Comment |
| ------ | --- |
| .pmx   | Physicsは未対応 |
| .pmd   | Read only |
| .vrm   | https://github.com/qmuntal/gltf 用のエクステンションです |
| .mqo   | UTF-8以外の.mqoを読むと文字化けします |

データを見ながら雰囲気で実装してるので，おかしな挙動をするかもしれません．

## Usage:

PMX/PMDを.mqoに変換するサンプルプログラム．(上記の表の全フォーマットが扱えるわけではないです)

```bash
cd modelconv
go get -d ./...
go build ./cmd/modelconv
./modelconv "path_to.pmx"
```

# License

MIT License
