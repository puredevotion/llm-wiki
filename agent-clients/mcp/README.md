# MCP Agent Clients

Agents connect to the backend through MCP Streamable HTTP.

Default endpoint:

```text
http://localhost:8080/mcp
```

Initial tools:

- `kb.search`
- `kb.capture`
- `kb.link`
- `kb.review_queue`
- `kb.summarize_thread`

Only `kb.search` is scaffolded in code. The remaining tools should be implemented after service interfaces and authorization policy are defined.
