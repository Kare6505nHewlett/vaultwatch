# vaultwatch

A CLI tool that monitors HashiCorp Vault secret expiry and sends alerts before leases or tokens expire.

---

## Installation

```bash
go install github.com/youruser/vaultwatch@latest
```

Or build from source:

```bash
git clone https://github.com/youruser/vaultwatch.git
cd vaultwatch && go build -o vaultwatch .
```

---

## Usage

Set your Vault address and token, then run `vaultwatch` with a warning threshold:

```bash
export VAULT_ADDR="https://vault.example.com"
export VAULT_TOKEN="s.xxxxxxxxxxxxxxxx"

# Alert if any lease or token expires within 24 hours
vaultwatch watch --threshold 24h
```

**Example output:**

```
[WARN] Token s.abc123 expires in 18h32m  (path: auth/token/lookup-self)
[WARN] Lease lease/database/creds/my-role expires in 6h10m
[OK]   All other secrets are healthy.
```

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--threshold` | `24h` | Warn if expiry is within this duration |
| `--interval` | `5m` | How often to poll Vault |
| `--alert-webhook` | — | Slack/webhook URL for notifications |
| `--log-format` | `text` | Output format: `text` or `json` |

---

## Configuration

`vaultwatch` respects standard Vault environment variables (`VAULT_ADDR`, `VAULT_TOKEN`, `VAULT_CACERT`, etc.) and can also be configured via a `vaultwatch.yaml` file in the working directory.

---

## License

MIT © 2024 youruser