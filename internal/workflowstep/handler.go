package workflowstep

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Handler はワークフローステップ作成APIを提供します。
type Handler struct {
	repository *Repository
}

// NewHandler はHTTPハンドラーを作成します。
func NewHandler(repository *Repository) *Handler {
	return &Handler{repository: repository}
}

// CreateWorkflowStep はPOST /workflow-stepsでステップを作成します。
func (h *Handler) CreateWorkflowStep(c *gin.Context) {
	var step WorkflowStep
	if err := c.ShouldBindJSON(&step); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	step.ID = uuid.NewString()
	if err := h.repository.Create(&step); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create workflow step"})
		return
	}

	c.JSON(http.StatusCreated, step)
}
