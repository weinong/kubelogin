# Work Routing

How to decide who handles what.

## Routing Table

| Work Type | Route To | Examples |
|-----------|----------|----------|
| Architecture & design | Morpheus | New auth-mode design, contract changes, scope/sequencing, refactor proposals |
| Implementation | Neo | Write/modify Go code, wire up auth flows, fix bugs, add tests for changed behavior |
| Code review | Trinity | Review diffs, approve/reject, demand missing tests, flag security issues |
| Scope & priorities | Morpheus | What to build next, trade-offs, irreversible decisions |
| Session logging | Scribe | Automatic — never needs routing |
| Backlog monitoring | Ralph | Work queue, keep-alive scans |

## Issue Routing

| Label | Action | Who |
|-------|--------|-----|
| `squad` | Triage: analyze issue, assign `squad:{member}` label | Morpheus |
| `squad:morpheus` | Architectural decisions, scope, design proposals | Morpheus |
| `squad:neo` | Implementation work, bug fixes, code changes | Neo |
| `squad:trinity` | Code-review-specific issues, security concerns, test gaps | Trinity |

### How Issue Assignment Works

1. When a GitHub issue gets the `squad` label, **Morpheus** triages it — analyzing content, assigning the right `squad:{member}` label, and commenting with triage notes.
2. When a `squad:{member}` label is applied, that member picks up the issue in their next session.
3. Members can reassign by removing their label and adding another member's label.
4. The `squad` label is the "inbox" — untriaged issues waiting for triage.

## Rules

1. **Eager by default** — spawn all agents who could usefully start work, including anticipatory downstream work.
2. **Scribe always runs** after substantial work, always as `mode: "background"`. Never blocks.
3. **Quick facts → coordinator answers directly.** Don't spawn an agent for "what Go version are we on?"
4. **When two agents could handle it**, pick the one whose domain is the primary concern.
5. **"Team, ..." → fan-out.** Spawn all relevant agents in parallel as `mode: "background"`.
6. **Anticipate downstream work.** If Neo is implementing a feature, queue Trinity to review as soon as a diff exists.
7. **Reviewer Rejection Protocol** — if Trinity rejects Neo's work, Neo is locked out of the revision. On a 2-coder team this means escalating back to Weinong (or spawning a new specialist). No self-revise.
8. **Issue-labeled work** — when a `squad:{member}` label is applied, route to that member. Morpheus handles all `squad` (base label) triage.
