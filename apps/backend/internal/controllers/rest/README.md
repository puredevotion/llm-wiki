# REST Controllers

Thin REST/JSON transport adapters.

Responsibilities:

- Decode requests and validate transport shape.
- Map request DTOs to service commands/queries.
- Call services.
- Map domain/service errors to HTTP responses.

No business logic, SQL, graph queries, or external provider SDK calls belong here.
