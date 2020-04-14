//+build !faker

package models

import (
	"fmt"

	uiConfigModels "github.com/monitoror/monitoror/api/config/models"
)

type (
	BuildParams struct {
		Owner      string `json:"owner" query:"owner"`
		Repository string `json:"repository" query:"repository"`
		Branch     string `json:"branch" query:"branch"`
	}
)

func (p *BuildParams) Validate(_ *uiConfigModels.ConfigVersion) *uiConfigModels.ConfigError {
	if p.Owner == "" {
		return &uiConfigModels.ConfigError{
			ID:      uiConfigModels.ConfigErrorMissingRequiredField,
			Message: fmt.Sprintf(`Required "owner" field is missing.`),
			Data:    uiConfigModels.ConfigErrorData{FieldName: "owner"},
		}
	}

	if p.Repository == "" {
		return &uiConfigModels.ConfigError{
			ID:      uiConfigModels.ConfigErrorMissingRequiredField,
			Message: fmt.Sprintf(`Required "repository" field is missing.`),
			Data:    uiConfigModels.ConfigErrorData{FieldName: "repository"},
		}
	}

	if p.Branch == "" {
		return &uiConfigModels.ConfigError{
			ID:      uiConfigModels.ConfigErrorMissingRequiredField,
			Message: fmt.Sprintf(`Required "branch" field is missing.`),
			Data:    uiConfigModels.ConfigErrorData{FieldName: "branch"},
		}
	}

	return nil
}

// Used by cache as identifier
func (p *BuildParams) String() string {
	return fmt.Sprintf("BUILD-%s-%s-%s", p.Owner, p.Repository, p.Branch)
}
