# Go Workflow Step Order Debug Lab

同一ワークフロー内で同じ順序番号のステップを登録できてしまう、業務ロジック上の一意性バグを再現・修正するための最小Goプロジェクトです。

> このリポジトリは学習用の再現コードです。実運用の認証、ワークフロー定義管理、マイグレーション運用は含みません。

## 扱うバグ

ワークフローの開始ステップと次ステップは `order` で決める設計です。このとき同じ `workflow_id` に `order = 1` を二つ登録すると、どちらが最初のステップかを一意に決められません。

| 操作 | バグ状態 | 修正後 |
| --- | --- | --- |
| 同一ワークフローに同じ順序を登録 | `201 Created` | `409 Conflict` |
| 同じ `(workflow_id, order)` の保存件数 | 2件 | 1件 |
| 別ワークフローに同じ順序を登録 | `201 Created` | `201 Created` |

## 前提条件

Go 1.23.11 と C コンパイラを用意してください。SQLiteドライバを使うため、テストは `CGO_ENABLED=1` で実行します。

```bash
CGO_ENABLED=1 go test ./...
```

## バグを再現する

バグ状態はコミット `e5695b9` に残しています。テストは実際の Gin HTTP ハンドラー、GORM、インメモリSQLiteを通して、HTTP応答と最終DB状態を確認します。

```bash
git checkout e5695b9
CGO_ENABLED=1 go test ./... -count=1 -v
```

次の失敗が観測されます。

```text
status: expected 409, actual 201
persisted count: expected 1, actual 2
```

## 修正を確認する

修正後の `main` では、`workflow_id` と `order` に複合一意制約を置き、SQLiteの制約違反をドメインエラーへ変換してHTTP 409を返します。

```bash
git switch main
CGO_ENABLED=1 go test ./... -count=1 -v
```

## ディレクトリ構成

| パス | 役割 |
| --- | --- |
| `internal/workflowstep/model.go` | `WorkflowStep` と複合一意制約を定義します。 |
| `internal/workflowstep/repository.go` | 永続化とSQLiteの一意制約違反の変換を担当します。 |
| `internal/workflowstep/handler.go` | `POST /workflow-steps` でHTTP 409を返します。 |
| `internal/workflowstep/handler_test.go` | HTTP応答、永続化状態、別ワークフロー境界を検証します。 |
| `docs/debugging-record.md` | 失敗観測、原因、修正、制約を記録します。 |

## 設計上の制約

このプロジェクトは順序重複だけを扱います。連番の強制、途中挿入時の再採番、完了済みワークフローの編集可否などは、別の業務ルールとして設計・テストしてください。

## 参考資料

- [GORM: Database Indexes](https://gorm.io/docs/indexes.html)
- [RFC 9110: 409 Conflict](https://www.rfc-editor.org/rfc/rfc9110#section-15.5.10)
