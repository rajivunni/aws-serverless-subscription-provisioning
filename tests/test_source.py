"""Offline packaging checks. These do not validate an AWS deployment."""

import base64
import hashlib
import hmac
import importlib.util
import json
from pathlib import Path
import re
import unittest


ROOT = Path(__file__).resolve().parents[1]


class SourceChecks(unittest.TestCase):
    def test_fixtures_are_minimal_synthetic_events(self):
        paths = [ROOT / "data/event.json", *sorted((ROOT / "internal/webhook/testdata").glob("*.json"))]
        self.assertEqual(len(paths), 3)
        allowed = {"id", "order_placed", "cancellation_scheduled_for", "email", "order_id",
                   "customer_id", "status", "log", "items", "first_name", "last_name"}
        for path in paths:
            event = json.loads(path.read_text())
            self.assertEqual(set(event), allowed)
            self.assertEqual(event["email"], "subscriber@example.invalid")
            self.assertTrue(event["order_id"].startswith("DEMO-ORDER-"))
            self.assertTrue(event["customer_id"].startswith("DEMO-CUSTOMER-"))
            self.assertEqual(event["log"], [])
            self.assertEqual(event["items"][0]["properties"][0]["value"], "+12025550123")

    def test_template_configuration_references_exist(self):
        template = (ROOT / "template.yaml").read_text()
        mappings = (ROOT / "mappings.yaml.example").read_text()
        keys = set(re.findall(r"^  (\w+):", mappings, re.MULTILINE))
        references = set(re.findall(r"!FindInMap\s*\[\s*AppConfig,\s*default,\s*(\w+)\s*\]", template))
        self.assertTrue(references)
        self.assertLessEqual(references, keys)
        self.assertIn("State: DISABLED", template)
        self.assertNotIn("State: ENABLED", template)
        self.assertFalse((ROOT / "mappings.yaml").exists())

    def test_example_endpoints_are_reserved(self):
        mappings = (ROOT / "mappings.yaml.example").read_text()
        for host in re.findall(r"https://([^/\s\"]+)", mappings):
            self.assertTrue(host.endswith(".example.invalid"))
        for account in re.findall(r"arn:aws:secretsmanager:[^:]+:(\d+):", mappings):
            self.assertEqual(account, "000000000000")

    def test_state_machine_json_targets(self):
        paths = sorted((ROOT / "step-functions").glob("*.json"))
        self.assertEqual(len(paths), 2)
        for path in paths:
            document = json.loads(path.read_text())
            states = document["States"]
            self.assertIn(document["StartAt"], states)
            def walk(value):
                if isinstance(value, dict):
                    for key, item in value.items():
                        if key in {"Next", "Default"}:
                            self.assertIn(item, states)
                        walk(item)
                elif isinstance(value, list):
                    for item in value:
                        walk(item)
            walk(states)

    def test_signature_helper_matches_standard_hmac(self):
        path = ROOT / "scripts/generate_webhook_signature.py"
        spec = importlib.util.spec_from_file_location("signature_helper", path)
        module = importlib.util.module_from_spec(spec)
        spec.loader.exec_module(module)
        payload = (ROOT / "data/event.json").read_bytes()
        key = "synthetic-signing-key-not-for-use"
        expected = base64.b64encode(hmac.new(key.encode(), payload, hashlib.sha256).digest()).decode()
        self.assertEqual(module.compute_signature(payload, key), expected)
        self.assertNotEqual(module.compute_signature(payload + b" ", key), expected)

    def test_module_and_build_targets(self):
        namespace = "github.com/rajivunni/aws-serverless-subscription-provisioning"
        self.assertEqual((ROOT / "go.mod").read_text().splitlines()[0], "module " + namespace)
        makefile = (ROOT / "Makefile").read_text()
        for entry in ("webhook", "provisioner", "provider-webhook", "email-sender"):
            self.assertIn("./cmd/" + entry, makefile)
            self.assertTrue((ROOT / "cmd" / entry / "main.go").is_file())

    def test_no_sensitive_or_binary_artifacts(self):
        forbidden = {"bootstrap", "samconfig.toml", "mappings.yaml", ".env", "credentials"}
        for path in ROOT.rglob("*"):
            if not path.is_file() or "__pycache__" in path.parts:
                continue
            self.assertNotIn(path.name, forbidden)
            self.assertNotIn(path.suffix, {".pem", ".key", ".p12", ".pfx", ".zip", ".log"})
            self.assertNotIn(b"\x00", path.read_bytes(), msg=str(path.relative_to(ROOT)))


if __name__ == "__main__":
    unittest.main()
