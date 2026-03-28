# jstats

JSON データに対する SPL スタイルの `stats` コマンド。JSON 配列または JSONL ストリームを stdin から読み込み、1つ以上のフィールドでグループ化した集計を行います。

English documentation: [README.md](README.md)

## インストール

[リリースページ](https://github.com/nlink-jp/jstats/releases) からプラットフォームに合ったバイナリをダウンロードしてください。

```sh
unzip jstats-<version>-<os>-<arch>.zip
mv jstats /usr/local/bin/
```

ソースからビルドする場合（Go 1.24 以上が必要）：

```sh
go install github.com/nlink-jp/jstats@latest
```

## 使用方法

```
jstats [flags] <expression>

Flags:
  -format string   出力形式: json（デフォルト）, text, md, csv
  -version         バージョン情報を表示して終了
```

### 式の構文

```
func1, func2(field) [as alias], ... [by field1, field2, ...]
```

## 関数一覧

| 関数 | 説明 |
|---|---|
| `count` | COUNT(*) — グループ内の総行数 |
| `count(field)` | NULL 以外の値の件数 |
| `sum(field)` | 合計 |
| `min(field)` | 最小値 |
| `max(field)` | 最大値 |
| `avg(field)` | 平均 |
| `median(field)` | 中央値（= p50）|
| `stdev(field)` | 標本標準偏差 |
| `var(field)` | 標本分散 |
| `range(field)` | max − min |
| `p<N>(field)` | N パーセンタイル（0–100）例: `p95(latency)` |
| `dc(field)` | distinct count（重複排除した件数）|
| `first(field)` | 最初の値（入力順）|
| `last(field)` | 最後の値（入力順）|
| `mode(field)` | 最頻値 |
| `values(field)` | 重複排除した値の配列 |
| `list(field)` | 全値の配列 |

## 使用例

```bash
# ステータスコード別の件数
cat access.json | jstats "count by status"

# サービス別レイテンシのパーセンタイル
cat metrics.json | jstats "count, avg(latency), p95(latency), p99(latency), stdev(latency) by service"

# distinct count・値一覧・最頻値
cat events.json | jstats -format text "dc(user_id), values(action), mode(action) by host"

# グループ化なし — 全データの集計
cat data.json | jstats "count, min(score), max(score), avg(score), median(score)"

# エイリアス指定
cat sales.json | jstats -format md "sum(amount) as total, avg(amount) as avg_sale by region"

# JSONL 入力にも対応
cat events.jsonl | jstats "count by type"
```

## 出力形式

| フラグ | 出力 |
|---|---|
| `json`（デフォルト）| JSON 配列 |
| `text` | ASCII テーブル |
| `md` | Markdown テーブル |
| `csv` | CSV |

## ビルド

```sh
git clone https://github.com/nlink-jp/jstats.git
cd jstats
make build        # 現在のプラットフォーム向けにビルド → dist/jstats
make build-all    # 全プラットフォーム向けにクロスコンパイル → dist/
make package      # ビルドして .zip アーカイブを作成
make test         # テストを実行
make clean        # dist/ を削除
```

## 処理の仕組み

```
stdin（JSON 配列または JSONL）
        │
        ▼
  parseInput()          バイト列を []map[string]interface{} にパース
        │               [{"k":"v"}] と {"k":"v"}\n{"k":"v"} の両方に対応
        ▼
  parseExpr()           stats 式をトークナイズしてパース
        │               例: "count, avg(latency) by host"
        │               → StatsQuery{Funcs: [...], ByFields: [...]}
        ▼
  computeStats()        ByFields でグループ化し、各 AggFunc を適用
        │               グループは最初の出現順を維持（出力順が安定）
        │
        ├── count / sum / min / max / avg / range / dc
        │     数値またはテキストの直接イテレーション
        │
        ├── stdev / var
        │     2パス: 平均計算 → 二乗偏差の和 / (n-1)
        │
        ├── median / p<N>
        │     フィールド値のコピーをソート → 線形補間
        │
        ├── mode
        │     頻度マップ。同数の場合は最初に出現した値を返す
        │
        └── values / list / first / last
              入力順を保ちながらスライスに蓄積
        │
        ▼
    render()             結果行を json / text / md / csv に整形
        │
        ▼
     stdout
```

NULL および存在しないフィールドの値は、すべての数値関数でサイレントにスキップされます。`count(field)` は NULL 以外の値のみをカウントします。引数なしの `count` は常にグループ内の全行をカウントします。
