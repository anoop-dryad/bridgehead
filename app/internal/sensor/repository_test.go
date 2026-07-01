package sensor_test

import (
	"context"
	"testing"

	"github.com/anoop-dryad/bridgehead/app/internal/sensor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/anoop-dryad/bridgehead/app/infra/testhelper"
)

func TestUpsertSensor_Insert(t *testing.T) {
	db := testhelper.NewTestDB(t)
	repo := sensor.NewRepository(db)

	err := repo.UpsertSensor(context.Background(), sensor.Sensor{
		EUI:      "641aba00000005aa",
		DeviceID: "sn-silvav3n805",
		AppID:    "8fcc36f0-6b93",
	})

	require.NoError(t, err)

	sensor, err := repo.GetByEUI(context.Background(), "641aba00000005aa")
	require.NoError(t, err)
	assert.Equal(t, "sn-silvav3n805", sensor.DeviceID)
}

func TestUpsertSensor_UpdateOnConflict(t *testing.T) {
	db := testhelper.NewTestDB(t)
	repo := sensor.NewRepository(db)

	// insert first
	repo.UpsertSensor(context.Background(), sensor.Sensor{
		EUI:      "641aba00000005aa",
		DeviceID: "old-device-id",
		AppID:    "8fcc36f0",
	})

	// upsert with new device_id
	repo.UpsertSensor(context.Background(), sensor.Sensor{
		EUI:      "641aba00000005aa",
		DeviceID: "new-device-id", // changed
		AppID:    "8fcc36f0",
	})

	sensor, _ := repo.GetByEUI(context.Background(), "641aba00000005aa")
	assert.Equal(t, "new-device-id", sensor.DeviceID) // updated
}

func TestUpsertMapping_GatewayChanges(t *testing.T) {
	db := testhelper.NewTestDB(t)
	repo := sensor.NewRepository(db)

	// seed sensor first (FK constraint)
	repo.UpsertSensor(context.Background(), sensor.Sensor{
		EUI: "641aba00000005aa", DeviceID: "d1", AppID: "a1",
	})

	// initial mapping
	repo.UpsertMapping(context.Background(), sensor.GatewayMapping{
		SensorEUI:  "641aba00000005aa",
		GatewayEUI: "0016c001f158d594", // old gateway
	})

	// gateway changes — higher RSSI on new gateway
	repo.UpsertMapping(context.Background(), sensor.GatewayMapping{
		SensorEUI:  "641aba00000005aa",
		GatewayEUI: "0016c001f157f502", // new gateway
	})

	mapping, _ := repo.GetMappingBySensorEUI(context.Background(), "641aba00000005aa")
	assert.Equal(t, "0016c001f157f502", mapping.GatewayEUI) // updated
}

func TestGetSensorsByGatewayEUI(t *testing.T) {
	db := testhelper.NewTestDB(t)
	repo := sensor.NewRepository(db)

	// seed two sensors on same gateway
	repo.UpsertSensor(context.Background(), sensor.Sensor{EUI: "sensor-1", DeviceID: "d1", AppID: "a1"})
	repo.UpsertSensor(context.Background(), sensor.Sensor{EUI: "sensor-2", DeviceID: "d2", AppID: "a1"})
	repo.UpsertMapping(context.Background(), sensor.GatewayMapping{SensorEUI: "sensor-1", GatewayEUI: "gw-1"})
	repo.UpsertMapping(context.Background(), sensor.GatewayMapping{SensorEUI: "sensor-2", GatewayEUI: "gw-1"})

	sensors, err := repo.GetSensorsByGatewayEUI(context.Background(), "gw-1")
	require.NoError(t, err)
	assert.Len(t, sensors, 2)
}
