# Operational routes

Health and diagnostics. These aren't business endpoints.

## 🟢 GET `/health`

Liveness check that returns a fixed greeting. Both APIs register it. It only proves the Lambda ran; it doesn't check the database or storage. No client calls it.

**Auth:** 🟢 Public.

**Request:** no parameters or body.

**Response `200`:**

```json
{ "greeting": "Server running" }
```

**Errors:** `400` with the plain-text body `Error parsing payload` if JSON marshaling fails, which can't realistically happen.

**Status:** ✅ Implemented.

**Open:** whether to add a separate readiness check later.

**Code:** route [`test_routes.go`](../../infrastructure/legacy/gateway/routes/test_routes.go) · handler [`test/ping/main.go`](../../infrastructure/legacy/lambda/api/test/ping/main.go)

## 🟢 GET `/database/test`

Opens a MySQL connection and runs `SELECT 1 + 1`. **It's dormant:** the route helper exists in `gateway/routes`, but neither API stack calls it, so the route isn't deployed and isn't one of the 18 active routes.

**Auth:** 🟢 It would be public if it were registered.

**Request:** no parameters or body.

**Response `200`:**

```json
{ "success": true, "result": 2 }
```

**Errors:** `500` `{"success": false, "error": "<reason>"}`.

**Status:** 🟨 Dormant: not registered by either API stack.

**Known issues:**

- Its test query is embedded in the handler rather than being an application query.
- It isn't part of the public API design. Equivalent checks belong in protected diagnostics or integration tests.

**Code:** route helper [`database_routes.go`](../../infrastructure/legacy/gateway/routes/database_routes.go) · integration [`test_database.go`](../../infrastructure/legacy/gateway/integrations/test_database.go) · handler [`database/test/main.go`](../../infrastructure/legacy/lambda/api/database/test/main.go)
