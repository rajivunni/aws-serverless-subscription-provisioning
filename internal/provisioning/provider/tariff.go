package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type tariff struct {
	Name               string `json:"name"`
	Number             int    `json:"number"`
	RequiresDataBoltOn bool   `json:"requiresDataBoltOn"`
}

type boltOn struct {
	Name         string `json:"name"`
	Number       int    `json:"number"`
	IsDataBoltOn bool   `json:"isDataBoltOn"`
}

func (api *Client) getTariffs(ctx context.Context, networkId int) ([]tariff, error) {
	u := fmt.Sprintf("%s/api/account/%d/network/%d/tariff", api.baseUrl.String(), api.accountId, networkId)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}

	resp, err := api.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request for url:%s and error is: %w", u, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code from tariff API: %d,url: %s", resp.StatusCode, u)
	}
	var res listResponse[tariff]
	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		return nil, err
	}
	tariffs := res.Results
	if tariffs == nil || len(tariffs) == 0 {
		return nil, fmt.Errorf("no tariffs found for network %d, url: %s", networkId, u)
	}
	return tariffs, nil
}

func (api *Client) getBoltOns(ctx context.Context, tariffId int) ([]boltOn, error) {
	u := fmt.Sprintf("%s/api/account/%d/tariff/%d/boltons", api.baseUrl.String(), api.accountId, tariffId)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	resp, err := api.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request for url:%s and error is: %w", u, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code from bolton API: %d, url: %s", resp.StatusCode, u)
	}
	var res listResponse[boltOn]
	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		return nil, err
	}
	boltons := res.Results
	if boltons == nil || len(boltons) == 0 {
		return nil, fmt.Errorf("no boltons found for tariff %d, url: %s", tariffId, u)
	}
	return boltons, nil
}
