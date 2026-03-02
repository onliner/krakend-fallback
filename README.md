# Krakend Fallback

Server plugin for **KrakenD** that allows configuring route-based rules to:

- Enforce required response fields (returning an error if missing)
- Inject default values for absent fields
- Extract and propagate backend errors (`error_*`) when required fields are missing

---

## Overview

The plugin works as a **server middleware**.

For configured routes it:

1. Intercepts the aggregated JSON response
2. If response is not `2xx` or not JSON → passes it through untouched
3. If JSON and successful:
   - Injects default fields (if missing)
   - Validates required fields
   - If required field is missing:
     - Looks for `error_*` keys in the response
     - Sorts them
     - Returns the first backend error
     - If no backend error found → returns `500`

---

## Example KrakenD Configuration

```json
{
  "plugin/http-server": {
    "name": ["onliner/krakend-fallback"],
    "onliner/krakend-fallback": {
      "routes": [
        {
          "path": "/products/{product}/positions",
          "required": ["product"],
          "default": {
            "positions": []
          }
        }
      ]
    }
  }
}
```

---

## Route Configuration

### `path`

Path template using `{param}` syntax.

Example:

```
/products/{product}/positions
```

---

### `required`

List of top-level JSON keys that must exist in the final aggregated response.

If any key is missing:

- Plugin searches for `error_*` fields
- If found → returns the first backend error
- If not found → returns `500 Internal Server Error`

---

### `default`

Map of fields to inject if they are missing.

Defaults are only applied for successful `2xx` JSON responses.

Example:

```json
"default": {
  "positions": [],
  "shops": null
}
```

---

## Backend Error Format

Backend errors must follow this structure inside the aggregated response:

```json
{
  "error_1": {
    "http_status_code": 500,
    "http_body": "{\"message\":\"backend failed\"}",
    "http_body_encoding": "application/json"
  }
}
```

### Fields

- `http_status_code` – HTTP status to return
- `http_body` – raw response body
- `http_body_encoding` – Content-Type header (optional, defaults to `text/plain`)

If multiple `error_*` keys exist, the plugin:

1. Sorts them lexicographically
2. Returns the first one

## Notes

- Plugin modifies only successful `2xx` JSON responses
- Original headers are preserved when modifying response
- `Content-Length` is recalculated when body is modified
- Thread-safe and race-tested
