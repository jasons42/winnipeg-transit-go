package integration_tests

import (
	"context"
	"testing"
)

func TestStops_search(t *testing.T) {
	q := "main"

	client := client(t)
	ctx := context.Background()

	stops, resp, err := client.Stops.Search(ctx, q)
	if err != nil {
		t.Errorf("error: %+v\nresp: %+v", err, resp)
	}

	if len(stops) == 0 {
		t.Errorf("expected to get stops but found none")
	}
}
