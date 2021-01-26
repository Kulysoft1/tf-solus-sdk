package solus

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPlansService_List(t *testing.T) {
	expected := PlansResponse{
		Data: []Plan{
			fakePlan,
		},
	}

	s := startTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/plans", r.URL.Path)
		assert.Equal(t, http.MethodGet, r.Method)

		writeJSON(t, w, http.StatusOK, expected)
	})
	defer s.Close()

	actual, err := createTestClient(t, s.URL).Plans.List(context.Background())
	require.NoError(t, err)
	actual.service = nil
	require.Equal(t, expected, actual)
}

func TestPlansService_Create(t *testing.T) {
	data := PlanCreateRequest{
		Name: "name",
		Params: PlanParams{
			Disk: 1,
			RAM:  2,
			VCPU: 3,
		},
		StorageType:        "storage type",
		ImageFormat:        "image format",
		IsVisible:          true,
		IsSnapshotsEnabled: true,
		Limits: PlanLimits{
			DiskBandwidth: DiskBandwidthPlanLimit{
				IsEnabled: true,
				Limit:     11,
				Unit:      DiskBandwidthPlanLimitUnitBps,
			},
			DiskIOPS: DiskIOPSPlanLimit{
				IsEnabled: true,
				Limit:     12,
				Unit:      DiskIOPSPlanLimitUnitOPS,
			},
			NetworkIncomingBandwidth: BandwidthPlanLimit{
				IsEnabled: true,
				Limit:     13,
				Unit:      BandwidthPlanLimitUnitKbps,
			},
			NetworkOutgoingBandwidth: BandwidthPlanLimit{
				IsEnabled: true,
				Limit:     14,
				Unit:      BandwidthPlanLimitUnitMbps,
			},
			NetworkIncomingTraffic: TrafficPlanLimit{
				IsEnabled: true,
				Limit:     15,
				Unit:      TrafficPlanLimitUnitTB,
			},
			NetworkOutgoingTraffic: TrafficPlanLimit{
				IsEnabled: true,
				Limit:     16,
				Unit:      TrafficPlanLimitUnitMB,
			},
			NetworkReduceBandwidth: BandwidthPlanLimit{},
		},
		TokensPerHour:    4,
		TokensPerMonth:   5,
		Position:         6,
		ResetLimitPolicy: PlanResetLimitPolicyVMCreatedDay,
	}

	s := startTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/plans", r.URL.Path)
		assert.Equal(t, http.MethodPost, r.Method)
		assertRequestBody(t, r, data)

		writeResponse(t, w, http.StatusCreated, fakePlan)
	})
	defer s.Close()

	actual, err := createTestClient(t, s.URL).Plans.Create(context.Background(), data)
	require.NoError(t, err)
	require.Equal(t, fakePlan, actual)
}

func TestPlansService_Delete(t *testing.T) {
	s := startTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/plans/10", r.URL.Path)
		assert.Equal(t, http.MethodDelete, r.Method)

		w.WriteHeader(204)
	})
	defer s.Close()

	err := createTestClient(t, s.URL).Plans.Delete(context.Background(), 10)
	require.NoError(t, err)
}
