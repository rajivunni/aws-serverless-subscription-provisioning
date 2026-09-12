package provisioning

import (
	"context"
	"github.com/rajivunni/aws-serverless-subscription-provisioning/internal/provisioning/provider"
	"github.com/rajivunni/aws-serverless-subscription-provisioning/internal/provisioning/store"
	"log"
)

type Provider struct {
	Api      *provider.Client
	PStore   *store.ProvisioningStore
	UnpStore *store.UnprovisioningStore
}

func (wf *Provider) Provision(ctx context.Context, msisdn string, tariffId, boltOn int, lookupTariff bool, meta *provider.ProvisionOrderMeta) error {
	result, err := wf.Api.Provision(ctx, msisdn, tariffId, boltOn, lookupTariff, meta)
	if err != nil {
		log.Printf("Provisioning operation operation failed; details omitted")
		return err
	}
	log.Printf("Provisioning operation status update")
	err = wf.PStore.InsertProvisioning(ctx, result, meta)
	if err != nil {
		log.Printf("Provisioning operation operation failed; details omitted")
		return err
	}
	log.Printf("Provisioning operation status update")
	return nil
}

func (wf *Provider) UnProvision(ctx context.Context, msisdn string, meta *provider.ProvisionOrderMeta) error {
	result, err := wf.Api.UnProvisioning(ctx, msisdn)
	if err != nil {
		log.Printf("Provisioning operation operation failed; details omitted")
		return err
	}
	log.Printf("Provisioning operation status update")
	err = wf.UnpStore.InsertUnprovisioning(ctx, result, meta)
	if err != nil {
		log.Printf("Provisioning operation operation failed; details omitted")
		return err
	}
	log.Printf("Provisioning operation status update")
	return nil
}

func (wf *Provider) SuspendSim(ctx context.Context, msisdn string, holdingTariffId int) error {
	result, err := wf.Api.SuspendSim(ctx, msisdn, holdingTariffId)
	if err != nil {
		log.Printf("Provisioning operation operation failed; details omitted")
		return err
	}
	log.Printf("Provisioning operation status update")
	err = wf.UnpStore.InsertUnprovisioning(ctx, result, &provider.ProvisionOrderMeta{})
	if err != nil {
		log.Printf("Provisioning operation operation failed; details omitted")
		return err
	}
	return nil
}
