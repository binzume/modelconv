# Experimental 3D model converter

Goで3Dモデルファイルを読み書きするライブラリ＆変換ツールです．

| format | read | write | comment |
| ------ | -- | -- | --- |
| .pmx | ○ | ○ | Physicsは未対応 |
| .pmd | ○ |  | Read only |
| .vrm | △ |  △ | https://github.com/qmuntal/gltf 用のエクステンションです |
| .mqo | ○ |  ○ | UTF-8以外の.mqoを読むと文字化けします |

VRM以外はデータを見ながら雰囲気で読み書きしてるので，おかしな挙動をするかもしれません．
いまのところテスト用なので実用には耐えません．

## Usage:

PMX/PMDを.mqoに変換するサンプルプログラム．

```bash
cd modelconv
go get -d ./...
go run ./cmd/modelconv "path_to.pmx"
```

# License

MIT License
