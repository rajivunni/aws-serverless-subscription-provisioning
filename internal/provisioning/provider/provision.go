package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type simStockItem struct {
	SerialNumber string `json:"SerialNumber,omitempty"`
	Msisdn       string `json:"msisdn,omitempty"`
}

type provisionRequest struct {
	StockItem                  []simStockItem `json:"stock"`
	PurchaseOrderNumber        string         `json:"purchaseOrderNumber"`
	TariffNumber               int            `json:"tariffNumber"`
	BoltOns                    []int          `json:"boltOns"`
	CreateNewProvisioningOrder bool           `json:"createNewProvisioningOrder"`
	ProvisioningOrderNumber    int            `json:"provisioningOrderNumber"`
}

func (pr *provisionRequest) addBoltOns(bolts []int) {
	pr.BoltOns = append(pr.BoltOns, bolts...)
}
func (pr *provisionRequest) validate() error {
	if len(pr.StockItem) == 0 {
		return fmt.Errorf("at least one stock item is required")
	}
	for _, si := range pr.StockItem {
		if si.SerialNumber == "" || si.Msisdn == "" {
			return fmt.Errorf("stock item must have serial number and msisdn")
		}
	}
	if pr.PurchaseOrderNumber == "" {
		return fmt.Errorf("purchase order number is required")
	}
	if pr.TariffNumber == 0 {
		return fmt.Errorf("tariff number is required")
	}
	return nil
}

type provisionResponse struct {
	TicketNumber string `json:"ticketNumber"`
}

func (api *Client) provisionSim(ctx context.Context, pr provisionRequest) (*provisionResponse, error) {
	err := pr.validate()
	if err != nil {
		return nil, err
	}
	body, err := json.Marshal(pr)
	u := api.baseUrl.ResolveReference(&url.URL{Path: "/api/SimSubmission/Provision"})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := api.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	var prResp provisionResponse
	err = json.NewDecoder(resp.Body).Decode(&prResp)
	if err != nil {
		return nil, err
	}
	return &prResp, nil
}

func (api *Client) unprovision(ctx context.Context, serialNumber string) (*provisionResponse, error) {
	if serialNumber == "" {
		return nil, fmt.Errorf("serial number is required")
	}
	r := map[string]interface{}{
		"serialNumbers": []string{serialNumber},
	}
	body, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}
	u := api.baseUrl.ResolveReference(&url.URL{Path: "/api/SimSubmission/Unprovision"})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := api.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	var prResp *provisionResponse
	err = json.NewDecoder(resp.Body).Decode(&prResp)
	if err != nil {
		return nil, err
	}
	return prResp, nil
}

func (api *Client) changeProvision(ctx context.Context, serialNumber string, tariffNumber int) (*provisionResponse, error) {
	if serialNumber == "" {
		return nil, fmt.Errorf("serial number is required")
	}
	r := struct {
		Stock        []simStockItem `json:"stock"`
		TariffNumber int            `json:"tariffNumber"`
	}{
		Stock:        []simStockItem{{SerialNumber: serialNumber}},
		TariffNumber: tariffNumber,
	}
	body, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}
	u := api.baseUrl.ResolveReference(&url.URL{Path: "/api/SimSubmission/ChangeProvision"})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := api.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	var prResp provisionResponse
	err = json.NewDecoder(resp.Body).Decode(&prResp)
	if err != nil {
		return nil, err
	}
	return &prResp, nil
}
