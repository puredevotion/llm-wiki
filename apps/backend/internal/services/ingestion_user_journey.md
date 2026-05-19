# User Journey: Conversation Ingestion via MCP

## Journey 1: Summarize and Ingest Conversation
**As an** LLM agent,
**I want to** submit a confirmed summary of a conversation to the knowledge base,
**so that** the knowledge is preserved for future sessions and linked to relevant projects and participants.

### Test Cases:
1.  **Ingest Summary (Happy Path)**:
    *   Input: Valid token, project name, participants list, timestamp, summary body, and key conclusions.
    *   Output: Success response with IDs of created Source and Zettel.
    *   Verification:
        *   A `Source` of kind `conversation` is created with the provided metadata.
        *   A `Zettel` of lifecycle `project` is created and linked to the source.
        *   Actors are created/resolved for the participants.

2.  **Authentication Failure**:
    *   Input: Invalid or missing token.
    *   Output: Error (401 Unauthorized or MCP equivalent).

3.  **Validation Failure**:
    *   Input: Missing required fields (e.g., empty summary or conclusions).
    *   Output: Error (400 Bad Request or MCP equivalent).

4.  **Token Retrieval**:
    *   Input: Request to `kb.get_token`.
    *   Output: The configured agent token (for now, as per simplified requirement).

## Journey 2: Participant Resolution
**As a** system,
**I want to** ensure that participants mentioned in a summary are correctly linked to Actor records,
**so that** I can track collaboration across the knowledge base.

### Test Cases:
1.  **Resolve Existing Actor**:
    *   Input: Participant name that already exists in the `actors` table.
    *   Result: The existing Actor ID is used for linkage.

2.  **Create New Actor**:
    *   Input: Participant name that does not exist.
    *   Result: A new Actor of kind `person` is created.
