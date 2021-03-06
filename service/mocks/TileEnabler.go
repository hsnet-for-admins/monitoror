// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import (
	configmodels "github.com/monitoror/monitoror/api/config/models"
	mock "github.com/stretchr/testify/mock"

	models "github.com/monitoror/monitoror/models"
)

// TileEnabler is an autogenerated mock type for the TileEnabler type
type TileEnabler struct {
	mock.Mock
}

// Enable provides a mock function with given fields: variantName, paramsValidator, routePath
func (_m *TileEnabler) Enable(variantName models.VariantName, paramsValidator configmodels.ParamsValidator, routePath string) {
	_m.Called(variantName, paramsValidator, routePath)
}
