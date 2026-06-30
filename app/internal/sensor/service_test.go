package sensor_test

import (
	"context"
	"errors"
	"testing"

	"github.com/anoop-dryad/bridgehead/app/internal/sensor"
	"github.com/anoop-dryad/bridgehead/app/internal/sensor/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap/zaptest"
)

func TestRecordUplink_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockRepositoryInterface(ctrl)
	svc := sensor.NewService(repo, zaptest.NewLogger(t))

	repo.EXPECT().
		WithTransaction(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, fn func(ctx context.Context) error) error {
			return fn(ctx)
		})

	repo.EXPECT().UpsertSensor(gomock.Any(), sensor.Sensor{
		EUI:      "641aba00000005aa",
		DeviceID: "sn-silvav3n805",
		AppID:    "8fcc36f0-6b93",
	}).Return(nil)

	repo.EXPECT().UpsertMapping(gomock.Any(), sensor.GatewayMapping{
		SensorEUI:  "641aba00000005aa",
		GatewayEUI: "0016c001f157f502",
	}).Return(nil)

	err := svc.RecordUplink(context.Background(), sensor.UplinkEvent{
		SensorEUI:  "641aba00000005aa",
		DeviceID:   "sn-silvav3n805",
		AppID:      "8fcc36f0-6b93",
		GatewayEUI: "0016c001f157f502",
	})

	assert.NoError(t, err)
}

func TestRecordUplink_EmptyGatewayEui(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockRepositoryInterface(ctrl)
	svc := sensor.NewService(repo, zaptest.NewLogger(t))

	repo.EXPECT().UpsertSensor(gomock.Any(), gomock.Any()).Times(0)
	repo.EXPECT().UpsertMapping(gomock.Any(), gomock.Any()).Times(0)

	err := svc.RecordUplink(context.Background(), sensor.UplinkEvent{
		SensorEUI:  "641aba00000005aa",
		DeviceID:   "sn-silvav3n805",
		AppID:      "8fcc36f0-6b93",
		GatewayEUI: "",
	})

	assert.ErrorIs(t, err, sensor.ErrNoGatewayInUplink)
}

func TestRecordUplink_SensorUpsertFails_RollsBack(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockRepositoryInterface(ctrl)
	svc := sensor.NewService(repo, zaptest.NewLogger(t))

	repo.EXPECT().
		WithTransaction(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		})

	repo.EXPECT().
		UpsertSensor(gomock.Any(), gomock.Any()).
		Return(errors.New("db error"))

	// mapping must NOT be called — transaction rolled back
	repo.EXPECT().UpsertMapping(gomock.Any(), gomock.Any()).Times(0)

	err := svc.RecordUplink(context.Background(), sensor.UplinkEvent{
		SensorEUI:  "641aba00000005aa",
		GatewayEUI: "0016c001f157f502",
	})

	assert.Error(t, err)
}
