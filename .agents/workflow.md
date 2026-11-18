# Agent Workflow

[Agent entry point](../agents.md)

## Explore selectively

Use selective automatic delegation for code exploration, investigation, and review when substantial independent work justifies the coordination cost. Keep small tasks and tightly dependent steps in the primary thread. Start with targeted searches and the relevant guide rather than reading the entire repository.

| Work | Routing |
| --- | --- |
| Code exploration | Delegate bounded file discovery, dependency mapping, entry-point identification, and data-flow tracing to `gpt-6-luna` with high reasoning effort |
| Investigation | Delegate bounded diagnostic questions, evidence gathering, and hypothesis checks to `gpt-6-luna` with high reasoning effort; the primary agent handles complex diagnosis and unresolved findings |
| Code review | Use one independent reviewer for substantial changes or meaningful correctness risk; omit model and reasoning overrides so review inherits the selected session model and reasoning |

If Luna is unavailable, use the selected main model and disclose the fallback. Do not make Luna the universal subagent default. Preserve configured agent limits, main model and reasoning settings, permissions, and more specific project instructions.

## Boundaries and ownership

- Give each subagent a clear, bounded scope and expected output. Require concise findings, supporting evidence, file references where applicable, and explicit uncertainties.
- Keep exploration, investigation, and review subagents read-only. They must not edit files or run commands modifying source, configuration, or external state. Return findings to the primary agent.
- The primary agent owns implementation, verifies findings, applies fixes, and performs final validation. Subagent conclusions are evidence, not automatic acceptance.
- Normally run at most two exploration or investigation subagents concurrently. Avoid duplicate scopes, unnecessary nesting, and work blocked on another agent's results.
- Run one reviewer when substantial or risky changes are ready. Require actionable issues and relevant test gaps, or an explicit report that no actionable issues were found.

## Validation and efficiency

Use the commands and prerequisites in the [development guide](development.md). Run checks appropriate to the change; broaden or repeat passed checks only when new changes, failures, or unresolved concerns justify it. Preserve unrelated changes and report material limitations.

When validating the delegation workflow in fresh sessions, confirm substantial exploration and bounded investigation use Luna with high reasoning, while review inherits the selected model and reasoning, including after model changes. Confirm small tasks remain local, subagents remain read-only, and unresolved findings return to the primary agent.

On representative tasks, compare elapsed time, finding quality, and token use when measurements are available. Delegation may improve completion time for independent work, but speed is not guaranteed and additional token use is accepted. Avoid adding measurement overhead to routine small tasks.
