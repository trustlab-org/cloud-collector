# collector

Pulls security audit logs from SaaS platforms and produces them to Kafka, one event per message, ready for a SIEM pipeline.

**Sources:** Google Workspace (Admin SDK Reports API). Microsoft 365 and Zoho planned.

<p>
  <a href="https://github.com/warpstreamlabs/bento"><img src="https://raw.githubusercontent.com/warpstreamlabs/bento/main/icon.png" width="64" alt="Bento" align="left"></a>
  Built on <a href="https://github.com/warpstreamlabs/bento">Bento</a>, the MIT-licensed stream processor, used as a library.
  The collector adds the source plugins; Bento provides the Kafka output, batching, retries, metrics, config and graceful shutdown.
</p>
<br clear="left">

## What it guarantees

-  Each log type keeps a persistent cursor. The cursor advances only after Kafka acknowledges every event in the window.
- **Late events caught.** Every poll re-checks the last few minutes. A periodic sweep reaches back a per-log-type overlap based on Google's published lag times.
- **No duplicates.** Deterministic event ids are used for dedupe and as the Kafka message key.
- **Polite to the API.** Exponential backoff on 429 and 5xx responses.
- **Observable.** Prometheus metrics for events emitted, late arrivals and lateness, poll errors, and cursor lag per log type.
- **One static binary.** No runtime to install, which suits air-gapped servers.

## Quick start

Requires Go 1.27+.

```bash
go build -o bin/collector ./cmd/collector
go build -o bin/fakegws  ./cmd/fakegws

./bin/fakegws -rate 5 -late 0.2 -throttle 0.1     # fake Google API with synthetic logs
./bin/collector -c config/fake.yaml               
```

Metrics: `http://127.0.0.1:4195/metrics`

## Against a real Google Workspace tenant

1. Create a Google Cloud service account and download its JSON key.
2. In the Admin console, grant it domain-wide delegation with the scope `https://www.googleapis.com/auth/admin.reports.audit.readonly`.
3. Run:

```bash
export GWS_CREDENTIALS_FILE=/run/secrets/google-sa.json
export GWS_ADMIN_SUBJECT=admin@univ.ac.in
export GWS_DOMAIN=univ.ac.in
export COLLECTOR_OUTPUT=kafka KAFKA_BROKERS=kafka:9094
./bin/collector -c config/google-workspace.yaml
```

Mount the key at runtime; never bake it into an image.

## Output

```json
{
  "meta": {"org": "…", "server": "univ.ac.in", "source_program": "google_workspace_login", "event_id": "9af4…"},
  "json": { "id": {…}, "actor": {…}, "ipAddress": "…", "events": [ {"name": "login_failure", "parameters": […]} ] }
}
```

`json` keeps Google's original schema, so existing detection rules (e.g. Sigma) apply unchanged.



