package solus

import (
	"context"
	"fmt"

	"gopkg.in/guregu/null.v4"
)

type LocationsService service

type LocationCreateRequest struct {
	Name             string   `json:"name"`
	Description      string   `json:"description"`
	IconID           null.Int `json:"icon_id"`
	IsDefault        bool     `json:"is_default"`
	IsVisible        bool     `json:"is_visible"`
	ComputeResources []int    `json:"compute_resources,omitempty"`
}

type Location struct {
	ID               int               `json:"id"`
	Name             string            `json:"name"`
	Icon             Icon              `json:"icon"`
	Description      string            `json:"description"`
	IsDefault        bool              `json:"is_default"`
	IsVisible        bool              `json:"is_visible"`
	ComputeResources []ComputeResource `json:"compute_resources"`
}

type LocationsResponse struct {
	paginatedResponse

	Data []Location `json:"data"`
}

type locationResponse struct {
	Data Location `json:"data"`
}

func (s *LocationsService) Create(ctx context.Context, data LocationCreateRequest) (Location, error) {
	var resp locationResponse
	return resp.Data, s.client.create(ctx, "locations", data, &resp)
}

func (s *LocationsService) List(ctx context.Context, filter *FilterLocations) (LocationsResponse, error) {
	resp := LocationsResponse{
		paginatedResponse: paginatedResponse{
			service: (*service)(s),
		},
	}
	return resp, s.client.list(ctx, "locations", &resp, withFilter(filter.data))
}

func (s *LocationsService) Get(ctx context.Context, id int) (Location, error) {
	var resp locationResponse
	return resp.Data, s.client.get(ctx, fmt.Sprintf("locations/%d", id), &resp)
}

func (s *LocationsService) Update(ctx context.Context, id int, data LocationCreateRequest) (Location, error) {
	var resp locationResponse
	return resp.Data, s.client.update(ctx, fmt.Sprintf("locations/%d", id), data, &resp)
}

func (s *LocationsService) Delete(ctx context.Context, id int) error {
	return s.client.delete(ctx, fmt.Sprintf("locations/%d", id))
}
