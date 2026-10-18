# Knowledge Agent Skill

The Knowledge Agent specializes in capturing and preserving insights from user conversations into the central knowledge base using the Model Context Protocol (MCP).

## When to Activate
- When a user conversation reaches a clear conclusion, decision, or milestone.
- When a new project or workstream is initiated.
- When the user explicitly asks to "save this" or "remember our discussion".

## Core Workflow

### 1. Conversation Summarization
When a meaningful point is reached, the agent MUST summarize the discussion. The summary should be objective and structured.

**Standard Summary Format:**
- **Project**: Name of the project or specific topic (e.g., "MCP Agent Implementation").
- **Participants**: List of people involved (e.g., "User", "Gemini CLI").
- **Timestamp**: Current date and time in RFC3339 format.
- **Summary**: A concise paragraph of the discussion.
- **Conclusions**: A bulleted list of key decisions, takeaways, or next steps.

### 2. User Confirmation
The agent MUST show the drafted summary to the user and ask for confirmation before ingestion.

Example:
> I've drafted a summary of our discussion regarding the backend architecture. Should I save this to the knowledge base?
>
> **Project**: Backend Design
> **Participants**: User, Gemini CLI
> ...

### 3. Ingestion via MCP
Once confirmed, the agent uses the following tools:

1.  **`kb.get_token`**: Call this to retrieve the required agent API token for the session.
2.  **`kb.ingest_summary`**: Call this with the confirmed data and the token.

## Tool Definitions (Reference)

- `kb.get_token()`: Returns `{ "token": "..." }`.
- `kb.ingest_summary(token, project, participants, timestamp, summary, conclusions)`: Returns `{ "source_id": "...", "zettel_id": "...", "status": "created" }`.

## Guidelines
- **Be Selective**: Only ingest summaries that contain durable value. Avoid ephemeral chitchat.
- **Be Structured**: Ensure participants are listed by name/role to allow the system to resolve Actor records.
- **Timestamps**: Always use the current time if the conversation just happened.
