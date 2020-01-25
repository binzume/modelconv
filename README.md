# Experimental 3D model converter

Goで3Dモデルファイルを読み書きするライブラリ＆変換ツールです．

3DモデルをglTFやVRMに変換して，ブラウザ上で動かすために作ったものです．
いまのところテスト用なので実用には耐えません．

| format | read | write | comment |
| ------ | -- | -- | --- |
| .pmx | ○ | △ | ボーンやモーフはまだ書き出せません |
| .pmd | ○ |  | Read only |
| .vrm | △ |  △ | https://github.com/qmuntal/gltf 用のエクステンションです |
| .mqo | △ |  ○ | UTF-8以外の.mqoは読み込み失敗します |

VRM以外はデータを見ながら雰囲気で読み書きしてるので，おかしな挙動をするかもしれません．

## Usage:

PMX/PMDを.mqoに変換するサンプルプログラム．

```bash
cd modelconv
go get -d ./...
go run mmd2mqo.go "path_to.pmx"
```

# License

MIT License
