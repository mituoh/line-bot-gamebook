# LINEでゲームブック
LINEで選択肢を選んだり、位置情報や画像を送ることで進行するゲームを作ることができるフレームワーク。

## トークスクリプト
LINEトークを制御する独自のトークスクリプト・フォーマット

### 機能
[-] botに喋らせる
[-] 分岐(GOTO)させる
[ ] 指定待ち時間後に喋らせる機能

### 文法
- "@" は予約語を設定
  - @buttons : 質問して分岐させる。@endまでの選択肢を抽出。
- "[]" で実行
  - [*hoge] : *hogeまでジャンプ
- "*" でアドレス設定

### 記述例
```
*start
こんにちは
これは文法の例です
1行1行、200msくらい待ちながら出力します。
@buttons 分岐機能もできます
[*state1] 分岐1に行きます
[*state2] 分岐2に行きます
@end

*state1
普通のルートです
あまりに普通です
[*common]

*state2
やばいルートです
危険すぎます
[*common]

*common
おしまいです
```
