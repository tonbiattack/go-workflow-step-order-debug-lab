package workflowstep

import "time"

// WorkflowStep はワークフロー内で実行する一つの業務ステップです。
type WorkflowStep struct {
	ID          string    `gorm:"type:char(36);primaryKey" json:"id"`
	WorkflowID  string    `gorm:"type:char(36);not null;uniqueIndex:idx_workflow_step_order,priority:1" json:"workflow_id"`
	Name        string    `gorm:"type:varchar(255);not null" json:"name"`
	Description string    `gorm:"type:text" json:"description"`
	Order       int       `gorm:"not null;uniqueIndex:idx_workflow_step_order,priority:2" json:"order"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
