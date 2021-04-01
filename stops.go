package transit

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

type StopsService service

type StopList struct {
	Items     []*Stop `json:"stops"`
	QueryTime string  `json:"query-time"`
}

type Stop struct {
	Key         int64  `json:"key"`
	Name        string `json:"name"`
	Number      int64  `json:"number"`
	Direction   string `json:"direction"`
	Side        string `json:"side"`
	Street      Street `json:"street"`
	CrossStreet Street `json:"cross-street"`
	Centre      Centre `json:"centre"`
}

type Centre struct {
	UTM        UTM        `json:"utm"`
	Geographic Geographic `json:"geographic"`
}

type Geographic struct {
	Latitude  string `json:"latitude"`
	Longitude string `json:"longitude"`
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

func (s StopsService) Search(ctx context.Context, query string) ([]*Stop, *http.Response, error) {
	u := url.QueryEscape(fmt.Sprintf("stops:%v", query))
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
