package pingdom

import (
	"os"
	"testing"

	"github.com/monitoror/monitoror/internal/pkg/monitorable/test"
	"github.com/stretchr/testify/assert"
)

func TestNewMonitorable(t *testing.T) {
	// init Store
	store, mockMonitorableHelper := test.InitMockAndStore()

	// init Env
	// OK
	_ = os.Setenv("MO_MONITORABLE_PINGDOM_VARIANT0_TOKEN", "xxx")
	// Missing Token
	_ = os.Setenv("MO_MONITORABLE_PINGDOM_VARIANT1_URL", "https://pingdom.example.com")
	// Url broken
	_ = os.Setenv("MO_MONITORABLE_PINGDOM_VARIANT2_URL", "url%spingdom.example.com")

	// NewMonitorable
	monitorable := NewMonitorable(store)
	assert.NotNil(t, monitorable)

	// GetDisplayName
	assert.NotNil(t, monitorable.GetDisplayName())

	// GetVariantNames and check
	if assert.Len(t, monitorable.GetVariantNames(), 4) {
		_, err := monitorable.Validate("variant1")
		assert.Error(t, err)
		_, err = monitorable.Validate("variant2")
		assert.Error(t, err)
	}

	// Enable
	for _, variantName := range monitorable.GetVariantNames() {
		if valid, _ := monitorable.Validate(variantName); valid {
			monitorable.Enable(variantName)
		}
	}

	// Test calls
	mockMonitorableHelper.RouterAssertNumberOfCalls(t, 1, 1)
	mockMonitorableHelper.TileSettingsManagerAssertNumberOfCalls(t, 1, 1, 1, 1)
}
