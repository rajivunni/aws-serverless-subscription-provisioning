# Security and safe use

This package is an anonymized reference. It is not approved for a production deployment. Do not populate fixtures with customer data or test handlers using production credentials.

## Publication preparation

Client branding, customer fixtures, private endpoints, account identifiers, and embedded provider configuration identifiers are removed or replaced by synthetic examples. Credentials are not included. Runtime logs must not contain shared secrets, authentication responses, headers, payloads, customer metadata, or identity-derived execution names. Error details can themselves contain identifiers, so generic failure messages are used where appropriate.

The original MIT copyright notice is retained. Repository content review and source-publication permission are separate requirements. Confirm authorization to publish any derived intellectual property before uploading.

## Authentication limitations

Subscription requests use HMAC-SHA256 with constant-time comparison, but there is no signed freshness or replay window. Event-record deduplication is not a substitute for replay protection. Provider callbacks use a static shared header with constant-time comparison and reject empty values, but do not verify a signed payload. Secrets are loaded during initialization, so automatic refresh for warm processes is not demonstrated.

Several error paths acknowledge requests with HTTP 200. Authentication and retry behavior must be contract-tested with the actual upstream service. Do not assume failed deliveries will be retried or that every accepted response represents completed work.

## Deployment review still required

- Replace broad table CRUD grants with permissions justified by each handler. Secret resources are mapping-scoped, but example mappings do not prove a deployed policy is safe.
- Configure log retention, DynamoDB recovery and retention, API abuse protection, alarm coverage, and a failure/reconciliation process.
- Review records and persisted error messages for personal data. The provider metadata currently uses customer contact information, so retention and third-party transfer decisions remain necessary.
- Resolve non-atomic workflow/email processing, request retry safety, cancellation timing, and suspension fallback before real service changes.
- Keep populated mappings, environment files, credentials, logs, and binaries outside Git history. A scanner is an additional check, not proof that every sensitive value has been found.

If a credential was exposed in another copy or log, remove it from shared content and revoke or rotate it through the credential owner. Deleting a file does not revoke a credential.
