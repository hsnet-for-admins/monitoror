//+build !faker

package http

import (
	"github.com/monitoror/monitoror/api/config/versions"
	pkgMonitorable "github.com/monitoror/monitoror/internal/pkg/monitorable"
	coreModels "github.com/monitoror/monitoror/models"
	"github.com/monitoror/monitoror/monitorables/http/api"
	httpDelivery "github.com/monitoror/monitoror/monitorables/http/api/delivery/http"
	httpModels "github.com/monitoror/monitoror/monitorables/http/api/models"
	httpRepository "github.com/monitoror/monitoror/monitorables/http/api/repository"
	httpUsecase "github.com/monitoror/monitoror/monitorables/http/api/usecase"
	httpConfig "github.com/monitoror/monitoror/monitorables/http/config"
	"github.com/monitoror/monitoror/service/registry"
	"github.com/monitoror/monitoror/service/store"
)

type Monitorable struct {
	store *store.Store

	config map[coreModels.VariantName]*httpConfig.HTTP

	// Config tile settings
	statusTileEnabler    registry.TileEnabler
	rawTileEnabler       registry.TileEnabler
	formattedTileEnabler registry.TileEnabler
}

func NewMonitorable(store *store.Store) *Monitorable {
	m := &Monitorable{}
	m.store = store
	m.config = make(map[coreModels.VariantName]*httpConfig.HTTP)

	// Load core config from env
	pkgMonitorable.LoadConfig(&m.config, httpConfig.Default)

	// Register Monitorable Tile in config manager
	m.statusTileEnabler = store.Registry.RegisterTile(api.HTTPStatusTileType, versions.MinimalVersion, m.GetVariantNames())
	m.rawTileEnabler = store.Registry.RegisterTile(api.HTTPRawTileType, versions.MinimalVersion, m.GetVariantNames())
	m.formattedTileEnabler = store.Registry.RegisterTile(api.HTTPFormattedTileType, versions.MinimalVersion, m.GetVariantNames())

	return m
}

func (m *Monitorable) GetDisplayName() string {
	return "HTTP"
}

func (m *Monitorable) GetVariantNames() []coreModels.VariantName {
	return pkgMonitorable.GetVariants(m.config)
}

func (m *Monitorable) Validate(_ coreModels.VariantName) (bool, error) {
	return true, nil
}

func (m *Monitorable) Enable(variantName coreModels.VariantName) {
	conf := m.config[variantName]

	repository := httpRepository.NewHTTPRepository(conf)
	usecase := httpUsecase.NewHTTPUsecase(repository, m.store.CacheStore, m.store.CoreConfig.UpstreamCacheExpiration)
	delivery := httpDelivery.NewHTTPDelivery(usecase)

	// EnableTile route to echo
	routeGroup := m.store.MonitorableRouter.Group("/http", variantName)
	routeStatus := routeGroup.GET("/status", delivery.GetHTTPStatus)
	routeRaw := routeGroup.GET("/raw", delivery.GetHTTPRaw)
	routeJSON := routeGroup.GET("/formatted", delivery.GetHTTPFormatted)

	// EnableTile data for config hydration
	m.statusTileEnabler.Enable(variantName, &httpModels.HTTPStatusParams{}, routeStatus.Path)
	m.rawTileEnabler.Enable(variantName, &httpModels.HTTPRawParams{}, routeRaw.Path)
	m.formattedTileEnabler.Enable(variantName, &httpModels.HTTPFormattedParams{}, routeJSON.Path)
}
