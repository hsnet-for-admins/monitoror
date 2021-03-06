//+build faker

package models

import (
	"time"

	uiConfigModels "github.com/monitoror/monitoror/api/config/models"
	"github.com/monitoror/monitoror/models"
)

type (
	BuildParams struct {
		Project    string  `json:"project" query:"project"`
		Definition *int    `json:"definition" query:"definition"`
		Branch     *string `json:"branch,omitempty" query:"branch"`

		AuthorName      string `json:"authorName" query:"authorName"`
		AuthorAvatarURL string `json:"authorAvatarURL" query:"authorAvatarURL"`

		Status            models.TileStatus `json:"status" query:"status"`
		PreviousStatus    models.TileStatus `json:"previousStatus" query:"previousStatus"`
		StartedAt         time.Time         `json:"startedAt" query:"startedAt"`
		FinishedAt        time.Time         `json:"finishedAt" query:"finishedAt"`
		Duration          int64             `json:"duration" query:"duration"`
		EstimatedDuration int64             `json:"estimatedDuration" query:"estimatedDuration"`
	}
)

func (p *BuildParams) Validate(_ *uiConfigModels.ConfigVersion) *uiConfigModels.ConfigError {
	// TODO

	if p.Project == "" {
		return &uiConfigModels.ConfigError{}
	}

	if p.Definition == nil {
		return &uiConfigModels.ConfigError{}
	}

	return nil
}
