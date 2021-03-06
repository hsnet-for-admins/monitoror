package usecase

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/monitoror/monitoror/api/config/models"
	"github.com/monitoror/monitoror/api/config/versions"
	coreModels "github.com/monitoror/monitoror/models"
	"github.com/monitoror/monitoror/service/registry"
)

func (cu *configUsecase) Verify(configBag *models.ConfigBag) {
	if configBag.Config.Version == nil {
		configBag.AddErrors(models.ConfigError{
			ID:      models.ConfigErrorMissingRequiredField,
			Message: fmt.Sprintf(`Required "version" field is missing. Current config version is: %s.`, versions.CurrentVersion),
			Data: models.ConfigErrorData{
				FieldName: "version",
			},
		})
		return
	}

	if configBag.Config.Version.IsLessThan(versions.MinimalVersion) || configBag.Config.Version.IsGreaterThan(versions.CurrentVersion) {
		configBag.AddErrors(models.ConfigError{
			ID:      models.ConfigErrorUnsupportedVersion,
			Message: fmt.Sprintf(`Unsupported configuration version. Minimal supported version is %q. Current config version is: %q`, versions.MinimalVersion, versions.CurrentVersion),
			Data: models.ConfigErrorData{
				FieldName: "version",
				Value:     stringify(configBag.Config.Version),
				Expected:  fmt.Sprintf(`%q <= version <= %q`, versions.MinimalVersion, versions.CurrentVersion),
			},
		})
		return
	}

	if configBag.Config.Columns == nil {
		configBag.AddErrors(models.ConfigError{
			ID:      models.ConfigErrorMissingRequiredField,
			Message: fmt.Sprintf(`Required "columns" field is missing. Must be a positive integer.`),
			Data: models.ConfigErrorData{
				FieldName: "columns",
			},
		})
	} else if *configBag.Config.Columns <= 0 {
		configBag.AddErrors(models.ConfigError{
			ID:      models.ConfigErrorInvalidFieldValue,
			Message: fmt.Sprintf(`Invalid "columns" field. Must be a positive integer.`),
			Data: models.ConfigErrorData{
				FieldName: "columns",
				Value:     stringify(configBag.Config.Columns),
				Expected:  "columns > 0",
			},
		})
	}

	if configBag.Config.Zoom != nil && (*configBag.Config.Zoom <= 0 || *configBag.Config.Zoom > 10) {
		configBag.AddErrors(models.ConfigError{
			ID:      models.ConfigErrorInvalidFieldValue,
			Message: `Invalid "zoom" field. Must be a positive float between 0 and 10.`,
			Data: models.ConfigErrorData{
				FieldName: "zoom",
				Value:     stringify(configBag.Config.Zoom),
				Expected:  "0 < zoom <= 10",
			},
		})
	}

	if configBag.Config.Tiles == nil {
		configBag.AddErrors(models.ConfigError{
			ID:      models.ConfigErrorMissingRequiredField,
			Message: `Missing "tiles" field. Must be a non-empty array.`,
			Data: models.ConfigErrorData{
				FieldName: "tiles",
			},
		})
	} else if len(configBag.Config.Tiles) == 0 {
		configBag.AddErrors(models.ConfigError{
			ID:      models.ConfigErrorInvalidFieldValue,
			Message: `Invalid "tiles" field. Must be a non-empty array.`,
			Data: models.ConfigErrorData{
				FieldName:     "tiles",
				ConfigExtract: stringify(configBag.Config),
			},
		})
	} else {
		// Iterating through every config tiles
		for _, tile := range configBag.Config.Tiles {
			cu.verifyTile(configBag, &tile, nil)
		}
	}
}

func (cu *configUsecase) verifyTile(configBag *models.ConfigBag, tile *models.TileConfig, groupTile *models.TileConfig) {
	if tile.ColumnSpan != nil && *tile.ColumnSpan <= 0 {
		configBag.AddErrors(models.ConfigError{
			ID:      models.ConfigErrorInvalidFieldValue,
			Message: `Invalid "columnSpan" field. Must be a positive integer.`,
			Data: models.ConfigErrorData{
				FieldName:     "columnSpan",
				Expected:      "columnSpan > 0",
				ConfigExtract: stringify(tile),
			},
		})
		return
	}

	if tile.RowSpan != nil && *tile.RowSpan <= 0 {
		configBag.AddErrors(models.ConfigError{
			ID:      models.ConfigErrorInvalidFieldValue,
			Message: `Invalid "rowSpan" field. Must be a positive integer.`,
			Data: models.ConfigErrorData{
				FieldName:     "rowSpan",
				Expected:      "rowSpan > 0",
				ConfigExtract: stringify(tile),
			},
		})
		return
	}

	// Empty tile, skip
	if tile.Type == EmptyTileType {
		if groupTile != nil {
			configBag.AddErrors(models.ConfigError{
				ID:      models.ConfigErrorUnauthorizedSubtileType,
				Message: fmt.Sprintf(`Unauthorized %q type in %s tile.`, EmptyTileType, GroupTileType),
				Data: models.ConfigErrorData{
					ConfigExtract:          stringify(groupTile),
					ConfigExtractHighlight: stringify(tile),
				},
			})
		}
		return
	}

	// Group tile, parse and call verifyTile for each grouped tile
	if tile.Type == GroupTileType {
		if groupTile != nil {
			configBag.AddErrors(models.ConfigError{
				ID:      models.ConfigErrorUnauthorizedSubtileType,
				Message: fmt.Sprintf(`Unauthorized %q type in %s tile.`, GroupTileType, GroupTileType),
				Data: models.ConfigErrorData{
					ConfigExtract:          stringify(groupTile),
					ConfigExtractHighlight: stringify(tile),
				},
			})
			return
		}

		if tile.Params != nil {
			configBag.AddErrors(models.ConfigError{
				ID:      models.ConfigErrorUnauthorizedField,
				Message: fmt.Sprintf(`Unauthorized "params" key in %s tile definition.`, tile.Type),
				Data: models.ConfigErrorData{
					FieldName:     "params",
					ConfigExtract: stringify(tile),
				},
			})
			return
		}

		if tile.Tiles == nil {
			configBag.AddErrors(models.ConfigError{
				ID:      models.ConfigErrorMissingRequiredField,
				Message: fmt.Sprintf(`Missing "tiles" field in %s tile definition. Must be a non-empty array.`, tile.Type),
				Data: models.ConfigErrorData{
					FieldName:     "tiles",
					ConfigExtract: stringify(tile),
				},
			})
		} else if len(tile.Tiles) == 0 {
			configBag.AddErrors(models.ConfigError{
				ID:      models.ConfigErrorInvalidFieldValue,
				Message: fmt.Sprintf(`Invalid "tiles" field in %s tile definition. Must be a non-empty array.`, tile.Type),
				Data: models.ConfigErrorData{
					FieldName:     "tiles",
					ConfigExtract: stringify(tile),
				},
			})
			return
		}

		for _, groupTile := range tile.Tiles {
			cu.verifyTile(configBag, &groupTile, tile)
		}

		return
	}

	// Set ConfigVariant to DefaultVariant if empty
	if tile.ConfigVariant == "" {
		tile.ConfigVariant = coreModels.DefaultVariant
	}

	// Get the metadataExplorer for current tile
	var metadataExplorer registry.TileMetadataExplorer
	var exists bool
	if tile.Type.IsGenerator() {
		// This tile type is a generator tile type
		metadataExplorer, exists = cu.registry.GeneratorMetadata[tile.Type]
		if !exists {
			configBag.AddErrors(models.ConfigError{
				ID:      models.ConfigErrorUnknownGeneratorTileType,
				Message: fmt.Sprintf(`Unknown %q generator type in tile definition. Must be %s`, tile.Type, keys(cu.registry.GeneratorMetadata)),
				Data: models.ConfigErrorData{
					FieldName:     "type",
					ConfigExtract: stringify(tile),
					Expected:      keys(cu.registry.GeneratorMetadata),
				},
			})
			return
		}
	} else {
		// This tile type is a normal tile type
		metadataExplorer, exists = cu.registry.TileMetadata[tile.Type]
		if !exists {
			configBag.AddErrors(models.ConfigError{
				ID:      models.ConfigErrorUnknownTileType,
				Message: fmt.Sprintf(`Unknown %q generator type in tile definition. Must be %s`, tile.Type, keys(cu.registry.TileMetadata)),
				Data: models.ConfigErrorData{
					FieldName:     "type",
					ConfigExtract: stringify(tile),
					Expected:      keys(cu.registry.TileMetadata),
				},
			})
			return
		}
	}

	if configBag.Config.Version.IsLessThan(metadataExplorer.GetMinimalVersion()) {
		configBag.AddErrors(models.ConfigError{
			ID: models.ConfigErrorTileNotSupportedInThisVersion,
			Message: fmt.Sprintf(`%q tile type is not supported in version %q. Minimal supported version is %q`,
				tile.Type, configBag.Config.Version, metadataExplorer.GetMinimalVersion()),
			Data: models.ConfigErrorData{
				FieldName:     "type",
				ConfigExtract: stringify(tile),
				Expected:      fmt.Sprintf(`%q <= version`, metadataExplorer.GetMinimalVersion()),
			},
		})
		return
	}

	variantMetadataExplorer, exists := metadataExplorer.GetVariant(tile.ConfigVariant)
	if !exists {
		configBag.AddErrors(models.ConfigError{
			ID: models.ConfigErrorUnknownVariant,
			Message: fmt.Sprintf(`Unknown %q variant for %s type in tile definition. Must be %s`,
				tile.ConfigVariant, tile.Type, stringify(metadataExplorer.GetVariantNames())),
			Data: models.ConfigErrorData{
				FieldName:     "configVariant",
				Value:         stringify(tile.ConfigVariant),
				Expected:      stringify(metadataExplorer.GetVariantNames()),
				ConfigExtract: stringify(tile),
			},
		})
		return
	}

	if !variantMetadataExplorer.IsEnabled() {
		configBag.AddErrors(models.ConfigError{
			ID:      models.ConfigErrorDisabledVariant,
			Message: fmt.Sprintf(`Variant %q is disabled for %s type. Check errors on the server side for the reason`, tile.ConfigVariant, tile.Type),
			Data: models.ConfigErrorData{
				FieldName:     "configVariant",
				Value:         stringify(tile.ConfigVariant),
				ConfigExtract: stringify(tile),
			},
		})
		return
	}

	if tile.Params == nil {
		configBag.AddErrors(models.ConfigError{
			ID:      models.ConfigErrorMissingRequiredField,
			Message: fmt.Sprintf(`Missing "params" key in %s tile definition.`, tile.Type),
			Data: models.ConfigErrorData{
				FieldName:     "params",
				ConfigExtract: stringify(tile),
			},
		})
		return
	}

	// Create new validator by reflexion
	rType := reflect.TypeOf(variantMetadataExplorer.GetValidator())
	rInstance := reflect.New(rType.Elem()).Interface()

	// Marshal / Unmarshal the map[string]interface{} struct in new instance of ParamsValidator
	bytesParams, _ := json.Marshal(tile.Params)
	if unmarshalErr := json.Unmarshal(bytesParams, &rInstance); unmarshalErr != nil {
		configBag.AddErrors(models.ConfigError{
			ID:      models.ConfigErrorUnexpectedError,
			Message: unmarshalErr.Error(),
			Data: models.ConfigErrorData{
				FieldName:     "params",
				ConfigExtract: stringify(tile),
			},
		})
		return
	}

	// Marshal / Unmarshal instance of validator into map[string]interface{} to identify unknown fields
	structParams := make(map[string]interface{})
	bytesParamInstance, _ := json.Marshal(rInstance)
	_ = json.Unmarshal(bytesParamInstance, &structParams)

	// Check if struct has unknown fields
	for field := range tile.Params {
		if _, ok := structParams[field]; !ok {
			configBag.AddErrors(models.ConfigError{
				ID:      models.ConfigErrorUnknownField,
				Message: fmt.Sprintf(`Unknown %q tile params field.`, field),
				Data: models.ConfigErrorData{
					FieldName:     field,
					ConfigExtract: stringify(tile),
					Expected:      keys(structParams),
				},
			})
			return
		}
	}

	// Validate config with the config file version
	if err := rInstance.(models.ParamsValidator).Validate(configBag.Config.Version); err != nil {
		// TODO
		//configBag.AddErrors(*err)

		configBag.AddErrors(models.ConfigError{
			ID:      models.ConfigErrorInvalidFieldValue,
			Message: fmt.Sprintf(`Invalid params definition for %q: %q.`, tile.Type, string(bytesParams)),
			Data: models.ConfigErrorData{
				FieldName:     "params",
				ConfigExtract: stringify(tile),
			},
		})
	}
}
