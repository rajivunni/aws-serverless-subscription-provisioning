# Architecture and implementation boundaries

This document describes the supplied reference code, not an audited deployment.

## Request flow

1. A subscription event reaches API Gateway and the webhook Lambda. The handler verifies an HMAC, parses the event, writes an event record, and queues a workflow item.
2. The provisioning state machine invokes a Lambda that reads due workflow items and submits provider operations. Result records link subsequent provider callbacks to workflow data.
3. A provider callback checks a configured shared header, updates result state, and queues an email command. The email state machine invokes the Brevo API through an email Lambda.

The template declares four Lambda functions, five DynamoDB tables, two Step Functions state machines, and a disabled-by-default 15-minute EventBridge polling rule. The JSON state-machine definitions contain retry policies for Lambda invocation failures. They do not independently retry each failed work item.

## Data and adapters

The fixtures contain only invented events, example.invalid addresses, and a reserved fictional telephone number. Generic field names such as customer_id and ticketNumber remain because they define the integration schema. No real record is needed to understand the control flow.

The implementation retains a concrete Cellhire-style provider adapter and Brevo email adapter. Provider URL, authentication URL, account/company IDs, tariff IDs, and provisioning-order ID are configuration inputs. They must come from an authorized test environment before any integration test. Example endpoints are intentionally non-operational.

## Reliability limitations

- An event receipt and its workflow insertion are separate writes. Duplicate processing checks do not provide an end-to-end transaction.
- Workflow items are not conditionally claimed. Concurrent executions can operate on the same item, and some status-update errors are not propagated.
- Cancellation currently schedules fixed delays. Parsing a scheduled cancellation time does not mean scheduling uses it. Reactivation does not demonstrate cancellation of all older deferred actions.
- Suspension failure falls back to unprovisioning. This behavior needs an explicit business decision before live use.
- Provider retries reuse request bodies without a demonstrated rewind strategy and retry modifying requests without a proven idempotency key.
- Email command deduplication is non-atomic. Ticket lookup has a limited scan without complete pagination. Failed sends are marked failed, while the batch handler can still return success.
- A successful provider submission is an accepted request, not proof of a completed service change.

The preparation removes the duplicate direct suspension call after successful dispatch. It does not resolve the broader state-machine or provider-idempotency limitations above.

## Operating boundary

Only local tests and compilation checks can be established for this package. No claim is made about current third-party API compatibility, delivery guarantees, capacity, monitoring coverage, AWS costs, or regulatory compliance. A deployment needs a dedicated test account, provider sandbox or mock, failure-injection tests, resource-level IAM review, and an operator-approved lifecycle policy.
