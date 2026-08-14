package workflowstep

import (
	"errors"
	"strings"

	"gorm.io/gorm"
)

// Repository はワークフローステップの永続化を担当します。
type Repository struct {
	db *gorm.DB
}

// NewRepository は永続化リポジトリを作成します。
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// Create はワークフローステップを保存します。
func (r *Repository) Create(step *WorkflowStep) error {
	if err := r.db.Create(step).Error; err != nil {
		if isDuplicateWorkflowStepOrderError(err) {
			return ErrWorkflowStepOrderAlreadyExists
		}
		return err
	}
	return nil
}

// isDuplicateWorkflowStepOrderError は、SQLiteの複合一意制約違反をドメインエラーとして扱います。
func isDuplicateWorkflowStepOrderError(err error) bool {
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return true
	}
	return strings.Contains(err.Error(), "UNIQUE constraint failed: workflow_steps.workflow_id, workflow_steps.order")
}

// CountByWorkflowAndOrder はワークフロー内の指定順序にあるステップ数を返します。
func (r *Repository) CountByWorkflowAndOrder(workflowID string, order int) (int64, error) {
	var count int64
	err := r.db.Model(&WorkflowStep{}).
		Where("workflow_id = ? AND `order` = ?", workflowID, order).
		Count(&count).Error
	return count, err
}
