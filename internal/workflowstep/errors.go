package workflowstep

import "errors"

// ErrWorkflowStepOrderAlreadyExists は、同一ワークフロー内に同じ順序がすでに存在する場合に返します。
var ErrWorkflowStepOrderAlreadyExists = errors.New("workflow step order already exists")
