# Repository Guidelines

## Project Structure & Module Organization
This repository is currently minimal and centered on agent workflow metadata.

- `.agents/` stores skill surfaces used by Codex-compatible tooling.
- `.codex/` stores Codex runtime/project configuration (agent roles, MCP setup, and local behavior).
- Add application code under feature-first folders (for example, `src/<feature>/...`) and keep tests near implementation or in `tests/`.
- Keep workflow assets (skills, prompts, command shims) grouped by purpose, not by file type.

## Build, Test, and Development Commands
Use the following baseline commands from the repo root:

- `rg --files` — fast file inventory.
- `rg "<pattern>"` — fast code/content search.
- `git status --short` — quick change review before commits.
- `codex mcp list` — verify configured MCP servers are available.

When a runtime is added (Node/Python/etc.), define project scripts in a single canonical entry point (`package.json` or `Makefile`) and document them here.

## Coding Style & Naming Conventions
- Prefer ASCII text unless a file already requires Unicode.
- Use small, focused files; avoid broad utility dumping.
- Prefer immutable updates over in-place mutation.
- Naming:
  - Directories: kebab-case (`user-profile/`)
  - Files: kebab-case (`auth-service.ts`, `test-api.sh`)
  - Types/classes: PascalCase
  - Variables/functions: camelCase
- Use repository-standard formatters/linters once introduced (for example, ESLint/Prettier or Ruff/Black).

## Testing Guidelines
- Follow TDD when practical: failing test, minimal fix, refactor.
- Target **80%+ coverage** for active modules.
- Test naming:
  - Unit: `*.unit.test.*`
  - Integration: `*.int.test.*`
  - E2E: `*.e2e.test.*`
- Keep tests deterministic; mock network and time where relevant.

## Commit & Pull Request Guidelines
- Commit format: `<type>: <description>` (for example, `feat: add skill registry validator`).
- Common types: `feat`, `fix`, `refactor`, `docs`, `test`, `chore`, `perf`, `ci`.
- PRs should include:
  - What changed and why
  - Test evidence (commands + results)
  - Any risk/security impact
  - Screenshots/log snippets when UI/CLI behavior changes

## Security & Configuration Tips
- Never commit secrets; use environment variables.
- Validate external inputs at boundaries.
- Prefer least-privilege defaults in `.codex/config.toml` and review MCP/server changes in PRs.
