package kinesis

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func loadTestData(t *testing.T, filename string) []byte {
	t.Helper()
	data, err := os.ReadFile("testdata/" + filename)
	assert.NoError(t, err)
	return data
}

func TestTTIUplink_Unmarshal(t *testing.T) {
	var uplink TTIUplink
	err := json.Unmarshal(loadTestData(t, "uplink_two_gateways.json"), &uplink)

	assert.NoError(t, err)
	assert.Equal(t, "641ABA00000005AA", uplink.EndDeviceIDs.DevEUI)
	assert.Len(t, uplink.UplinkMessage.RxMetadata, 2)
}

func TestBestGateway_PickHighestRSSI(t *testing.T) {
	var uplink TTIUplink
	json.Unmarshal(loadTestData(t, "uplink_two_gateways.json"), &uplink)
	best := bestGateway(uplink.UplinkMessage.RxMetadata)
	assert.NotNil(t, best)
	assert.Equal(t, "0016C001F157F502", best.GatewayIDs.EUI)
}

func TestBestGateway_NoMetadata(t *testing.T) {
	var uplink TTIUplink
	json.Unmarshal(loadTestData(t, "uplink_no_metadata.json"), &uplink)
	best := bestGateway(uplink.UplinkMessage.RxMetadata)
	assert.Nil(t, best)
}
