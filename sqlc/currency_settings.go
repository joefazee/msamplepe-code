package db

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/timchuks/monieverse/internal/mapper"
	"github.com/timchuks/monieverse/internal/settings"
)

type dbCurrencySettings struct {
	store  Store
	mapper mapper.ConfigMapper
}

// NewSettings creates a new instance of CurrencySettings which is backed by a database
// if you want to write a different implementation, you can implement the Settings interface
func NewSettings(store Store, mapper mapper.ConfigMapper) settings.Settings {
	return &dbCurrencySettings{
		store:  store,
		mapper: mapper,
	}
}

func (d dbCurrencySettings) GetCurrencyConfigurations(ctx context.Context, currencyID int32, userID uuid.UUID) (settings.CurrencyConstraints, error) {
	configs := settings.CurrencyConstraints{}

	mapConfigs := make(map[string]string)

	rows, err := d.store.GetCurrencySetting(ctx, GetCurrencySettingParams{
		CurrencyID: currencyID,
		UserID:     userID,
	})
	if err != nil {
		return configs, err
	}

	for _, row := range rows {
		mapConfigs[row.ConfigKey] = row.ConfigValue
	}

	err = d.mapper.Map(mapConfigs, &configs)
	if err != nil {
		return configs, fmt.Errorf("GetCurrencyConfigurations: failed to map configs: %w", err)
	}

	return configs, nil

}

func (d dbCurrencySettings) GetSystemSettings(ctx context.Context) (settings.SystemSettings, error) {
	configs := settings.SystemSettings{}

	mapConfigs := make(map[string]string)
	if d.store == nil {
		return configs, nil
	}

	rows, err := d.store.GetSystemSettings(ctx)
	if err != nil {
		return configs, err
	}

	for _, row := range rows {
		mapConfigs[row.ConfigKey] = row.ConfigValue
	}

	err = d.mapper.Map(mapConfigs, &configs)
	if err != nil {
		return configs, fmt.Errorf("GetSystemSettings: failed to map configs: %w", err)
	}

	return configs, nil
}
