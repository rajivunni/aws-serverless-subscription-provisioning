package orchestration

import (
	"context"
	"fmt"
	"github.com/rajivunni/aws-serverless-subscription-provisioning/internal/provisioning"
	"github.com/rajivunni/aws-serverless-subscription-provisioning/internal/provisioning/provider"
	"log"
	"sync"
	"time"
)

type Orchestrator struct {
	HoldingTariffId int
	Wfs             *WorkflowStore
	Ch              *provisioning.Provider
}

type provisioningParams struct {
	msisdn       string
	tariffId     int
	boltOn       int
	lookupTariff bool
}

type unprovisioningParams struct {
	msisdn string
}

func (item *WorkflowItem) asProvisioningParams() (*provisioningParams, error) {
	data := item.Data
	log.Printf("Provisioning workflow status update")
	if data == nil {
		return nil, fmt.Errorf("item[%s] has no data", item.Id)
	}
	msisdn, ok := data["msisdn"].(string)
	if !ok {
		return nil, fmt.Errorf("item[%s] has no msisdn", item.Id)
	}
	v, ok := data["tariffId"].(float64)
	if !ok {
		return nil, fmt.Errorf("item[%s] has no tariffId", item.Id)
	}
	tariffId := int(v)
	v, ok = data["boltOn"].(float64)
	if !ok {
		return nil, fmt.Errorf("workflow boltOn must be numeric")
	}
	boltOn := int(v)
	return &provisioningParams{msisdn: msisdn, tariffId: tariffId, boltOn: boltOn, lookupTariff: true}, nil
}

func (item *WorkflowItem) makeOrderMeta() (*provider.ProvisionOrderMeta, error) {
	data := item.Data
	if data == nil {
		return nil, fmt.Errorf("item[%s] has no data", item.Id)
	}
	orderId, ok := data["orderId"].(string)
	if !ok {
		return nil, fmt.Errorf("item[%s] has no orderId", item.Id)
	}
	customerName, _ := data["customerName"].(string)
	email, _ := data["customerEmail"].(string)
	log.Printf("Provisioning workflow status update")
	return &provider.ProvisionOrderMeta{OrderId: orderId, CustomerName: customerName, CustomerEmail: email}, nil
}

func (item *WorkflowItem) asUnProvisionParams() (*unprovisioningParams, error) {
	data := item.Data
	if data == nil {
		return nil, fmt.Errorf("item[%s] has no data", item.Id)
	}
	msisdn, ok := data["msisdn"].(string)
	if !ok {
		return nil, fmt.Errorf("item[%s] has no msisdn", item.Id)
	}
	return &unprovisioningParams{msisdn: msisdn}, nil
}

func (orch *Orchestrator) Provision(ctx context.Context, item WorkflowItem) error {
	if item.State != Pending {
		log.Printf("Provisioning workflow status update")
		return fmt.Errorf("cannot provision item in state %s", item.State)
	}
	if item.State == Failure {
		log.Printf("Provisioning workflow status update")
		return fmt.Errorf("cannot provision item in state %s", item.State)
	}
	err := orch.Wfs.UpdateState(ctx, item.Id, Running, "")
	if err != nil {
		log.Printf("Provisioning workflow operation failed; details omitted")
		return err
	}
	log.Printf("Provisioning workflow status update")
	parms, err := item.asProvisioningParams()
	if err != nil {
		log.Printf("Provisioning workflow operation failed; details omitted")
		if uptErr := orch.Wfs.UpdateState(ctx, item.Id, Failure, err.Error()); uptErr != nil {
			log.Printf("Provisioning workflow operation failed; details omitted")
		}
		return err
	}
	meta, err := item.makeOrderMeta()
	if err != nil {
		log.Printf("Provisioning workflow operation failed; details omitted")
		if uptErr := orch.Wfs.UpdateState(ctx, item.Id, Failure, err.Error()); uptErr != nil {
			log.Printf("Provisioning workflow operation failed; details omitted")
		}
	}
	log.Printf("Provisioning workflow status update")
	err = orch.Ch.Provision(ctx, parms.msisdn, parms.tariffId, parms.boltOn, parms.lookupTariff, meta)
	if err != nil {
		log.Printf("Provisioning workflow operation failed; details omitted")
		if uptErr := orch.Wfs.UpdateState(ctx, item.Id, Failure, err.Error()); uptErr != nil {
			log.Printf("Provisioning workflow operation failed; details omitted")
		}
	} else {
		log.Printf("Provisioning workflow status update")
		if uptErr := orch.Wfs.UpdateState(ctx, item.Id, Success, ""); uptErr != nil {
			log.Printf("Provisioning workflow operation failed; details omitted")
		}
	}
	return err
}

func (orch *Orchestrator) Unprovision(ctx context.Context, item WorkflowItem) error {
	if item.State != Pending {
		log.Printf("Provisioning workflow status update")
		return fmt.Errorf("cannot provision item in state %s", item.State)
	}
	if item.State == Failure {
		log.Printf("Provisioning workflow status update")
		return fmt.Errorf("cannot provision item in state %s", item.State)
	}
	err := orch.Wfs.UpdateState(ctx, item.Id, Running, "")
	if err != nil {
		log.Printf("Provisioning workflow operation failed; details omitted")
		return err
	}
	log.Printf("Provisioning workflow status update")
	params, err := item.asUnProvisionParams()
	if err != nil {
		log.Printf("Provisioning workflow operation failed; details omitted")
		if uptErr := orch.Wfs.UpdateState(ctx, item.Id, Failure, err.Error()); uptErr != nil {
			log.Printf("Provisioning workflow operation failed; details omitted")
		}
		return err
	}
	meta, err := item.makeOrderMeta()
	if err != nil {
		log.Printf("Provisioning workflow operation failed; details omitted")
		if uptErr := orch.Wfs.UpdateState(ctx, item.Id, Failure, err.Error()); uptErr != nil {
			log.Printf("Provisioning workflow operation failed; details omitted")
		}
	}
	log.Printf("Provisioning workflow status update")
	err = orch.Ch.UnProvision(ctx, params.msisdn, meta)
	if err != nil {
		log.Printf("Provisioning workflow operation failed; details omitted")
		if uptErr := orch.Wfs.UpdateState(ctx, item.Id, Failure, err.Error()); uptErr != nil {
			log.Printf("Provisioning workflow operation failed; details omitted")
		}
		return err
	}
	log.Printf("Provisioning workflow status update")
	if uptErr := orch.Wfs.UpdateState(ctx, item.Id, Success, ""); uptErr != nil {
		log.Printf("Provisioning workflow operation failed; details omitted")
	}
	return nil
}

type ItemFailure struct {
	Item  WorkflowItem
	Error error
}

type Result struct {
	Failures  []ItemFailure
	Successes []WorkflowItem
}

func (orch *Orchestrator) doProvision(ctx context.Context, item WorkflowItem) *ItemFailure {
	err := orch.Provision(ctx, item)
	if err != nil {
		return &ItemFailure{Item: item, Error: err}
	}
	return nil
}
func (orch *Orchestrator) doUnProvision(ctx context.Context, item WorkflowItem) *ItemFailure {
	err := orch.Unprovision(ctx, item)
	if err != nil {
		return &ItemFailure{Item: item, Error: err}
	}
	return nil
}

func (orch *Orchestrator) doRun(ctx context.Context, item WorkflowItem) *ItemFailure {
	var failure *ItemFailure
	if item.Action == Provision {
		failure = orch.doProvision(ctx, item)
	} else if item.Action == Suspend {
		failure = &ItemFailure{Item: item, Error: orch.Suspend(ctx, item)}
		if failure.Error == nil {
			failure = nil
		}
	} else {
		failure = orch.doUnProvision(ctx, item)
	}
	return failure
}

func dueForExecution(item WorkflowItem, now int64) bool {
	if item.ExecuteAfter == 0 {
		return true
	}
	return item.ExecuteAfter <= now
}

func (orch *Orchestrator) Run(ctx context.Context) (*Result, error) {
	items, err := orch.Wfs.GetPendingItems(ctx)

	if err != nil {
		log.Printf("Provisioning workflow operation failed; details omitted")
		return nil, fmt.Errorf("failed to get pending items, reason: %s", err.Error())
	}

	if len(items) == 0 {
		log.Printf("Provisioning workflow status update")
		return &Result{}, nil
	}

	now := time.Now().Unix()
	dueItems := make([]WorkflowItem, 0, len(items))
	for _, item := range items {
		if !dueForExecution(item, now) {
			log.Printf("Provisioning workflow status update")
			continue
		}
		dueItems = append(dueItems, item)
	}

	if len(dueItems) == 0 {
		log.Printf("Provisioning workflow status update")
		return &Result{}, nil
	}

	var (
		failures  []ItemFailure
		successes []WorkflowItem
		mu        sync.Mutex
	)
	sem := make(chan struct{}, 2)
	var wg sync.WaitGroup
	for _, i := range dueItems {
		item := i
		wg.Add(1)
		sem <- struct{}{}
		go func() {
			defer wg.Done()
			defer func() { <-sem }()
			failure := orch.doRun(ctx, item)
			mu.Lock()
			if failure != nil {
				failures = append(failures, *failure)
			} else {
				successes = append(successes, item)
			}
			mu.Unlock()
		}()
	}
	wg.Wait()

	return &Result{failures, successes}, nil

}

func (orch *Orchestrator) Suspend(ctx context.Context, item WorkflowItem) error {
	if item.State != Pending {
		return fmt.Errorf("cannot suspend item in state %s", item.State)
	}
	err := orch.Wfs.UpdateState(ctx, item.Id, Running, "")
	if err != nil {
		return err
	}
	params, err := item.asUnProvisionParams()
	if err != nil {
		if uptErr := orch.Wfs.UpdateState(ctx, item.Id, Failure, err.Error()); uptErr != nil {
			log.Printf("Provisioning workflow operation failed; details omitted")
		}
		return err
	}
	meta, _ := item.makeOrderMeta()
	log.Printf("Provisioning workflow status update")
	suspendErr := orch.Ch.SuspendSim(ctx, params.msisdn, orch.HoldingTariffId)
	if suspendErr != nil {
		log.Printf("Provisioning workflow operation failed; details omitted")
		unprovErr := orch.Ch.UnProvision(ctx, params.msisdn, meta)
		if unprovErr != nil {
			log.Printf("Provisioning workflow operation failed; details omitted")
			if uptErr := orch.Wfs.UpdateState(ctx, item.Id, Failure, fmt.Sprintf("suspend: %s; unprovision: %s", suspendErr.Error(), unprovErr.Error())); uptErr != nil {
				log.Printf("Provisioning workflow operation failed; details omitted")
			}
			return unprovErr
		}
		log.Printf("Provisioning workflow status update")
	} else {
		log.Printf("Provisioning workflow status update")
		followUp := WorkflowItem{
			Id:           fmt.Sprintf("unprov_%s_%d", item.Id, time.Now().Unix()),
			Action:       Unprovision,
			State:        Pending,
			Source:       Source("auto"),
			Error:        "",
			Data:         item.Data,
			CreatedAt:    time.Now().Unix(),
			UpdatedAt:    time.Now().Unix(),
			ExecuteAfter: time.Now().Unix() + (30 * 24 * 60 * 60),
		}
		if addErr := orch.Wfs.Add(ctx, followUp); addErr != nil {
			log.Printf("Provisioning workflow operation failed; details omitted")
		} else {
			log.Printf("Provisioning workflow status update")
		}
	}
	if uptErr := orch.Wfs.UpdateState(ctx, item.Id, Success, ""); uptErr != nil {
		log.Printf("Provisioning workflow operation failed; details omitted")
	}
	return nil
}
