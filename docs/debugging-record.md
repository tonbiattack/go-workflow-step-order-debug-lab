# デバッグ記録：ワークフローステップ順序の重複を拒否する

## 目的

同一ワークフロー内では `order` が一意であることを業務契約とします。この順序は開始ステップと進行先を決めるため、同じ順序の複数ステップを保存してはいけません。HTTP APIが重複を拒否し、最終DB状態に重複行を残さないことを確認します。

## 最初に観測した事実

バグ状態では、`approval-flow` に `order = 1` の「申請」が存在する状態で、同じ順序の「承認」を `POST /workflow-steps` へ送信しました。期待するHTTP応答は409、該当する `(workflow_id, order)` の行数は1件です。

| 観測項目 | 実際の結果 | 根拠 |
| --- | --- | --- |
| HTTP応答 | `201 Created` | `TestCreateWorkflowStep_DuplicateOrderMustBeRejected` |
| 最終DB状態 | 同じ `(workflow_id, order)` の行が2件 | 同テストでの `CountByWorkflowAndOrder` |
| 別ワークフローで同じ順序を使う要求 | `201 Created` | `TestCreateWorkflowStep_SameOrderAcrossWorkflowsIsAllowed` |

```text
status: expected 409, actual 201
persisted count: expected 1, actual 2
```

この再現ではHTTP応答だけでなく、テスト後にDBを読み直して件数を確認します。したがって、APIが成功を返したことと、重複データが最終的に保存されたことを分けて確認できます。

## 仮説と切り分け

| 仮説 | 確認方法 | 結果 |
| --- | --- | --- |
| JSONバインドまたはルーティングが登録処理を妨げている | HTTP応答とDB件数を確認する | 否定。201が返り、重複行も2件保存された。 |
| `order` は全ワークフローで一意にすべき | 別ワークフローで `order = 1` を作成する | 否定。別ワークフローでは同じ順序を許可する。 |
| DBに同一ワークフロー内の順序を守る制約がない | バグ状態のモデルタグと失敗テストを確認する | 採用。複合一意制約がなかった。 |

## 原因

バグ状態の `WorkflowStep` は `WorkflowID` と `Order` を必須フィールドとしているだけで、組として一意にする制約がありませんでした。ハンドラーはリクエストをバインド後にそのまま保存するため、DBは重複行を受け入れ、APIは201を返しました。

GORMでは、同じ名前の `uniqueIndex` を複数フィールドに指定することで複合一意インデックスを作成できます。[GORM Database Indexes](https://gorm.io/docs/indexes.html)

## 修正

`WorkflowID` と `Order` に同じ複合一意インデックス名を指定しました。SQLiteの一意制約違反を `ErrWorkflowStepOrderAlreadyExists` に変換し、ハンドラーはHTTP 409を返します。

```go
type WorkflowStep struct {
    WorkflowID string `gorm:"type:char(36);not null;uniqueIndex:idx_workflow_step_order,priority:1"`
    Order      int    `gorm:"not null;uniqueIndex:idx_workflow_step_order,priority:2"`
}

if errors.Is(err, ErrWorkflowStepOrderAlreadyExists) {
    c.JSON(http.StatusConflict, gin.H{"error": "workflow step order already exists"})
    return
}
```

アプリケーション層だけの事前検索ではなくDB制約を使うことで、将来追加される別の書き込み経路でも順序の重複保存を防げます。

## 再発防止テスト

`TestCreateWorkflowStep_DuplicateOrderMustBeRejected` は、重複登録が409になることと、DBの該当行数が1件のままであることを確認します。`TestCreateWorkflowStep_SameOrderAcrossWorkflowsIsAllowed` は、全体一意制約へ過剰に修正していないことを確認します。

## 再現手順

```bash
git checkout e5695b9
CGO_ENABLED=1 go test ./... -count=1 -v

git switch main
CGO_ENABLED=1 go test ./... -count=1 -v
```

## 制約

この教材で扱うのは順序重複だけです。欠番、途中挿入、並べ替え、ワークフロー完了後の変更可否は別の業務ルールであり、追加のユースケースと回帰テストが必要です。
