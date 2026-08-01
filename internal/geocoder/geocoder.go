package geocoder

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type NominatimResponse struct {
	DisplayName string `json:"display_name"`
	Lat         string `json:"lat"`
	Lon         string `json:"lon"`
}

type Geocoder struct {
	client *http.Client
}

func NewGeocoder() *Geocoder {
	return &Geocoder{
		client: &http.Client{Timeout: 5 * time.Second},
	}
}

// ValidateAndNormalizeAddress checks if an address exists on the map.
// Returns the formatted address if found, or an error if not found.
func (g *Geocoder) ValidateAndNormalizeAddress(ctx context.Context, address string) (string, error) {
	endpoint := fmt.Sprintf("https://nominatim.openstreetmap.org/search?q=%s&format=json&limit=1", url.QueryEscape(address))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", err
	}

	// Nominatim requires a custom User-Agent header
	req.Header.Set("User-Agent", "LibraryApp/1.0")

	resp, err := g.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to reach map service: %w", err)
	}
	defer resp.Body.Close()

	var results []NominatimResponse
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return "", err
	}

	if len(results) == 0 {
		return "", errors.New("address does not exist on the map")
	}

	// Returns the official formatted address string from the map
	return results[0].DisplayName, nil
}
