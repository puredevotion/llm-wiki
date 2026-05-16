# Ingestion

Capture and processing pipeline for URLs, PDFs, conversations, pasted text, and files.

Pipeline:

```text
source capture -> raw artifact -> extraction -> chunking -> classification -> zettel proposals -> review queue
```
