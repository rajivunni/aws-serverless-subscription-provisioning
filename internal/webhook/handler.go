package webhook

import (
	"context"
	"errors"
	"fmt"
	"github.com/rajivunni/aws-serverless-subscription-provisioning/internal/orchestration"
	"log"
	"time"

	"github.com/aws/aws-secretsmanager-caching-go/secretcache"
)

type HandleResult int

const (
	SignatureVerificationFailed HandleResult = iota
	ParsingFailed
	UnknownError
	AlreadyHandled
	Submitted
)

type Handler struct {
	verifier *Verifier
	store    *EventStore
	wfs      *orchestration.WorkflowStore
	st       *StateMachine
	tariffId int
	boltOnId int
}

func NewHandler(signatureSecretId string, eventStore *EventStore, wfs *orchestration.WorkflowStore, st *StateMachine, tariffId, boltOnId int) (*Handler, error) {
	cache, err := secretcache.New()
	if err != nil {
		return nil, err
	}
	secret, err := cache.GetSecretString(signatureSecretId)
	if err != nil {
		return nil, err
	}
	verifier, err := NewVerifier(secret)
	if err != nil {
		return nil, err
	}
	return &Handler{
		verifier: verifier,
		store:    eventStore,
		wfs:      wfs,
		st:       st,
		tariffId: tariffId,
		boltOnId: boltOnId,
	}, nil
}

func sevenDaysFromNow() int64 {
	return time.Now().AddDate(0, 0, 7).Unix()
}

func clampToNow(ts int64) int64 {
	now := time.Now().Unix()
	if ts < now {
		return now
	}
	return ts
}

func (h *Handler) findSubscriptionExpirationTime(event *Event) (time.Time, error) {
	return event.CancellationExpirationTime()
}

func (h *Handler) executeAfterForEvent(event *Event) (int64, error) {
	now := time.Now().Unix()
	switch event.Status {
	case Active, Expired:
		return now, nil
	case Cancelled:
		// This reference uses a fixed 30-day delay, not the parsed cancellation timestamp.
		return time.Now().Add(30 * 24 * time.Hour).Unix(), nil
	default:
		return 0, fmt.Errorf("unknown status %s", event.Status)
	}
}

func (h *Handler) Handle(ctx context.Context, signature string, payload []byte) (HandleResult, error) {
	log.Printf("Subscription event status update")
	err := h.verifier.Verify(signature, payload)
	if err != nil {
		log.Printf("Subscription event operation failed; details omitted")
		return SignatureVerificationFailed, nil
	}
	event, err := Parse(payload)
	if err != nil {
		log.Printf("Subscription event operation failed; details omitted")
		return ParsingFailed, nil
	}
	if event == nil {
		log.Printf("Subscription event status update")
		return UnknownError, errors.New("unknown error while handling webhook event")
	}
	msisdn, err := event.Msisdn()
	if err != nil {
		log.Printf("Subscription event operation failed; details omitted")
		return ParsingFailed, nil
	}
	log.Printf("Subscription event status update")

	key := event.Key()
	existing, err := h.store.FindById(ctx, key)

	if existing != nil {
		log.Printf("Subscription event status update")
		return AlreadyHandled, nil
	}

	dbE := EventRecord{
		Id:           key,
		ReceivedAt:   time.Now().Unix(),
		Email:        event.Email,
		OrderId:      event.OrderId,
		Status:       event.Status,
		Msisdn:       msisdn,
		Ttl:          sevenDaysFromNow(),
		CustomerName: event.CustomerName(),
	}

	r, err := h.store.MarkReceived(ctx, dbE)

	if err != nil {
		log.Printf("Subscription event operation failed; details omitted")
		return UnknownError, errors.New("failed to store event in database")
	}
	log.Printf("Subscription event status update")
	err = h.AddFromWebhook(ctx, event, &dbE)
	if err != nil {
		log.Printf("Subscription event operation failed; details omitted")
		return UnknownError, errors.New("failed to add workflow item for event")
	}
	if r == DuplicateEvent {
		log.Printf("Subscription event status update")
		return AlreadyHandled, nil
	}
	log.Printf("Subscription event status update")
	err = h.st.startProvisioner(ctx, &dbE)
	if err != nil {
		log.Printf("Subscription event operation failed; details omitted")
		return UnknownError, errors.New("failed to start provisioning workflow")
	}
	return Submitted, nil
}

func (h *Handler) AddFromWebhook(ctx context.Context, parsedEvent *Event, event *EventRecord) error {
	var action orchestration.Action
	switch event.Status {
	case Active:
		action = orchestration.Provision
	case Cancelled:
		action = orchestration.Suspend
	case Expired:
		action = orchestration.Unprovision
	default:
		err := fmt.Errorf("unknown status %s", event.Status)
		return err
	}

	executeAfter, err := h.executeAfterForEvent(parsedEvent)
	if err != nil {
		return err
	}

	now := time.Now().Unix()
	item := &orchestration.WorkflowItem{
		Id:           event.Id,
		Action:       action,
		State:        orchestration.Pending,
		Source:       orchestration.Webhook,
		Error:        "",
		CreatedAt:    now,
		UpdatedAt:    now,
		ExecuteAfter: executeAfter,
		Data: map[string]interface{}{
			"msisdn":        event.Msisdn,
			"tariffId":      h.tariffId,
			"boltOn":        h.boltOnId,
			"customerEmail": event.Email,
			"customerName":  event.CustomerName,
			"orderId":       event.OrderId,
			"status":        event.Status,
			"receivedAt":    event.ReceivedAt,
		},
	}
	return h.wfs.Add(ctx, *item)
}
