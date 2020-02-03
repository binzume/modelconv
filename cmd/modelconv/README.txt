
3Dモデルファイルを変換するプログラムです．
使い方はQiitaの記事を参考にしてください．
https://qiita.com/binzume/items/d29cd21b9860809f72cf

Windows(x64) 以外のバイナリが必要な場合はお手数ですがご自身でビルドしてください．
https://github.com/binzume/modelconv

## Usage

### .pmd,.pmx → .mqo
./modelconv -rot180 -scale 80 input.pmx output.mqo

-rot180 : Y軸回りに100°回転します
-scale : 変換時にスケールをかけます
入力ファイル以外は省略可能なので，ファイルを実行ファイルにドラッグ＆ドロップすれば変換されます．

### .glb+.json → .vrm
./modelconv -vrmconfig input.vrmconfig.json input.glb output.vrm

input.vrmconfig.json の書き方はQiitaの記事を参考にしてください．

### .mqo → .glb
./modelconv input.mqo outout.glb

モーフ未対応．
メタセコイアはmm単位，glTFはm単位として扱います(デフォルトで1/1000のサイズになります)

# License

https://github.com/binzume/modelconv
  MIT License (https://github.com/binzume/modelconv/blob/master/LICENSE)

glTFの読み書きに以下のライブラリを使用しています．

qmuntal/gltf https://github.com/qmuntal/gltf
  BSD 2-Clause (https://github.com/qmuntal/gltf/blob/master/LICENSE)
