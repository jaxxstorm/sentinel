## Context

Sentinel currently has implementation coverage for onboarding, realtime IPNBus observation, event diffing, and sink delivery, but operator-facing documentation is fragmented. The repository needs documentation that is easy to browse and maintain, plus a root README that points readers to the right operational guides quickly.

The requested outcome is a Docsify-renderable `docs/` tree that documents configuration and sink behavior, and a simplified README without emoji-heavy or generated-sounding prose.

## Goals / Non-Goals

**Goals:**
- Add a `docs/` structure that works with Docsify and supports local rendering.
- Provide clear configuration documentation, including sink/route structure and environment interpolation behavior for webhook URLs.
- Document sink delivery behavior, including stdout/debug output, webhook retries, and success/failure logging.
- Provide practical troubleshooting guidance for missing webhook deliveries and idempotency-related suppression.
- Replace root README with a concise, conventional project overview and quick start.

**Non-Goals:**
- No runtime behavioral changes beyond documentation and small example alignment.
- No redesign of CLI command semantics.
- No external docs hosting/deployment automation in this change.

## Decisions

### 1. Use Docsify-native file layout under `docs/`
Decision:
- Create `docs/README.md` as docs landing page and use Docsify conventions (`docs/_sidebar.md`, optional `docs/_coverpage.md` if needed).
- Organize content by operator workflow: install/run, configuration, sinks, troubleshooting.

Rationale:
- Keeps docs simple markdown-first while enabling incremental growth.
- Matches user requirement for Docsify rendering without introducing static-site build complexity.

Alternatives considered:
- MkDocs or Docusaurus: rejected due to heavier setup and additional dependencies for this scope.

### 2. Treat configuration and sink docs as normative operator reference
Decision:
- Add a dedicated configuration reference page with field-level explanations and realistic examples based on `config.example.yaml`.
- Add a sink behavior page covering route matching, stdout/debug output shape, webhook retries, and common failure modes.

Rationale:
- Most support burden comes from ambiguous config/sink behavior.
- Centralizing this information improves reproducibility and troubleshooting.

Alternatives considered:
- Keep all docs in README only: rejected because the README would become too long and hard to navigate.

### 3. Keep README short and index-like
Decision:
- Rewrite root `README.md` in a conventional structure: project summary, quick start, key commands, docs links, development/test entry points.
- Avoid emoji and marketing-style language.

Rationale:
- Repository root should orient contributors quickly, not duplicate full operator docs.

Alternatives considered:
- Long-form README containing all docs: rejected as hard to maintain and less friendly for future expansion.

### 4. Keep implementation aligned with existing runtime semantics
Decision:
- Documentation must reflect current code behavior: default stdout-debug sink, route fallback behavior, env expansion for sink URLs, webhook success/failure logs, and idempotency state effects.

Rationale:
- Documentation drift is a major source of operator confusion.
- This change should reduce support/debug cycles by documenting what Sentinel actually does.

## Risks / Trade-offs

- [Risk: Docs drift from runtime behavior over time] -> Mitigation: anchor examples to `config.example.yaml` and document behavior already enforced by tests.
- [Risk: README becomes too terse for first-time users] -> Mitigation: include direct links to configuration and troubleshooting pages from quick start.
- [Risk: Docsify-specific files may be incomplete for some local setups] -> Mitigation: include local preview instructions and minimal required docsify structure.
- [Risk: Over-documenting internal implementation details] -> Mitigation: focus on operator-observable behavior and commands.

## Migration Plan

1. Add Docsify-compatible `docs/` skeleton and navigation.
2. Add pages for configuration reference, sink behavior, and troubleshooting.
3. Update root README with concise quick start and links to docs pages.
4. Verify examples match current runtime behavior and config field names.
5. Validate formatting and link integrity.

Rollback:
- Revert `docs/` additions and restore previous README in one commit if needed.

## Open Questions

- Should we include a docs page specifically for OpenSpec workflow or keep docs focused strictly on runtime operations?
- Should docs include sample RequestBin/Pipedream workflows, or remain vendor-neutral?
