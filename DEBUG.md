# Debugging

## See every request

Pass a `Logger` in the options. The client logs at `Debug` level:

| Message | Fields |
|---|---|
| `Making request` | `method`, `url` |
| `Request completed` | `status_code`, `response_size` |
| `Retrying request` | `attempt`, `wait_time`, `url` |

```go
type stdLogger struct{}

func (stdLogger) Debug(msg string, f ...openbymadata.LogField) { log.Println("DEBUG", msg, f) }
func (stdLogger) Info(msg string, f ...openbymadata.LogField)  { log.Println("INFO", msg, f) }
func (stdLogger) Warn(msg string, f ...openbymadata.LogField)  { log.Println("WARN", msg, f) }
func (stdLogger) Error(msg string, f ...openbymadata.LogField) { log.Println("ERROR", msg, f) }

opts := openbymadata.DefaultClientOptions()
opts.Logger = stdLogger{}
client := openbymadata.NewClient(opts)
```

`NoOpLogger` (the default) discards everything.

## Common surprises

| Symptom | Cause |
|---|---|
| `GetHistory` returns no bars, no error | the symbol needs a settlement suffix: `"GGAL 24HS"` |
| `GetCurrentQuote` says "no quote available" | `Settlement48HS` (T+1 replaced it), or the market is closed: check `IsMarketOpen` |
| A field is always zero or empty | BYMA renamed it. Run the drift check below |
| `HTTP error 400` after a few seconds | BYMA failed the request on every retry; it does this under load. Try again later |
| `HTTP error 401` | BYMA moved the endpoint behind a login (e.g. `GetCompanyBalance`) |
| Same values for 5 minutes | the cache; call `ClearCache()` or disable it |

## Check whether BYMA changed something

From a clone of this repository:

```bash
# Every public method against the live API
INTEGRATION_TEST=true go test -tags integration -run Integration -v ./...

# Fields BYMA stopped (or started) sending, compared with the recorded fixtures
go test -tags record -run TestRecordFixtures -v ./internal/api/

# Endpoints added or removed in BYMA's web app
go run ./scripts/discover -probe
```

[docs/MAINTAINING.md](docs/MAINTAINING.md) explains how to read and fix each.

## Look at a raw response

Send the same request the client does, including the `Options` header from
`sectionHeader` in `internal/api/endpoints.go`:

```bash
curl -s -X POST https://open.bymadata.com.ar/vanoms-be-core/rest/api/bymadata/free/leading-equity \
  -H 'Content-Type: application/json' -H 'Options: renta-variable' \
  -H 'Origin: https://open.bymadata.com.ar' \
  -d '{"excludeZeroPxAndQty":false,"T2":false,"T1":true,"T0":false,"Content-Type":"application/json"}' | head -c 800
```
