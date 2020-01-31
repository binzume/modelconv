
3Dモデルファイルを変換するプログラムです．
使い方はQiitaの記事を参考にしてください．
https://qiita.com/binzume TODO: URL

Windows(x64) 以外のバイナリが必要な場合はお手数ですがご自身でビルドしてください．
https://github.com/binzume/modelconv

## Usage

### .pmd,.pmx → .mqo
./modelconv -rot180 -scale 80 input.pmx output.mqo

-rot180 : Y軸回りに100°回転します
-scale : 変換時にスケールをかけます
入力ファイル以外は省略可能なので，ファイルを実行ファイルにドラッグ＆ドロップすれば変換されます．

### .glb+.json → .vrm
./modelconv -config input.vrmconfig.json input.glb output.vrm

input.vrmconfig.json の書き方はQiitaの記事を参考にしてください．

# License

https://github.com/binzume/modelconv
  MIT License (https://github.com/binzume/modelconv/blob/master/LICENSE)

glTFの読み書きに以下のライブラリを使用しています．

qmuntal/gltf https://github.com/qmuntal/gltf
  BSD 2-Clause (https://github.com/qmuntal/gltf/blob/master/LICENSE)
