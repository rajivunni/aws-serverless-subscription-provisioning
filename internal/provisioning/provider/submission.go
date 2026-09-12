package provider

import (
	"context"
	"fmt"
	"log"
	"strings"
)

type ProvisioningResult struct {
	TicketNumber string
	TariffNumber int
	TariffName   string
	BoltOnNumber int
	BoltOnName   string
	SerialNumber string
	Msisdn       string
	Iccid        string
	Imei         string
	NetworkId    int
}
type UnprovisioningResult struct {
	TicketNumber string
	SerialNumber string
	Iccid        string
	TariffName   string
	Msisdn       string
	Imei         string
	NetworkId    int
}
type ProvisionOrderMeta struct {
	CustomerName  string
	CustomerEmail string
	OrderId       string
}

// toPurchaseOrderNumber formats optional metadata for the concrete provider adapter.
// Contact data is sent to the provider, so a deployment needs a data-sharing review.
func (meta *ProvisionOrderMeta) toPurchaseOrderNumber() string {
	if meta == nil {
		return ""
	}
	var items []string
	for _, item := range []string{meta.CustomerName, meta.CustomerEmail} {
		if item != "" {
			items = append(items, item)
		}
	}
	return strings.Join(items, ",")
}

func (api *Client) Provision(ctx context.Context, msisdn string, tariffId, boltonId int, lookupTariff bool, meta *ProvisionOrderMeta) (*ProvisioningResult, error) {
	sim, err := api.findSim(ctx, msisdn)
	if err != nil {
		log.Printf("Provider submission operation failed; details omitted")
		return nil, err
	}
	if sim.Tariff != "" {
		log.Printf("Provider submission status update")
		return nil, fmt.Errorf("SIM %s already has tariff %s", sim.SerialNumber, sim.Tariff)
	}
	stI := simStockItem{
		SerialNumber: sim.SerialNumber,
		Msisdn:       sim.Msisdn,
	}
	var trf *tariff
	if lookupTariff {
		trfs, err := api.getTariffs(ctx, sim.NetworkId)
		if err != nil {
			log.Printf("Provider submission operation failed; details omitted")
			return nil, err
		}
		if len(trfs) == 0 {
			return nil, fmt.Errorf("no tariffs found for SIM %s", sim.SerialNumber)
		}
		for _, t := range trfs {
			if t.Number == tariffId {
				trf = &t
				break
			}
		}
		if trf == nil {
			return nil, fmt.Errorf("tariff %d not found", tariffId)
		}
	} else {
		trf = &tariff{Number: tariffId, Name: fmt.Sprintf("Tariff %d", tariffId)}
	}
	var boltOns []int
	var bolt *boltOn
	if trf.RequiresDataBoltOn {
		if boltonId == 0 {
			return nil, fmt.Errorf("tariff %d requires data bolt-on, but no bolt-on id provided", tariffId)
		}
		log.Printf("Provider submission status update")
		bolts, err := api.getBoltOns(ctx, tariffId)
		if err != nil {
			return nil, err
		}
		if len(bolts) == 0 {
			return nil, fmt.Errorf("no bolts available for tariff %d", tariffId)
		}
		for _, b := range bolts {
			if b.Number == boltonId {
				bolt = &b
				break
			}
		}
		if bolt == nil {
			return nil, fmt.Errorf("bolt-on %d not found", boltonId)
		}
		boltOns = append(boltOns, bolt.Number)
	}
	pon := meta.toPurchaseOrderNumber()
	if pon == "" {
		pon = fmt.Sprintf("PO-%s-%d", sim.SerialNumber, sim.NetworkId)
	}
	pr := provisionRequest{
		StockItem:                  []simStockItem{stI},
		PurchaseOrderNumber:        pon,
		TariffNumber:               trf.Number,
		BoltOns:                    boltOns,
		CreateNewProvisioningOrder: false,
		ProvisioningOrderNumber:    api.provisioningOrderId,
	}
	resp, err := api.provisionSim(ctx, pr)
	if err != nil {
		log.Printf("Provider submission operation failed; details omitted")
		return nil, err
	}
	var boltName string
	var boltNumber int
	if bolt != nil {
		boltName = bolt.Name
		boltNumber = bolt.Number
	}
	return &ProvisioningResult{
		TicketNumber: resp.TicketNumber,
		TariffNumber: trf.Number,
		TariffName:   trf.Name,
		BoltOnNumber: boltNumber,
		BoltOnName:   boltName,
		SerialNumber: sim.SerialNumber,
		Msisdn:       sim.Msisdn,
		NetworkId:    sim.NetworkId,
		Iccid:        sim.Iccid,
		Imei:         sim.Imei,
	}, nil
}

func (api *Client) UnProvisioning(ctx context.Context, msisdn string) (*UnprovisioningResult, error) {
	sim, err := api.findSim(ctx, msisdn)
	if err != nil {
		log.Printf("Provider submission operation failed; details omitted")
		return nil, err
	}
	if sim.Tariff == "" {
		return nil, fmt.Errorf("tariff for SIM %s is empty", sim.SerialNumber)
	}
	resp, err := api.unprovision(ctx, sim.SerialNumber)
	if err != nil {
		log.Printf("Provider submission operation failed; details omitted")
		return nil, err
	}
	return &UnprovisioningResult{
		TicketNumber: resp.TicketNumber,
		SerialNumber: sim.SerialNumber,
		Iccid:        sim.Iccid,
		TariffName:   sim.Tariff,
		Msisdn:       sim.Msisdn,
		Imei:         sim.Imei,
		NetworkId:    sim.NetworkId,
	}, nil
}

func (api *Client) SuspendSim(ctx context.Context, msisdn string, holdingTariffId int) (*UnprovisioningResult, error) {
	sim, err := api.findSim(ctx, msisdn)
	if err != nil {
		return nil, err
	}
	if sim.Tariff == "" {
		return nil, fmt.Errorf("SIM %s has no active tariff, cannot suspend", sim.SerialNumber)
	}
	resp, err := api.changeProvision(ctx, sim.SerialNumber, holdingTariffId)
	if err != nil {
		return nil, fmt.Errorf("failed to change provision for SIM %s: %v", sim.SerialNumber, err)
	}
	return &UnprovisioningResult{
		TicketNumber: resp.TicketNumber,
		SerialNumber: sim.SerialNumber,
		Iccid:        sim.Iccid,
		TariffName:   sim.Tariff,
		Msisdn:       sim.Msisdn,
		NetworkId:    sim.NetworkId,
	}, nil
}
