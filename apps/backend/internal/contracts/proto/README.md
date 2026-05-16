# Protobuf Contracts

Protobuf API contract boundary for sync, capture, search, and administration.

Generated Go code should stay at this boundary. Services should receive internal command/query types rather than generated protobuf messages directly.
