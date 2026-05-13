# Codex Collaboration Workflow

## How To Ask

When you start a new task, include one of these:

- `按 layered-architecture skill 执行`
- `按 docs/ARCHITECTURE.md 执行`

This ensures Codex follows the exact project layering constraints.

## Non-Negotiable Constraints

- Keep `handler -> service -> repository(interface) -> repository/mysql`.
- Do not put SQL in handler/service.
- Do not put HTTP parsing in service/repository.
- Keep reusable mappers as shared components, not service-to-service method calls.

## Review Mode

If you ask for review, Codex should prioritize:

- Architecture violations.
- Business regressions.
- Missing validation and edge cases.
- Missing tests.
