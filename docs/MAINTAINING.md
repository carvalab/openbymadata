# Keeping up with the BYMA API

BYMA changes the Open Data API without notice: fields get renamed, endpoints
start requiring headers or a login, new endpoints show up in their web app.
This is the routine for catching that and deciding what to **add**,
**remove** or **edit**.

## What runs on its own

| Workflow | When | Catches |
|---|---|---|
| **Integration** | weekdays 14:00 ART | a public method failing against the live API (`integration_test.go`), or a field BYMA stopped sending (drift check in `internal/api/record_test.go`) |
| **Discover** | Mondays 12:00 ART | endpoints BYMA's web app calls that we don't, endpoints it stopped calling, and `Options` header changes. Opens or updates the issue "BYMA API: endpoints to add, remove or fix", and closes it once the diff is clean |

A red **Live API** badge or an open Discover issue means there's work.

The Integration run asks `IsMarketOpen` first. Checks that need live prices
(traded volume, a GGAL quote, the server clock inside trading hours) only run
while the market is open, so holidays and late runs don't fail. Everything
else runs every time.

## The discovery tool

BYMA's web app (open.bymadata.com.ar) is the reference client. Its
JavaScript bundle lists every endpoint, the `Options` header it needs and the
request body it sends. `scripts/discover` reads the bundle and compares it
with `internal/api/endpoints.go`:

```bash
go run ./scripts/discover                    # diff only
go run ./scripts/discover -probe             # also call each new endpoint live
go run ./scripts/discover -probe -recheck    # also re-call the skipped ones
```

The report has three sections:

- **New endpoints**: with their header, request body and, with `-probe`,
  the live status and field names. Either implement each one or add it to
  `scripts/discover/skipped.txt` with a reason and date.
- **Endpoints the web app no longer calls**: candidates for removal.
- **Options header mismatches**: fix `sectionHeader`.

It exits 1 when any section is non-empty, 0 when there's nothing to do.

## Adding an endpoint

1. **Look at the real response.** `-probe` shows the status and fields. For
   one endpoint, call it with the client's headers:
   ```bash
   curl -s -X POST https://open.bymadata.com.ar/vanoms-be-core/rest/api/bymadata/free/cauciones \
     -H 'Content-Type: application/json' -H 'Options: renta-fija' \
     -H 'Origin: https://open.bymadata.com.ar' -d '{"Content-Type":"application/json"}' | head -c 600
   ```
   If the probe returns 405, the endpoint is a GET with query parameters
   (chart endpoints work this way). If it returns an empty list, try the
   body the bundle shows before giving up.
2. **Add the path** as a `path…` constant and its section to
   `sectionHeader` in `internal/api/endpoints.go`.
3. **Parse it.** Reuse `parseSecurity`, `parseBond`, `parseOption` or
   `parseIndex` when the record has the usual market fields. Use
   `parseAPIResponse`, which accepts both bare lists and `{"data": [...]}`.
4. **Expose it.** Add the method to the `Client` interface in `types.go`.
   Market-wide lists get a cached wrapper in `client.go`; per-symbol calls
   are served by the embedded `*api.Client`.
5. **Record a fixture.** Add the call to `TestRecordFixtures`, then:
   ```bash
   RECORD_WRITE=1 go test -tags record -run TestRecordFixtures ./internal/api/
   git checkout -- internal/api/testdata   # keep existing fixtures; new files stay
   ```
6. **Test it**: a unit test against the fixture, and a live assertion in
   `integration_test.go`.
7. **Document it** in both READMEs.
8. **Release a minor version.** New `Client` methods break custom
   implementations of the interface.

## Removing an endpoint

When Discover lists an endpoint under "no longer calls", or a method starts
failing for good (login, 404):

1. Confirm with a few live calls over a day or two; BYMA also fails
   transiently.
2. Keep the method, mark it `// Deprecated:` with the reason, and make the
   docs say so. Delete it only in a later minor version.
3. Add a "Known limitation" note to the release.

## Editing (a field or header changed)

1. Find what changed:
   ```bash
   INTEGRATION_TEST=true go test -tags integration -run Integration -v ./...
   go test -tags record -run TestRecordFixtures -v ./internal/api/
   go run ./scripts/discover
   ```
   `DRIFT` lines name the file and field; Discover shows header changes.
2. Fix the parser, accepting both the old and the new field where it's
   cheap (see `totalNegotiated`).
3. Re-record the affected fixtures with `RECORD_WRITE=1`. Prices will
   differ, so update the expected values in `internal/api/api_test.go`
   and `client_fixture_test.go`.
4. Add a live assertion in `integration_test.go` so it can't regress.
5. Release a patch version, with upgrade notes at the top of the release.

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
- Use the request body from the bundle, not a guessed one. `cauciones`
  returned an empty list with the securities body and 128 records with the
  web app's.
- Skipped endpoints and why: `scripts/discover/skipped.txt`.
- Some endpoints are GET only: `market-open`, `server-time`, the instrument
  catalog and every `chart/...` series. POST gets HTTP 405.
- Chart endpoints report failures in the body: `{"s":"error","errmsg":...}`
  with HTTP 200. `GetHistory` turns that into an error.
- `server-time` sends Buenos Aires local time without a zone.
- `Change`/`imbalance` is a fraction (0.0123 = +1.23%).
- `market-open` is true only during trading hours, false on holidays: use it
  rather than a calendar to know whether prices should be moving.
- `denominationCcy` can be `"EXT"` besides `"ARS"`; meaning unconfirmed.
