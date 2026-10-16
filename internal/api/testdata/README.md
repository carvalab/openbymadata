# Test fixtures

Responses recorded from the live BYMA Open Data API on 2026-10-16, trimmed to
the first three records. File names are the endpoint path after
`/vanoms-be-core/rest/api/bymadata/free/` with `/` replaced by `_`.

Two files are hand-written because the live API had nothing to record:

- `index-future.json`: BYMA returned an empty list (no futures listed).
- `bnown_bolsar_balances.json`: the endpoint now redirects to an OAuth login.

To refresh, re-record with the same request bodies the client sends and
update the expected values in `api_test.go`.
