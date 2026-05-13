# Jiezhang Backend Architecture

## Core Rule

All business requests must follow this path:

`handler -> service -> repository(interface) -> repository/mysql(implementation)`

No layer should skip over the next layer.

## Layer Responsibilities

### `internal/http/handler`

- Parse and validate request params/body.
- Read auth context (`current_user`, `account_book`).
- Call exactly one service workflow per endpoint (or a small orchestrated set).
- Convert service errors to HTTP response.
- Keep transport DTOs in `internal/http/dto`.

Handler must not:

- Write SQL.
- Touch `gorm.DB` directly.
- Build complex business aggregation.
- Encode business policy decisions.

### `internal/service`

- Hold business rules and use-case orchestration.
- Compose repository calls.
- Validate domain-level input.
- Return domain/service DTOs.

Service must not:

- Parse HTTP-only payload concerns.
- Write raw SQL.
- Depend on `gin.Context`.

### `internal/repository`

- Define interfaces, filters, and storage records.
- Keep persistence contracts stable.

### `internal/repository/mysql`

- Implement repository interfaces with GORM/SQL.
- Only persistence and query concerns.

MySQL repository must not:

- Produce HTTP response models.
- Carry request parsing logic.

## Dependency Injection

- All wiring is done in `internal/bootstrap/modules`.
- `internal/bootstrap/app.go` only composes modules.
- Shared utilities (e.g. URL builder, statement row mapper) are built once in modules and injected.

## Shared Mapping Rule

- Cross-service output mapping should be extracted into reusable mapper components.
- Current example: `internal/service/statement/mapper.go`.
- Do not call methods from one service inside another service just for mapping.

## File Placement Rules

- HTTP payload/request structs: `internal/http/dto`.
- Service input/output structs: `internal/service/<domain>` (or dedicated package).
- Repository filter/record structs: `internal/repository`.
- DB row temp structs: `internal/repository/mysql`.

## PR/Change Checklist

- Does handler avoid SQL and business aggregation?
- Does service own business policy?
- Does repository interface describe persistence boundary clearly?
- Is mysql package only implementing repository interfaces?
- Are new structs placed in the correct layer?
- Is module wiring updated, without leaking infra details into handler?
- Did you run `gofmt` and `go test ./...`?
