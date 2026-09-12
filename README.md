# AWS serverless subscription provisioning

An anonymized source-code reference for processing subscription webhooks, scheduling provider operations, and sending status notifications. All included event fixtures and configuration examples are synthetic.

This is a portfolio reference, not a deploy-ready product or a production-readiness claim. No cloud deployment or live provider request was performed during preparation. See [SECURITY.md](SECURITY.md) before adapting it.

## What the code demonstrates

- Go Lambda handlers for subscription events, provider operations, provider callbacks, and email dispatch.
- DynamoDB-backed event, workflow, result, and email-command records.
- Two Step Functions definitions and an EventBridge polling schedule.
- HMAC-SHA256 verification for subscription payloads and a shared-header check for provider callbacks.

The provider package contains a concrete Cellhire-style API adapter. Its endpoints and request shapes are integration-specific, not a universal telecom interface. The email sender uses the Brevo transactional-email API, not AWS SES. A different provider requires adapter and contract-test changes.

## Local checks

Python 3 can run the packaging checks without dependencies or network access:

```sh
python3 -m unittest discover -s tests -v
```

With the Go version declared in `go.mod` or a compatible newer toolchain:

```sh
go test ./internal/...
go vet ./...
go build ./...
```

These commands compile or test source. They do not deploy or run Lambda entrypoints. The first Go invocation may download module dependencies. Do not run the `cmd` programs against real credentials for a local test.

`scripts/generate_webhook_signature.py` calculates a signature over the exact bytes of a local payload. It reads a signing key only from a named environment variable. The default name is `WEBHOOK_SIGNING_SECRET`. Use a throwaway value and synthetic payload for demonstrations. It does not load a production profile or call an endpoint.

## Configuration and deployment boundary

`mappings.yaml.example` contains reserved example domains and explicit placeholders. An operator would need an ignored local `mappings.yaml`, scoped Secrets Manager entries, provider account/tariff/order identifiers, and a verified email sender. No credentials, populated mapping, deployment profile, compiled binary, or automated deployment workflow is included.

The SAM template is retained to explain the resource relationships. It has not been deployed or validated against a configured AWS account. Its recurring schedule is disabled by default. Review runtime support, resource naming, IAM, retention, failure handling, and provider contracts before any deployment. Enabling it can create billable resources and issue real provider or email requests.

## Known limitations

Workflow updates and email deduplication are not atomic. Overlapping executions and partial writes can lose or duplicate work. Cancellation uses fixed delays rather than the parsed cancellation timestamp, and stale deferred work is not demonstrably cancelled on reactivation. Provider submission acceptance is not confirmation that provisioning completed.

HTTP retries do not establish provider idempotency. Failed individual emails are marked failed without being retried by the surrounding state machine. The callback authentication and operational controls also need further work. These limitations are described in [ARCHITECTURE.md](ARCHITECTURE.md) and [SECURITY.md](SECURITY.md).

## License

The supplied MIT license and copyright notice are preserved in [LICENSE](LICENSE). Removing identifying information does not establish permission to publish third-party or client-derived intellectual property. Confirm those rights separately before uploading a derived package.
