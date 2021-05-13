package transit

import (
	"context"
	"net/http"
	"net/url"
)

type StopsService service

type StopList struct {
	Items     []*Stop `json:"stops"`
	QueryTime string  `json:"query-time"`
}

type StopListOptions struct {
	Street   int64   `url:"street,omitempty"`
	Route    int64   `url:"route,omitempty"`
	Variant  int64   `url:"variant,omitempty"`
	X        int64   `url:"x,omitempty"`
	Y        int64   `url:"y,omitempty"`
	Lat      float64 `url:"lat,omitempty"`
	Lon      float64 `url:"lon,omitempty"`
	Distance int64   `url:"distance,omitempty"`
	Walking  bool    `url:"walking,omitempty"`
}

type Stop struct {
	Key         int64  `json:"key,omitempty"`          // A unique identifier for this stop.
	Name        string `json:"name,omitempty"`         // The stop name.
	Number      int64  `json:"number,omitempty"`       // The stop number.
	Direction   string `json:"direction,omitempty"`    // Specifies which direction buses which service the stop are heading.
	Side        string `json:"side,omitempty"`         // Specifies which side of the intersection the stop lies on.
	Street      Street `json:"street,omitempty"`       // The street the stop is on.
	CrossStreet Street `json:"cross-street,omitempty"` // The nearest cross-street to the stop.
	Centre      Centre `json:"centre,omitempty"`       // A geographical point describing where the stop is located. Both UTM and geographic coordinate systems are provided.
}

type Centre struct {
	UTM        UTM        `json:"utm"`
	Geographic Geographic `json:"geographic"`
}

type Geographic struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type UTM struct {
	Zone string `json:"zone"`
	X    int64  `json:"x"`
	Y    int64  `json:"y"`
}

type Street struct {
	Key  int64  `json:"key"`
	Name string `json:"name"`
	Type string `json:"type"`
}

func (s StopsService) List(ctx context.Context, opts *StopListOptions) ([]*Stop, *http.Response, error) {
	u, err := addOptions("stops", opts)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest(u)
	if err != nil {
		return nil, nil, err
	}

	var stops StopList
	resp, err := s.client.Do(ctx, req, &stops)
	if err != nil {
		return nil, resp, err
	}

	return stops.Items, resp, err
}

func (s StopsService) SearchWildcard(ctx context.Context, verbose bool, query string) ([]*Stop, *http.Response, error) {
	u := url.QueryEscape("stops:" + query)
	req, err := s.client.NewRequest(u)
	if err != nil {
		return nil, nil, err
	}

	var stops StopList
	resp, err := s.client.Do(ctx, req, &stops)
	if err != nil {
		return nil, resp, err
	}

	return stops.Items, resp, err
}
