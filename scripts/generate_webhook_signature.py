#!/usr/bin/env python3
"""Generate an HMAC header locally, without network access or saved credentials."""

import argparse
import base64
import hashlib
import hmac
import os
from pathlib import Path


def compute_signature(payload: bytes, signing_key: str) -> str:
    digest = hmac.new(signing_key.encode(), payload, hashlib.sha256).digest()
    return base64.b64encode(digest).decode("ascii")


def main() -> None:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("payload", type=Path)
    parser.add_argument("--secret-env", default="WEBHOOK_SIGNING_SECRET",
                        help="Name of the environment variable holding the signing key")
    parser.add_argument("--header-name", default="X-Platform-Hmac-Sha256")
    args = parser.parse_args()
    signing_key = os.environ.get(args.secret_env)
    if not signing_key:
        parser.error("The requested signing-key environment variable is not set")
    signature = compute_signature(args.payload.read_bytes(), signing_key)
    print(f"{args.header_name}: {signature}")


if __name__ == "__main__":
    main()
