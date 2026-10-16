# Keeping up with the BYMA API

BYMA changes the Open Data API without notice: fields get renamed, endpoints
start requiring headers or a login. This is the routine for catching and
fixing that.

## What runs on its own

The **Integration** workflow runs every weekday at 14:00 ART, while the
market is open:

1. `integration_test.go` calls every public method against the live API.
2. The drift check (`internal/api/record_test.go`) compares each live
   response with the recorded fixture. It fails when a field the fixtures
   have is gone, and logs fields that are new.

A red **Live API** badge means one of the two broke. Start from its log.

## When something breaks

1. **Find what changed.**
   ```bash
   INTEGRATION_TEST=true go test -tags integration -run Integration -v ./...
   go test -tags record -run TestRecordFixtures -v ./internal/api/
   ```
   Lines starting with `DRIFT` name the file and field.

2. **If an endpoint returns 400/401, look at the web app.** BYMA's own site
   is the reference client.
   ```bash
   scripts/byma-webapp-endpoints.sh
   ```
   It prints the `Options` header the site sends per endpoint (compare with
   `sectionHeader` in `internal/api/endpoints.go`) and the endpoints this
   library doesn't call yet. For request bodies, search the downloaded
   `main.*.js` for the endpoint name.

3. **Probe by hand** with the same headers the client sends, for example:
   ```bash
   curl -s -X POST https://open.bymadata.com.ar/vanoms-be-core/rest/api/bymadata/free/bnown/byma-ads \
     -H 'Content-Type: application/json' -H 'Options: notices-and-events' \
     -H 'Origin: https://open.bymadata.com.ar' -d '{"Content-Type":"application/json"}'
   ```

4. **Fix the parser**, then re-record the fixtures so the tests use the new
   shape:
   ```bash
   RECORD_WRITE=1 go test -tags record -run TestRecordFixtures ./internal/api/
   go test ./...
   ```
   Prices in the new fixtures will differ, so update the expected values
   in `internal/api/api_test.go` and `client_fixture_test.go`.

5. **Add a live assertion** for the fix in `integration_test.go`, so it
   can't silently regress.

6. **Release.** Open a PR, wait for Test and Live API, merge, tag
   `vX.Y.Z`, and put upgrade notes at the top of the GitHub release.

## Things already learned

- Every request needs the `Options: <section>` header the web app sends.
  `byma-ads` (news) enforces it since 2026-10-16.
- History symbols need a settlement suffix: `"GGAL 24HS"`. A bare `"GGAL"`
  returns `no_data` without an error.
- 48HS no longer quotes since T+1 settlement; use 24HS or CI.
- `market-time` sends `timezone: "GMT-03:00"`, not an IANA name.
- File ids (`descarga`, `archivo`) are numbers. Download via
  `free/sba/download/{id}`; the `bolsar.com` links BYMA returns are dead.
- BYMA returns HTTP 400 at random (seen on `total-negotiated`); the same
  request succeeds on retry, so the client retries 400s.
- `bnown/bolsar/balances` redirects to an OAuth login (401).
- Futures prices come per 1000 and the client scales them.
- The free intraday endpoint returns delayed data only, a few bars a day.
