# Backend Architecture

This backend uses a pragmatic DDD-style layout. The goal is not "pure DDD",
but a structure where a new feature has an obvious place to live.

## Layers

`server/cmd`

The executable entrypoint. Keep it small: load config, open storage, create the
router, and start the process.

`internal/app`

Application composition. This package wires HTTP routes to handlers and creates
shared services. It should not contain business rules.

`internal/<feature>/handlers`

HTTP transport layer. Handlers parse requests, call services, and format
responses. They should not contain SQL or business calculations.

`internal/<feature>/service`

Application/domain behavior. Services validate use cases, coordinate
repositories, and own business rules.

`internal/<feature>/repo`

Repository interfaces owned by the feature. Services depend on these
interfaces, not on SQLite directly.

`internal/<feature>/models`

Feature-owned DTOs/entities used by handlers, services, and repositories.

`internal/storage`

Infrastructure adapter. This package implements repository interfaces using
SQLite. Keep it as one Go package unless you also split the `Storage` type into
separate adapter structs. In Go, files in subdirectories are different packages.

## Adding A Feature

1. Create `internal/<feature>/models`.
2. Define the repository interface in `internal/<feature>/repo`.
3. Implement business logic in `internal/<feature>/service`.
4. Implement HTTP endpoints in `internal/<feature>/handlers`.
5. Add SQLite tables and repository methods in `internal/storage`.
6. Register routes in `internal/app/router.go`.
7. Add focused tests near the layer being changed.

## Rules Of Thumb

- Handlers should not know database details.
- Services should not know Gin details.
- Storage should not import handlers.
- Feature packages should own their repository interfaces.
- Shared helpers belong in `internal/utils` only when they are truly generic.
- Prefer small comments that explain why code exists, not what every line does.
