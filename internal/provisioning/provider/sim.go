package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type simDetail struct {
	SerialNumber string `json:"serialNumber"`
	Msisdn       string `json:"tel"`
	Tariff       string `json:"tariff"`
	Imei         string `json:"imei"`
	NetworkId    int    `json:"networkId"`
	Iccid        string `json:"iccid"`
}

func (api *Client) findSim(ctx context.Context, tel string) (*simDetail, error) {
	u := fmt.Sprintf("%s/api/company/%d/account/%d/sim?tel=%s", api.baseUrl.String(), api.companyId, api.accountId, tel)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	resp, err := api.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %v, url: %s", err, u)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d,url: %s", resp.StatusCode, u)
	}
	var res listResponse[simDetail]
	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	if res.Count == 0 || res.Results == nil || len(res.Results) == 0 {
		return nil, fmt.Errorf("no sim found for tel: %s", tel)
	}
	return &res.Results[0], nil

}
