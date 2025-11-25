# Designing hooks for token usage tracking
_Exported on 11/25/2025 at 07:33:59 EST from Cursor (2.1.32)_

---

**User**

I am interested in extending ctd to help me keep track of token usage. There are numerous hooks available in cursor that we could probably use to get at this data 



https://cursor.com/docs/agent/hooks # Hooks



Hooks let you observe, control, and extend the agent loop using custom scripts. Hooks are spawned processes that communicate over stdio using JSON in both directions. They run before or after defined stages of the agent loop and can observe, block, or modify behavior.



With hooks, you can:



- Run formatters after edits

- Add analytics for events

- Scan for PII or secrets

- Gate risky operations (e.g., SQL writes)



## Agent and Tab Support



Hooks work with both **Cursor Agent** (Cmd+K/Agent Chat) and **Cursor Tab** (inline completions), but they use different hook events:



**Agent (Cmd+K/Agent Chat)** uses the standard hooks:



- `beforeShellExecution` / `afterShellExecution` - Control shell commands

- `beforeMCPExecution` / `afterMCPExecution` - Control MCP tool usage

- `beforeReadFile` / `afterFileEdit` - Control file access and edits

- `beforeSubmitPrompt` - Validate prompts before submission

- `stop` - Handle agent completion

- `afterAgentResponse` / `afterAgentThought` - Track agent responses



**Tab (inline completions)** uses specialized hooks:



- `beforeTabFileRead` - Control file access for Tab completions

- `afterTabFileEdit` - Post-process Tab edits



These separate hooks allow different policies for autonomous Tab operations versus user-directed Agent operations.



## Quickstart



Create a `hooks.json` file. You can create it at the project level (`<project>/.cursor/hooks.json`) or in your home directory (`~/.cursor/hooks.json`). Project-level hooks apply only to that specific project, while home directory hooks apply globally.



```

{

  "version": 1,

  "hooks": {

    "afterFileEdit": [{ "command": "./hooks/format.sh" }]

  }

}

```



Create your hook script at `~/.cursor/hooks/format.sh`:



```

#!/bin/bash

# Read input, do something, exit 0

cat > /dev/null

exit 0

```



Make it executable:



```

chmod +x ~/.cursor/hooks/format.sh

```



Restart Cursor. Your hook now runs after every file edit.



## Examples



hooks.jsonaudit.shblock-git.shredact-secrets.shformat-tab.shredact-secrets-tab.sh```

{

  "version": 1,

  "hooks": {

    "beforeShellExecution": [

      {

        "command": "./hooks/audit.sh"

      },

      {

        "command": "./hooks/block-git.sh"

      }

    ],

    "beforeMCPExecution": [

      {

        "command": "./hooks/audit.sh"

      }

    ],

    "afterShellExecution": [

      {

        "command": "./hooks/audit.sh"

      }

    ],

    "afterMCPExecution": [

      {

        "command": "./hooks/audit.sh"

      }

    ],

    "beforeReadFile": [

      {

        "command": "./hooks/redact-secrets.sh"

      }

    ],

    "afterFileEdit": [

      {

        "command": "./hooks/audit.sh"

      }

    ],

    "beforeSubmitPrompt": [

      {

        "command": "./hooks/audit.sh"

      }

    ],

    "stop": [

      {

        "command": "./hooks/audit.sh"

      }

    ],

    "beforeTabFileRead": [

      {

        "command": "./hooks/redact-secrets-tab.sh"

      }

    ],

    "afterTabFileEdit": [

      {

        "command": "./hooks/format-tab.sh"

      }

    ]

  }

}

```



## Configuration



Define hooks in a `hooks.json` file. Configuration can exist at multiple levels; higher-priority sources override lower ones:



```

~/.cursor/

├── hooks.json

└── hooks/

    ├── audit.sh

    ├── block-git.sh

    └── redact-secrets.sh

```



- Project (Project-specific):

- `<project>/.cursor/hooks.json`

- Home Directory (User-specific):

- `~/.cursor/hooks.json`

- Global (Enterprise-managed):

- macOS: `/Library/Application Support/Cursor/hooks.json`

- Linux/WSL: `/etc/cursor/hooks.json`

- Windows: `C:\\ProgramData\\Cursor\\hooks.json`



The `hooks` object maps hook names to arrays of hook definitions. Each definition currently supports a `command` property that can be a shell string, an absolute path, or a path relative to the `hooks.json` file.



### Configuration file



```

{

  "version": 1,

  "hooks": {

    "beforeShellExecution": [{ "command": "./script.sh" }],

    "afterShellExecution": [{ "command": "./script.sh" }],

    "afterMCPExecution": [{ "command": "./script.sh" }],

    "afterFileEdit": [{ "command": "./format.sh" }],

    "beforeTabFileRead": [{ "command": "./redact-secrets-tab.sh" }],

    "afterTabFileEdit": [{ "command": "./format-tab.sh" }]

  }

}

```



The Agent hooks (`beforeShellExecution`, `afterShellExecution`, `beforeMCPExecution`, `afterMCPExecution`, `beforeReadFile`, `afterFileEdit`, `beforeSubmitPrompt`, `stop`, `afterAgentResponse`, `afterAgentThought`) apply to Cmd+K and Agent Chat operations. The Tab hooks (`beforeTabFileRead`, `afterTabFileEdit`) apply specifically to inline Tab completions.



## Team Distribution



Hooks can be distributed to team members using MDM tools or Cursor's cloud distribution system.



### MDM Distribution



Distribute hooks across your organization using Mobile Device Management (MDM) tools. Place the `hooks.json` file and hook scripts in the target directories on each machine.



**User home directory** (per-user distribution):



- `~/.cursor/hooks.json`

- `~/.cursor/hooks/` (for hook scripts)



**Global directories** (system-wide distribution):



- macOS: `/Library/Application Support/Cursor/hooks.json`

- Linux/WSL: `/etc/cursor/hooks.json`

- Windows: `C:\\ProgramData\\Cursor\\hooks.json`



Note: MDM-based distribution is fully managed by your organization. Cursor does not deploy or manage files through your MDM solution. Ensure your internal IT or security team handles configuration, deployment, and updates in accordance with your organization's policies.



### Cloud Distribution (Enterprise Only)



Enterprise teams can use Cursor's native cloud distribution to automatically sync hooks to all team members. Configure hooks in the [web dashboard](https://cursor.com/dashboard?tab=team-content&section=hooks). Cursor automatically delivers configured hooks to all client machines when team members log in.



Cloud distribution provides:



- Automatic synchronization to all team members (every thirty minutes)

- Operating system targeting for platform-specific hooks

- Centralized management through the dashboard



Enterprise administrators can create, edit, and manage team hooks from the dashboard without requiring access to individual machines.



## Reference



### Common schema



#### Input (all hooks)



```

{

  "conversation_id": "string",

  "generation_id": "string",

  "hook_event_name": "string",

  "workspace_roots": ["<path>"]

}

```



### Hook events



#### beforeShellExecution / beforeMCPExecution



Called before any shell command or MCP tool is executed. Return a permission decision.



```

// beforeShellExecution input

{

  "command": "<full terminal command>",

  "cwd": "<current working directory>"

}



// beforeMCPExecution input

{

  "tool_name": "<tool name>",

  "tool_input": "<json params>"

}

// Plus either:

{ "url": "<server url>" }

// Or:

{ "command": "<command string>" }



// Output

{

  "permission": "allow" | "deny" | "ask",

  "user_message": "<message shown in client>",

  "agent_message": "<message sent to agent>"

}

```



#### afterShellExecution



Fires after a shell command executes; useful for auditing or collecting metrics from command output.



```

// Input

{

  "command": "<full terminal command>",

  "output": "<full terminal output>"

}

```



#### afterMCPExecution



Fires after an MCP tool executes; includes the tool's input parameters and full JSON result.



```

// Input

{

  "tool_name": "<tool name>",

  "tool_input": "<json params>",

  "result_json": "<tool result json>"

}

```



#### afterFileEdit



Fires after the Agent edits a file; useful for formatters or accounting of agent-written code.



```

// Input

{

  "file_path": "<absolute path>",

  "edits": [{ "old_string": "<search>", "new_string": "<replace>" }]

}

```



#### beforeReadFile



Enable redaction or access control before the Agent reads a file. Includes any prompt attachments for auditing rules inclusion.



```

// Input

{

  "file_path": "<absolute path>",

  "content": "<file contents>",

  "attachments": [

    {

      "type": "rule",

      "file_path": "<absolute path>"

    }

  ]

}



// Output

{

  "permission": "allow" | "deny"

}

```



#### beforeTabFileRead



Called before Tab (inline completions) reads a file. Enable redaction or access control before Tab accesses file contents.



**Key differences from `beforeReadFile`:**



- Only triggered by Tab, not Agent

- Does not include `attachments` field (Tab doesn't use prompt attachments)

- Useful for applying different policies to autonomous Tab operations



```

// Input

{

  "file_path": "<absolute path>",

  "content": "<file contents>"

}



// Output

{

  "permission": "allow" | "deny"

}

```



#### afterTabFileEdit



Called after Tab (inline completions) edits a file. Useful for formatters or auditing of Tab-written code.



**Key differences from `afterFileEdit`:**



- Only triggered by Tab, not Agent

- Includes detailed edit information: `range`, `old_line`, and `new_line` for precise edit tracking

- Useful for fine-grained formatting or analysis of Tab edits



```

// Input

{

  "file_path": "<absolute path>",

  "edits": [

    {

      "old_string": "<search>",

      "new_string": "<replace>",

      "range": {

        "start_line_number": 10,

        "start_column": 5,

        "end_line_number": 10,

        "end_column": 20

      },

      "old_line": "<line before edit>",

      "new_line": "<line after edit>"

    }

  ]

}



// Output

{

  // No output fields currently supported

}

```



#### beforeSubmitPrompt



Called right after user hits send but before backend request. Can prevent submission.



```

// Input

{

  "prompt": "<user prompt text>",

  "attachments": [

    {

      "type": "file" | "rule",

      "filePath": "<absolute path>"

    }

  ]

}



// Output

{

  "continue": true | false

}

```



#### afterAgentResponse



Called after the agent has completed an assistant message.



```

// Input

{

  "text": "<assistant final text>"

}

```



#### stop



Called when the agent loop ends. Can optionally auto-submit a follow-up user message to keep iterating.



```

// Input

{

  "status": "completed" | "aborted" | "error",

  "loop_count": 0

}

```



```

// Output

{

  "followup_message": "<message text>"

}

```



- The optional `followup_message` is a string. When provided and non-empty, Cursor will automatically submit it as the next user message. This enables loop-style flows (e.g., iterate until a goal is met).

- The `loop_count` field indicates how many times the stop hook has already triggered an automatic follow-up for this conversation (starts at 0). To prevent infinite loops, a maximum of 5 auto follow-ups is enforced.



## Troubleshooting



**How to confirm hooks are active**



There is a Hooks tab in Cursor Settings to debug configured and executed hooks, as well as a Hooks output channel to see errors.



**If hooks are not working**



- Restart Cursor to ensure the hooks service is running.

- Ensure hook script paths are relative to `hooks.json` when using relative paths.



Lets think about how we could implement this - lets kick a design around for a bit 

---

**Cursor**

I'll first scan the existing `ctd` architecture to see how it streams and stores data, then we can design how hooks-based token tracking would integrate (data model, new events, storage, visualization). I’ll keep this phase high-level and then we can iterate on details (e.g., exact hook payloads, schemas, and where to surface this in the extension).

### Goals and context

**Goal**: Use Cursor hooks to measure (or approximate) token usage per “generation” and feed that into `ctd`, so you can inspect usage across time/conversations/models in the same way you currently inspect thoughts.

You already have a nice pipeline:

- Cursor → hook script → `ThoughtService.CreateThought` → Postgres → streaming to the extension.

Token tracking should ideally be an **additive** extension of this, not a completely separate system.

---

### High‑level architecture (with hooks + token usage)

Here’s an end‑to‑end shape that keeps `ctd` at the center:

```mermaid
sequenceDiagram
    participant Cursor as Cursor Agent
    participant Hooks as Cursor Hooks Runtime
    participant TokenHook as Token Usage Hook (binary/script)
    participant CTD as ctd (ThoughtService)
    participant DB as Postgres
    participant Ext as Cursor/VSCode Extension

    Cursor->>Hooks: User sends prompt / agent runs
    Hooks->>TokenHook: JSON (beforeSubmitPrompt / afterAgentResponse / ...)
    TokenHook->>TokenHook: Compute token counts (prompt / completion)
    TokenHook->>CTD: CreateThought (with token metadata)
    CTD->>DB: INSERT thoughts row
    CTD-->>Ext: StreamThoughts (with token metadata)
    Ext->>User: Show thoughts + token usage
```

**Key ideas**:

- Reuse the **existing `Thought` pipeline** but enrich it with token fields.
- Use **Cursor hooks** as the only integration point with Cursor.
- Do token counting on the **hook process side**, since that’s what sees raw text.

---

### What you can realistically get from Cursor hooks

Based on the docs you pasted, hooks give you:

- A **common envelope**: `conversation_id`, `generation_id`, `hook_event_name`, `workspace_roots`, …
- Event‑specific payloads, e.g.:
  - `beforeSubmitPrompt`: `{ prompt, attachments[] }`
  - `afterAgentResponse`: `{ text }`
  - `afterMCPExecution`: `{ tool_name, tool_input, result_json }`
  - `afterAgentThought`: (not fully shown here, but likely a text “thought” payload)

They **do not (yet)** expose explicit `prompt_tokens` / `completion_tokens`. So design‑wise:

- **Baseline**: estimate tokens yourself by running a tokenizer library in the hook process (e.g. tiktoken or similar, model‑aware if you know the model).
- **Forward‑compatible**: if Cursor later adds token usage fields, you just stop estimating and copy their numbers.

---

### Data model options in `ctd`

You have a clean `Thought` type and a `thoughts` table today:

- `Thought` (Go): `ConversationID`, `GenerationID`, `Text`, `DurationMs`, `Model`, `HookEventName`, `CursorVersion`, `WorkspaceRoots`, `UserEmail`, `Timestamp`
- DB: one row per `generation_id`, `ON CONFLICT (generation_id) DO NOTHING`

You now want token usage. Two main design options:

#### Option A: Extend `Thought` with token fields (simple, integrated)

Add optional fields:

- **Domain / proto / DB**:
  - `prompt_token_count int`
  - `completion_token_count int`
  - `total_token_count int`
  - (optionally) `prompt_char_count`, `completion_char_count` as cheap fallbacks

Effect:

- **Proto**: add these to `message Thought` as new fields (high numbered, optional semantics).
- **Domain struct**: mirror those fields.
- **DB schema**: new nullable integer columns on `thoughts` (via a new migration).
- **Handlers / repo**: plumb the fields through `mapProtoToThought`, `mapThoughtToProto`, and the insert queries.

Then your token usage “event” is just **another flavor of Thought**:

- `hook_event_name` gives you which hook produced it (`beforeSubmitPrompt`, `afterAgentResponse`, `afterMCPExecution`, `afterAgentThought`, …).
- Tokens show up as additional metadata on that thought.

Pros:

- Uses all existing infrastructure (Connect service, broadcaster, extension plumbing).
- Easy to render in the extension alongside existing thought metadata.
- Queries/statistics can be added later by aggregating from `thoughts`.

Cons / wrinkle:

- Today `generation_id` is **unique**. If you want to log multiple token‑related events for the *same* generation (prompt + completion + tools), you either:
  - **Aggregate** into a single row per generation (see next section), or
  - Loosen uniqueness (e.g. unique on `(generation_id, hook_event_name)` or introduce an `event_id`).

#### Option B: Separate `TokenUsage` entity and service (cleaner separation)

Add a new type:

- `TokenUsage`: `{ conversation_id, generation_id, prompt_tokens, completion_tokens, total_tokens, model, timestamp, user_email, workspace_roots, ... }`
- New table `token_usage` and maybe a new `UsageService` in the proto.

Pros:

- Keeps “thought logs” and “billing/usage metrics” distinct.
- You can evolve usage independently (e.g. add price, cost, context window stats).

Cons:

- More moving pieces: new service, new repo, new extension plumbing.
- Hooks have to decide when to call ThoughtService vs UsageService (or both).

Given where `ctd` is today, I’d lean to **Option A** as the pragmatic first step, and only split out a dedicated `TokenUsage` service later if the semantics grow.

---

### Hook‑side design: how to get good token numbers

We want, ideally, **one `Thought` per generation** that has all token numbers. But some details:

- `beforeSubmitPrompt` sees **prompt** text.
- `afterAgentResponse` sees **completion** text.
- Hooks are **separate processes per event**, so you can’t just keep an in‑memory map.

You have three broad strategies:

#### Strategy 1: Aggregate across events inside the hook (most accurate)

Design:

- Configure a **single hook binary** for several events:

  ```json
  {
    "version": 1,
    "hooks": {
      "beforeSubmitPrompt": [{ "command": "~/.cursor/hooks/ctd-token-usage" }],
      "afterAgentResponse": [{ "command": "~/.cursor/hooks/ctd-token-usage" }],
      "afterAgentThought": [{ "command": "~/.cursor/hooks/ctd-token-usage" }]
    }
  }
  ```

- The `ctd-token-usage` binary:
  - Reads the JSON input from stdin.
  - Uses `hook_event_name` (and the event‑specific payload) to decide what it’s seeing.
  - Uses a **small local state store** keyed by `(conversation_id, generation_id)`:
    - Could be a `~/.cursor/ctd-token-usage/state.sqlite` or a simple JSON file you read/write with locking.
  - Per event:
    - `beforeSubmitPrompt`:
      - Compute `prompt_token_count` (and maybe `prompt_char_count`).
      - Store partial record `{ prompt_tokens, timestamp, model?, workspace_roots, user_email? }` under that `(conversation_id, generation_id)`.
    - `afterAgentResponse`:
      - Compute `completion_token_count` from `text`.
      - Look up the matching partial record.
      - Combine into a **single logical record** `{ prompt_tokens, completion_tokens, total_tokens }`.
      - Call `ThoughtService.CreateThought` **once** for that generation, filling in both token counts and other metadata.
      - Delete the partial state entry.
    - `afterAgentThought` (optional):
      - You may just treat these as “visibility thoughts” and not compute tokens, or compute tokens for internal thoughts separately.

This plays nicely with the existing DB constraint `generation_id UNIQUE` because **only the final “aggregated” event does the insert**.

Tradeoffs/notes:

- Requires some on‑disk state so that events can be paired. Volume is tiny; complexity is moderate but very doable.
- If `afterAgentResponse` never arrives (canceled request), you can either:
  - Expire stale partial entries after N minutes, or
  - Insert a row with only prompt tokens (and zero completion tokens).

#### Strategy 2: Log per‑event rows, relax uniqueness (simpler hooks, more DB rows)

Design:

- Modify schema so that **multiple rows per `generation_id`** are allowed:
  - E.g. drop the `UNIQUE` constraint or replace with `(generation_id, hook_event_name)`.
- Make the hook process very simple:
  - For each event, compute whatever tokens you can see **right now** and call `CreateThought`.
  - Example:
    - `beforeSubmitPrompt`: `prompt_token_count` filled; `completion_token_count` zero.
    - `afterAgentResponse`: `completion_token_count` filled; maybe `prompt_token_count` unknown.
    - `afterAgentThought`: maybe record `thought_token_count` if you care.

Now you see **multiple thought rows per generation**, distinguished by `hook_event_name`.

Pros:

- Hook binary is stateless and trivial.
- No local state / pairing logic, each hook event stands alone.

Cons:

- For high‑level token accounting you now have to **aggregate per generation + per event** in SQL or in analysis tooling.
- The semantic of “one row per generation” is relaxed; you now have “multiple rows per generation/event”.

If you’re comfortable with that, this is a very practical choice.

#### Strategy 3: Only log at one event (e.g. afterAgentResponse) (minimal but lossy)

Design:

- Pick a single event as your “usage moment” – usually `afterAgentResponse`.
- Compute:
  - `completion_token_count` from the assistant text.
  - **Approximate** `prompt_token_count` using a heuristic (e.g. store last user prompt text keyed by `conversation_id` in a local file, then use it here).
- Insert one row per generation with both counts, but accept that prompt tokens are fuzzy.

This avoids multi‑event pairing, at the cost of **less accurate prompt accounting**. Might be good enough if you mostly care about “order‑of‑magnitude” usage.

---

### Token counting inside the hook

Whichever strategy you pick, inside `ctd-token-usage` you’ll roughly:

- Parse JSON input from stdin.
- Identify:
  - `conversation_id`
  - `generation_id`
  - `hook_event_name`
  - Relevant text (`prompt` for `beforeSubmitPrompt`, `text` for `afterAgentResponse`, etc.).
  - `workspace_roots`, `user_email`, `cursor_version`, `model` if present.
- Run a tokenizer:
  - If the hook payload includes `model`, pick the right encoding for that model.
  - If not, either:
    - Default to a “close enough” encoding (e.g. a reasonably similar GPT‑4 tokenizer), or
    - Just track character counts and label them as approximate tokens.
- Build a `Thought` message with:
  - `Text`: either:
    - The raw content (prompt/response/thought), or
    - A **short summary** like `"Token usage: prompt=123, completion=456 (total=579)"` if you don’t want to store the full text here.
  - `HookEventName`: the hook event name.
  - `Model`, `CursorVersion`, `WorkspaceRoots`, `UserEmail`, `Timestamp`.
  - New token fields: `prompt_token_count`, `completion_token_count`, `total_token_count`.

Then call `CreateThought` via HTTP/JSON or gRPC using the generated Connect client (tiny CLI wrapper around the generated client would be ideal).

For hooks that expect an output (e.g. permission checks), you must still **return a valid JSON result** (`permission: "allow"`/etc.) and exit 0; token logging should be best‑effort and never block or fail the agent.

---

### Surfacing token usage in the extension

Once `Thought` carries token metadata, you can enhance the existing extension with no new endpoints:

- **Per‑thought view**:
  - Next to model/duration, show:
    - `Prompt: N tokens`, `Completion: M tokens`, `Total: N+M`.
- **Filtering / quick stats** (follow‑up work):
  - Add a simple client‑side aggregate for the **currently streamed session**:
    - “This session: 12,345 tokens (prompt 4,321 / completion 8,024)”.
  - Later, if you add a query API in `ctd`, you could compute per‑day or per‑workspace aggregates.

The nice part: the webview already gets a `Thought` stream via `StreamThoughts`; you only need to adjust the generated `Thought` TS type and UI components.

---

### Proposal to converge on

If you’re happy with the general direction, I’d suggest we:

1. **Choose data model approach**:
   - My recommendation: **Option A + Strategy 1**:
     - Extend `Thought` with token fields.
     - Keep `generation_id UNIQUE`.
     - Use a small local state store in the hook to aggregate `beforeSubmitPrompt` + `afterAgentResponse` into a single insert.
2. Then we can sketch:
   - The **exact proto changes** to `Thought`.
   - The **DB migration** for `thoughts`.
   - A concrete `hooks.json` snippet.
   - A minimal `ctd-token-usage` hook (probably in Go, consistent with the daemon) that you can drop into `~/.cursor/hooks`.

If you tell me which option/strategy combination resonates most with how you use `ctd`, I can next draft the concrete schema changes and a first‑pass hook binary design.

```plaintext
sequenceDiagram
    participant Cursor as Cursor Agent
    participant Hooks as Cursor Hooks Runtime
    participant TokenHook as Token Usage Hook (binary/script)
    participant CTD as ctd (ThoughtService)
    participant DB as Postgres
    participant Ext as Cursor/VSCode Extension

    Cursor->>Hooks: User sends prompt / agent runs
    Hooks->>TokenHook: JSON (beforeSubmitPrompt / afterAgentResponse / ...)
    TokenHook->>TokenHook: Compute token counts (prompt / completion)
    TokenHook->>CTD: CreateThought (with token metadata)
    CTD->>DB: INSERT thoughts row
    CTD-->>Ext: StreamThoughts (with token metadata)
    Ext->>User: Show thoughts + token usage
```

```json
  {
    "version": 1,
    "hooks": {
      "beforeSubmitPrompt": [{ "command": "~/.cursor/hooks/ctd-token-usage" }],
      "afterAgentResponse": [{ "command": "~/.cursor/hooks/ctd-token-usage" }],
      "afterAgentThought": [{ "command": "~/.cursor/hooks/ctd-token-usage" }]
    }
  }
```

---

**User**

Lets go with Option B for the data model 

Lets architect a new go binary hook which can be dropped in ~/.cursor/hooks that does all the necessary agregation and then sends it to the server over gRPC. Enhance the hook so that it also takes care of the current functionality. It should emit logs like the daemon does. The go hook should be super performant and scalable and non-blocking. 

Develop a design document in the .cursor directory in accordance with the rules of this project

---

**Cursor**

```markdown
# Cursor Token Usage Hook and TokenUsage Service Design

## Overview

This document specifies the design for extending the Cursor Thought Daemon (`ctd`) ecosystem with:

- A **Go-based Cursor hook binary** that can be installed in `~/.cursor/hooks`.
- A new **TokenUsage domain model, service, and storage** path separate from existing `Thought` records (Option B data model).

The hook will:

- Observe Cursor agent and tab lifecycle events via the Cursor hooks JSON protocol.
- Aggregate prompt and completion information across events into structured token usage records.
- Forward both:
  - **Token usage** to a new `TokenUsageService` in `ctd`.
  - **Thought-like content** to the existing `ThoughtService` to preserve current behavior.
- Operate in a **non-blocking, best-effort** manner so that failures in the telemetry path never block the Cursor agent.

The design targets high throughput and low per-event latency while preserving the existing `ctd` architecture and logging conventions.

## Goals and Non-Goals

### Goals

- Record token usage for Cursor interactions with the following properties:
  - Associated with `conversation_id`, `generation_id`, model, user, and workspace roots.
  - Captures at minimum prompt and completion token counts, with room for additional metrics.
- Introduce a **separate TokenUsage model and service** in `ctd` without changing existing `Thought` semantics.
- Provide a **single Go binary hook** that:
  - Can be referenced from `hooks.json` in `~/.cursor` or project-level `.cursor/hooks.json`.
  - Handles both existing "thought logging" and new token accounting.
  - Uses the same logging conventions as the daemon (`pkg/logger` + `log/slog`).
- Ensure the hook is:
  - **Fast**: minimal CPU and memory overhead per event.
  - **Non-blocking**: does not block agent operations; on failure it degrades to no-op with logging.
  - **Scalable**: safe for high-frequency events and long-running sessions.

### Non-Goals

- Building a full-featured historical analytics API or UI for usage data (out of scope; this design enables it later).
- Computing precise currency cost; the design focuses on **token counts and related metrics**, with cost derivation as a possible follow-up.
- Enforcing quota or rate limiting decisions inside the hook; this iteration is observability-only.

## High-Level Architecture

### System Context

```mermaid
graph TB
    subgraph Cursor
        User[User] --> Agent[Cursor Agent / Tab]
        Agent --> HooksRuntime[Cursor Hooks Runtime]
    end

    HooksRuntime -->|hook_event + JSON over stdio| TokenHook[ctd-hook (Go binary)]

    subgraph Local Daemons
        TokenHook --> CTD[ctd (Thought + TokenUsage services)]
    end

    CTD --> PG[(PostgreSQL)]

    subgraph IDE
        CTD --> Ext[VS Code / Cursor Extension]
    end
```

The new `ctd-hook` binary is invoked by the Cursor hooks runtime as a short-lived process. It receives hook input via stdin and writes hook decisions (for the events that expect them) to stdout. It independently calls the `ctd` daemon over Connect-RPC (HTTP/1.1 or h2c) to persist token usage and thought records.

### Sequence Flow per Generation

```mermaid
sequenceDiagram
    participant User
    participant Cursor as Cursor Agent
    participant Hooks as Cursor Hooks Runtime
    participant Hook as ctd-hook (Go)
    participant CTD as ctd Daemon
    participant DB as Postgres

    User->>Cursor: Edit prompt / run agent
    Cursor->>Hooks: beforeSubmitPrompt event
    Hooks->>Hook: JSON (prompt, conversation_id, generation_id, workspace_roots, ...)
    Hook->>Hook: Compute prompt token count
    Hook->>CTD: RecordTokenUsage (prompt metrics)
    CTD->>DB: INSERT token_usage row

    Cursor->>Hooks: afterAgentResponse event
    Hooks->>Hook: JSON (assistant text, ids, model, ...)
    Hook->>Hook: Compute completion token count
    Hook->>CTD: RecordTokenUsage (completion metrics)
    CTD->>DB: INSERT token_usage row (same generation_id, different stage)

    Note over Hook,CTD: Optionally also send ThoughtService.CreateThought<br/>entries for visibility (current functionality)
```

This design intentionally uses **multiple TokenUsage rows per `generation_id`** distinguished by stage (for example, `prompt`, `completion`, `tool`, `system`). Aggregation into per-generation or per-session totals is a downstream concern for querying and analytics.

## Data Model (Option B)

### Protobuf: TokenUsageService

Add a new proto definition alongside `ThoughtService` in `api/proto/v1/token_usage.proto` (file name and package aligned with existing layout):

- **Service**
  - `TokenUsageService`
    - `rpc CreateTokenUsage(CreateTokenUsageRequest) returns (CreateTokenUsageResponse);`

- **Messages**
  - `TokenUsageStage` (enum)
    - `TOKEN_USAGE_STAGE_UNSPECIFIED`
    - `TOKEN_USAGE_STAGE_PROMPT`
    - `TOKEN_USAGE_STAGE_COMPLETION`
    - `TOKEN_USAGE_STAGE_TOOL`
    - `TOKEN_USAGE_STAGE_SYSTEM`
  - `TokenUsage`
    - `string conversation_id`
    - `string generation_id`
    - `string model`
    - `TokenUsageStage stage`
    - `int64 prompt_tokens`
    - `int64 completion_tokens`
    - `int64 total_tokens`
    - `int64 prompt_chars`
    - `int64 completion_chars`
    - `repeated string workspace_roots`
    - `string user_email`
    - `google.protobuf.Timestamp timestamp`
  - `CreateTokenUsageRequest`
    - `TokenUsage usage`
  - `CreateTokenUsageResponse`
    - `string id` (server-generated usage record identifier)

#### Mapping Principles

- `conversation_id` and `generation_id` are copied from the Cursor hook envelope.
- `stage` is derived from the hook event name and payload:
  - `beforeSubmitPrompt` → `TOKEN_USAGE_STAGE_PROMPT`
  - `afterAgentResponse` → `TOKEN_USAGE_STAGE_COMPLETION`
  - `afterMCPExecution` → `TOKEN_USAGE_STAGE_TOOL`
  - Other events can be mapped in future iterations.
- `prompt_tokens` and `completion_tokens` are populated conservatively:
  - For a prompt-stage event, `prompt_tokens` is set, `completion_tokens` is zero.
  - For a completion-stage event, `completion_tokens` is set, `prompt_tokens` is zero.
  - `total_tokens` is always `prompt_tokens + completion_tokens`.

### Domain Model

Add a new domain type in `internal/domain/token_usage.go`:

- `TokenUsage`:
  - `ID uuid.UUID`
  - `ConversationID uuid.UUID`
  - `GenerationID uuid.UUID`
  - `Model string`
  - `Stage TokenUsageStage` (Go enum type)
  - `PromptTokens int64`
  - `CompletionTokens int64`
  - `TotalTokens int64`
  - `PromptChars int64`
  - `CompletionChars int64`
  - `WorkspaceRoots []string`
  - `UserEmail string`
  - `Timestamp time.Time`

Define a `TokenUsageRepository` interface:

- `Store(ctx context.Context, usage *TokenUsage) error`
- `StoreBatch(ctx context.Context, usages []*TokenUsage) error`
- `Close() error`
- `HealthCheck(ctx context.Context) error`

### Database Schema

Create a new `token_usage` table via a migration (for example, `002_token_usage.sql`):

- Columns:
  - `id UUID PRIMARY KEY DEFAULT gen_random_uuid()`
  - `conversation_id UUID NOT NULL`
  - `generation_id UUID NOT NULL`
  - `model VARCHAR(255) NOT NULL`
  - `stage VARCHAR(32) NOT NULL`
  - `prompt_tokens BIGINT NOT NULL DEFAULT 0`
  - `completion_tokens BIGINT NOT NULL DEFAULT 0`
  - `total_tokens BIGINT NOT NULL DEFAULT 0`
  - `prompt_chars BIGINT NOT NULL DEFAULT 0`
  - `completion_chars BIGINT NOT NULL DEFAULT 0`
  - `workspace_roots TEXT[] NOT NULL`
  - `user_email VARCHAR(255) NOT NULL`
  - `timestamp TIMESTAMPTZ NOT NULL`
  - `created_at TIMESTAMPTZ DEFAULT NOW()`

- Indexes:
  - `idx_token_usage_conversation_id (conversation_id)`
  - `idx_token_usage_generation_id (generation_id)`
  - `idx_token_usage_user_email (user_email)`
  - `idx_token_usage_timestamp (timestamp)`
  - Optional composite index for frequent queries, for example `(user_email, timestamp)`.

The schema allows multiple rows per `(conversation_id, generation_id)` pair, differentiated by `stage`. This directly reflects individual hook events without requiring cross-event aggregation in the hook process.

## Server-Side Components

### TokenUsageService Implementation

Add a new service implementation under `internal/server` or a dedicated package (for example, `internal/server/tokenusage.go`):

- Struct:
  - `TokenUsageServer` with:
    - `repo domain.TokenUsageRepository`
    - `log *logger.Logger`

- Methods:
  - `CreateTokenUsage(ctx context.Context, req *connect.Request[apiv1.CreateTokenUsageRequest]) (*connect.Response[apiv1.CreateTokenUsageResponse], error)`
    - Validate fields (UUID parsing, stage, non-negative token counts).
    - Map proto `TokenUsage` to domain `TokenUsage`.
    - Call `TokenUsageRepository.Store`.
    - Return generated `id`.

### Repository Implementation

Implement `TokenUsageRepository` in `internal/repository/postgres_token_usage.go`:

- Use `pgxpool.Pool` consistent with existing `PostgresRepository`.
- Insert query uses parameterized SQL with context-aware logging.
- `StoreBatch` uses `pgx.Batch` similar to the existing `StoreBatch` for `Thought`.
- Health checks piggyback on shared pool connectivity.

### Application Wiring

Extend `internal/app/app.go`:

- Construct a `TokenUsageRepository` alongside the existing `ThoughtRepository` using the same database configuration.
- Register the new `TokenUsageService` handler on the HTTP mux via generated Connect bindings:
  - For example, `apiv1connect.NewTokenUsageServiceHandler`.
- Ensure `cleanupFunc` closes both repositories.

## Hook Binary Design (`ctd-hook`)

### Responsibilities

- Run as a **short-lived process** invoked by Cursor hooks.
- Read a **single JSON document** from stdin, matching the Cursor hook schema:
  - Base fields: `conversation_id`, `generation_id`, `hook_event_name`, `workspace_roots`, etc.
  - Event-specific fields (depending on `hook_event_name`):
    - `beforeSubmitPrompt`: `{ "prompt": "<string>", "attachments": [...] }`
    - `afterAgentResponse`: `{ "text": "<assistant text>" }`
    - `afterMCPExecution`: `{ "tool_name": "...", "tool_input": {...}, "result_json": {...} }`
- Compute:
  - Token counts using a model-aware or approximate tokenizer.
  - Character counts as a fallback metric.
- Emit a **valid hook response** on stdout for events that require output:
  - `beforeSubmitPrompt`: `{ "continue": true }` (or `false` if policy changes are introduced).
  - `beforeShellExecution` / `beforeMCPExecution`: `{ "permission": "allow" }` by default.
- Send telemetry to `ctd`:
  - `TokenUsageService.CreateTokenUsage` for token metrics.
  - `ThoughtService.CreateThought` for thought content where appropriate (mirroring current behavior).

### Binary Location and Invocation

- Binary name: `ctd-hook` (configurable via documentation; the name itself is not enforced by Cursor).
- Installed path for user-level hooks:
  - `~/.cursor/hooks/ctd-hook`
- Example `~/.cursor/hooks.json` snippet:

```json
{
  "version": 1,
  "hooks": {
    "beforeSubmitPrompt": [
      { "command": "~/.cursor/hooks/ctd-hook beforeSubmitPrompt" }
    ],
    "afterAgentResponse": [
      { "command": "~/.cursor/hooks/ctd-hook afterAgentResponse" }
    ],
    "afterMCPExecution": [
      { "command": "~/.cursor/hooks/ctd-hook afterMCPExecution" }
    ],
    "afterAgentThought": [
      { "command": "~/.cursor/hooks/ctd-hook afterAgentThought" }
    ]
  }
}
```

The first argument after the binary name indicates the **expected event type**, which is used for validation and to select the appropriate decoder.

### Internal Structure

```mermaid
graph LR
    stdin[stdin JSON] --> Decoder[HookInput Decoder]
    Decoder --> Router[Event Router]
    Router --> PromptHandler[Prompt Event Handler]
    Router --> CompletionHandler[Completion Event Handler]
    Router --> ToolHandler[Tool Event Handler]

    PromptHandler --> Metrics[Token/Char Metrics]
    CompletionHandler --> Metrics
    ToolHandler --> Metrics

    Metrics --> Mapper[Mapper: Hook -> TokenUsage]
    Mapper --> Client[CTD gRPC Client]
    Client -->|CreateTokenUsage| CTD

    Router --> ThoughtMapper[Mapper: Hook -> Thought]
    ThoughtMapper --> Client
```

Key components:

- `HookInput` type representing the common envelope plus a raw `payload` field.
- Event-specific payload structs (for example, `BeforeSubmitPromptPayload`, `AfterAgentResponsePayload`).
- A `Tokenizer` abstraction to encapsulate model-aware token counting.
- A `CTDClient` facade exposing:
  - `CreateTokenUsage(ctx context.Context, usage TokenUsage) error`
  - `CreateThought(ctx context.Context, thought Thought) error`
- `Logger` using the existing `pkg/logger` package configured for console output.

### Performance and Non-Blocking Behavior

- **Fast startup and teardown**:
  - The binary avoids global state initialization beyond a logger and configuration parsing.
  - Connect client construction and HTTP transport reuse is kept minimal and efficient.
- **Tight timeouts** on outgoing calls to `ctd`:
  - Per-call deadline (for example, 250–500 ms) to avoid blocking Cursor operations.
  - On timeout or error, the hook:
    - Logs the failure with structured context.
    - Still returns a success hook decision (`allow` / `continue: true`) to Cursor.
- **Best-effort semantics**:
  - Telemetry failures never prevent the agent from proceeding.
  - No retries inside the hook binary to keep latency bounded.
- **Resource usage**:
  - Tokenization operates on in-memory strings; for large prompts or responses, the tokenizer operates linearly in input size.
  - No background goroutines survive process exit; the process lifetime is tightly scoped to the single hook invocation.

### Tokenization Strategy

- Introduce a small `Tokenizer` interface:

  - `CountTokens(model string, text string) (int64, error)`

- Provide an implementation that:
  - Uses a configurable default model encoding when the Cursor model name is unknown.
  - Can be swapped in tests to avoid heavy dependencies.
  - Optionally falls back to a simple heuristic (for example, characters divided by a constant) if a precise tokenizer is not available.

The specific tokenizer library is not mandated by this design; it must be:

- Pure Go (or otherwise easy to vendor) to avoid external runtime dependencies.
- Safe for untrusted input strings.

### Logging

- The hook uses the existing `pkg/logger` package so that log formatting and levels match the daemon.
- Logging principles:
  - Info-level events for major lifecycle edges:
    - Hook start and end.
    - Successful calls to `CreateTokenUsage` and `CreateThought`.
  - Debug-level events for detailed diagnostics:
    - Parsed event fields.
    - Token counts and character counts.
  - Error-level events for failures:
    - Network errors reaching `ctd`.
    - Tokenization failures.
- Logs are written to stderr to avoid interfering with the JSON response expected on stdout.

## Maintaining Current Functionality

The existing design expects `ctd` to receive **thought content** that is rendered by the extension. The new hook will:

- Continue to emit `ThoughtService.CreateThought` calls for:
  - Events that represent internal reasoning (`afterAgentThought`).
  - Optionally, summarized views of prompts and completions for additional visibility.
- Preserve existing fields:
  - `conversation_id`, `generation_id`, `text`, `duration_ms`, `model`, `hook_event_name`, `cursor_version`, `workspace_roots`, `user_email`, `timestamp`.

This ensures that the new TokenUsage path is additive and does not regress the current visualization behavior.

## Configuration

The hook binary reads configuration in the following precedence order:

1. Environment variables:
   - `CTD_HOOK_SERVER_URL` (for example, `http://127.0.0.1:36912`)
   - `CTD_HOOK_LOG_LEVEL` (for example, `info`, `debug`)
   - `CTD_HOOK_DEFAULT_MODEL` (fallback tokenizer model name)
2. Optional config file:
   - `$XDG_CONFIG_HOME/ctd/hook.yaml` with keys:
     - `server_url`
     - `log_level`
     - `default_model`
3. Built-in defaults:
   - Server URL: `http://127.0.0.1:36912`
   - Log level: `info`

The hook does not read from the main `ctd` server configuration to avoid tight coupling between the daemon deployment and user-specific Cursor configuration.

## Testing Strategy

### Hook Binary Tests

- **Unit tests**:
  - JSON decoding for each supported hook event.
  - `Tokenizer` behavior for representative inputs.
  - Mapping from hook input to domain `TokenUsage` and `Thought`.
  - Enforcement that hook output JSON remains valid when telemetry fails.
- **Integration tests**:
  - Run `ctd` in test mode using existing helpers.
  - Execute the hook binary with recorded hook JSON fixtures via stdin.
  - Assert that:
    - `TokenUsage` rows appear in the database with expected fields.
    - `Thought` rows are created where expected.
- **Performance tests**:
  - Stress test with synthetic events similar to existing performance tests for `CreateThought`.
  - Ensure per-event latency and memory usage remain within acceptable bounds.

### Server-Side Tests

- New tests for:
  - `TokenUsageServer.CreateTokenUsage` covering:
    - Valid input path.
    - Invalid UUIDs.
    - Negative token counts (rejected).
  - `PostgresTokenUsageRepository`:
    - Happy path stores.
    - Batch inserts.
    - Health checks.

All tests integrate with existing tooling and coverage requirements (80% minimum).

## Security and Privacy Considerations

- The hook processes **potentially sensitive content** (prompts and completions). The design mandates:
  - No logging of full prompt or completion text at info-level or higher.
  - Optional debug logs must be disabled by default and clearly documented.
- The token usage path stores **metadata plus aggregate metrics** rather than full content, which reduces stored sensitive data.
- Communications between the hook and `ctd`:
  - Use localhost by default.
  - May be extended to support TLS in future iterations if `ctd` is deployed remotely.
- The hook treats all errors as non-fatal to the Cursor agent, avoiding denial of service risk due to telemetry issues.

## Future Extensions

Potential follow-up work enabled by this design:

- Add query endpoints to `TokenUsageService`:
  - Per-user, per-day token usage summaries.
  - Per-workspace or per-model aggregates.
- Surface aggregate usage in the VS Code / Cursor extension:
  - “This session” token count.
  - Breakdown by model or workspace.
- Introduce cost calculation using per-model pricing configuration.
- Add more event-specific metrics (for example, tool call token usage, system prompt size).

This design establishes a dedicated, extensible TokenUsage path while preserving and enhancing the existing thought logging behavior.


```

### Design doc created

I have added a detailed design document at:

- **`.cursor/plans/token-usage-hook-design.plan.md`**

It covers:

- A separate **TokenUsage** data model and `TokenUsageService` (Option B).
- A new **Postgres-backed `token_usage` table** and repository interface.
- A **Go hook binary (`ctd-hook`)** architecture that:
  - Handles multiple Cursor hook events (`beforeSubmitPrompt`, `afterAgentResponse`, `afterMCPExecution`, `afterAgentThought`).
  - Computes token and character counts per event.
  - Sends **TokenUsage** records to the new service and **Thought** records to the existing `ThoughtService`.
  - Uses `pkg/logger` so logs match the daemon and remains **best-effort and non-blocking** for Cursor.
- Configuration, testing strategy, performance, and security considerations.
- Mermaid diagrams for system context, sequence, and internal hook structure.

If you want, next steps can be: (1) define the new `token_usage.proto`, (2) add the domain + repository + service skeletons, and (3) scaffold the `ctd-hook` binary with a basic tokenizer stub.

---

**User**

there is a mistake in the System Context Mermaid diagram 

---

**Cursor**

```markdown
# Cursor Token Usage Hook and TokenUsage Service Design

## Overview

This document specifies the design for extending the Cursor Thought Daemon (`ctd`) ecosystem with:

- A **Go-based Cursor hook binary** that can be installed in `~/.cursor/hooks`.
- A new **TokenUsage domain model, service, and storage** path separate from existing `Thought` records (Option B data model).

The hook will:

- Observe Cursor agent and tab lifecycle events via the Cursor hooks JSON protocol.
- Aggregate prompt and completion information across events into structured token usage records.
- Forward both:
  - **Token usage** to a new `TokenUsageService` in `ctd`.
  - **Thought-like content** to the existing `ThoughtService` to preserve current behavior.
- Operate in a **non-blocking, best-effort** manner so that failures in the telemetry path never block the Cursor agent.

The design targets high throughput and low per-event latency while preserving the existing `ctd` architecture and logging conventions.

## Goals and Non-Goals

### Goals

- Record token usage for Cursor interactions with the following properties:
  - Associated with `conversation_id`, `generation_id`, model, user, and workspace roots.
  - Captures at minimum prompt and completion token counts, with room for additional metrics.
- Introduce a **separate TokenUsage model and service** in `ctd` without changing existing `Thought` semantics.
- Provide a **single Go binary hook** that:
  - Can be referenced from `hooks.json` in `~/.cursor` or project-level `.cursor/hooks.json`.
  - Handles both existing "thought logging" and new token accounting.
  - Uses the same logging conventions as the daemon (`pkg/logger` + `log/slog`).
- Ensure the hook is:
  - **Fast**: minimal CPU and memory overhead per event.
  - **Non-blocking**: does not block agent operations; on failure it degrades to no-op with logging.
  - **Scalable**: safe for high-frequency events and long-running sessions.

### Non-Goals

- Building a full-featured historical analytics API or UI for usage data (out of scope; this design enables it later).
- Computing precise currency cost; the design focuses on **token counts and related metrics**, with cost derivation as a possible follow-up.
- Enforcing quota or rate limiting decisions inside the hook; this iteration is observability-only.

## High-Level Architecture

### System Context

```mermaid
graph TB
    subgraph Cursor
        User[User] --> Agent[Cursor Agent / Tab]
        Agent --> HooksRuntime[Cursor Hooks Runtime]
    end

    HooksRuntime -->|hook_event + JSON over stdio| TokenHook[ctd-hook (Go binary)]

    subgraph "Local Daemon"
        CTD[ctd (Thought + TokenUsage services)]
    end

    TokenHook --> CTD
    CTD --> PG[(PostgreSQL)]

    subgraph IDE
        Ext[VS Code / Cursor Extension]
    end

    CTD --> Ext
```

The new `ctd-hook` binary is invoked by the Cursor hooks runtime as a short-lived process. It receives hook input via stdin and writes hook decisions (for the events that expect them) to stdout. It independently calls the `ctd` daemon over Connect-RPC (HTTP/1.1 or h2c) to persist token usage and thought records.

### Sequence Flow per Generation

```mermaid
sequenceDiagram
    participant User
    participant Cursor as Cursor Agent
    participant Hooks as Cursor Hooks Runtime
    participant Hook as ctd-hook (Go)
    participant CTD as ctd Daemon
    participant DB as Postgres

    User->>Cursor: Edit prompt / run agent
    Cursor->>Hooks: beforeSubmitPrompt event
    Hooks->>Hook: JSON (prompt, conversation_id, generation_id, workspace_roots, ...)
    Hook->>Hook: Compute prompt token count
    Hook->>CTD: RecordTokenUsage (prompt metrics)
    CTD->>DB: INSERT token_usage row

    Cursor->>Hooks: afterAgentResponse event
    Hooks->>Hook: JSON (assistant text, ids, model, ...)
    Hook->>Hook: Compute completion token count
    Hook->>CTD: RecordTokenUsage (completion metrics)
    CTD->>DB: INSERT token_usage row (same generation_id, different stage)

    Note over Hook,CTD: Optionally also send ThoughtService.CreateThought<br/>entries for visibility (current functionality)
```

This design intentionally uses **multiple TokenUsage rows per `generation_id`** distinguished by stage (for example, `prompt`, `completion`, `tool`, `system`). Aggregation into per-generation or per-session totals is a downstream concern for querying and analytics.

## Data Model (Option B)

### Protobuf: TokenUsageService

Add a new proto definition alongside `ThoughtService` in `api/proto/v1/token_usage.proto` (file name and package aligned with existing layout):

- **Service**
  - `TokenUsageService`
    - `rpc CreateTokenUsage(CreateTokenUsageRequest) returns (CreateTokenUsageResponse);`

- **Messages**
  - `TokenUsageStage` (enum)
    - `TOKEN_USAGE_STAGE_UNSPECIFIED`
    - `TOKEN_USAGE_STAGE_PROMPT`
    - `TOKEN_USAGE_STAGE_COMPLETION`
    - `TOKEN_USAGE_STAGE_TOOL`
    - `TOKEN_USAGE_STAGE_SYSTEM`
  - `TokenUsage`
    - `string conversation_id`
    - `string generation_id`
    - `string model`
    - `TokenUsageStage stage`
    - `int64 prompt_tokens`
    - `int64 completion_tokens`
    - `int64 total_tokens`
    - `int64 prompt_chars`
    - `int64 completion_chars`
    - `repeated string workspace_roots`
    - `string user_email`
    - `google.protobuf.Timestamp timestamp`
  - `CreateTokenUsageRequest`
    - `TokenUsage usage`
  - `CreateTokenUsageResponse`
    - `string id` (server-generated usage record identifier)

#### Mapping Principles

- `conversation_id` and `generation_id` are copied from the Cursor hook envelope.
- `stage` is derived from the hook event name and payload:
  - `beforeSubmitPrompt` → `TOKEN_USAGE_STAGE_PROMPT`
  - `afterAgentResponse` → `TOKEN_USAGE_STAGE_COMPLETION`
  - `afterMCPExecution` → `TOKEN_USAGE_STAGE_TOOL`
  - Other events can be mapped in future iterations.
- `prompt_tokens` and `completion_tokens` are populated conservatively:
  - For a prompt-stage event, `prompt_tokens` is set, `completion_tokens` is zero.
  - For a completion-stage event, `completion_tokens` is set, `prompt_tokens` is zero.
  - `total_tokens` is always `prompt_tokens + completion_tokens`.

### Domain Model

Add a new domain type in `internal/domain/token_usage.go`:

- `TokenUsage`:
  - `ID uuid.UUID`
  - `ConversationID uuid.UUID`
  - `GenerationID uuid.UUID`
  - `Model string`
  - `Stage TokenUsageStage` (Go enum type)
  - `PromptTokens int64`
  - `CompletionTokens int64`
  - `TotalTokens int64`
  - `PromptChars int64`
  - `CompletionChars int64`
  - `WorkspaceRoots []string`
  - `UserEmail string`
  - `Timestamp time.Time`

Define a `TokenUsageRepository` interface:

- `Store(ctx context.Context, usage *TokenUsage) error`
- `StoreBatch(ctx context.Context, usages []*TokenUsage) error`
- `Close() error`
- `HealthCheck(ctx context.Context) error`

### Database Schema

Create a new `token_usage` table via a migration (for example, `002_token_usage.sql`):

- Columns:
  - `id UUID PRIMARY KEY DEFAULT gen_random_uuid()`
  - `conversation_id UUID NOT NULL`
  - `generation_id UUID NOT NULL`
  - `model VARCHAR(255) NOT NULL`
  - `stage VARCHAR(32) NOT NULL`
  - `prompt_tokens BIGINT NOT NULL DEFAULT 0`
  - `completion_tokens BIGINT NOT NULL DEFAULT 0`
  - `total_tokens BIGINT NOT NULL DEFAULT 0`
  - `prompt_chars BIGINT NOT NULL DEFAULT 0`
  - `completion_chars BIGINT NOT NULL DEFAULT 0`
  - `workspace_roots TEXT[] NOT NULL`
  - `user_email VARCHAR(255) NOT NULL`
  - `timestamp TIMESTAMPTZ NOT NULL`
  - `created_at TIMESTAMPTZ DEFAULT NOW()`

- Indexes:
  - `idx_token_usage_conversation_id (conversation_id)`
  - `idx_token_usage_generation_id (generation_id)`
  - `idx_token_usage_user_email (user_email)`
  - `idx_token_usage_timestamp (timestamp)`
  - Optional composite index for frequent queries, for example `(user_email, timestamp)`.

The schema allows multiple rows per `(conversation_id, generation_id)` pair, differentiated by `stage`. This directly reflects individual hook events without requiring cross-event aggregation in the hook process.

## Server-Side Components

### TokenUsageService Implementation

Add a new service implementation under `internal/server` or a dedicated package (for example, `internal/server/tokenusage.go`):

- Struct:
  - `TokenUsageServer` with:
    - `repo domain.TokenUsageRepository`
    - `log *logger.Logger`

- Methods:
  - `CreateTokenUsage(ctx context.Context, req *connect.Request[apiv1.CreateTokenUsageRequest]) (*connect.Response[apiv1.CreateTokenUsageResponse], error)`
    - Validate fields (UUID parsing, stage, non-negative token counts).
    - Map proto `TokenUsage` to domain `TokenUsage`.
    - Call `TokenUsageRepository.Store`.
    - Return generated `id`.

### Repository Implementation

Implement `TokenUsageRepository` in `internal/repository/postgres_token_usage.go`:

- Use `pgxpool.Pool` consistent with existing `PostgresRepository`.
- Insert query uses parameterized SQL with context-aware logging.
- `StoreBatch` uses `pgx.Batch` similar to the existing `StoreBatch` for `Thought`.
- Health checks piggyback on shared pool connectivity.

### Application Wiring

Extend `internal/app/app.go`:

- Construct a `TokenUsageRepository` alongside the existing `ThoughtRepository` using the same database configuration.
- Register the new `TokenUsageService` handler on the HTTP mux via generated Connect bindings:
  - For example, `apiv1connect.NewTokenUsageServiceHandler`.
- Ensure `cleanupFunc` closes both repositories.

## Hook Binary Design (`ctd-hook`)

### Responsibilities

- Run as a **short-lived process** invoked by Cursor hooks.
- Read a **single JSON document** from stdin, matching the Cursor hook schema:
  - Base fields: `conversation_id`, `generation_id`, `hook_event_name`, `workspace_roots`, etc.
  - Event-specific fields (depending on `hook_event_name`):
    - `beforeSubmitPrompt`: `{ "prompt": "<string>", "attachments": [...] }`
    - `afterAgentResponse`: `{ "text": "<assistant text>" }`
    - `afterMCPExecution`: `{ "tool_name": "...", "tool_input": {...}, "result_json": {...} }`
- Compute:
  - Token counts using a model-aware or approximate tokenizer.
  - Character counts as a fallback metric.
- Emit a **valid hook response** on stdout for events that require output:
  - `beforeSubmitPrompt`: `{ "continue": true }` (or `false` if policy changes are introduced).
  - `beforeShellExecution` / `beforeMCPExecution`: `{ "permission": "allow" }` by default.
- Send telemetry to `ctd`:
  - `TokenUsageService.CreateTokenUsage` for token metrics.
  - `ThoughtService.CreateThought` for thought content where appropriate (mirroring current behavior).

### Binary Location and Invocation

- Binary name: `ctd-hook` (configurable via documentation; the name itself is not enforced by Cursor).
- Installed path for user-level hooks:
  - `~/.cursor/hooks/ctd-hook`
- Example `~/.cursor/hooks.json` snippet:

```json
{
  "version": 1,
  "hooks": {
    "beforeSubmitPrompt": [
      { "command": "~/.cursor/hooks/ctd-hook beforeSubmitPrompt" }
    ],
    "afterAgentResponse": [
      { "command": "~/.cursor/hooks/ctd-hook afterAgentResponse" }
    ],
    "afterMCPExecution": [
      { "command": "~/.cursor/hooks/ctd-hook afterMCPExecution" }
    ],
    "afterAgentThought": [
      { "command": "~/.cursor/hooks/ctd-hook afterAgentThought" }
    ]
  }
}
```

The first argument after the binary name indicates the **expected event type**, which is used for validation and to select the appropriate decoder.

### Internal Structure

```mermaid
graph LR
    stdin[stdin JSON] --> Decoder[HookInput Decoder]
    Decoder --> Router[Event Router]
    Router --> PromptHandler[Prompt Event Handler]
    Router --> CompletionHandler[Completion Event Handler]
    Router --> ToolHandler[Tool Event Handler]

    PromptHandler --> Metrics[Token/Char Metrics]
    CompletionHandler --> Metrics
    ToolHandler --> Metrics

    Metrics --> Mapper[Mapper: Hook -> TokenUsage]
    Mapper --> Client[CTD gRPC Client]
    Client -->|CreateTokenUsage| CTD

    Router --> ThoughtMapper[Mapper: Hook -> Thought]
    ThoughtMapper --> Client
```

Key components:

- `HookInput` type representing the common envelope plus a raw `payload` field.
- Event-specific payload structs (for example, `BeforeSubmitPromptPayload`, `AfterAgentResponsePayload`).
- A `Tokenizer` abstraction to encapsulate model-aware token counting.
- A `CTDClient` facade exposing:
  - `CreateTokenUsage(ctx context.Context, usage TokenUsage) error`
  - `CreateThought(ctx context.Context, thought Thought) error`
- `Logger` using the existing `pkg/logger` package configured for console output.

### Performance and Non-Blocking Behavior

- **Fast startup and teardown**:
  - The binary avoids global state initialization beyond a logger and configuration parsing.
  - Connect client construction and HTTP transport reuse is kept minimal and efficient.
- **Tight timeouts** on outgoing calls to `ctd`:
  - Per-call deadline (for example, 250–500 ms) to avoid blocking Cursor operations.
  - On timeout or error, the hook:
    - Logs the failure with structured context.
    - Still returns a success hook decision (`allow` / `continue: true`) to Cursor.
- **Best-effort semantics**:
  - Telemetry failures never prevent the agent from proceeding.
  - No retries inside the hook binary to keep latency bounded.
- **Resource usage**:
  - Tokenization operates on in-memory strings; for large prompts or responses, the tokenizer operates linearly in input size.
  - No background goroutines survive process exit; the process lifetime is tightly scoped to the single hook invocation.

### Tokenization Strategy

- Introduce a small `Tokenizer` interface:

  - `CountTokens(model string, text string) (int64, error)`

- Provide an implementation that:
  - Uses a configurable default model encoding when the Cursor model name is unknown.
  - Can be swapped in tests to avoid heavy dependencies.
  - Optionally falls back to a simple heuristic (for example, characters divided by a constant) if a precise tokenizer is not available.

The specific tokenizer library is not mandated by this design; it must be:

- Pure Go (or otherwise easy to vendor) to avoid external runtime dependencies.
- Safe for untrusted input strings.

### Logging

- The hook uses the existing `pkg/logger` package so that log formatting and levels match the daemon.
- Logging principles:
  - Info-level events for major lifecycle edges:
    - Hook start and end.
    - Successful calls to `CreateTokenUsage` and `CreateThought`.
  - Debug-level events for detailed diagnostics:
    - Parsed event fields.
    - Token counts and character counts.
  - Error-level events for failures:
    - Network errors reaching `ctd`.
    - Tokenization failures.
- Logs are written to stderr to avoid interfering with the JSON response expected on stdout.

## Maintaining Current Functionality

The existing design expects `ctd` to receive **thought content** that is rendered by the extension. The new hook will:

- Continue to emit `ThoughtService.CreateThought` calls for:
  - Events that represent internal reasoning (`afterAgentThought`).
  - Optionally, summarized views of prompts and completions for additional visibility.
- Preserve existing fields:
  - `conversation_id`, `generation_id`, `text`, `duration_ms`, `model`, `hook_event_name`, `cursor_version`, `workspace_roots`, `user_email`, `timestamp`.

This ensures that the new TokenUsage path is additive and does not regress the current visualization behavior.

## Configuration

The hook binary reads configuration in the following precedence order:

1. Environment variables:
   - `CTD_HOOK_SERVER_URL` (for example, `http://127.0.0.1:36912`)
   - `CTD_HOOK_LOG_LEVEL` (for example, `info`, `debug`)
   - `CTD_HOOK_DEFAULT_MODEL` (fallback tokenizer model name)
2. Optional config file:
   - `$XDG_CONFIG_HOME/ctd/hook.yaml` with keys:
     - `server_url`
     - `log_level`
     - `default_model`
3. Built-in defaults:
   - Server URL: `http://127.0.0.1:36912`
   - Log level: `info`

The hook does not read from the main `ctd` server configuration to avoid tight coupling between the daemon deployment and user-specific Cursor configuration.

## Testing Strategy

### Hook Binary Tests

- **Unit tests**:
  - JSON decoding for each supported hook event.
  - `Tokenizer` behavior for representative inputs.
  - Mapping from hook input to domain `TokenUsage` and `Thought`.
  - Enforcement that hook output JSON remains valid when telemetry fails.
- **Integration tests**:
  - Run `ctd` in test mode using existing helpers.
  - Execute the hook binary with recorded hook JSON fixtures via stdin.
  - Assert that:
    - `TokenUsage` rows appear in the database with expected fields.
    - `Thought` rows are created where expected.
- **Performance tests**:
  - Stress test with synthetic events similar to existing performance tests for `CreateThought`.
  - Ensure per-event latency and memory usage remain within acceptable bounds.

### Server-Side Tests

- New tests for:
  - `TokenUsageServer.CreateTokenUsage` covering:
    - Valid input path.
    - Invalid UUIDs.
    - Negative token counts (rejected).
  - `PostgresTokenUsageRepository`:
    - Happy path stores.
    - Batch inserts.
    - Health checks.

All tests integrate with existing tooling and coverage requirements (80% minimum).

## Security and Privacy Considerations

- The hook processes **potentially sensitive content** (prompts and completions). The design mandates:
  - No logging of full prompt or completion text at info-level or higher.
  - Optional debug logs must be disabled by default and clearly documented.
- The token usage path stores **metadata plus aggregate metrics** rather than full content, which reduces stored sensitive data.
- Communications between the hook and `ctd`:
  - Use localhost by default.
  - May be extended to support TLS in future iterations if `ctd` is deployed remotely.
- The hook treats all errors as non-fatal to the Cursor agent, avoiding denial of service risk due to telemetry issues.

## Future Extensions

Potential follow-up work enabled by this design:

- Add query endpoints to `TokenUsageService`:
  - Per-user, per-day token usage summaries.
  - Per-workspace or per-model aggregates.
- Surface aggregate usage in the VS Code / Cursor extension:
  - “This session” token count.
  - Breakdown by model or workspace.
- Introduce cost calculation using per-model pricing configuration.
- Add more event-specific metrics (for example, tool call token usage, system prompt size).

This design establishes a dedicated, extensible TokenUsage path while preserving and enhancing the existing thought logging behavior.



```

You are right: the diagram previously showed `ctd-hook` inside the “Local Daemons” subgraph even though it is a short‑lived hook process, not a daemon.

I have updated the **System Context** mermaid diagram so that:

- `ctd-hook` now sits **outside** the daemon subgraph.
- A new subgraph **"Local Daemon"** contains only `ctd (Thought + TokenUsage services)`.
- Edges are:
  - `HooksRuntime --> TokenHook`
  - `TokenHook --> CTD`
  - `CTD --> PG`
  - `CTD --> Ext` with `Ext` inside the `IDE` subgraph.

This should now match the intended architecture and terminology.

---

**User**

Parse error on line 7: ... TokenHook[ctd-hook (Go binary)] su -----------------------^ Expecting 'SQE', 'DOUBLECIRCLEEND', 'PE', '-)', 'STADIUMEND', 'SUBROUTINEEND', 'PIPE', 'CYLINDEREND', 'DIAMOND_STOP', 'TAGEND', 'TRAPEND', 'INVTRAPEND', 'UNICODE_TEXT', 'TEXT', 'TAGSTART', got 'PS'





---

**Cursor**

```markdown
# Cursor Token Usage Hook and TokenUsage Service Design

## Overview

This document specifies the design for extending the Cursor Thought Daemon (`ctd`) ecosystem with:

- A **Go-based Cursor hook binary** that can be installed in `~/.cursor/hooks`.
- A new **TokenUsage domain model, service, and storage** path separate from existing `Thought` records (Option B data model).

The hook will:

- Observe Cursor agent and tab lifecycle events via the Cursor hooks JSON protocol.
- Aggregate prompt and completion information across events into structured token usage records.
- Forward both:
  - **Token usage** to a new `TokenUsageService` in `ctd`.
  - **Thought-like content** to the existing `ThoughtService` to preserve current behavior.
- Operate in a **non-blocking, best-effort** manner so that failures in the telemetry path never block the Cursor agent.

The design targets high throughput and low per-event latency while preserving the existing `ctd` architecture and logging conventions.

## Goals and Non-Goals

### Goals

- Record token usage for Cursor interactions with the following properties:
  - Associated with `conversation_id`, `generation_id`, model, user, and workspace roots.
  - Captures at minimum prompt and completion token counts, with room for additional metrics.
- Introduce a **separate TokenUsage model and service** in `ctd` without changing existing `Thought` semantics.
- Provide a **single Go binary hook** that:
  - Can be referenced from `hooks.json` in `~/.cursor` or project-level `.cursor/hooks.json`.
  - Handles both existing "thought logging" and new token accounting.
  - Uses the same logging conventions as the daemon (`pkg/logger` + `log/slog`).
- Ensure the hook is:
  - **Fast**: minimal CPU and memory overhead per event.
  - **Non-blocking**: does not block agent operations; on failure it degrades to no-op with logging.
  - **Scalable**: safe for high-frequency events and long-running sessions.

### Non-Goals

- Building a full-featured historical analytics API or UI for usage data (out of scope; this design enables it later).
- Computing precise currency cost; the design focuses on **token counts and related metrics**, with cost derivation as a possible follow-up.
- Enforcing quota or rate limiting decisions inside the hook; this iteration is observability-only.

## High-Level Architecture

### System Context

```mermaid
graph TB
    subgraph Cursor
        User[User] --> Agent[Cursor Agent / Tab]
        Agent --> HooksRuntime[Cursor Hooks Runtime]
    end

    HooksRuntime -->|hook_event + JSON over stdio| TokenHook
    TokenHook[ctd-hook Go binary]

    subgraph "Local Daemon"
        CTD[ctd (Thought + TokenUsage services)]
    end

    TokenHook --> CTD
    CTD --> PG[(PostgreSQL)]

    subgraph IDE
        Ext[VS Code / Cursor Extension]
    end

    CTD --> Ext
```

The new `ctd-hook` binary is invoked by the Cursor hooks runtime as a short-lived process. It receives hook input via stdin and writes hook decisions (for the events that expect them) to stdout. It independently calls the `ctd` daemon over Connect-RPC (HTTP/1.1 or h2c) to persist token usage and thought records.

### Sequence Flow per Generation

```mermaid
sequenceDiagram
    participant User
    participant Cursor as Cursor Agent
    participant Hooks as Cursor Hooks Runtime
    participant Hook as ctd-hook (Go)
    participant CTD as ctd Daemon
    participant DB as Postgres

    User->>Cursor: Edit prompt / run agent
    Cursor->>Hooks: beforeSubmitPrompt event
    Hooks->>Hook: JSON (prompt, conversation_id, generation_id, workspace_roots, ...)
    Hook->>Hook: Compute prompt token count
    Hook->>CTD: RecordTokenUsage (prompt metrics)
    CTD->>DB: INSERT token_usage row

    Cursor->>Hooks: afterAgentResponse event
    Hooks->>Hook: JSON (assistant text, ids, model, ...)
    Hook->>Hook: Compute completion token count
    Hook->>CTD: RecordTokenUsage (completion metrics)
    CTD->>DB: INSERT token_usage row (same generation_id, different stage)

    Note over Hook,CTD: Optionally also send ThoughtService.CreateThought<br/>entries for visibility (current functionality)
```

This design intentionally uses **multiple TokenUsage rows per `generation_id`** distinguished by stage (for example, `prompt`, `completion`, `tool`, `system`). Aggregation into per-generation or per-session totals is a downstream concern for querying and analytics.

## Data Model (Option B)

### Protobuf: TokenUsageService

Add a new proto definition alongside `ThoughtService` in `api/proto/v1/token_usage.proto` (file name and package aligned with existing layout):

- **Service**
  - `TokenUsageService`
    - `rpc CreateTokenUsage(CreateTokenUsageRequest) returns (CreateTokenUsageResponse);`

- **Messages**
  - `TokenUsageStage` (enum)
    - `TOKEN_USAGE_STAGE_UNSPECIFIED`
    - `TOKEN_USAGE_STAGE_PROMPT`
    - `TOKEN_USAGE_STAGE_COMPLETION`
    - `TOKEN_USAGE_STAGE_TOOL`
    - `TOKEN_USAGE_STAGE_SYSTEM`
  - `TokenUsage`
    - `string conversation_id`
    - `string generation_id`
    - `string model`
    - `TokenUsageStage stage`
    - `int64 prompt_tokens`
    - `int64 completion_tokens`
    - `int64 total_tokens`
    - `int64 prompt_chars`
    - `int64 completion_chars`
    - `repeated string workspace_roots`
    - `string user_email`
    - `google.protobuf.Timestamp timestamp`
  - `CreateTokenUsageRequest`
    - `TokenUsage usage`
  - `CreateTokenUsageResponse`
    - `string id` (server-generated usage record identifier)

#### Mapping Principles

- `conversation_id` and `generation_id` are copied from the Cursor hook envelope.
- `stage` is derived from the hook event name and payload:
  - `beforeSubmitPrompt` → `TOKEN_USAGE_STAGE_PROMPT`
  - `afterAgentResponse` → `TOKEN_USAGE_STAGE_COMPLETION`
  - `afterMCPExecution` → `TOKEN_USAGE_STAGE_TOOL`
  - Other events can be mapped in future iterations.
- `prompt_tokens` and `completion_tokens` are populated conservatively:
  - For a prompt-stage event, `prompt_tokens` is set, `completion_tokens` is zero.
  - For a completion-stage event, `completion_tokens` is set, `prompt_tokens` is zero.
  - `total_tokens` is always `prompt_tokens + completion_tokens`.

### Domain Model

Add a new domain type in `internal/domain/token_usage.go`:

- `TokenUsage`:
  - `ID uuid.UUID`
  - `ConversationID uuid.UUID`
  - `GenerationID uuid.UUID`
  - `Model string`
  - `Stage TokenUsageStage` (Go enum type)
  - `PromptTokens int64`
  - `CompletionTokens int64`
  - `TotalTokens int64`
  - `PromptChars int64`
  - `CompletionChars int64`
  - `WorkspaceRoots []string`
  - `UserEmail string`
  - `Timestamp time.Time`

Define a `TokenUsageRepository` interface:

- `Store(ctx context.Context, usage *TokenUsage) error`
- `StoreBatch(ctx context.Context, usages []*TokenUsage) error`
- `Close() error`
- `HealthCheck(ctx context.Context) error`

### Database Schema

Create a new `token_usage` table via a migration (for example, `002_token_usage.sql`):

- Columns:
  - `id UUID PRIMARY KEY DEFAULT gen_random_uuid()`
  - `conversation_id UUID NOT NULL`
  - `generation_id UUID NOT NULL`
  - `model VARCHAR(255) NOT NULL`
  - `stage VARCHAR(32) NOT NULL`
  - `prompt_tokens BIGINT NOT NULL DEFAULT 0`
  - `completion_tokens BIGINT NOT NULL DEFAULT 0`
  - `total_tokens BIGINT NOT NULL DEFAULT 0`
  - `prompt_chars BIGINT NOT NULL DEFAULT 0`
  - `completion_chars BIGINT NOT NULL DEFAULT 0`
  - `workspace_roots TEXT[] NOT NULL`
  - `user_email VARCHAR(255) NOT NULL`
  - `timestamp TIMESTAMPTZ NOT NULL`
  - `created_at TIMESTAMPTZ DEFAULT NOW()`

- Indexes:
  - `idx_token_usage_conversation_id (conversation_id)`
  - `idx_token_usage_generation_id (generation_id)`
  - `idx_token_usage_user_email (user_email)`
  - `idx_token_usage_timestamp (timestamp)`
  - Optional composite index for frequent queries, for example `(user_email, timestamp)`.

The schema allows multiple rows per `(conversation_id, generation_id)` pair, differentiated by `stage`. This directly reflects individual hook events without requiring cross-event aggregation in the hook process.

## Server-Side Components

### TokenUsageService Implementation

Add a new service implementation under `internal/server` or a dedicated package (for example, `internal/server/tokenusage.go`):

- Struct:
  - `TokenUsageServer` with:
    - `repo domain.TokenUsageRepository`
    - `log *logger.Logger`

- Methods:
  - `CreateTokenUsage(ctx context.Context, req *connect.Request[apiv1.CreateTokenUsageRequest]) (*connect.Response[apiv1.CreateTokenUsageResponse], error)`
    - Validate fields (UUID parsing, stage, non-negative token counts).
    - Map proto `TokenUsage` to domain `TokenUsage`.
    - Call `TokenUsageRepository.Store`.
    - Return generated `id`.

### Repository Implementation

Implement `TokenUsageRepository` in `internal/repository/postgres_token_usage.go`:

- Use `pgxpool.Pool` consistent with existing `PostgresRepository`.
- Insert query uses parameterized SQL with context-aware logging.
- `StoreBatch` uses `pgx.Batch` similar to the existing `StoreBatch` for `Thought`.
- Health checks piggyback on shared pool connectivity.

### Application Wiring

Extend `internal/app/app.go`:

- Construct a `TokenUsageRepository` alongside the existing `ThoughtRepository` using the same database configuration.
- Register the new `TokenUsageService` handler on the HTTP mux via generated Connect bindings:
  - For example, `apiv1connect.NewTokenUsageServiceHandler`.
- Ensure `cleanupFunc` closes both repositories.

## Hook Binary Design (`ctd-hook`)

### Responsibilities

- Run as a **short-lived process** invoked by Cursor hooks.
- Read a **single JSON document** from stdin, matching the Cursor hook schema:
  - Base fields: `conversation_id`, `generation_id`, `hook_event_name`, `workspace_roots`, etc.
  - Event-specific fields (depending on `hook_event_name`):
    - `beforeSubmitPrompt`: `{ "prompt": "<string>", "attachments": [...] }`
    - `afterAgentResponse`: `{ "text": "<assistant text>" }`
    - `afterMCPExecution`: `{ "tool_name": "...", "tool_input": {...}, "result_json": {...} }`
- Compute:
  - Token counts using a model-aware or approximate tokenizer.
  - Character counts as a fallback metric.
- Emit a **valid hook response** on stdout for events that require output:
  - `beforeSubmitPrompt`: `{ "continue": true }` (or `false` if policy changes are introduced).
  - `beforeShellExecution` / `beforeMCPExecution`: `{ "permission": "allow" }` by default.
- Send telemetry to `ctd`:
  - `TokenUsageService.CreateTokenUsage` for token metrics.
  - `ThoughtService.CreateThought` for thought content where appropriate (mirroring current behavior).

### Binary Location and Invocation

- Binary name: `ctd-hook` (configurable via documentation; the name itself is not enforced by Cursor).
- Installed path for user-level hooks:
  - `~/.cursor/hooks/ctd-hook`
- Example `~/.cursor/hooks.json` snippet:

```json
{
  "version": 1,
  "hooks": {
    "beforeSubmitPrompt": [
      { "command": "~/.cursor/hooks/ctd-hook beforeSubmitPrompt" }
    ],
    "afterAgentResponse": [
      { "command": "~/.cursor/hooks/ctd-hook afterAgentResponse" }
    ],
    "afterMCPExecution": [
      { "command": "~/.cursor/hooks/ctd-hook afterMCPExecution" }
    ],
    "afterAgentThought": [
      { "command": "~/.cursor/hooks/ctd-hook afterAgentThought" }
    ]
  }
}
```

The first argument after the binary name indicates the **expected event type**, which is used for validation and to select the appropriate decoder.

### Internal Structure

```mermaid
graph LR
    stdin[stdin JSON] --> Decoder[HookInput Decoder]
    Decoder --> Router[Event Router]
    Router --> PromptHandler[Prompt Event Handler]
    Router --> CompletionHandler[Completion Event Handler]
    Router --> ToolHandler[Tool Event Handler]

    PromptHandler --> Metrics[Token/Char Metrics]
    CompletionHandler --> Metrics
    ToolHandler --> Metrics

    Metrics --> Mapper[Mapper: Hook -> TokenUsage]
    Mapper --> Client[CTD gRPC Client]
    Client -->|CreateTokenUsage| CTD

    Router --> ThoughtMapper[Mapper: Hook -> Thought]
    ThoughtMapper --> Client
```

Key components:

- `HookInput` type representing the common envelope plus a raw `payload` field.
- Event-specific payload structs (for example, `BeforeSubmitPromptPayload`, `AfterAgentResponsePayload`).
- A `Tokenizer` abstraction to encapsulate model-aware token counting.
- A `CTDClient` facade exposing:
  - `CreateTokenUsage(ctx context.Context, usage TokenUsage) error`
  - `CreateThought(ctx context.Context, thought Thought) error`
- `Logger` using the existing `pkg/logger` package configured for console output.

### Performance and Non-Blocking Behavior

- **Fast startup and teardown**:
  - The binary avoids global state initialization beyond a logger and configuration parsing.
  - Connect client construction and HTTP transport reuse is kept minimal and efficient.
- **Tight timeouts** on outgoing calls to `ctd`:
  - Per-call deadline (for example, 250–500 ms) to avoid blocking Cursor operations.
  - On timeout or error, the hook:
    - Logs the failure with structured context.
    - Still returns a success hook decision (`allow` / `continue: true`) to Cursor.
- **Best-effort semantics**:
  - Telemetry failures never prevent the agent from proceeding.
  - No retries inside the hook binary to keep latency bounded.
- **Resource usage**:
  - Tokenization operates on in-memory strings; for large prompts or responses, the tokenizer operates linearly in input size.
  - No background goroutines survive process exit; the process lifetime is tightly scoped to the single hook invocation.

### Tokenization Strategy

- Introduce a small `Tokenizer` interface:

  - `CountTokens(model string, text string) (int64, error)`

- Provide an implementation that:
  - Uses a configurable default model encoding when the Cursor model name is unknown.
  - Can be swapped in tests to avoid heavy dependencies.
  - Optionally falls back to a simple heuristic (for example, characters divided by a constant) if a precise tokenizer is not available.

The specific tokenizer library is not mandated by this design; it must be:

- Pure Go (or otherwise easy to vendor) to avoid external runtime dependencies.
- Safe for untrusted input strings.

### Logging

- The hook uses the existing `pkg/logger` package so that log formatting and levels match the daemon.
- Logging principles:
  - Info-level events for major lifecycle edges:
    - Hook start and end.
    - Successful calls to `CreateTokenUsage` and `CreateThought`.
  - Debug-level events for detailed diagnostics:
    - Parsed event fields.
    - Token counts and character counts.
  - Error-level events for failures:
    - Network errors reaching `ctd`.
    - Tokenization failures.
- Logs are written to stderr to avoid interfering with the JSON response expected on stdout.

## Maintaining Current Functionality

The existing design expects `ctd` to receive **thought content** that is rendered by the extension. The new hook will:

- Continue to emit `ThoughtService.CreateThought` calls for:
  - Events that represent internal reasoning (`afterAgentThought`).
  - Optionally, summarized views of prompts and completions for additional visibility.
- Preserve existing fields:
  - `conversation_id`, `generation_id`, `text`, `duration_ms`, `model`, `hook_event_name`, `cursor_version`, `workspace_roots`, `user_email`, `timestamp`.

This ensures that the new TokenUsage path is additive and does not regress the current visualization behavior.

## Configuration

The hook binary reads configuration in the following precedence order:

1. Environment variables:
   - `CTD_HOOK_SERVER_URL` (for example, `http://127.0.0.1:36912`)
   - `CTD_HOOK_LOG_LEVEL` (for example, `info`, `debug`)
   - `CTD_HOOK_DEFAULT_MODEL` (fallback tokenizer model name)
2. Optional config file:
   - `$XDG_CONFIG_HOME/ctd/hook.yaml` with keys:
     - `server_url`
     - `log_level`
     - `default_model`
3. Built-in defaults:
   - Server URL: `http://127.0.0.1:36912`
   - Log level: `info`

The hook does not read from the main `ctd` server configuration to avoid tight coupling between the daemon deployment and user-specific Cursor configuration.

## Testing Strategy

### Hook Binary Tests

- **Unit tests**:
  - JSON decoding for each supported hook event.
  - `Tokenizer` behavior for representative inputs.
  - Mapping from hook input to domain `TokenUsage` and `Thought`.
  - Enforcement that hook output JSON remains valid when telemetry fails.
- **Integration tests**:
  - Run `ctd` in test mode using existing helpers.
  - Execute the hook binary with recorded hook JSON fixtures via stdin.
  - Assert that:
    - `TokenUsage` rows appear in the database with expected fields.
    - `Thought` rows are created where expected.
- **Performance tests**:
  - Stress test with synthetic events similar to existing performance tests for `CreateThought`.
  - Ensure per-event latency and memory usage remain within acceptable bounds.

### Server-Side Tests

- New tests for:
  - `TokenUsageServer.CreateTokenUsage` covering:
    - Valid input path.
    - Invalid UUIDs.
    - Negative token counts (rejected).
  - `PostgresTokenUsageRepository`:
    - Happy path stores.
    - Batch inserts.
    - Health checks.

All tests integrate with existing tooling and coverage requirements (80% minimum).

## Security and Privacy Considerations

- The hook processes **potentially sensitive content** (prompts and completions). The design mandates:
  - No logging of full prompt or completion text at info-level or higher.
  - Optional debug logs must be disabled by default and clearly documented.
- The token usage path stores **metadata plus aggregate metrics** rather than full content, which reduces stored sensitive data.
- Communications between the hook and `ctd`:
  - Use localhost by default.
  - May be extended to support TLS in future iterations if `ctd` is deployed remotely.
- The hook treats all errors as non-fatal to the Cursor agent, avoiding denial of service risk due to telemetry issues.

## Future Extensions

Potential follow-up work enabled by this design:

- Add query endpoints to `TokenUsageService`:
  - Per-user, per-day token usage summaries.
  - Per-workspace or per-model aggregates.
- Surface aggregate usage in the VS Code / Cursor extension:
  - “This session” token count.
  - Breakdown by model or workspace.
- Introduce cost calculation using per-model pricing configuration.
- Add more event-specific metrics (for example, tool call token usage, system prompt size).

This design establishes a dedicated, extensible TokenUsage path while preserving and enhancing the existing thought logging behavior.



```

You were hitting a Mermaid parse issue on the `TokenHook[ctd-hook (Go binary)]` line.

I’ve updated the diagram to avoid the problematic inline label and instead:

```mermaid
graph TB
    subgraph Cursor
        User[User] --> Agent[Cursor Agent / Tab]
        Agent --> HooksRuntime[Cursor Hooks Runtime]
    end

    HooksRuntime -->|hook_event + JSON over stdio| TokenHook
    TokenHook[ctd-hook Go binary]

    subgraph "Local Daemon"
        CTD[ctd (Thought + TokenUsage services)]
    end

    TokenHook --> CTD
    CTD --> PG[(PostgreSQL)]

    subgraph IDE
        Ext[VS Code / Cursor Extension]
    end

    CTD --> Ext
```

This keeps the same semantics but uses a separate node definition line for `TokenHook`, which should resolve the parse error.

```plaintext
graph TB
    subgraph Cursor
        User[User] --> Agent[Cursor Agent / Tab]
        Agent --> HooksRuntime[Cursor Hooks Runtime]
    end

    HooksRuntime -->|hook_event + JSON over stdio| TokenHook
    TokenHook[ctd-hook Go binary]

    subgraph "Local Daemon"
        CTD[ctd (Thought + TokenUsage services)]
    end

    TokenHook --> CTD
    CTD --> PG[(PostgreSQL)]

    subgraph IDE
        Ext[VS Code / Cursor Extension]
    end

    CTD --> Ext
```

---

**User**

Parse error on line 11: ...on" CTD[ctd (Thought + TokenUsag ----------------------^ Expecting 'SQE', 'DOUBLECIRCLEEND', 'PE', '-)', 'STADIUMEND', 'SUBROUTINEEND', 'PIPE', 'CYLINDEREND', 'DIAMOND_STOP', 'TAGEND', 'TRAPEND', 'INVTRAPEND', 'UNICODE_TEXT', 'TEXT', 'TAGSTART', got 'PS'





---

**Cursor**

```markdown
# Cursor Token Usage Hook and TokenUsage Service Design

## Overview

This document specifies the design for extending the Cursor Thought Daemon (`ctd`) ecosystem with:

- A **Go-based Cursor hook binary** that can be installed in `~/.cursor/hooks`.
- A new **TokenUsage domain model, service, and storage** path separate from existing `Thought` records (Option B data model).

The hook will:

- Observe Cursor agent and tab lifecycle events via the Cursor hooks JSON protocol.
- Aggregate prompt and completion information across events into structured token usage records.
- Forward both:
  - **Token usage** to a new `TokenUsageService` in `ctd`.
  - **Thought-like content** to the existing `ThoughtService` to preserve current behavior.
- Operate in a **non-blocking, best-effort** manner so that failures in the telemetry path never block the Cursor agent.

The design targets high throughput and low per-event latency while preserving the existing `ctd` architecture and logging conventions.

## Goals and Non-Goals

### Goals

- Record token usage for Cursor interactions with the following properties:
  - Associated with `conversation_id`, `generation_id`, model, user, and workspace roots.
  - Captures at minimum prompt and completion token counts, with room for additional metrics.
- Introduce a **separate TokenUsage model and service** in `ctd` without changing existing `Thought` semantics.
- Provide a **single Go binary hook** that:
  - Can be referenced from `hooks.json` in `~/.cursor` or project-level `.cursor/hooks.json`.
  - Handles both existing "thought logging" and new token accounting.
  - Uses the same logging conventions as the daemon (`pkg/logger` + `log/slog`).
- Ensure the hook is:
  - **Fast**: minimal CPU and memory overhead per event.
  - **Non-blocking**: does not block agent operations; on failure it degrades to no-op with logging.
  - **Scalable**: safe for high-frequency events and long-running sessions.

### Non-Goals

- Building a full-featured historical analytics API or UI for usage data (out of scope; this design enables it later).
- Computing precise currency cost; the design focuses on **token counts and related metrics**, with cost derivation as a possible follow-up.
- Enforcing quota or rate limiting decisions inside the hook; this iteration is observability-only.

## High-Level Architecture

### System Context

```mermaid
graph TB
    subgraph Cursor
        User[User] --> Agent[Cursor Agent / Tab]
        Agent --> HooksRuntime[Cursor Hooks Runtime]
    end

    HooksRuntime -->|hook_event + JSON over stdio| TokenHook
    TokenHook[ctd-hook Go binary]

    subgraph LocalDaemon
        CTD[ctd token services]
    end

    TokenHook --> CTD
    CTD --> PG[(PostgreSQL)]

    subgraph IDE
        Ext[VS Code / Cursor Extension]
    end

    CTD --> Ext
```

The new `ctd-hook` binary is invoked by the Cursor hooks runtime as a short-lived process. It receives hook input via stdin and writes hook decisions (for the events that expect them) to stdout. It independently calls the `ctd` daemon over Connect-RPC (HTTP/1.1 or h2c) to persist token usage and thought records.

### Sequence Flow per Generation

```mermaid
sequenceDiagram
    participant User
    participant Cursor as Cursor Agent
    participant Hooks as Cursor Hooks Runtime
    participant Hook as ctd-hook (Go)
    participant CTD as ctd Daemon
    participant DB as Postgres

    User->>Cursor: Edit prompt / run agent
    Cursor->>Hooks: beforeSubmitPrompt event
    Hooks->>Hook: JSON (prompt, conversation_id, generation_id, workspace_roots, ...)
    Hook->>Hook: Compute prompt token count
    Hook->>CTD: RecordTokenUsage (prompt metrics)
    CTD->>DB: INSERT token_usage row

    Cursor->>Hooks: afterAgentResponse event
    Hooks->>Hook: JSON (assistant text, ids, model, ...)
    Hook->>Hook: Compute completion token count
    Hook->>CTD: RecordTokenUsage (completion metrics)
    CTD->>DB: INSERT token_usage row (same generation_id, different stage)

    Note over Hook,CTD: Optionally also send ThoughtService.CreateThought<br/>entries for visibility (current functionality)
```

This design intentionally uses **multiple TokenUsage rows per `generation_id`** distinguished by stage (for example, `prompt`, `completion`, `tool`, `system`). Aggregation into per-generation or per-session totals is a downstream concern for querying and analytics.

## Data Model (Option B)

### Protobuf: TokenUsageService

Add a new proto definition alongside `ThoughtService` in `api/proto/v1/token_usage.proto` (file name and package aligned with existing layout):

- **Service**
  - `TokenUsageService`
    - `rpc CreateTokenUsage(CreateTokenUsageRequest) returns (CreateTokenUsageResponse);`

- **Messages**
  - `TokenUsageStage` (enum)
    - `TOKEN_USAGE_STAGE_UNSPECIFIED`
    - `TOKEN_USAGE_STAGE_PROMPT`
    - `TOKEN_USAGE_STAGE_COMPLETION`
    - `TOKEN_USAGE_STAGE_TOOL`
    - `TOKEN_USAGE_STAGE_SYSTEM`
  - `TokenUsage`
    - `string conversation_id`
    - `string generation_id`
    - `string model`
    - `TokenUsageStage stage`
    - `int64 prompt_tokens`
    - `int64 completion_tokens`
    - `int64 total_tokens`
    - `int64 prompt_chars`
    - `int64 completion_chars`
    - `repeated string workspace_roots`
    - `string user_email`
    - `google.protobuf.Timestamp timestamp`
  - `CreateTokenUsageRequest`
    - `TokenUsage usage`
  - `CreateTokenUsageResponse`
    - `string id` (server-generated usage record identifier)

#### Mapping Principles

- `conversation_id` and `generation_id` are copied from the Cursor hook envelope.
- `stage` is derived from the hook event name and payload:
  - `beforeSubmitPrompt` → `TOKEN_USAGE_STAGE_PROMPT`
  - `afterAgentResponse` → `TOKEN_USAGE_STAGE_COMPLETION`
  - `afterMCPExecution` → `TOKEN_USAGE_STAGE_TOOL`
  - Other events can be mapped in future iterations.
- `prompt_tokens` and `completion_tokens` are populated conservatively:
  - For a prompt-stage event, `prompt_tokens` is set, `completion_tokens` is zero.
  - For a completion-stage event, `completion_tokens` is set, `prompt_tokens` is zero.
  - `total_tokens` is always `prompt_tokens + completion_tokens`.

### Domain Model

Add a new domain type in `internal/domain/token_usage.go`:

- `TokenUsage`:
  - `ID uuid.UUID`
  - `ConversationID uuid.UUID`
  - `GenerationID uuid.UUID`
  - `Model string`
  - `Stage TokenUsageStage` (Go enum type)
  - `PromptTokens int64`
  - `CompletionTokens int64`
  - `TotalTokens int64`
  - `PromptChars int64`
  - `CompletionChars int64`
  - `WorkspaceRoots []string`
  - `UserEmail string`
  - `Timestamp time.Time`

Define a `TokenUsageRepository` interface:

- `Store(ctx context.Context, usage *TokenUsage) error`
- `StoreBatch(ctx context.Context, usages []*TokenUsage) error`
- `Close() error`
- `HealthCheck(ctx context.Context) error`

### Database Schema

Create a new `token_usage` table via a migration (for example, `002_token_usage.sql`):

- Columns:
  - `id UUID PRIMARY KEY DEFAULT gen_random_uuid()`
  - `conversation_id UUID NOT NULL`
  - `generation_id UUID NOT NULL`
  - `model VARCHAR(255) NOT NULL`
  - `stage VARCHAR(32) NOT NULL`
  - `prompt_tokens BIGINT NOT NULL DEFAULT 0`
  - `completion_tokens BIGINT NOT NULL DEFAULT 0`
  - `total_tokens BIGINT NOT NULL DEFAULT 0`
  - `prompt_chars BIGINT NOT NULL DEFAULT 0`
  - `completion_chars BIGINT NOT NULL DEFAULT 0`
  - `workspace_roots TEXT[] NOT NULL`
  - `user_email VARCHAR(255) NOT NULL`
  - `timestamp TIMESTAMPTZ NOT NULL`
  - `created_at TIMESTAMPTZ DEFAULT NOW()`

- Indexes:
  - `idx_token_usage_conversation_id (conversation_id)`
  - `idx_token_usage_generation_id (generation_id)`
  - `idx_token_usage_user_email (user_email)`
  - `idx_token_usage_timestamp (timestamp)`
  - Optional composite index for frequent queries, for example `(user_email, timestamp)`.

The schema allows multiple rows per `(conversation_id, generation_id)` pair, differentiated by `stage`. This directly reflects individual hook events without requiring cross-event aggregation in the hook process.

## Server-Side Components

### TokenUsageService Implementation

Add a new service implementation under `internal/server` or a dedicated package (for example, `internal/server/tokenusage.go`):

- Struct:
  - `TokenUsageServer` with:
    - `repo domain.TokenUsageRepository`
    - `log *logger.Logger`

- Methods:
  - `CreateTokenUsage(ctx context.Context, req *connect.Request[apiv1.CreateTokenUsageRequest]) (*connect.Response[apiv1.CreateTokenUsageResponse], error)`
    - Validate fields (UUID parsing, stage, non-negative token counts).
    - Map proto `TokenUsage` to domain `TokenUsage`.
    - Call `TokenUsageRepository.Store`.
    - Return generated `id`.

### Repository Implementation

Implement `TokenUsageRepository` in `internal/repository/postgres_token_usage.go`:

- Use `pgxpool.Pool` consistent with existing `PostgresRepository`.
- Insert query uses parameterized SQL with context-aware logging.
- `StoreBatch` uses `pgx.Batch` similar to the existing `StoreBatch` for `Thought`.
- Health checks piggyback on shared pool connectivity.

### Application Wiring

Extend `internal/app/app.go`:

- Construct a `TokenUsageRepository` alongside the existing `ThoughtRepository` using the same database configuration.
- Register the new `TokenUsageService` handler on the HTTP mux via generated Connect bindings:
  - For example, `apiv1connect.NewTokenUsageServiceHandler`.
- Ensure `cleanupFunc` closes both repositories.

## Hook Binary Design (`ctd-hook`)

### Responsibilities

- Run as a **short-lived process** invoked by Cursor hooks.
- Read a **single JSON document** from stdin, matching the Cursor hook schema:
  - Base fields: `conversation_id`, `generation_id`, `hook_event_name`, `workspace_roots`, etc.
  - Event-specific fields (depending on `hook_event_name`):
    - `beforeSubmitPrompt`: `{ "prompt": "<string>", "attachments": [...] }`
    - `afterAgentResponse`: `{ "text": "<assistant text>" }`
    - `afterMCPExecution`: `{ "tool_name": "...", "tool_input": {...}, "result_json": {...} }`
- Compute:
  - Token counts using a model-aware or approximate tokenizer.
  - Character counts as a fallback metric.
- Emit a **valid hook response** on stdout for events that require output:
  - `beforeSubmitPrompt`: `{ "continue": true }` (or `false` if policy changes are introduced).
  - `beforeShellExecution` / `beforeMCPExecution`: `{ "permission": "allow" }` by default.
- Send telemetry to `ctd`:
  - `TokenUsageService.CreateTokenUsage` for token metrics.
  - `ThoughtService.CreateThought` for thought content where appropriate (mirroring current behavior).

### Binary Location and Invocation

- Binary name: `ctd-hook` (configurable via documentation; the name itself is not enforced by Cursor).
- Installed path for user-level hooks:
  - `~/.cursor/hooks/ctd-hook`
- Example `~/.cursor/hooks.json` snippet:

```json
{
  "version": 1,
  "hooks": {
    "beforeSubmitPrompt": [
      { "command": "~/.cursor/hooks/ctd-hook beforeSubmitPrompt" }
    ],
    "afterAgentResponse": [
      { "command": "~/.cursor/hooks/ctd-hook afterAgentResponse" }
    ],
    "afterMCPExecution": [
      { "command": "~/.cursor/hooks/ctd-hook afterMCPExecution" }
    ],
    "afterAgentThought": [
      { "command": "~/.cursor/hooks/ctd-hook afterAgentThought" }
    ]
  }
}
```

The first argument after the binary name indicates the **expected event type**, which is used for validation and to select the appropriate decoder.

### Internal Structure

```mermaid
graph LR
    stdin[stdin JSON] --> Decoder[HookInput Decoder]
    Decoder --> Router[Event Router]
    Router --> PromptHandler[Prompt Event Handler]
    Router --> CompletionHandler[Completion Event Handler]
    Router --> ToolHandler[Tool Event Handler]

    PromptHandler --> Metrics[Token/Char Metrics]
    CompletionHandler --> Metrics
    ToolHandler --> Metrics

    Metrics --> Mapper[Mapper: Hook -> TokenUsage]
    Mapper --> Client[CTD gRPC Client]
    Client -->|CreateTokenUsage| CTD

    Router --> ThoughtMapper[Mapper: Hook -> Thought]
    ThoughtMapper --> Client
```

Key components:

- `HookInput` type representing the common envelope plus a raw `payload` field.
- Event-specific payload structs (for example, `BeforeSubmitPromptPayload`, `AfterAgentResponsePayload`).
- A `Tokenizer` abstraction to encapsulate model-aware token counting.
- A `CTDClient` facade exposing:
  - `CreateTokenUsage(ctx context.Context, usage TokenUsage) error`
  - `CreateThought(ctx context.Context, thought Thought) error`
- `Logger` using the existing `pkg/logger` package configured for console output.

### Performance and Non-Blocking Behavior

- **Fast startup and teardown**:
  - The binary avoids global state initialization beyond a logger and configuration parsing.
  - Connect client construction and HTTP transport reuse is kept minimal and efficient.
- **Tight timeouts** on outgoing calls to `ctd`:
  - Per-call deadline (for example, 250–500 ms) to avoid blocking Cursor operations.
  - On timeout or error, the hook:
    - Logs the failure with structured context.
    - Still returns a success hook decision (`allow` / `continue: true`) to Cursor.
- **Best-effort semantics**:
  - Telemetry failures never prevent the agent from proceeding.
  - No retries inside the hook binary to keep latency bounded.
- **Resource usage**:
  - Tokenization operates on in-memory strings; for large prompts or responses, the tokenizer operates linearly in input size.
  - No background goroutines survive process exit; the process lifetime is tightly scoped to the single hook invocation.

### Tokenization Strategy

- Introduce a small `Tokenizer` interface:

  - `CountTokens(model string, text string) (int64, error)`

- Provide an implementation that:
  - Uses a configurable default model encoding when the Cursor model name is unknown.
  - Can be swapped in tests to avoid heavy dependencies.
  - Optionally falls back to a simple heuristic (for example, characters divided by a constant) if a precise tokenizer is not available.

The specific tokenizer library is not mandated by this design; it must be:

- Pure Go (or otherwise easy to vendor) to avoid external runtime dependencies.
- Safe for untrusted input strings.

### Logging

- The hook uses the existing `pkg/logger` package so that log formatting and levels match the daemon.
- Logging principles:
  - Info-level events for major lifecycle edges:
    - Hook start and end.
    - Successful calls to `CreateTokenUsage` and `CreateThought`.
  - Debug-level events for detailed diagnostics:
    - Parsed event fields.
    - Token counts and character counts.
  - Error-level events for failures:
    - Network errors reaching `ctd`.
    - Tokenization failures.
- Logs are written to stderr to avoid interfering with the JSON response expected on stdout.

## Maintaining Current Functionality

The existing design expects `ctd` to receive **thought content** that is rendered by the extension. The new hook will:

- Continue to emit `ThoughtService.CreateThought` calls for:
  - Events that represent internal reasoning (`afterAgentThought`).
  - Optionally, summarized views of prompts and completions for additional visibility.
- Preserve existing fields:
  - `conversation_id`, `generation_id`, `text`, `duration_ms`, `model`, `hook_event_name`, `cursor_version`, `workspace_roots`, `user_email`, `timestamp`.

This ensures that the new TokenUsage path is additive and does not regress the current visualization behavior.

## Configuration

The hook binary reads configuration in the following precedence order:

1. Environment variables:
   - `CTD_HOOK_SERVER_URL` (for example, `http://127.0.0.1:36912`)
   - `CTD_HOOK_LOG_LEVEL` (for example, `info`, `debug`)
   - `CTD_HOOK_DEFAULT_MODEL` (fallback tokenizer model name)
2. Optional config file:
   - `$XDG_CONFIG_HOME/ctd/hook.yaml` with keys:
     - `server_url`
     - `log_level`
     - `default_model`
3. Built-in defaults:
   - Server URL: `http://127.0.0.1:36912`
   - Log level: `info`

The hook does not read from the main `ctd` server configuration to avoid tight coupling between the daemon deployment and user-specific Cursor configuration.

## Testing Strategy

### Hook Binary Tests

- **Unit tests**:
  - JSON decoding for each supported hook event.
  - `Tokenizer` behavior for representative inputs.
  - Mapping from hook input to domain `TokenUsage` and `Thought`.
  - Enforcement that hook output JSON remains valid when telemetry fails.
- **Integration tests**:
  - Run `ctd` in test mode using existing helpers.
  - Execute the hook binary with recorded hook JSON fixtures via stdin.
  - Assert that:
    - `TokenUsage` rows appear in the database with expected fields.
    - `Thought` rows are created where expected.
- **Performance tests**:
  - Stress test with synthetic events similar to existing performance tests for `CreateThought`.
  - Ensure per-event latency and memory usage remain within acceptable bounds.

### Server-Side Tests

- New tests for:
  - `TokenUsageServer.CreateTokenUsage` covering:
    - Valid input path.
    - Invalid UUIDs.
    - Negative token counts (rejected).
  - `PostgresTokenUsageRepository`:
    - Happy path stores.
    - Batch inserts.
    - Health checks.

All tests integrate with existing tooling and coverage requirements (80% minimum).

## Security and Privacy Considerations

- The hook processes **potentially sensitive content** (prompts and completions). The design mandates:
  - No logging of full prompt or completion text at info-level or higher.
  - Optional debug logs must be disabled by default and clearly documented.
- The token usage path stores **metadata plus aggregate metrics** rather than full content, which reduces stored sensitive data.
- Communications between the hook and `ctd`:
  - Use localhost by default.
  - May be extended to support TLS in future iterations if `ctd` is deployed remotely.
- The hook treats all errors as non-fatal to the Cursor agent, avoiding denial of service risk due to telemetry issues.

## Future Extensions

Potential follow-up work enabled by this design:

- Add query endpoints to `TokenUsageService`:
  - Per-user, per-day token usage summaries.
  - Per-workspace or per-model aggregates.
- Surface aggregate usage in the VS Code / Cursor extension:
  - “This session” token count.
  - Breakdown by model or workspace.
- Introduce cost calculation using per-model pricing configuration.
- Add more event-specific metrics (for example, tool call token usage, system prompt size).

This design establishes a dedicated, extensible TokenUsage path while preserving and enhancing the existing thought logging behavior.



```

The parse error was coming from the label `CTD[ctd (Thought + TokenUsage services)]` inside the `subgraph "Local Daemon"` block.

I have simplified that part of the diagram to avoid parentheses and the `+` symbol and also removed the quoted subgraph name, so the block now reads:

```mermaid
graph TB
    subgraph Cursor
        User[User] --> Agent[Cursor Agent / Tab]
        Agent --> HooksRuntime[Cursor Hooks Runtime]
    end

    HooksRuntime -->|hook_event + JSON over stdio| TokenHook
    TokenHook[ctd-hook Go binary]

    subgraph LocalDaemon
        CTD[ctd token services]
    end

    TokenHook --> CTD
    CTD --> PG[(PostgreSQL)]

    subgraph IDE
        Ext[VS Code / Cursor Extension]
    end

    CTD --> Ext
```

This should satisfy the Mermaid parser and still convey the intended structure.

```plaintext
graph TB
    subgraph Cursor
        User[User] --> Agent[Cursor Agent / Tab]
        Agent --> HooksRuntime[Cursor Hooks Runtime]
    end

    HooksRuntime -->|hook_event + JSON over stdio| TokenHook
    TokenHook[ctd-hook Go binary]

    subgraph LocalDaemon
        CTD[ctd token services]
    end

    TokenHook --> CTD
    CTD --> PG[(PostgreSQL)]

    subgraph IDE
        Ext[VS Code / Cursor Extension]
    end

    CTD --> Ext
```

---

**User**

come up with a detailed implementation plan for this design 

Implement the plan as specified, it is attached for your reference. Do NOT edit the plan file itself.

To-do's from the plan have already been created. Do not create them again. Mark them as in_progress as you work, starting with the first one. Don't stop until you have completed all the to-dos.

---

**Cursor**

```plaintext
syntax = "proto3";

package api.v1;

import "google/protobuf/timestamp.proto";

option go_package = "github.com/ainslies/ctd/gen/api/proto/v1;apiv1";

// TokenUsageService defines the API for recording token usage metrics.
service TokenUsageService {
  // CreateTokenUsage records a new token usage entry.
  rpc CreateTokenUsage(CreateTokenUsageRequest) returns (CreateTokenUsageResponse);
}

// TokenUsageStage indicates the stage of token usage within a generation.
enum TokenUsageStage {
  // Unspecified stage.
  TOKEN_USAGE_STAGE_UNSPECIFIED = 0;
  // Prompt tokens (user input).
  TOKEN_USAGE_STAGE_PROMPT = 1;
  // Completion tokens (assistant output).
  TOKEN_USAGE_STAGE_COMPLETION = 2;
  // Tool usage tokens (MCP tool calls).
  TOKEN_USAGE_STAGE_TOOL = 3;
  // System tokens (system prompt, context).
  TOKEN_USAGE_STAGE_SYSTEM = 4;
  // File edit tokens.
  TOKEN_USAGE_STAGE_FILE_EDIT = 5;
  // Stop event marker.
  TOKEN_USAGE_STAGE_STOP = 6;
}

// CreateTokenUsageRequest is the request to create a new token usage entry.
message CreateTokenUsageRequest {
  // The token usage data.
  TokenUsage usage = 1;
}

// CreateTokenUsageResponse is the response after creating a token usage entry.
message CreateTokenUsageResponse {
  // The ID of the created token usage record.
  string id = 1;
}

// TokenUsage represents token usage metrics for a single event.
message TokenUsage {
  // The ID of the conversation.
  string conversation_id = 1;
  // The unique ID of the generation.
  string generation_id = 2;
  // The AI model used.
  string model = 3;
  // The stage of token usage.
  TokenUsageStage stage = 4;
  // Number of prompt tokens.
  int64 prompt_tokens = 5;
  // Number of completion tokens.
  int64 completion_tokens = 6;
  // Total number of tokens (prompt + completion).
  int64 total_tokens = 7;
  // Number of prompt characters.
  int64 prompt_chars = 8;
  // Number of completion characters.
  int64 completion_chars = 9;
  // The workspace roots.
  repeated string workspace_roots = 10;
  // The user's email.
  string user_email = 11;
  // The timestamp when the usage was recorded.
  google.protobuf.Timestamp timestamp = 12;
  // The hook event name that triggered this usage record.
  string hook_event_name = 13;
}

```

```sql
-- Create token_usage table for storing token usage metrics
CREATE TABLE IF NOT EXISTS token_usage (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    conversation_id UUID NOT NULL,
    generation_id UUID NOT NULL,
    model VARCHAR(255) NOT NULL,
    stage VARCHAR(32) NOT NULL,
    prompt_tokens BIGINT NOT NULL DEFAULT 0,
    completion_tokens BIGINT NOT NULL DEFAULT 0,
    total_tokens BIGINT NOT NULL DEFAULT 0,
    prompt_chars BIGINT NOT NULL DEFAULT 0,
    completion_chars BIGINT NOT NULL DEFAULT 0,
    workspace_roots TEXT[] NOT NULL,
    user_email VARCHAR(255) NOT NULL,
    hook_event_name VARCHAR(255) NOT NULL,
    timestamp TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Create indexes for common query patterns
CREATE INDEX IF NOT EXISTS idx_token_usage_conversation_id ON token_usage(conversation_id);
CREATE INDEX IF NOT EXISTS idx_token_usage_generation_id ON token_usage(generation_id);
CREATE INDEX IF NOT EXISTS idx_token_usage_user_email ON token_usage(user_email);
CREATE INDEX IF NOT EXISTS idx_token_usage_timestamp ON token_usage(timestamp);
CREATE INDEX IF NOT EXISTS idx_token_usage_created_at ON token_usage(created_at);

-- Composite index for user-specific time-based queries
CREATE INDEX IF NOT EXISTS idx_token_usage_user_timestamp ON token_usage(user_email, timestamp);

-- Index for stage-based filtering
CREATE INDEX IF NOT EXISTS idx_token_usage_stage ON token_usage(stage);

```

```go
package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// TokenUsageStage represents the stage of token usage within a generation.
type TokenUsageStage string

const (
	// TokenUsageStageUnspecified indicates an unspecified stage.
	TokenUsageStageUnspecified TokenUsageStage = "unspecified"
	// TokenUsageStagePrompt indicates prompt tokens (user input).
	TokenUsageStagePrompt TokenUsageStage = "prompt"
	// TokenUsageStageCompletion indicates completion tokens (assistant output).
	TokenUsageStageCompletion TokenUsageStage = "completion"
	// TokenUsageStageTool indicates tool usage tokens (MCP tool calls).
	TokenUsageStageTool TokenUsageStage = "tool"
	// TokenUsageStageSystem indicates system tokens (system prompt, context).
	TokenUsageStageSystem TokenUsageStage = "system"
	// TokenUsageStageFileEdit indicates file edit tokens.
	TokenUsageStageFileEdit TokenUsageStage = "file_edit"
	// TokenUsageStageStop indicates a stop event marker.
	TokenUsageStageStop TokenUsageStage = "stop"
)

// TokenUsage represents token usage metrics for a single event.
type TokenUsage struct {
	ID               uuid.UUID       `json:"id"`
	ConversationID   uuid.UUID       `json:"conversation_id"`
	GenerationID     uuid.UUID       `json:"generation_id"`
	Model            string          `json:"model"`
	Stage            TokenUsageStage `json:"stage"`
	PromptTokens     int64           `json:"prompt_tokens"`
	CompletionTokens int64           `json:"completion_tokens"`
	TotalTokens      int64           `json:"total_tokens"`
	PromptChars      int64           `json:"prompt_chars"`
	CompletionChars  int64           `json:"completion_chars"`
	WorkspaceRoots   []string        `json:"workspace_roots"`
	UserEmail        string          `json:"user_email"`
	HookEventName    string          `json:"hook_event_name"`
	Timestamp        time.Time       `json:"timestamp"`
}

// TokenUsageRepository defines the interface for persisting token usage records.
type TokenUsageRepository interface {
	// Store saves a single token usage record to the repository.
	Store(ctx context.Context, usage *TokenUsage) error

	// StoreBatch saves multiple token usage records in a single transaction.
	StoreBatch(ctx context.Context, usages []*TokenUsage) error

	// Close closes the repository and releases resources.
	Close() error

	// HealthCheck verifies the repository is accessible.
	HealthCheck(ctx context.Context) error
}

```

```go
package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/ainslies/ctd/internal/config"
	"github.com/ainslies/ctd/internal/domain"
	"github.com/ainslies/ctd/pkg/logger"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresTokenUsageRepository implements the TokenUsageRepository interface using PostgreSQL.
type PostgresTokenUsageRepository struct {
	pool *pgxpool.Pool
	log  *logger.Logger
}

const insertTokenUsageQuery = `
	INSERT INTO token_usage (
		id, conversation_id, generation_id, model, stage,
		prompt_tokens, completion_tokens, total_tokens,
		prompt_chars, completion_chars, workspace_roots,
		user_email, hook_event_name, timestamp
	) VALUES (
		$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
	)
`

// NewPostgresTokenUsageRepository creates a new PostgreSQL token usage repository.
func NewPostgresTokenUsageRepository(
	ctx context.Context, cfg *config.DatabaseConfig, log *logger.Logger,
) (*PostgresTokenUsageRepository, error) {
	poolConfig, err := pgxpool.ParseConfig(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database URL: %w", err)
	}

	// Apply connection pool settings with overflow checks
	maxConns, err := safeIntToInt32(cfg.MaxConns, "max_conns")
	if err != nil {
		return nil, err
	}
	minConns, err := safeIntToInt32(cfg.MinConns, "min_conns")
	if err != nil {
		return nil, err
	}

	poolConfig.MaxConns = maxConns
	poolConfig.MinConns = minConns
	poolConfig.MaxConnLifetime = cfg.MaxConnLifetime
	poolConfig.MaxConnIdleTime = cfg.MaxConnIdleTime

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test the connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &PostgresTokenUsageRepository{
		pool: pool,
		log:  log,
	}, nil
}

// NewPostgresTokenUsageRepositoryWithPool creates a new PostgreSQL token usage repository
// using an existing connection pool. This is useful for sharing a pool between repositories.
func NewPostgresTokenUsageRepositoryWithPool(
	pool *pgxpool.Pool, log *logger.Logger,
) *PostgresTokenUsageRepository {
	return &PostgresTokenUsageRepository{
		pool: pool,
		log:  log,
	}
}

// Store saves a single token usage record to the database.
func (r *PostgresTokenUsageRepository) Store(ctx context.Context, usage *domain.TokenUsage) error {
	r.log.DebugContext(ctx, "storing token usage",
		"id", usage.ID,
		"generation_id", usage.GenerationID,
		"stage", usage.Stage,
	)

	_, err := r.pool.Exec(ctx, insertTokenUsageQuery,
		usage.ID,
		usage.ConversationID,
		usage.GenerationID,
		usage.Model,
		string(usage.Stage),
		usage.PromptTokens,
		usage.CompletionTokens,
		usage.TotalTokens,
		usage.PromptChars,
		usage.CompletionChars,
		usage.WorkspaceRoots,
		usage.UserEmail,
		usage.HookEventName,
		usage.Timestamp,
	)
	if err != nil {
		r.log.ErrorContext(ctx, "failed to insert token usage", "error", err, "id", usage.ID)
		return fmt.Errorf("failed to insert token usage: %w", err)
	}

	return nil
}

// StoreBatch saves multiple token usage records in a single transaction.
func (r *PostgresTokenUsageRepository) StoreBatch(ctx context.Context, usages []*domain.TokenUsage) error {
	if len(usages) == 0 {
		return nil
	}

	r.log.DebugContext(ctx, "storing batch token usage", "count", len(usages))

	transaction, err := r.pool.Begin(ctx)
	if err != nil {
		r.log.ErrorContext(ctx, "failed to begin transaction", "error", err)
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if rbErr := transaction.Rollback(ctx); rbErr != nil && !errors.Is(rbErr, pgx.ErrTxClosed) {
			r.log.ErrorContext(ctx, "failed to rollback transaction", "error", rbErr)
		}
	}()

	batch := &pgx.Batch{}

	for _, usage := range usages {
		batch.Queue(insertTokenUsageQuery,
			usage.ID,
			usage.ConversationID,
			usage.GenerationID,
			usage.Model,
			string(usage.Stage),
			usage.PromptTokens,
			usage.CompletionTokens,
			usage.TotalTokens,
			usage.PromptChars,
			usage.CompletionChars,
			usage.WorkspaceRoots,
			usage.UserEmail,
			usage.HookEventName,
			usage.Timestamp,
		)
	}

	batchResults := transaction.SendBatch(ctx, batch)
	defer func() {
		if closeErr := batchResults.Close(); closeErr != nil {
			r.log.ErrorContext(ctx, "failed to close batch results", "error", closeErr)
		}
	}()

	// Execute all queued commands
	for range usages {
		_, err := batchResults.Exec()
		if err != nil {
			return fmt.Errorf("failed to execute batch insert: %w", err)
		}
	}

	if err := batchResults.Close(); err != nil {
		return fmt.Errorf("failed to close batch results: %w", err)
	}

	if err := transaction.Commit(ctx); err != nil {
		r.log.ErrorContext(ctx, "failed to commit transaction", "error", err)
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// Close closes the repository and releases resources.
func (r *PostgresTokenUsageRepository) Close() error {
	r.pool.Close()
	return nil
}

// HealthCheck verifies the repository is accessible.
func (r *PostgresTokenUsageRepository) HealthCheck(ctx context.Context) error {
	if err := r.pool.Ping(ctx); err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}
	return nil
}

```

```go
package server

import (
	"context"
	"errors"
	"fmt"
	"time"

	"connectrpc.com/connect"
	apiv1 "github.com/ainslies/ctd/gen/api/proto/v1"
	"github.com/ainslies/ctd/internal/domain"
	"github.com/ainslies/ctd/pkg/logger"
	"github.com/google/uuid"
)

// TokenUsageServer implements the TokenUsageService API.
type TokenUsageServer struct {
	repo domain.TokenUsageRepository
	log  *logger.Logger
}

// NewTokenUsageServer creates a new TokenUsageServer.
func NewTokenUsageServer(repo domain.TokenUsageRepository, log *logger.Logger) *TokenUsageServer {
	return &TokenUsageServer{
		repo: repo,
		log:  log,
	}
}

// CreateTokenUsage records a new token usage entry.
func (s *TokenUsageServer) CreateTokenUsage(
	ctx context.Context,
	req *connect.Request[apiv1.CreateTokenUsageRequest],
) (*connect.Response[apiv1.CreateTokenUsageResponse], error) {
	s.log.InfoContext(ctx, "received CreateTokenUsage request")

	if req.Msg.GetUsage() == nil {
		s.log.WarnContext(ctx, "usage is required")
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("usage is required"))
	}

	usage, err := mapProtoToTokenUsage(req.Msg.GetUsage())
	if err != nil {
		s.log.WarnContext(ctx, "invalid token usage data", "error", err)
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid token usage data: %w", err))
	}

	// Validate token counts are non-negative
	if usage.PromptTokens < 0 || usage.CompletionTokens < 0 || usage.TotalTokens < 0 {
		s.log.WarnContext(ctx, "token counts must be non-negative")
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("token counts must be non-negative"))
	}

	s.log.DebugContext(ctx, "mapped token usage",
		"id", usage.ID,
		"conversation_id", usage.ConversationID,
		"stage", usage.Stage,
	)

	if err := s.repo.Store(ctx, usage); err != nil {
		s.log.ErrorContext(ctx, "failed to store token usage", "error", err)
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to store token usage: %w", err))
	}

	s.log.InfoContext(ctx, "successfully created token usage", "id", usage.ID)

	res := connect.NewResponse(&apiv1.CreateTokenUsageResponse{
		Id: usage.ID.String(),
	})
	return res, nil
}

// mapProtoToTokenUsage converts a proto TokenUsage to a domain TokenUsage.
func mapProtoToTokenUsage(protoUsage *apiv1.TokenUsage) (*domain.TokenUsage, error) {
	conversationID, err := uuid.Parse(protoUsage.GetConversationId())
	if err != nil {
		return nil, fmt.Errorf("invalid conversation_id: %w", err)
	}

	generationID, err := uuid.Parse(protoUsage.GetGenerationId())
	if err != nil {
		return nil, fmt.Errorf("invalid generation_id: %w", err)
	}

	// Default timestamp to now if missing or zero
	timestamp := time.Now()
	if protoUsage.GetTimestamp() != nil {
		timestamp = protoUsage.GetTimestamp().AsTime()
	}

	stage := mapProtoStageToTokenUsageStage(protoUsage.GetStage())

	return &domain.TokenUsage{
		ID:               uuid.New(),
		ConversationID:   conversationID,
		GenerationID:     generationID,
		Model:            protoUsage.GetModel(),
		Stage:            stage,
		PromptTokens:     protoUsage.GetPromptTokens(),
		CompletionTokens: protoUsage.GetCompletionTokens(),
		TotalTokens:      protoUsage.GetTotalTokens(),
		PromptChars:      protoUsage.GetPromptChars(),
		CompletionChars:  protoUsage.GetCompletionChars(),
		WorkspaceRoots:   protoUsage.GetWorkspaceRoots(),
		UserEmail:        protoUsage.GetUserEmail(),
		HookEventName:    protoUsage.GetHookEventName(),
		Timestamp:        timestamp,
	}, nil
}

// mapProtoStageToTokenUsageStage converts a proto TokenUsageStage to a domain TokenUsageStage.
func mapProtoStageToTokenUsageStage(protoStage apiv1.TokenUsageStage) domain.TokenUsageStage {
	switch protoStage {
	case apiv1.TokenUsageStage_TOKEN_USAGE_STAGE_PROMPT:
		return domain.TokenUsageStagePrompt
	case apiv1.TokenUsageStage_TOKEN_USAGE_STAGE_COMPLETION:
		return domain.TokenUsageStageCompletion
	case apiv1.TokenUsageStage_TOKEN_USAGE_STAGE_TOOL:
		return domain.TokenUsageStageTool
	case apiv1.TokenUsageStage_TOKEN_USAGE_STAGE_SYSTEM:
		return domain.TokenUsageStageSystem
	case apiv1.TokenUsageStage_TOKEN_USAGE_STAGE_FILE_EDIT:
		return domain.TokenUsageStageFileEdit
	case apiv1.TokenUsageStage_TOKEN_USAGE_STAGE_STOP:
		return domain.TokenUsageStageStop
	default:
		return domain.TokenUsageStageUnspecified
	}
}

// mapDomainStageToProtoStage converts a domain TokenUsageStage to a proto TokenUsageStage.
func mapDomainStageToProtoStage(stage domain.TokenUsageStage) apiv1.TokenUsageStage {
	switch stage {
	case domain.TokenUsageStagePrompt:
		return apiv1.TokenUsageStage_TOKEN_USAGE_STAGE_PROMPT
	case domain.TokenUsageStageCompletion:
		return apiv1.TokenUsageStage_TOKEN_USAGE_STAGE_COMPLETION
	case domain.TokenUsageStageTool:
		return apiv1.TokenUsageStage_TOKEN_USAGE_STAGE_TOOL
	case domain.TokenUsageStageSystem:
		return apiv1.TokenUsageStage_TOKEN_USAGE_STAGE_SYSTEM
	case domain.TokenUsageStageFileEdit:
		return apiv1.TokenUsageStage_TOKEN_USAGE_STAGE_FILE_EDIT
	case domain.TokenUsageStageStop:
		return apiv1.TokenUsageStage_TOKEN_USAGE_STAGE_STOP
	default:
		return apiv1.TokenUsageStage_TOKEN_USAGE_STAGE_UNSPECIFIED
	}
}

```

```go
// Package app provides application setup and initialization logic for the Cursor Thought Daemon.
package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"connectrpc.com/connect"
	"github.com/ainslies/ctd/gen/api/proto/v1/apiv1connect"
	"github.com/ainslies/ctd/internal/config"
	"github.com/ainslies/ctd/internal/domain"
	"github.com/ainslies/ctd/internal/handlers"
	"github.com/ainslies/ctd/internal/middleware"
	"github.com/ainslies/ctd/internal/repository"
	"github.com/ainslies/ctd/internal/server"
	"github.com/ainslies/ctd/pkg/logger"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

const (
	// DefaultReadHeaderTimeout is the default timeout for reading request headers.
	DefaultReadHeaderTimeout = 5 * time.Second
	// DefaultReadTimeout is the default timeout for reading the entire request.
	DefaultReadTimeout = 5 * time.Minute
	// DefaultIdleTimeout is the default idle timeout for keep-alive connections.
	DefaultIdleTimeout = 120 * time.Second
	// DefaultBroadcasterBufferSize is the default channel buffer size for the broadcaster.
	DefaultBroadcasterBufferSize = 50
	// DefaultShutdownTimeout is the default timeout for graceful shutdown.
	DefaultShutdownTimeout = 10 * time.Second
)

// Repositories holds all repository instances for the application.
type Repositories struct {
	Thought    domain.ThoughtRepository
	TokenUsage domain.TokenUsageRepository
}

// SetupDatabase connects to the database and verifies the connection.
func SetupDatabase(
	ctx context.Context, cfg *config.DatabaseConfig, log *logger.Logger,
) (domain.ThoughtRepository, error) {
	log.Info("Connecting to database", "max_conns", cfg.MaxConns)
	repo, err := repository.NewPostgresRepository(ctx, cfg, log)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Health check
	if err := repo.HealthCheck(ctx); err != nil {
		if closeErr := repo.Close(); closeErr != nil {
			log.Error("Failed to close repository", "error", closeErr)
		}
		return nil, fmt.Errorf("database health check failed: %w", err)
	}
	log.Info("Database connection established")

	return repo, nil
}

// SetupRepositories connects to the database and creates all repository instances.
func SetupRepositories(
	ctx context.Context, cfg *config.DatabaseConfig, log *logger.Logger,
) (*Repositories, error) {
	log.Info("Connecting to database", "max_conns", cfg.MaxConns)

	// Create thought repository
	thoughtRepo, err := repository.NewPostgresRepository(ctx, cfg, log)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database for thoughts: %w", err)
	}

	// Health check
	if err := thoughtRepo.HealthCheck(ctx); err != nil {
		if closeErr := thoughtRepo.Close(); closeErr != nil {
			log.Error("Failed to close thought repository", "error", closeErr)
		}
		return nil, fmt.Errorf("database health check failed: %w", err)
	}

	// Create token usage repository
	tokenUsageRepo, err := repository.NewPostgresTokenUsageRepository(ctx, cfg, log)
	if err != nil {
		if closeErr := thoughtRepo.Close(); closeErr != nil {
			log.Error("Failed to close thought repository", "error", closeErr)
		}
		return nil, fmt.Errorf("failed to connect to database for token usage: %w", err)
	}

	log.Info("Database connections established")

	return &Repositories{
		Thought:    thoughtRepo,
		TokenUsage: tokenUsageRepo,
	}, nil
}

// Close closes all repositories.
func (r *Repositories) Close() error {
	var errs []error
	if err := r.Thought.Close(); err != nil {
		errs = append(errs, fmt.Errorf("failed to close thought repository: %w", err))
	}
	if err := r.TokenUsage.Close(); err != nil {
		errs = append(errs, fmt.Errorf("failed to close token usage repository: %w", err))
	}
	if len(errs) > 0 {
		return fmt.Errorf("close errors: %v", errs)
	}
	return nil
}

// CleanupFunc is a function that performs cleanup on shutdown.
type CleanupFunc func(context.Context) error

// SetupRoutes configures the HTTP routes and middleware.
// Returns the handler and a cleanup function.
func SetupRoutes(repo domain.ThoughtRepository, log *logger.Logger) (http.Handler, CleanupFunc) {
	// Initialize API server with default broadcaster config
	// In the future, we can pass this config in
	broadcasterCfg := server.BroadcasterConfig{
		ChannelBufferSize: DefaultBroadcasterBufferSize,
	}
	thoughtServer := server.NewThoughtServer(repo, log, broadcasterCfg)

	// Create logging interceptor
	loggingInterceptor := middleware.NewLoggingInterceptor(log)

	path, handler := apiv1connect.NewThoughtServiceHandler(
		thoughtServer,
		connect.WithInterceptors(loggingInterceptor),
	)

	mux := http.NewServeMux()
	mux.Handle(path, handler)

	// Add health check endpoint
	mux.Handle("/health", handlers.NewHealthCheckHandler(repo, log))

	// Add ping endpoint
	mux.HandleFunc("/ping", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, `{"status":"ok"}`)
	})

	// Add CORS middleware wrapper
	corsMiddleware := middleware.NewCORSMiddleware()

	cleanupFunc := func(ctx context.Context) error {
		return thoughtServer.Shutdown(ctx)
	}

	return corsMiddleware(mux), cleanupFunc
}

// SetupServer creates the HTTP server.
func SetupServer(cfg *config.ServerConfig, handler http.Handler) *http.Server {
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	return &http.Server{
		Addr:              addr,
		Handler:           h2c.NewHandler(handler, &http2.Server{}),
		ReadHeaderTimeout: DefaultReadHeaderTimeout,
		ReadTimeout:       DefaultReadTimeout,
		WriteTimeout:      0, // Disable write timeout for streaming
		IdleTimeout:       DefaultIdleTimeout,
	}
}

// Run starts the server and handles graceful shutdown.
func Run(ctx context.Context, srv *http.Server, log *logger.Logger, cleanup func(context.Context) error) error {
	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Start server
	go func() {
		log.Info("Starting HTTP/gRPC server", "address", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("Server failed", "error", err)
			// We can't easily exit here without os.Exit, but the main loop will catch the signal or context cancellation
		}
	}()

	// Wait for shutdown signal
	select {
	case <-sigChan:
		log.Info("Received shutdown signal, stopping gracefully")
	case <-ctx.Done():
		log.Info("Context canceled, stopping gracefully")
	}

	// Shutdown server with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(ctx, DefaultShutdownTimeout)
	defer shutdownCancel()

	var errs []error

	// Perform custom cleanup
	if cleanup != nil {
		if err := cleanup(shutdownCtx); err != nil {
			errs = append(errs, fmt.Errorf("cleanup failed: %w", err))
		}
	}

	// Shutdown HTTP server
	if err := srv.Shutdown(shutdownCtx); err != nil {
		errs = append(errs, fmt.Errorf("server shutdown failed: %w", err))
	}

	if len(errs) > 0 {
		return fmt.Errorf("shutdown errors: %v", errs)
	}

	return nil
}

```

```go
// Package app provides application setup and initialization logic for the Cursor Thought Daemon.
package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"connectrpc.com/connect"
	"github.com/ainslies/ctd/gen/api/proto/v1/apiv1connect"
	"github.com/ainslies/ctd/internal/config"
	"github.com/ainslies/ctd/internal/domain"
	"github.com/ainslies/ctd/internal/handlers"
	"github.com/ainslies/ctd/internal/middleware"
	"github.com/ainslies/ctd/internal/repository"
	"github.com/ainslies/ctd/internal/server"
	"github.com/ainslies/ctd/pkg/logger"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

const (
	// DefaultReadHeaderTimeout is the default timeout for reading request headers.
	DefaultReadHeaderTimeout = 5 * time.Second
	// DefaultReadTimeout is the default timeout for reading the entire request.
	DefaultReadTimeout = 5 * time.Minute
	// DefaultIdleTimeout is the default idle timeout for keep-alive connections.
	DefaultIdleTimeout = 120 * time.Second
	// DefaultBroadcasterBufferSize is the default channel buffer size for the broadcaster.
	DefaultBroadcasterBufferSize = 50
	// DefaultShutdownTimeout is the default timeout for graceful shutdown.
	DefaultShutdownTimeout = 10 * time.Second
)

// Repositories holds all repository instances for the application.
type Repositories struct {
	Thought    domain.ThoughtRepository
	TokenUsage domain.TokenUsageRepository
}

// SetupDatabase connects to the database and verifies the connection.
func SetupDatabase(
	ctx context.Context, cfg *config.DatabaseConfig, log *logger.Logger,
) (domain.ThoughtRepository, error) {
	log.Info("Connecting to database", "max_conns", cfg.MaxConns)
	repo, err := repository.NewPostgresRepository(ctx, cfg, log)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Health check
	if err := repo.HealthCheck(ctx); err != nil {
		if closeErr := repo.Close(); closeErr != nil {
			log.Error("Failed to close repository", "error", closeErr)
		}
		return nil, fmt.Errorf("database health check failed: %w", err)
	}
	log.Info("Database connection established")

	return repo, nil
}

// SetupRepositories connects to the database and creates all repository instances.
func SetupRepositories(
	ctx context.Context, cfg *config.DatabaseConfig, log *logger.Logger,
) (*Repositories, error) {
	log.Info("Connecting to database", "max_conns", cfg.MaxConns)

	// Create thought repository
	thoughtRepo, err := repository.NewPostgresRepository(ctx, cfg, log)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database for thoughts: %w", err)
	}

	// Health check
	if err := thoughtRepo.HealthCheck(ctx); err != nil {
		if closeErr := thoughtRepo.Close(); closeErr != nil {
			log.Error("Failed to close thought repository", "error", closeErr)
		}
		return nil, fmt.Errorf("database health check failed: %w", err)
	}

	// Create token usage repository
	tokenUsageRepo, err := repository.NewPostgresTokenUsageRepository(ctx, cfg, log)
	if err != nil {
		if closeErr := thoughtRepo.Close(); closeErr != nil {
			log.Error("Failed to close thought repository", "error", closeErr)
		}
		return nil, fmt.Errorf("failed to connect to database for token usage: %w", err)
	}

	log.Info("Database connections established")

	return &Repositories{
		Thought:    thoughtRepo,
		TokenUsage: tokenUsageRepo,
	}, nil
}

// Close closes all repositories.
func (r *Repositories) Close() error {
	var errs []error
	if err := r.Thought.Close(); err != nil {
		errs = append(errs, fmt.Errorf("failed to close thought repository: %w", err))
	}
	if err := r.TokenUsage.Close(); err != nil {
		errs = append(errs, fmt.Errorf("failed to close token usage repository: %w", err))
	}
	if len(errs) > 0 {
		return fmt.Errorf("close errors: %v", errs)
	}
	return nil
}

// CleanupFunc is a function that performs cleanup on shutdown.
type CleanupFunc func(context.Context) error

// SetupRoutes configures the HTTP routes and middleware.
// Returns the handler and a cleanup function.
func SetupRoutes(repo domain.ThoughtRepository, log *logger.Logger) (http.Handler, CleanupFunc) {
	// Initialize API server with default broadcaster config
	// In the future, we can pass this config in
	broadcasterCfg := server.BroadcasterConfig{
		ChannelBufferSize: DefaultBroadcasterBufferSize,
	}
	thoughtServer := server.NewThoughtServer(repo, log, broadcasterCfg)

	// Create logging interceptor
	loggingInterceptor := middleware.NewLoggingInterceptor(log)

	path, handler := apiv1connect.NewThoughtServiceHandler(
		thoughtServer,
		connect.WithInterceptors(loggingInterceptor),
	)

	mux := http.NewServeMux()
	mux.Handle(path, handler)

	// Add health check endpoint
	mux.Handle("/health", handlers.NewHealthCheckHandler(repo, log))

	// Add ping endpoint
	mux.HandleFunc("/ping", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, `{"status":"ok"}`)
	})

	// Add CORS middleware wrapper
	corsMiddleware := middleware.NewCORSMiddleware()

	cleanupFunc := func(ctx context.Context) error {
		return thoughtServer.Shutdown(ctx)
	}

	return corsMiddleware(mux), cleanupFunc
}

// SetupRoutesWithRepositories configures the HTTP routes and middleware with all services.
// Returns the handler and a cleanup function.
func SetupRoutesWithRepositories(repos *Repositories, log *logger.Logger) (http.Handler, CleanupFunc) {
	// Initialize thought API server with default broadcaster config
	broadcasterCfg := server.BroadcasterConfig{
		ChannelBufferSize: DefaultBroadcasterBufferSize,
	}
	thoughtServer := server.NewThoughtServer(repos.Thought, log, broadcasterCfg)

	// Initialize token usage API server
	tokenUsageServer := server.NewTokenUsageServer(repos.TokenUsage, log)

	// Create logging interceptor
	loggingInterceptor := middleware.NewLoggingInterceptor(log)

	mux := http.NewServeMux()

	// Register ThoughtService
	thoughtPath, thoughtHandler := apiv1connect.NewThoughtServiceHandler(
		thoughtServer,
		connect.WithInterceptors(loggingInterceptor),
	)
	mux.Handle(thoughtPath, thoughtHandler)

	// Register TokenUsageService
	tokenUsagePath, tokenUsageHandler := apiv1connect.NewTokenUsageServiceHandler(
		tokenUsageServer,
		connect.WithInterceptors(loggingInterceptor),
	)
	mux.Handle(tokenUsagePath, tokenUsageHandler)

	// Add health check endpoint
	mux.Handle("/health", handlers.NewHealthCheckHandler(repos.Thought, log))

	// Add ping endpoint
	mux.HandleFunc("/ping", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, `{"status":"ok"}`)
	})

	// Add CORS middleware wrapper
	corsMiddleware := middleware.NewCORSMiddleware()

	cleanupFunc := func(ctx context.Context) error {
		return thoughtServer.Shutdown(ctx)
	}

	return corsMiddleware(mux), cleanupFunc
}

// SetupServer creates the HTTP server.
func SetupServer(cfg *config.ServerConfig, handler http.Handler) *http.Server {
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	return &http.Server{
		Addr:              addr,
		Handler:           h2c.NewHandler(handler, &http2.Server{}),
		ReadHeaderTimeout: DefaultReadHeaderTimeout,
		ReadTimeout:       DefaultReadTimeout,
		WriteTimeout:      0, // Disable write timeout for streaming
		IdleTimeout:       DefaultIdleTimeout,
	}
}

// Run starts the server and handles graceful shutdown.
func Run(ctx context.Context, srv *http.Server, log *logger.Logger, cleanup func(context.Context) error) error {
	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Start server
	go func() {
		log.Info("Starting HTTP/gRPC server", "address", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("Server failed", "error", err)
			// We can't easily exit here without os.Exit, but the main loop will catch the signal or context cancellation
		}
	}()

	// Wait for shutdown signal
	select {
	case <-sigChan:
		log.Info("Received shutdown signal, stopping gracefully")
	case <-ctx.Done():
		log.Info("Context canceled, stopping gracefully")
	}

	// Shutdown server with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(ctx, DefaultShutdownTimeout)
	defer shutdownCancel()

	var errs []error

	// Perform custom cleanup
	if cleanup != nil {
		if err := cleanup(shutdownCtx); err != nil {
			errs = append(errs, fmt.Errorf("cleanup failed: %w", err))
		}
	}

	// Shutdown HTTP server
	if err := srv.Shutdown(shutdownCtx); err != nil {
		errs = append(errs, fmt.Errorf("server shutdown failed: %w", err))
	}

	if len(errs) > 0 {
		return fmt.Errorf("shutdown errors: %v", errs)
	}

	return nil
}

```

Now I need to update the main.go to use the new setup functions:

```go
// Package main provides the entry point for the Cursor Thought Daemon server.
package main

import (
	"context"
	"fmt"
	"net/url"
	"os"

	"github.com/ainslies/ctd/internal/app"
	"github.com/ainslies/ctd/internal/config"
	"github.com/ainslies/ctd/pkg/logger"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// Initialize logger
	log := logger.New()
	log.Info("Starting Cursor Thought Daemon")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Error("Failed to load configuration", "error", err)
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		log.Error("Invalid configuration", "error", err)
		return fmt.Errorf("invalid configuration: %w", err)
	}

	log.Info("Configuration loaded successfully")
	log.Info("Server configured", "host", cfg.Server.Host, "port", cfg.Server.Port)
	log.Info("Database connection configured", "url", maskDatabaseURL(cfg.Database.URL))

	// Create context that can be cancelled
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize database and repositories
	repos, err := app.SetupRepositories(ctx, &cfg.Database, log)
	if err != nil {
		log.Error("Failed to setup database", "error", err)
		return fmt.Errorf("failed to setup database: %w", err)
	}
	defer func() {
		if err := repos.Close(); err != nil {
			log.Error("Failed to close repositories", "error", err)
		}
	}()

	// Setup routes and server with all services
	handler, cleanup := app.SetupRoutesWithRepositories(repos, log)
	srv := app.SetupServer(&cfg.Server, handler)

	// Run server
	if err := app.Run(ctx, srv, log, cleanup); err != nil {
		log.Error("Server error", "error", err)
		return fmt.Errorf("server error: %w", err)
	}

	// Cancel context to stop any background tasks
	cancel()
	log.Info("Shutdown complete")
	return nil
}

// maskDatabaseURL masks sensitive information in database URL.
func maskDatabaseURL(rawURL string) string {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "***"
	}
	if parsedURL.User != nil {
		parsedURL.User = url.UserPassword("***", "***")
	}
	return parsedURL.String()
}

```

```go
// Package hook provides the Cursor hook binary functionality for token usage tracking.
package hook

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/adrg/xdg"
	"github.com/spf13/viper"
)

const (
	// DefaultServerURL is the default ctd server URL.
	DefaultServerURL = "http://127.0.0.1:36912"
	// DefaultLogLevel is the default log level.
	DefaultLogLevel = "info"
	// DefaultModel is the default model for token counting.
	DefaultModel = "gpt-4"
	// DefaultTimeout is the default timeout for ctd calls in milliseconds.
	DefaultTimeout = 500

	// EnvPrefix is the prefix for environment variables.
	EnvPrefix = "CTD_HOOK"
)

// Config holds the hook configuration.
type Config struct {
	// ServerURL is the URL of the ctd server.
	ServerURL string `mapstructure:"server_url"`
	// LogLevel is the log level (debug, info, warn, error).
	LogLevel string `mapstructure:"log_level"`
	// DefaultModel is the default model for token counting when not specified.
	DefaultModel string `mapstructure:"default_model"`
	// TimeoutMs is the timeout for ctd calls in milliseconds.
	TimeoutMs int `mapstructure:"timeout_ms"`
}

// LoadConfig loads the hook configuration from environment variables and config file.
func LoadConfig() (*Config, error) {
	v := viper.New()

	// Set defaults
	v.SetDefault("server_url", DefaultServerURL)
	v.SetDefault("log_level", DefaultLogLevel)
	v.SetDefault("default_model", DefaultModel)
	v.SetDefault("timeout_ms", DefaultTimeout)

	// Environment variables
	v.SetEnvPrefix(EnvPrefix)
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Config file paths
	configDir := filepath.Join(xdg.ConfigHome, "ctd")
	v.SetConfigName("hook")
	v.SetConfigType("yaml")
	v.AddConfigPath(configDir)

	// Try to read config file (ignore if not found)
	if err := v.ReadInConfig(); err != nil {
		// Config file not found is not an error
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			// Only return error if it's not a "file not found" error
			// and the file actually exists but couldn't be read
			if _, statErr := os.Stat(filepath.Join(configDir, "hook.yaml")); statErr == nil {
				return nil, err
			}
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// Validate validates the configuration.
func (c *Config) Validate() error {
	// Basic validation - server URL must be set
	if c.ServerURL == "" {
		c.ServerURL = DefaultServerURL
	}
	if c.LogLevel == "" {
		c.LogLevel = DefaultLogLevel
	}
	if c.DefaultModel == "" {
		c.DefaultModel = DefaultModel
	}
	if c.TimeoutMs <= 0 {
		c.TimeoutMs = DefaultTimeout
	}
	return nil
}

```

```go
package hook

import (
	"encoding/json"
	"fmt"
	"io"
)

// HookEventName represents the type of hook event.
type HookEventName string

const (
	// EventBeforeSubmitPrompt is fired before a prompt is submitted.
	EventBeforeSubmitPrompt HookEventName = "beforeSubmitPrompt"
	// EventAfterAgentResponse is fired after the agent responds.
	EventAfterAgentResponse HookEventName = "afterAgentResponse"
	// EventAfterAgentThought is fired after the agent generates a thought.
	EventAfterAgentThought HookEventName = "afterAgentThought"
	// EventAfterMCPExecution is fired after an MCP tool is executed.
	EventAfterMCPExecution HookEventName = "afterMCPExecution"
	// EventAfterFileEdit is fired after a file is edited.
	EventAfterFileEdit HookEventName = "afterFileEdit"
	// EventStop is fired when the agent loop ends.
	EventStop HookEventName = "stop"
	// EventBeforeShellExecution is fired before a shell command is executed.
	EventBeforeShellExecution HookEventName = "beforeShellExecution"
	// EventAfterShellExecution is fired after a shell command is executed.
	EventAfterShellExecution HookEventName = "afterShellExecution"
	// EventBeforeMCPExecution is fired before an MCP tool is executed.
	EventBeforeMCPExecution HookEventName = "beforeMCPExecution"
	// EventBeforeReadFile is fired before a file is read.
	EventBeforeReadFile HookEventName = "beforeReadFile"
)

// HookInput represents the common envelope for all hook events.
type HookInput struct {
	// ConversationID is the unique identifier for the conversation.
	ConversationID string `json:"conversation_id"`
	// GenerationID is the unique identifier for the generation.
	GenerationID string `json:"generation_id"`
	// HookEventName is the name of the hook event.
	HookEventName HookEventName `json:"hook_event_name"`
	// WorkspaceRoots are the workspace root directories.
	WorkspaceRoots []string `json:"workspace_roots"`

	// RawPayload holds the raw JSON for event-specific fields.
	// This is populated after initial parsing.
	RawPayload json.RawMessage `json:"-"`
}

// BeforeSubmitPromptPayload is the payload for beforeSubmitPrompt events.
type BeforeSubmitPromptPayload struct {
	// Prompt is the user's prompt text.
	Prompt string `json:"prompt"`
	// Attachments are files or rules attached to the prompt.
	Attachments []Attachment `json:"attachments"`
}

// Attachment represents a file or rule attachment.
type Attachment struct {
	// Type is the attachment type (file, rule).
	Type string `json:"type"`
	// FilePath is the path to the attached file.
	FilePath string `json:"filePath,omitempty"`
}

// AfterAgentResponsePayload is the payload for afterAgentResponse events.
type AfterAgentResponsePayload struct {
	// Text is the assistant's response text.
	Text string `json:"text"`
}

// AfterAgentThoughtPayload is the payload for afterAgentThought events.
type AfterAgentThoughtPayload struct {
	// Text is the thought text.
	Text string `json:"text"`
}

// AfterMCPExecutionPayload is the payload for afterMCPExecution events.
type AfterMCPExecutionPayload struct {
	// ToolName is the name of the MCP tool.
	ToolName string `json:"tool_name"`
	// ToolInput is the JSON input to the tool.
	ToolInput json.RawMessage `json:"tool_input"`
	// ResultJSON is the JSON result from the tool.
	ResultJSON json.RawMessage `json:"result_json"`
}

// BeforeMCPExecutionPayload is the payload for beforeMCPExecution events.
type BeforeMCPExecutionPayload struct {
	// ToolName is the name of the MCP tool.
	ToolName string `json:"tool_name"`
	// ToolInput is the JSON input to the tool.
	ToolInput json.RawMessage `json:"tool_input"`
	// URL is the server URL (optional).
	URL string `json:"url,omitempty"`
	// Command is the command string (optional).
	Command string `json:"command,omitempty"`
}

// FileEdit represents a single file edit operation.
type FileEdit struct {
	// OldString is the text to replace.
	OldString string `json:"old_string"`
	// NewString is the replacement text.
	NewString string `json:"new_string"`
	// Range is the edit range (optional, for Tab edits).
	Range *EditRange `json:"range,omitempty"`
	// OldLine is the line before edit (optional, for Tab edits).
	OldLine string `json:"old_line,omitempty"`
	// NewLine is the line after edit (optional, for Tab edits).
	NewLine string `json:"new_line,omitempty"`
}

// EditRange represents the range of an edit.
type EditRange struct {
	StartLineNumber int `json:"start_line_number"`
	StartColumn     int `json:"start_column"`
	EndLineNumber   int `json:"end_line_number"`
	EndColumn       int `json:"end_column"`
}

// AfterFileEditPayload is the payload for afterFileEdit events.
type AfterFileEditPayload struct {
	// FilePath is the absolute path to the edited file.
	FilePath string `json:"file_path"`
	// Edits are the edit operations performed.
	Edits []FileEdit `json:"edits"`
}

// StopPayload is the payload for stop events.
type StopPayload struct {
	// Status is the completion status (completed, aborted, error).
	Status string `json:"status"`
	// LoopCount is the number of times the stop hook has triggered auto follow-ups.
	LoopCount int `json:"loop_count"`
}

// BeforeShellExecutionPayload is the payload for beforeShellExecution events.
type BeforeShellExecutionPayload struct {
	// Command is the full terminal command.
	Command string `json:"command"`
	// Cwd is the current working directory.
	Cwd string `json:"cwd"`
}

// AfterShellExecutionPayload is the payload for afterShellExecution events.
type AfterShellExecutionPayload struct {
	// Command is the full terminal command.
	Command string `json:"command"`
	// Output is the full terminal output.
	Output string `json:"output"`
}

// BeforeReadFilePayload is the payload for beforeReadFile events.
type BeforeReadFilePayload struct {
	// FilePath is the absolute path to the file.
	FilePath string `json:"file_path"`
	// Content is the file contents.
	Content string `json:"content"`
	// Attachments are any prompt attachments for auditing rules inclusion.
	Attachments []Attachment `json:"attachments"`
}

// ParseInput parses the hook input from a reader.
func ParseInput(r io.Reader) (*HookInput, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read input: %w", err)
	}

	var input HookInput
	if err := json.Unmarshal(data, &input); err != nil {
		return nil, fmt.Errorf("failed to parse input JSON: %w", err)
	}

	// Store the raw payload for later parsing of event-specific fields
	input.RawPayload = data

	return &input, nil
}

// ParsePayload parses the event-specific payload from the raw input.
func (h *HookInput) ParsePayload() (any, error) {
	switch h.HookEventName {
	case EventBeforeSubmitPrompt:
		var payload BeforeSubmitPromptPayload
		if err := json.Unmarshal(h.RawPayload, &payload); err != nil {
			return nil, fmt.Errorf("failed to parse beforeSubmitPrompt payload: %w", err)
		}
		return &payload, nil

	case EventAfterAgentResponse:
		var payload AfterAgentResponsePayload
		if err := json.Unmarshal(h.RawPayload, &payload); err != nil {
			return nil, fmt.Errorf("failed to parse afterAgentResponse payload: %w", err)
		}
		return &payload, nil

	case EventAfterAgentThought:
		var payload AfterAgentThoughtPayload
		if err := json.Unmarshal(h.RawPayload, &payload); err != nil {
			return nil, fmt.Errorf("failed to parse afterAgentThought payload: %w", err)
		}
		return &payload, nil

	case EventAfterMCPExecution:
		var payload AfterMCPExecutionPayload
		if err := json.Unmarshal(h.RawPayload, &payload); err != nil {
			return nil, fmt.Errorf("failed to parse afterMCPExecution payload: %w", err)
		}
		return &payload, nil

	case EventBeforeMCPExecution:
		var payload BeforeMCPExecutionPayload
		if err := json.Unmarshal(h.RawPayload, &payload); err != nil {
			return nil, fmt.Errorf("failed to parse beforeMCPExecution payload: %w", err)
		}
		return &payload, nil

	case EventAfterFileEdit:
		var payload AfterFileEditPayload
		if err := json.Unmarshal(h.RawPayload, &payload); err != nil {
			return nil, fmt.Errorf("failed to parse afterFileEdit payload: %w", err)
		}
		return &payload, nil

	case EventStop:
		var payload StopPayload
		if err := json.Unmarshal(h.RawPayload, &payload); err != nil {
			return nil, fmt.Errorf("failed to parse stop payload: %w", err)
		}
		return &payload, nil

	case EventBeforeShellExecution:
		var payload BeforeShellExecutionPayload
		if err := json.Unmarshal(h.RawPayload, &payload); err != nil {
			return nil, fmt.Errorf("failed to parse beforeShellExecution payload: %w", err)
		}
		return &payload, nil

	case EventAfterShellExecution:
		var payload AfterShellExecutionPayload
		if err := json.Unmarshal(h.RawPayload, &payload); err != nil {
			return nil, fmt.Errorf("failed to parse afterShellExecution payload: %w", err)
		}
		return &payload, nil

	case EventBeforeReadFile:
		var payload BeforeReadFilePayload
		if err := json.Unmarshal(h.RawPayload, &payload); err != nil {
			return nil, fmt.Errorf("failed to parse beforeReadFile payload: %w", err)
		}
		return &payload, nil

	default:
		return nil, fmt.Errorf("unknown hook event: %s", h.HookEventName)
	}
}

```

```go
package hook

import (
	"strings"

	"github.com/pkoukk/tiktoken-go"
)

// Tokenizer is the interface for counting tokens in text.
type Tokenizer interface {
	// CountTokens counts the number of tokens in the given text for the specified model.
	CountTokens(model, text string) (int64, error)
}

// TiktokenTokenizer uses tiktoken for accurate token counting.
type TiktokenTokenizer struct {
	defaultModel string
}

// NewTiktokenTokenizer creates a new tiktoken-based tokenizer.
func NewTiktokenTokenizer(defaultModel string) *TiktokenTokenizer {
	return &TiktokenTokenizer{
		defaultModel: defaultModel,
	}
}

// CountTokens counts tokens using tiktoken.
func (t *TiktokenTokenizer) CountTokens(model, text string) (int64, error) {
	if model == "" {
		model = t.defaultModel
	}

	// Map common model names to tiktoken encoding names
	encoding := mapModelToEncoding(model)

	enc, err := tiktoken.GetEncoding(encoding)
	if err != nil {
		// Fall back to cl100k_base if encoding not found
		enc, err = tiktoken.GetEncoding("cl100k_base")
		if err != nil {
			return 0, err
		}
	}

	tokens := enc.Encode(text, nil, nil)
	return int64(len(tokens)), nil
}

// HeuristicTokenizer uses a simple character-based heuristic for token counting.
type HeuristicTokenizer struct {
	// CharsPerToken is the average number of characters per token.
	// Default is 4, which is a reasonable approximation for English text.
	CharsPerToken int
}

// NewHeuristicTokenizer creates a new heuristic-based tokenizer.
func NewHeuristicTokenizer() *HeuristicTokenizer {
	return &HeuristicTokenizer{
		CharsPerToken: 4,
	}
}

// CountTokens counts tokens using a character-based heuristic.
func (t *HeuristicTokenizer) CountTokens(_, text string) (int64, error) {
	if t.CharsPerToken <= 0 {
		t.CharsPerToken = 4
	}
	charCount := len(text)
	tokenCount := charCount / t.CharsPerToken
	if charCount%t.CharsPerToken > 0 {
		tokenCount++
	}
	return int64(tokenCount), nil
}

// CompositeTokenizer tries tiktoken first, then falls back to heuristic.
type CompositeTokenizer struct {
	tiktoken  *TiktokenTokenizer
	heuristic *HeuristicTokenizer
}

// NewCompositeTokenizer creates a new composite tokenizer.
func NewCompositeTokenizer(defaultModel string) *CompositeTokenizer {
	return &CompositeTokenizer{
		tiktoken:  NewTiktokenTokenizer(defaultModel),
		heuristic: NewHeuristicTokenizer(),
	}
}

// CountTokens tries tiktoken first, falls back to heuristic on error.
func (t *CompositeTokenizer) CountTokens(model, text string) (int64, error) {
	count, err := t.tiktoken.CountTokens(model, text)
	if err != nil {
		// Fall back to heuristic
		return t.heuristic.CountTokens(model, text)
	}
	return count, nil
}

// mapModelToEncoding maps model names to tiktoken encoding names.
func mapModelToEncoding(model string) string {
	model = strings.ToLower(model)

	// GPT-4 and GPT-3.5 turbo models use cl100k_base
	if strings.Contains(model, "gpt-4") ||
		strings.Contains(model, "gpt-3.5") ||
		strings.Contains(model, "text-embedding") {
		return "cl100k_base"
	}

	// Claude models - use cl100k_base as approximation
	if strings.Contains(model, "claude") {
		return "cl100k_base"
	}

	// Anthropic models - use cl100k_base as approximation
	if strings.Contains(model, "anthropic") {
		return "cl100k_base"
	}

	// Gemini models - use cl100k_base as approximation
	if strings.Contains(model, "gemini") {
		return "cl100k_base"
	}

	// GPT-4o models use o200k_base
	if strings.Contains(model, "gpt-4o") || strings.Contains(model, "o1") {
		return "o200k_base"
	}

	// Default to cl100k_base for unknown models
	return "cl100k_base"
}

// CountChars returns the character count of the text.
func CountChars(text string) int64 {
	return int64(len(text))
}

```

```go
package hook

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"connectrpc.com/connect"
	apiv1 "github.com/ainslies/ctd/gen/api/proto/v1"
	"github.com/ainslies/ctd/gen/api/proto/v1/apiv1connect"
	"github.com/ainslies/ctd/internal/domain"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// CTDClient is a client for communicating with the ctd daemon.
type CTDClient struct {
	thoughtClient    apiv1connect.ThoughtServiceClient
	tokenUsageClient apiv1connect.TokenUsageServiceClient
	timeout          time.Duration
}

// NewCTDClient creates a new CTD client.
func NewCTDClient(serverURL string, timeout time.Duration) *CTDClient {
	httpClient := &http.Client{
		Timeout: timeout,
	}

	return &CTDClient{
		thoughtClient:    apiv1connect.NewThoughtServiceClient(httpClient, serverURL),
		tokenUsageClient: apiv1connect.NewTokenUsageServiceClient(httpClient, serverURL),
		timeout:          timeout,
	}
}

// CreateTokenUsage sends a token usage record to ctd.
func (c *CTDClient) CreateTokenUsage(ctx context.Context, usage *domain.TokenUsage) error {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	protoUsage := &apiv1.TokenUsage{
		ConversationId:   usage.ConversationID.String(),
		GenerationId:     usage.GenerationID.String(),
		Model:            usage.Model,
		Stage:            mapDomainStageToProto(usage.Stage),
		PromptTokens:     usage.PromptTokens,
		CompletionTokens: usage.CompletionTokens,
		TotalTokens:      usage.TotalTokens,
		PromptChars:      usage.PromptChars,
		CompletionChars:  usage.CompletionChars,
		WorkspaceRoots:   usage.WorkspaceRoots,
		UserEmail:        usage.UserEmail,
		HookEventName:    usage.HookEventName,
		Timestamp:        timestamppb.New(usage.Timestamp),
	}

	req := connect.NewRequest(&apiv1.CreateTokenUsageRequest{
		Usage: protoUsage,
	})

	_, err := c.tokenUsageClient.CreateTokenUsage(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to create token usage: %w", err)
	}

	return nil
}

// CreateThought sends a thought record to ctd.
func (c *CTDClient) CreateThought(ctx context.Context, thought *domain.Thought) error {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	protoThought := &apiv1.Thought{
		ConversationId: thought.ConversationID.String(),
		GenerationId:   thought.GenerationID.String(),
		Text:           thought.Text,
		DurationMs:     int32(thought.DurationMs),
		Model:          thought.Model,
		HookEventName:  thought.HookEventName,
		CursorVersion:  thought.CursorVersion,
		WorkspaceRoots: thought.WorkspaceRoots,
		UserEmail:      thought.UserEmail,
		Timestamp:      timestamppb.New(thought.Timestamp),
	}

	req := connect.NewRequest(&apiv1.CreateThoughtRequest{
		Thought: protoThought,
	})

	_, err := c.thoughtClient.CreateThought(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to create thought: %w", err)
	}

	return nil
}

// mapDomainStageToProto converts a domain TokenUsageStage to a proto TokenUsageStage.
func mapDomainStageToProto(stage domain.TokenUsageStage) apiv1.TokenUsageStage {
	switch stage {
	case domain.TokenUsageStagePrompt:
		return apiv1.TokenUsageStage_TOKEN_USAGE_STAGE_PROMPT
	case domain.TokenUsageStageCompletion:
		return apiv1.TokenUsageStage_TOKEN_USAGE_STAGE_COMPLETION
	case domain.TokenUsageStageTool:
		return apiv1.TokenUsageStage_TOKEN_USAGE_STAGE_TOOL
	case domain.TokenUsageStageSystem:
		return apiv1.TokenUsageStage_TOKEN_USAGE_STAGE_SYSTEM
	case domain.TokenUsageStageFileEdit:
		return apiv1.TokenUsageStage_TOKEN_USAGE_STAGE_FILE_EDIT
	case domain.TokenUsageStageStop:
		return apiv1.TokenUsageStage_TOKEN_USAGE_STAGE_STOP
	default:
		return apiv1.TokenUsageStage_TOKEN_USAGE_STAGE_UNSPECIFIED
	}
}

```

```go
package hook

import (
	"encoding/json"
	"io"
)

// HookOutput represents the output to be written to stdout.
type HookOutput struct {
	// Permission is the permission decision (allow, deny, ask).
	// Used for beforeShellExecution, beforeMCPExecution, beforeReadFile.
	Permission string `json:"permission,omitempty"`

	// Continue indicates whether to continue with the operation.
	// Used for beforeSubmitPrompt.
	Continue *bool `json:"continue,omitempty"`

	// UserMessage is a message shown to the user in the client.
	UserMessage string `json:"user_message,omitempty"`

	// AgentMessage is a message sent to the agent.
	AgentMessage string `json:"agent_message,omitempty"`

	// FollowupMessage is an optional follow-up message for stop events.
	FollowupMessage string `json:"followup_message,omitempty"`
}

// NewAllowOutput creates an output that allows the operation.
func NewAllowOutput() *HookOutput {
	return &HookOutput{
		Permission: "allow",
	}
}

// NewContinueOutput creates an output that continues the operation.
func NewContinueOutput() *HookOutput {
	cont := true
	return &HookOutput{
		Continue: &cont,
	}
}

// NewEmptyOutput creates an empty output for events that don't require a response.
func NewEmptyOutput() *HookOutput {
	return &HookOutput{}
}

// WriteOutput writes the hook output to the given writer as JSON.
func WriteOutput(w io.Writer, output *HookOutput) error {
	encoder := json.NewEncoder(w)
	return encoder.Encode(output)
}

```

```go
package hook

import (
	"context"
	"time"

	"github.com/ainslies/ctd/internal/domain"
	"github.com/ainslies/ctd/pkg/logger"
	"github.com/google/uuid"
)

// Handler processes a hook event and returns the appropriate output.
type Handler interface {
	// Handle processes the hook input and returns the output.
	Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error)
}

// HandlerDependencies contains the dependencies for handlers.
type HandlerDependencies struct {
	Client    *CTDClient
	Tokenizer Tokenizer
	Log       *logger.Logger
	Config    *Config
}

// Router routes hook events to the appropriate handler.
type Router struct {
	deps     *HandlerDependencies
	handlers map[HookEventName]Handler
}

// NewRouter creates a new router with all handlers registered.
func NewRouter(deps *HandlerDependencies) *Router {
	r := &Router{
		deps:     deps,
		handlers: make(map[HookEventName]Handler),
	}

	// Register handlers
	r.handlers[EventBeforeSubmitPrompt] = &BeforeSubmitPromptHandler{deps: deps}
	r.handlers[EventAfterAgentResponse] = &AfterAgentResponseHandler{deps: deps}
	r.handlers[EventAfterAgentThought] = &AfterAgentThoughtHandler{deps: deps}
	r.handlers[EventAfterMCPExecution] = &AfterMCPExecutionHandler{deps: deps}
	r.handlers[EventAfterFileEdit] = &AfterFileEditHandler{deps: deps}
	r.handlers[EventStop] = &StopHandler{deps: deps}
	r.handlers[EventBeforeShellExecution] = &BeforeShellExecutionHandler{deps: deps}
	r.handlers[EventAfterShellExecution] = &AfterShellExecutionHandler{deps: deps}
	r.handlers[EventBeforeMCPExecution] = &BeforeMCPExecutionHandler{deps: deps}
	r.handlers[EventBeforeReadFile] = &BeforeReadFileHandler{deps: deps}

	return r
}

// Route routes the input to the appropriate handler.
func (r *Router) Route(ctx context.Context, input *HookInput) (*HookOutput, error) {
	handler, ok := r.handlers[input.HookEventName]
	if !ok {
		// Unknown event, return empty output
		r.deps.Log.Warn("unknown hook event", "event", input.HookEventName)
		return NewEmptyOutput(), nil
	}

	payload, err := input.ParsePayload()
	if err != nil {
		r.deps.Log.Error("failed to parse payload", "error", err, "event", input.HookEventName)
		// Return appropriate default output even on parse error
		return getDefaultOutput(input.HookEventName), nil
	}

	return handler.Handle(ctx, input, payload)
}

// getDefaultOutput returns the default output for an event type.
func getDefaultOutput(event HookEventName) *HookOutput {
	switch event {
	case EventBeforeSubmitPrompt:
		return NewContinueOutput()
	case EventBeforeShellExecution, EventBeforeMCPExecution, EventBeforeReadFile:
		return NewAllowOutput()
	default:
		return NewEmptyOutput()
	}
}

// BeforeSubmitPromptHandler handles beforeSubmitPrompt events.
type BeforeSubmitPromptHandler struct {
	deps *HandlerDependencies
}

// Handle processes the beforeSubmitPrompt event.
func (h *BeforeSubmitPromptHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	p, ok := payload.(*BeforeSubmitPromptPayload)
	if !ok {
		return NewContinueOutput(), nil
	}

	// Count tokens
	tokenCount, err := h.deps.Tokenizer.CountTokens(h.deps.Config.DefaultModel, p.Prompt)
	if err != nil {
		h.deps.Log.Error("failed to count tokens", "error", err)
		tokenCount = 0
	}

	charCount := CountChars(p.Prompt)

	h.deps.Log.Debug("prompt metrics",
		"tokens", tokenCount,
		"chars", charCount,
		"conversation_id", input.ConversationID,
	)

	// Create token usage record
	usage := &domain.TokenUsage{
		ID:               uuid.New(),
		ConversationID:   parseUUID(input.ConversationID),
		GenerationID:     parseUUID(input.GenerationID),
		Model:            h.deps.Config.DefaultModel,
		Stage:            domain.TokenUsageStagePrompt,
		PromptTokens:     tokenCount,
		CompletionTokens: 0,
		TotalTokens:      tokenCount,
		PromptChars:      charCount,
		CompletionChars:  0,
		WorkspaceRoots:   input.WorkspaceRoots,
		UserEmail:        "",
		HookEventName:    string(input.HookEventName),
		Timestamp:        time.Now(),
	}

	// Send to ctd (best-effort)
	if err := h.deps.Client.CreateTokenUsage(ctx, usage); err != nil {
		h.deps.Log.Error("failed to send token usage", "error", err)
	} else {
		h.deps.Log.Info("token usage recorded", "stage", "prompt", "tokens", tokenCount)
	}

	return NewContinueOutput(), nil
}

// AfterAgentResponseHandler handles afterAgentResponse events.
type AfterAgentResponseHandler struct {
	deps *HandlerDependencies
}

// Handle processes the afterAgentResponse event.
func (h *AfterAgentResponseHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	p, ok := payload.(*AfterAgentResponsePayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	// Count tokens
	tokenCount, err := h.deps.Tokenizer.CountTokens(h.deps.Config.DefaultModel, p.Text)
	if err != nil {
		h.deps.Log.Error("failed to count tokens", "error", err)
		tokenCount = 0
	}

	charCount := CountChars(p.Text)

	h.deps.Log.Debug("completion metrics",
		"tokens", tokenCount,
		"chars", charCount,
		"conversation_id", input.ConversationID,
	)

	// Create token usage record
	usage := &domain.TokenUsage{
		ID:               uuid.New(),
		ConversationID:   parseUUID(input.ConversationID),
		GenerationID:     parseUUID(input.GenerationID),
		Model:            h.deps.Config.DefaultModel,
		Stage:            domain.TokenUsageStageCompletion,
		PromptTokens:     0,
		CompletionTokens: tokenCount,
		TotalTokens:      tokenCount,
		PromptChars:      0,
		CompletionChars:  charCount,
		WorkspaceRoots:   input.WorkspaceRoots,
		UserEmail:        "",
		HookEventName:    string(input.HookEventName),
		Timestamp:        time.Now(),
	}

	// Send token usage to ctd (best-effort)
	if err := h.deps.Client.CreateTokenUsage(ctx, usage); err != nil {
		h.deps.Log.Error("failed to send token usage", "error", err)
	} else {
		h.deps.Log.Info("token usage recorded", "stage", "completion", "tokens", tokenCount)
	}

	// Also send as thought for visibility
	thought := &domain.Thought{
		ID:             uuid.New(),
		ConversationID: parseUUID(input.ConversationID),
		GenerationID:   parseUUID(input.GenerationID),
		Text:           p.Text,
		DurationMs:     0,
		Model:          h.deps.Config.DefaultModel,
		HookEventName:  string(input.HookEventName),
		CursorVersion:  "",
		WorkspaceRoots: input.WorkspaceRoots,
		UserEmail:      "",
		Timestamp:      time.Now(),
	}

	if err := h.deps.Client.CreateThought(ctx, thought); err != nil {
		h.deps.Log.Error("failed to send thought", "error", err)
	}

	return NewEmptyOutput(), nil
}

// AfterAgentThoughtHandler handles afterAgentThought events.
type AfterAgentThoughtHandler struct {
	deps *HandlerDependencies
}

// Handle processes the afterAgentThought event.
func (h *AfterAgentThoughtHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	p, ok := payload.(*AfterAgentThoughtPayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	// Send as thought for visibility
	thought := &domain.Thought{
		ID:             uuid.New(),
		ConversationID: parseUUID(input.ConversationID),
		GenerationID:   parseUUID(input.GenerationID),
		Text:           p.Text,
		DurationMs:     0,
		Model:          h.deps.Config.DefaultModel,
		HookEventName:  string(input.HookEventName),
		CursorVersion:  "",
		WorkspaceRoots: input.WorkspaceRoots,
		UserEmail:      "",
		Timestamp:      time.Now(),
	}

	if err := h.deps.Client.CreateThought(ctx, thought); err != nil {
		h.deps.Log.Error("failed to send thought", "error", err)
	} else {
		h.deps.Log.Info("thought recorded", "event", input.HookEventName)
	}

	return NewEmptyOutput(), nil
}

// AfterMCPExecutionHandler handles afterMCPExecution events.
type AfterMCPExecutionHandler struct {
	deps *HandlerDependencies
}

// Handle processes the afterMCPExecution event.
func (h *AfterMCPExecutionHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	p, ok := payload.(*AfterMCPExecutionPayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	// Count tokens in tool input and result
	inputTokens, _ := h.deps.Tokenizer.CountTokens(h.deps.Config.DefaultModel, string(p.ToolInput))
	resultTokens, _ := h.deps.Tokenizer.CountTokens(h.deps.Config.DefaultModel, string(p.ResultJSON))
	totalTokens := inputTokens + resultTokens

	inputChars := CountChars(string(p.ToolInput))
	resultChars := CountChars(string(p.ResultJSON))

	h.deps.Log.Debug("tool metrics",
		"tool", p.ToolName,
		"input_tokens", inputTokens,
		"result_tokens", resultTokens,
		"conversation_id", input.ConversationID,
	)

	// Create token usage record
	usage := &domain.TokenUsage{
		ID:               uuid.New(),
		ConversationID:   parseUUID(input.ConversationID),
		GenerationID:     parseUUID(input.GenerationID),
		Model:            h.deps.Config.DefaultModel,
		Stage:            domain.TokenUsageStageTool,
		PromptTokens:     inputTokens,
		CompletionTokens: resultTokens,
		TotalTokens:      totalTokens,
		PromptChars:      inputChars,
		CompletionChars:  resultChars,
		WorkspaceRoots:   input.WorkspaceRoots,
		UserEmail:        "",
		HookEventName:    string(input.HookEventName),
		Timestamp:        time.Now(),
	}

	if err := h.deps.Client.CreateTokenUsage(ctx, usage); err != nil {
		h.deps.Log.Error("failed to send token usage", "error", err)
	} else {
		h.deps.Log.Info("token usage recorded", "stage", "tool", "tool", p.ToolName, "tokens", totalTokens)
	}

	return NewEmptyOutput(), nil
}

// AfterFileEditHandler handles afterFileEdit events.
type AfterFileEditHandler struct {
	deps *HandlerDependencies
}

// Handle processes the afterFileEdit event.
func (h *AfterFileEditHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	p, ok := payload.(*AfterFileEditPayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	// Count characters in edits
	var oldChars, newChars int64
	for _, edit := range p.Edits {
		oldChars += CountChars(edit.OldString)
		newChars += CountChars(edit.NewString)
	}

	h.deps.Log.Debug("file edit metrics",
		"file", p.FilePath,
		"old_chars", oldChars,
		"new_chars", newChars,
		"edit_count", len(p.Edits),
	)

	// Create token usage record for file edits
	usage := &domain.TokenUsage{
		ID:               uuid.New(),
		ConversationID:   parseUUID(input.ConversationID),
		GenerationID:     parseUUID(input.GenerationID),
		Model:            h.deps.Config.DefaultModel,
		Stage:            domain.TokenUsageStageFileEdit,
		PromptTokens:     0,
		CompletionTokens: 0,
		TotalTokens:      0,
		PromptChars:      oldChars,
		CompletionChars:  newChars,
		WorkspaceRoots:   input.WorkspaceRoots,
		UserEmail:        "",
		HookEventName:    string(input.HookEventName),
		Timestamp:        time.Now(),
	}

	if err := h.deps.Client.CreateTokenUsage(ctx, usage); err != nil {
		h.deps.Log.Error("failed to send token usage", "error", err)
	} else {
		h.deps.Log.Info("file edit recorded", "file", p.FilePath, "edits", len(p.Edits))
	}

	return NewEmptyOutput(), nil
}

// StopHandler handles stop events.
type StopHandler struct {
	deps *HandlerDependencies
}

// Handle processes the stop event.
func (h *StopHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	p, ok := payload.(*StopPayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	h.deps.Log.Info("agent loop stopped",
		"status", p.Status,
		"loop_count", p.LoopCount,
		"conversation_id", input.ConversationID,
	)

	// Create token usage record as session marker
	usage := &domain.TokenUsage{
		ID:               uuid.New(),
		ConversationID:   parseUUID(input.ConversationID),
		GenerationID:     parseUUID(input.GenerationID),
		Model:            h.deps.Config.DefaultModel,
		Stage:            domain.TokenUsageStageStop,
		PromptTokens:     0,
		CompletionTokens: 0,
		TotalTokens:      0,
		PromptChars:      0,
		CompletionChars:  0,
		WorkspaceRoots:   input.WorkspaceRoots,
		UserEmail:        "",
		HookEventName:    string(input.HookEventName),
		Timestamp:        time.Now(),
	}

	if err := h.deps.Client.CreateTokenUsage(ctx, usage); err != nil {
		h.deps.Log.Error("failed to send stop marker", "error", err)
	}

	return NewEmptyOutput(), nil
}

// BeforeShellExecutionHandler handles beforeShellExecution events.
type BeforeShellExecutionHandler struct {
	deps *HandlerDependencies
}

// Handle processes the beforeShellExecution event.
func (h *BeforeShellExecutionHandler) Handle(_ context.Context, input *HookInput, payload any) (*HookOutput, error) {
	p, ok := payload.(*BeforeShellExecutionPayload)
	if !ok {
		return NewAllowOutput(), nil
	}

	h.deps.Log.Debug("shell execution",
		"command", p.Command,
		"cwd", p.Cwd,
		"conversation_id", input.ConversationID,
	)

	// Always allow shell execution (observability only)
	return NewAllowOutput(), nil
}

// AfterShellExecutionHandler handles afterShellExecution events.
type AfterShellExecutionHandler struct {
	deps *HandlerDependencies
}

// Handle processes the afterShellExecution event.
func (h *AfterShellExecutionHandler) Handle(_ context.Context, input *HookInput, payload any) (*HookOutput, error) {
	p, ok := payload.(*AfterShellExecutionPayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	h.deps.Log.Debug("shell execution completed",
		"command", p.Command,
		"output_length", len(p.Output),
		"conversation_id", input.ConversationID,
	)

	return NewEmptyOutput(), nil
}

// BeforeMCPExecutionHandler handles beforeMCPExecution events.
type BeforeMCPExecutionHandler struct {
	deps *HandlerDependencies
}

// Handle processes the beforeMCPExecution event.
func (h *BeforeMCPExecutionHandler) Handle(_ context.Context, input *HookInput, payload any) (*HookOutput, error) {
	p, ok := payload.(*BeforeMCPExecutionPayload)
	if !ok {
		return NewAllowOutput(), nil
	}

	h.deps.Log.Debug("MCP execution",
		"tool", p.ToolName,
		"conversation_id", input.ConversationID,
	)

	// Always allow MCP execution (observability only)
	return NewAllowOutput(), nil
}

// BeforeReadFileHandler handles beforeReadFile events.
type BeforeReadFileHandler struct {
	deps *HandlerDependencies
}

// Handle processes the beforeReadFile event.
func (h *BeforeReadFileHandler) Handle(_ context.Context, input *HookInput, payload any) (*HookOutput, error) {
	p, ok := payload.(*BeforeReadFilePayload)
	if !ok {
		return NewAllowOutput(), nil
	}

	h.deps.Log.Debug("file read",
		"file", p.FilePath,
		"content_length", len(p.Content),
		"conversation_id", input.ConversationID,
	)

	// Always allow file read (observability only)
	return NewAllowOutput(), nil
}

// parseUUID parses a UUID string, returning a zero UUID on error.
func parseUUID(s string) uuid.UUID {
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil
	}
	return id
}

```

```go
// Package main provides the entry point for the Cursor hook binary.
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/ainslies/ctd/internal/hook"
	"github.com/ainslies/ctd/pkg/logger"
)

func main() {
	// Always exit 0 to avoid blocking Cursor
	// Errors are logged but don't prevent agent from proceeding
	if err := run(); err != nil {
		// Log to stderr (stdout is reserved for hook output)
		fmt.Fprintf(os.Stderr, "ctd-hook error: %v\n", err)
	}
	os.Exit(0)
}

func run() error {
	// Initialize logger to stderr
	log := logger.NewWithOutput(os.Stderr)

	// Load configuration
	cfg, err := hook.LoadConfig()
	if err != nil {
		log.Error("failed to load config", "error", err)
		// Use defaults on config error
		cfg = &hook.Config{
			ServerURL:    hook.DefaultServerURL,
			LogLevel:     hook.DefaultLogLevel,
			DefaultModel: hook.DefaultModel,
			TimeoutMs:    hook.DefaultTimeout,
		}
	}

	if err := cfg.Validate(); err != nil {
		log.Error("invalid config", "error", err)
	}

	log.Debug("hook started",
		"server_url", cfg.ServerURL,
		"default_model", cfg.DefaultModel,
		"timeout_ms", cfg.TimeoutMs,
	)

	// Parse input from stdin
	input, err := hook.ParseInput(os.Stdin)
	if err != nil {
		log.Error("failed to parse input", "error", err)
		// Write default output and exit
		return writeDefaultOutput(os.Stdout)
	}

	log.Debug("received hook event",
		"event", input.HookEventName,
		"conversation_id", input.ConversationID,
		"generation_id", input.GenerationID,
	)

	// Create client with timeout
	timeout := time.Duration(cfg.TimeoutMs) * time.Millisecond
	client := hook.NewCTDClient(cfg.ServerURL, timeout)

	// Create tokenizer
	tokenizer := hook.NewCompositeTokenizer(cfg.DefaultModel)

	// Create handler dependencies
	deps := &hook.HandlerDependencies{
		Client:    client,
		Tokenizer: tokenizer,
		Log:       log,
		Config:    cfg,
	}

	// Create router and handle the event
	router := hook.NewRouter(deps)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	output, err := router.Route(ctx, input)
	if err != nil {
		log.Error("handler error", "error", err, "event", input.HookEventName)
		// Return default output on handler error
		return writeDefaultOutput(os.Stdout)
	}

	// Write output to stdout
	if err := hook.WriteOutput(os.Stdout, output); err != nil {
		log.Error("failed to write output", "error", err)
		return err
	}

	log.Debug("hook completed", "event", input.HookEventName)
	return nil
}

// writeDefaultOutput writes a safe default output for unknown or error cases.
func writeDefaultOutput(w *os.File) error {
	// Default to allowing operations to avoid blocking the agent
	output := hook.NewAllowOutput()
	return hook.WriteOutput(w, output)
}

```

Now I need to add the `NewWithOutput` function to the logger package:

```go
// Package logger provides structured logging capabilities using slog.
package logger

import (
	"context"
	"io"
	"log/slog"
	"os"

	"github.com/phsym/console-slog"
)

// Logger provides structured logging capabilities using slog.
type Logger struct {
	slog *slog.Logger
}

// New creates a new structured logger with console output to stderr.
func New() *Logger {
	handler := console.NewHandler(os.Stderr, &console.HandlerOptions{
		Level:      slog.LevelInfo,
		AddSource:  false,
		TimeFormat: "15:04:05",
	})

	return &Logger{
		slog: slog.New(handler),
	}
}

// NewWithOutput creates a new structured logger with console output to the specified writer.
func NewWithOutput(w io.Writer) *Logger {
	handler := console.NewHandler(w, &console.HandlerOptions{
		Level:      slog.LevelInfo,
		AddSource:  false,
		TimeFormat: "15:04:05",
	})

	return &Logger{
		slog: slog.New(handler),
	}
}

// NewWithLevel creates a new logger with a specific log level.
func NewWithLevel(level slog.Level) *Logger {
	handler := console.NewHandler(os.Stderr, &console.HandlerOptions{
		Level:      level,
		AddSource:  level == slog.LevelDebug, // Add source location for debug
		TimeFormat: "15:04:05",
	})

	return &Logger{
		slog: slog.New(handler),
	}
}

// With returns a new Logger with the given attributes added to every log entry.
func (l *Logger) With(args ...any) *Logger {
	return &Logger{
		slog: l.slog.With(args...),
	}
}

// WithGroup returns a new Logger that starts a group for all attributes.
func (l *Logger) WithGroup(name string) *Logger {
	return &Logger{
		slog: l.slog.WithGroup(name),
	}
}

// Info logs an informational message with structured key-value pairs.
func (l *Logger) Info(msg string, args ...any) {
	l.slog.Info(msg, args...)
}

// InfoContext logs an informational message with context and structured key-value pairs.
func (l *Logger) InfoContext(ctx context.Context, msg string, args ...any) {
	l.slog.InfoContext(ctx, msg, args...)
}

// Warn logs a warning message with structured key-value pairs.
func (l *Logger) Warn(msg string, args ...any) {
	l.slog.Warn(msg, args...)
}

// WarnContext logs a warning message with context and structured key-value pairs.
func (l *Logger) WarnContext(ctx context.Context, msg string, args ...any) {
	l.slog.WarnContext(ctx, msg, args...)
}

// Error logs an error message with structured key-value pairs.
func (l *Logger) Error(msg string, args ...any) {
	l.slog.Error(msg, args...)
}

// ErrorContext logs an error message with context and structured key-value pairs.
func (l *Logger) ErrorContext(ctx context.Context, msg string, args ...any) {
	l.slog.ErrorContext(ctx, msg, args...)
}

// Debug logs a debug message with structured key-value pairs.
func (l *Logger) Debug(msg string, args ...any) {
	l.slog.Debug(msg, args...)
}

// DebugContext logs a debug message with context and structured key-value pairs.
func (l *Logger) DebugContext(ctx context.Context, msg string, args ...any) {
	l.slog.DebugContext(ctx, msg, args...)
}

// Fatal logs an error message.
// Note: This does not exit the program. Callers should handle exit logic in main() if needed.
func (l *Logger) Fatal(msg string, args ...any) {
	l.slog.Error(msg, args...)
}

// Slog returns the underlying slog.Logger for advanced usage.
func (l *Logger) Slog() *slog.Logger {
	return l.slog
}

```

```makefile
.PHONY: help build build-hook build-all run test lint lint-daemon lint-extension clean docker-up docker-down install-deps install install-hook uninstall uninstall-hook

help: ## Display this help message
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

gen: ## Generate Go code from Proto
	@echo "Generating Go code..."
	@buf generate
	@echo "Generated Go code in gen/"

build: ## Build the daemon binary
	@echo "Building ctd..."
	@go build -o bin/ctd ./cmd/thought-daemon

build-hook: ## Build the hook binary
	@echo "Building ctd-hook..."
	@go build -o bin/ctd-hook ./cmd/ctd-hook

build-all: build build-hook ## Build both daemon and hook binaries

install: build ## Install ctd to ~/.local/bin (no sudo required)
	@echo "Installing ctd..."
	@mkdir -p ~/.local/bin
	@cp bin/ctd ~/.local/bin/
	@chmod +x ~/.local/bin/ctd
	@echo " Installed to ~/.local/bin/ctd"
	@echo ""
	@if echo $$PATH | grep -q "$$HOME/.local/bin"; then \
		echo " ~/.local/bin is in your PATH"; \
		echo "  You can now run: ctd"; \
	else \
		echo " ~/.local/bin is not in your PATH"; \
		echo "  Add it by running:"; \
		echo "    echo 'export PATH=\"\$$HOME/.local/bin:\$$PATH\"' >> ~/.bashrc"; \
		echo "    source ~/.bashrc"; \
		echo "  (or ~/.zshrc, ~/.profile, etc. depending on your shell)"; \
	fi
	@echo ""
	@mkdir -p ~/.config/ctd
	@if [ ! -f ~/.config/ctd/config.yaml ]; then \
		cp config.yaml.example ~/.config/ctd/config.yaml; \
		echo " Created default config at ~/.config/ctd/config.yaml"; \
	else \
		echo "ℹ Config already exists at ~/.config/ctd/config.yaml"; \
	fi

extension-install: ## Build and install the VS Code extension to Cursor
	@echo "Installing VS Code extension dependencies..."
	@cd vscode-extension && npm install
	@echo "Building VS Code extension..."
	@cd vscode-extension && npm run build
	@echo "Packaging VS Code extension..."
	@cd vscode-extension && ./node_modules/.bin/vsce package
	@echo "Installing extension to Cursor..."
	@version=$$(cd vscode-extension && node -p "require('./package.json').version"); \
	cursor --uninstall-extension cursor-user.cursor-thought-stream 2>/dev/null || true; \
	cursor --install-extension vscode-extension/cursor-thought-stream-$$version.vsix --force; \
	echo "✅ Extension v$$version installed successfully!"; \
	echo "🔄 Please reload Cursor window (Cmd+Shift+P -> 'Developer: Reload Window')"

uninstall: ## Remove ctd from ~/.local/bin
	@echo "Uninstalling ctd..."
	@rm -f ~/.local/bin/ctd
	@echo " Removed ~/.local/bin/ctd"
	@echo "ℹ Config remains at ~/.config/ctd/ (remove manually if desired)"

run: ## Run the daemon locally
	@go run ./cmd/thought-daemon

test: test-daemon test-extension ## Run all tests (daemon + extension)

test-all: test-unit test-integration test-e2e ## Run all tests including E2E

test-race: ## Run all tests with race detector
	@echo "Running all tests with race detector..."
	@go test -v -race ./...

test-unit: ## Run only unit tests (fast)
	@echo "Running unit tests..."
	@go test -v -race -short ./internal/... ./pkg/... ./cmd/...

test-unit-verbose: ## Run unit tests with verbose output
	@echo "Running unit tests (verbose)..."
	@go test -v -race -short ./internal/... ./pkg/... ./cmd/...

test-integration: ## Run integration tests (requires Docker)
	@echo "Running integration tests..."
	@TESTCONTAINERS_RYUK_DISABLED=true go test -v -race -run Integration ./test/integration/...

test-e2e: ## Run end-to-end tests
	@echo "Running E2E tests..."
	@TESTCONTAINERS_RYUK_DISABLED=true go test -v -race ./test/e2e/...

test-performance: ## Run performance tests
	@echo "Running performance tests..."
	@TESTCONTAINERS_RYUK_DISABLED=true go test -v -run=Test ./test/performance/...

test-bench: ## Run benchmark tests
	@echo "Running benchmarks..."
	@go test -bench=. -benchmem -run=^Benchmark ./test/performance/...

test-coverage: ## Run tests with coverage report
	@echo "Running tests with coverage..."
	@go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
	@go tool cover -func=coverage.out | grep total
	@echo ""
	@echo "Coverage report saved to coverage.out"
	@echo "Run 'make test-coverage-html' to view in browser"

test-coverage-html: test-coverage ## Generate and open HTML coverage report
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report saved to coverage.html"
	@open coverage.html || xdg-open coverage.html || echo "Open coverage.html in your browser"

test-verbose: ## Run tests with verbose output
	@go test -v -race ./...

test-one: ## Run a specific test (usage: make test-one TEST=TestName)
	@if [ -z "$(TEST)" ]; then \
		echo "Error: TEST variable not set"; \
		echo "Usage: make test-one TEST=TestName"; \
		exit 1; \
	fi
	@go test -v -race -run $(TEST) ./...

test-watch: ## Watch for changes and run tests (requires entr)
	@echo "Watching for changes (requires 'entr' - install with brew install entr)..."
	@find . -name '*.go' | entr -c make test-unit

test-daemon: ## Run daemon tests (unit + integration)
	@echo "Running daemon tests..."
	@make test-unit
	@make test-integration

test-extension: ## Run extension tests
	@echo "Running extension tests..."
	@cd vscode-extension && npm test

lint: lint-daemon lint-extension ## Run all linters (daemon + extension)

lint-daemon: ## Run Go linter
	@echo "Linting Go code..."
	@golangci-lint run --config .golangci.yml --timeout=5m ./...
	@echo "✓ Go linting passed"

lint-extension: ## Run TypeScript/React linters
	@echo "Linting TypeScript/React code..."
	@cd vscode-extension && npm run lint
	@echo "✓ Extension linting passed"

clean: ## Clean build artifacts
	@rm -rf bin/
	@echo "Cleaned build artifacts"

docker-up: ## Start PostgreSQL container
	@docker-compose up -d
	@echo "Waiting for PostgreSQL to be ready..."
	@sleep 3

docker-down: ## Stop PostgreSQL container
	@docker-compose down

install-deps: ## Install Go dependencies
	@go mod download
	@go mod tidy

.DEFAULT_GOAL := help

```

```makefile
.PHONY: help build build-hook build-all run test lint lint-daemon lint-extension clean docker-up docker-down install-deps install install-hook uninstall uninstall-hook

help: ## Display this help message
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

gen: ## Generate Go code from Proto
	@echo "Generating Go code..."
	@buf generate
	@echo "Generated Go code in gen/"

build: ## Build the daemon binary
	@echo "Building ctd..."
	@go build -o bin/ctd ./cmd/thought-daemon

build-hook: ## Build the hook binary
	@echo "Building ctd-hook..."
	@go build -o bin/ctd-hook ./cmd/ctd-hook

build-all: build build-hook ## Build both daemon and hook binaries

install: build ## Install ctd to ~/.local/bin (no sudo required)
	@echo "Installing ctd..."
	@mkdir -p ~/.local/bin
	@cp bin/ctd ~/.local/bin/
	@chmod +x ~/.local/bin/ctd
	@echo " Installed to ~/.local/bin/ctd"
	@echo ""
	@if echo $$PATH | grep -q "$$HOME/.local/bin"; then \
		echo " ~/.local/bin is in your PATH"; \
		echo "  You can now run: ctd"; \
	else \
		echo " ~/.local/bin is not in your PATH"; \
		echo "  Add it by running:"; \
		echo "    echo 'export PATH=\"\$$HOME/.local/bin:\$$PATH\"' >> ~/.bashrc"; \
		echo "    source ~/.bashrc"; \
		echo "  (or ~/.zshrc, ~/.profile, etc. depending on your shell)"; \
	fi
	@echo ""
	@mkdir -p ~/.config/ctd
	@if [ ! -f ~/.config/ctd/config.yaml ]; then \
		cp config.yaml.example ~/.config/ctd/config.yaml; \
		echo " Created default config at ~/.config/ctd/config.yaml"; \
	else \
		echo "ℹ Config already exists at ~/.config/ctd/config.yaml"; \
	fi

extension-install: ## Build and install the VS Code extension to Cursor
	@echo "Installing VS Code extension dependencies..."
	@cd vscode-extension && npm install
	@echo "Building VS Code extension..."
	@cd vscode-extension && npm run build
	@echo "Packaging VS Code extension..."
	@cd vscode-extension && ./node_modules/.bin/vsce package
	@echo "Installing extension to Cursor..."
	@version=$$(cd vscode-extension && node -p "require('./package.json').version"); \
	cursor --uninstall-extension cursor-user.cursor-thought-stream 2>/dev/null || true; \
	cursor --install-extension vscode-extension/cursor-thought-stream-$$version.vsix --force; \
	echo "✅ Extension v$$version installed successfully!"; \
	echo "🔄 Please reload Cursor window (Cmd+Shift+P -> 'Developer: Reload Window')"

install-hook: build-hook ## Install ctd-hook to ~/.cursor/hooks
	@echo "Installing ctd-hook..."
	@mkdir -p ~/.cursor/hooks
	@cp bin/ctd-hook ~/.cursor/hooks/
	@chmod +x ~/.cursor/hooks/ctd-hook
	@echo "[OK] Installed to ~/.cursor/hooks/ctd-hook"
	@echo ""
	@echo "Add the following to ~/.cursor/hooks.json to enable the hook:"
	@echo '{'
	@echo '  "version": 1,'
	@echo '  "hooks": {'
	@echo '    "beforeSubmitPrompt": [{ "command": "~/.cursor/hooks/ctd-hook" }],'
	@echo '    "afterAgentResponse": [{ "command": "~/.cursor/hooks/ctd-hook" }],'
	@echo '    "afterAgentThought": [{ "command": "~/.cursor/hooks/ctd-hook" }],'
	@echo '    "afterMCPExecution": [{ "command": "~/.cursor/hooks/ctd-hook" }],'
	@echo '    "afterFileEdit": [{ "command": "~/.cursor/hooks/ctd-hook" }],'
	@echo '    "stop": [{ "command": "~/.cursor/hooks/ctd-hook" }]'
	@echo '  }'
	@echo '}'

uninstall: ## Remove ctd from ~/.local/bin
	@echo "Uninstalling ctd..."
	@rm -f ~/.local/bin/ctd
	@echo "[OK] Removed ~/.local/bin/ctd"
	@echo "[INFO] Config remains at ~/.config/ctd/ (remove manually if desired)"

uninstall-hook: ## Remove ctd-hook from ~/.cursor/hooks
	@echo "Uninstalling ctd-hook..."
	@rm -f ~/.cursor/hooks/ctd-hook
	@echo "[OK] Removed ~/.cursor/hooks/ctd-hook"
	@echo "[INFO] Remember to update ~/.cursor/hooks.json to remove hook references"

run: ## Run the daemon locally
	@go run ./cmd/thought-daemon

test: test-daemon test-extension ## Run all tests (daemon + extension)

test-all: test-unit test-integration test-e2e ## Run all tests including E2E

test-race: ## Run all tests with race detector
	@echo "Running all tests with race detector..."
	@go test -v -race ./...

test-unit: ## Run only unit tests (fast)
	@echo "Running unit tests..."
	@go test -v -race -short ./internal/... ./pkg/... ./cmd/...

test-unit-verbose: ## Run unit tests with verbose output
	@echo "Running unit tests (verbose)..."
	@go test -v -race -short ./internal/... ./pkg/... ./cmd/...

test-integration: ## Run integration tests (requires Docker)
	@echo "Running integration tests..."
	@TESTCONTAINERS_RYUK_DISABLED=true go test -v -race -run Integration ./test/integration/...

test-e2e: ## Run end-to-end tests
	@echo "Running E2E tests..."
	@TESTCONTAINERS_RYUK_DISABLED=true go test -v -race ./test/e2e/...

test-performance: ## Run performance tests
	@echo "Running performance tests..."
	@TESTCONTAINERS_RYUK_DISABLED=true go test -v -run=Test ./test/performance/...

test-bench: ## Run benchmark tests
	@echo "Running benchmarks..."
	@go test -bench=. -benchmem -run=^Benchmark ./test/performance/...

test-coverage: ## Run tests with coverage report
	@echo "Running tests with coverage..."
	@go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
	@go tool cover -func=coverage.out | grep total
	@echo ""
	@echo "Coverage report saved to coverage.out"
	@echo "Run 'make test-coverage-html' to view in browser"

test-coverage-html: test-coverage ## Generate and open HTML coverage report
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report saved to coverage.html"
	@open coverage.html || xdg-open coverage.html || echo "Open coverage.html in your browser"

test-verbose: ## Run tests with verbose output
	@go test -v -race ./...

test-one: ## Run a specific test (usage: make test-one TEST=TestName)
	@if [ -z "$(TEST)" ]; then \
		echo "Error: TEST variable not set"; \
		echo "Usage: make test-one TEST=TestName"; \
		exit 1; \
	fi
	@go test -v -race -run $(TEST) ./...

test-watch: ## Watch for changes and run tests (requires entr)
	@echo "Watching for changes (requires 'entr' - install with brew install entr)..."
	@find . -name '*.go' | entr -c make test-unit

test-daemon: ## Run daemon tests (unit + integration)
	@echo "Running daemon tests..."
	@make test-unit
	@make test-integration

test-extension: ## Run extension tests
	@echo "Running extension tests..."
	@cd vscode-extension && npm test

lint: lint-daemon lint-extension ## Run all linters (daemon + extension)

lint-daemon: ## Run Go linter
	@echo "Linting Go code..."
	@golangci-lint run --config .golangci.yml --timeout=5m ./...
	@echo "✓ Go linting passed"

lint-extension: ## Run TypeScript/React linters
	@echo "Linting TypeScript/React code..."
	@cd vscode-extension && npm run lint
	@echo "✓ Extension linting passed"

clean: ## Clean build artifacts
	@rm -rf bin/
	@echo "Cleaned build artifacts"

docker-up: ## Start PostgreSQL container
	@docker-compose up -d
	@echo "Waiting for PostgreSQL to be ready..."
	@sleep 3

docker-down: ## Stop PostgreSQL container
	@docker-compose down

install-deps: ## Install Go dependencies
	@go mod download
	@go mod tidy

.DEFAULT_GOAL := help

```

```go
package server

import (
	"context"
	"errors"
	"testing"
	"time"

	"connectrpc.com/connect"
	apiv1 "github.com/ainslies/ctd/gen/api/proto/v1"
	"github.com/ainslies/ctd/internal/domain"
	"github.com/ainslies/ctd/pkg/logger"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// MockTokenUsageRepository is a mock implementation of TokenUsageRepository.
type MockTokenUsageRepository struct {
	mock.Mock
}

func (m *MockTokenUsageRepository) Store(ctx context.Context, usage *domain.TokenUsage) error {
	args := m.Called(ctx, usage)
	return args.Error(0)
}

func (m *MockTokenUsageRepository) StoreBatch(ctx context.Context, usages []*domain.TokenUsage) error {
	args := m.Called(ctx, usages)
	return args.Error(0)
}

func (m *MockTokenUsageRepository) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockTokenUsageRepository) HealthCheck(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func TestNewTokenUsageServer(t *testing.T) {
	repo := new(MockTokenUsageRepository)
	log := logger.New()

	srv := NewTokenUsageServer(repo, log)

	assert.NotNil(t, srv)
	assert.NotNil(t, srv.repo)
	assert.NotNil(t, srv.log)
}

func TestMapProtoToTokenUsage_Valid(t *testing.T) {
	conversationID := uuid.New()
	generationID := uuid.New()

	protoUsage := &apiv1.TokenUsage{
		ConversationId:   conversationID.String(),
		GenerationId:     generationID.String(),
		Model:            "gpt-4",
		Stage:            apiv1.TokenUsageStage_TOKEN_USAGE_STAGE_PROMPT,
		PromptTokens:     100,
		CompletionTokens: 0,
		TotalTokens:      100,
		PromptChars:      400,
		CompletionChars:  0,
		WorkspaceRoots:   []string{"/test"},
		UserEmail:        "test@example.com",
		HookEventName:    "beforeSubmitPrompt",
		Timestamp:        timestamppb.Now(),
	}

	usage, err := mapProtoToTokenUsage(protoUsage)

	require.NoError(t, err)
	assert.Equal(t, conversationID, usage.ConversationID)
	assert.Equal(t, generationID, usage.GenerationID)
	assert.Equal(t, "gpt-4", usage.Model)
	assert.Equal(t, domain.TokenUsageStagePrompt, usage.Stage)
	assert.Equal(t, int64(100), usage.PromptTokens)
	assert.Equal(t, int64(0), usage.CompletionTokens)
	assert.Equal(t, int64(100), usage.TotalTokens)
	assert.Equal(t, int64(400), usage.PromptChars)
	assert.Equal(t, int64(0), usage.CompletionChars)
	assert.Equal(t, []string{"/test"}, usage.WorkspaceRoots)
	assert.Equal(t, "test@example.com", usage.UserEmail)
	assert.Equal(t, "beforeSubmitPrompt", usage.HookEventName)
	assert.NotEqual(t, uuid.Nil, usage.ID)
}

func TestMapProtoToTokenUsage_InvalidConversationID(t *testing.T) {
	protoUsage := &apiv1.TokenUsage{
		ConversationId: "invalid-uuid",
		GenerationId:   uuid.New().String(),
	}

	_, err := mapProtoToTokenUsage(protoUsage)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid conversation_id")
}

func TestMapProtoToTokenUsage_InvalidGenerationID(t *testing.T) {
	protoUsage := &apiv1.TokenUsage{
		ConversationId: uuid.New().String(),
		GenerationId:   "invalid-uuid",
	}

	_, err := mapProtoToTokenUsage(protoUsage)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid generation_id")
}

func TestMapProtoToTokenUsage_MissingTimestamp(t *testing.T) {
	protoUsage := &apiv1.TokenUsage{
		ConversationId: uuid.New().String(),
		GenerationId:   uuid.New().String(),
		Timestamp:      nil,
	}

	usage, err := mapProtoToTokenUsage(protoUsage)

	require.NoError(t, err)
	assert.WithinDuration(t, time.Now(), usage.Timestamp, time.Second)
}

func TestMapProtoStageToTokenUsageStage(t *testing.T) {
	tests := []struct {
		name     string
		input    apiv1.TokenUsageStage
		expected domain.TokenUsageStage
	}{
		{"unspecified", apiv1.TokenUsageStage_TOKEN_USAGE_STAGE_UNSPECIFIED, domain.TokenUsageStageUnspecified},
		{"prompt", apiv1.TokenUsageStage_TOKEN_USAGE_STAGE_PROMPT, domain.TokenUsageStagePrompt},
		{"completion", apiv1.TokenUsageStage_TOKEN_USAGE_STAGE_COMPLETION, domain.TokenUsageStageCompletion},
		{"tool", apiv1.TokenUsageStage_TOKEN_USAGE_STAGE_TOOL, domain.TokenUsageStageTool},
		{"system", apiv1.TokenUsageStage_TOKEN_USAGE_STAGE_SYSTEM, domain.TokenUsageStageSystem},
		{"file_edit", apiv1.TokenUsageStage_TOKEN_USAGE_STAGE_FILE_EDIT, domain.TokenUsageStageFileEdit},
		{"stop", apiv1.TokenUsageStage_TOKEN_USAGE_STAGE_STOP, domain.TokenUsageStageStop},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mapProtoStageToTokenUsageStage(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMapDomainStageToProtoStage(t *testing.T) {
	tests := []struct {
		name     string
		input    domain.TokenUsageStage
		expected apiv1.TokenUsageStage
	}{
		{"unspecified", domain.TokenUsageStageUnspecified, apiv1.TokenUsageStage_TOKEN_USAGE_STAGE_UNSPECIFIED},
		{"prompt", domain.TokenUsageStagePrompt, apiv1.TokenUsageStage_TOKEN_USAGE_STAGE_PROMPT},
		{"completion", domain.TokenUsageStageCompletion, apiv1.TokenUsageStage_TOKEN_USAGE_STAGE_COMPLETION},
		{"tool", domain.TokenUsageStageTool, apiv1.TokenUsageStage_TOKEN_USAGE_STAGE_TOOL},
		{"system", domain.TokenUsageStageSystem, apiv1.TokenUsageStage_TOKEN_USAGE_STAGE_SYSTEM},
		{"file_edit", domain.TokenUsageStageFileEdit, apiv1.TokenUsageStage_TOKEN_USAGE_STAGE_FILE_EDIT},
		{"stop", domain.TokenUsageStageStop, apiv1.TokenUsageStage_TOKEN_USAGE_STAGE_STOP},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mapDomainStageToProtoStage(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTokenUsageServer_CreateTokenUsage_Success(t *testing.T) {
	repo := new(MockTokenUsageRepository)
	log := logger.New()
	srv := NewTokenUsageServer(repo, log)

	repo.On("Store", mock.Anything, mock.AnythingOfType("*domain.TokenUsage")).Return(nil)

	req := connect.NewRequest(&apiv1.CreateTokenUsageRequest{
		Usage: &apiv1.TokenUsage{
			ConversationId:   uuid.New().String(),
			GenerationId:     uuid.New().String(),
			Model:            "gpt-4",
			Stage:            apiv1.TokenUsageStage_TOKEN_USAGE_STAGE_PROMPT,
			PromptTokens:     100,
			CompletionTokens: 0,
			TotalTokens:      100,
		},
	})

	res, err := srv.CreateTokenUsage(context.Background(), req)

	require.NoError(t, err)
	assert.NotEmpty(t, res.Msg.GetId())
	repo.AssertExpectations(t)
}

func TestTokenUsageServer_CreateTokenUsage_MissingUsage(t *testing.T) {
	repo := new(MockTokenUsageRepository)
	log := logger.New()
	srv := NewTokenUsageServer(repo, log)

	req := connect.NewRequest(&apiv1.CreateTokenUsageRequest{
		Usage: nil,
	})

	_, err := srv.CreateTokenUsage(context.Background(), req)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "usage is required")
}

func TestTokenUsageServer_CreateTokenUsage_InvalidUUID(t *testing.T) {
	repo := new(MockTokenUsageRepository)
	log := logger.New()
	srv := NewTokenUsageServer(repo, log)

	req := connect.NewRequest(&apiv1.CreateTokenUsageRequest{
		Usage: &apiv1.TokenUsage{
			ConversationId: "invalid-uuid",
			GenerationId:   uuid.New().String(),
		},
	})

	_, err := srv.CreateTokenUsage(context.Background(), req)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid token usage data")
}

func TestTokenUsageServer_CreateTokenUsage_NegativeTokens(t *testing.T) {
	repo := new(MockTokenUsageRepository)
	log := logger.New()
	srv := NewTokenUsageServer(repo, log)

	req := connect.NewRequest(&apiv1.CreateTokenUsageRequest{
		Usage: &apiv1.TokenUsage{
			ConversationId: uuid.New().String(),
			GenerationId:   uuid.New().String(),
			PromptTokens:   -100,
		},
	})

	_, err := srv.CreateTokenUsage(context.Background(), req)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "token counts must be non-negative")
}

func TestTokenUsageServer_CreateTokenUsage_RepositoryError(t *testing.T) {
	repo := new(MockTokenUsageRepository)
	log := logger.New()
	srv := NewTokenUsageServer(repo, log)

	repo.On("Store", mock.Anything, mock.AnythingOfType("*domain.TokenUsage")).Return(errors.New("database error"))

	req := connect.NewRequest(&apiv1.CreateTokenUsageRequest{
		Usage: &apiv1.TokenUsage{
			ConversationId: uuid.New().String(),
			GenerationId:   uuid.New().String(),
			PromptTokens:   100,
		},
	})

	_, err := srv.CreateTokenUsage(context.Background(), req)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to store token usage")
	repo.AssertExpectations(t)
}

```

```go
package repository

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPostgresTokenUsageRepository_InsertQuery(t *testing.T) {
	// Verify the insert query has the correct number of placeholders
	query := insertTokenUsageQuery

	// Count placeholders
	placeholderCount := 0
	for _, c := range query {
		if c == '$' {
			placeholderCount++
		}
	}

	// Should have 14 placeholders for 14 columns
	assert.Equal(t, 14, placeholderCount, "insert query should have 14 placeholders")
}

```

```go
package hook

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadConfig_Defaults(t *testing.T) {
	cfg, err := LoadConfig()

	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, DefaultServerURL, cfg.ServerURL)
	assert.Equal(t, DefaultLogLevel, cfg.LogLevel)
	assert.Equal(t, DefaultModel, cfg.DefaultModel)
	assert.Equal(t, DefaultTimeout, cfg.TimeoutMs)
}

func TestConfig_Validate_Defaults(t *testing.T) {
	cfg := &Config{}

	err := cfg.Validate()

	assert.NoError(t, err)
	assert.Equal(t, DefaultServerURL, cfg.ServerURL)
	assert.Equal(t, DefaultLogLevel, cfg.LogLevel)
	assert.Equal(t, DefaultModel, cfg.DefaultModel)
	assert.Equal(t, DefaultTimeout, cfg.TimeoutMs)
}

func TestConfig_Validate_PreservesValidValues(t *testing.T) {
	cfg := &Config{
		ServerURL:    "http://localhost:8080",
		LogLevel:     "debug",
		DefaultModel: "claude-3-opus",
		TimeoutMs:    1000,
	}

	err := cfg.Validate()

	assert.NoError(t, err)
	assert.Equal(t, "http://localhost:8080", cfg.ServerURL)
	assert.Equal(t, "debug", cfg.LogLevel)
	assert.Equal(t, "claude-3-opus", cfg.DefaultModel)
	assert.Equal(t, 1000, cfg.TimeoutMs)
}

func TestConfig_Validate_FixesInvalidTimeout(t *testing.T) {
	cfg := &Config{
		ServerURL:    "http://localhost:8080",
		LogLevel:     "info",
		DefaultModel: "gpt-4",
		TimeoutMs:    -100,
	}

	err := cfg.Validate()

	assert.NoError(t, err)
	assert.Equal(t, DefaultTimeout, cfg.TimeoutMs)
}

```

```go
package hook

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseInput_Valid(t *testing.T) {
	input := `{
		"conversation_id": "123e4567-e89b-12d3-a456-426614174000",
		"generation_id": "223e4567-e89b-12d3-a456-426614174000",
		"hook_event_name": "beforeSubmitPrompt",
		"workspace_roots": ["/test/workspace"],
		"prompt": "Hello, world!"
	}`

	result, err := ParseInput(strings.NewReader(input))

	require.NoError(t, err)
	assert.Equal(t, "123e4567-e89b-12d3-a456-426614174000", result.ConversationID)
	assert.Equal(t, "223e4567-e89b-12d3-a456-426614174000", result.GenerationID)
	assert.Equal(t, EventBeforeSubmitPrompt, result.HookEventName)
	assert.Equal(t, []string{"/test/workspace"}, result.WorkspaceRoots)
}

func TestParseInput_InvalidJSON(t *testing.T) {
	input := `{invalid json}`

	_, err := ParseInput(strings.NewReader(input))

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse input JSON")
}

func TestParseInput_EmptyInput(t *testing.T) {
	input := ``

	_, err := ParseInput(strings.NewReader(input))

	require.Error(t, err)
}

func TestHookInput_ParsePayload_BeforeSubmitPrompt(t *testing.T) {
	input := &HookInput{
		HookEventName: EventBeforeSubmitPrompt,
		RawPayload:    []byte(`{"prompt": "test prompt", "attachments": []}`),
	}

	payload, err := input.ParsePayload()

	require.NoError(t, err)
	p, ok := payload.(*BeforeSubmitPromptPayload)
	require.True(t, ok)
	assert.Equal(t, "test prompt", p.Prompt)
}

func TestHookInput_ParsePayload_AfterAgentResponse(t *testing.T) {
	input := &HookInput{
		HookEventName: EventAfterAgentResponse,
		RawPayload:    []byte(`{"text": "response text"}`),
	}

	payload, err := input.ParsePayload()

	require.NoError(t, err)
	p, ok := payload.(*AfterAgentResponsePayload)
	require.True(t, ok)
	assert.Equal(t, "response text", p.Text)
}

func TestHookInput_ParsePayload_AfterAgentThought(t *testing.T) {
	input := &HookInput{
		HookEventName: EventAfterAgentThought,
		RawPayload:    []byte(`{"text": "thought text"}`),
	}

	payload, err := input.ParsePayload()

	require.NoError(t, err)
	p, ok := payload.(*AfterAgentThoughtPayload)
	require.True(t, ok)
	assert.Equal(t, "thought text", p.Text)
}

func TestHookInput_ParsePayload_AfterMCPExecution(t *testing.T) {
	input := &HookInput{
		HookEventName: EventAfterMCPExecution,
		RawPayload:    []byte(`{"tool_name": "test_tool", "tool_input": {}, "result_json": {}}`),
	}

	payload, err := input.ParsePayload()

	require.NoError(t, err)
	p, ok := payload.(*AfterMCPExecutionPayload)
	require.True(t, ok)
	assert.Equal(t, "test_tool", p.ToolName)
}

func TestHookInput_ParsePayload_AfterFileEdit(t *testing.T) {
	input := &HookInput{
		HookEventName: EventAfterFileEdit,
		RawPayload:    []byte(`{"file_path": "/test/file.go", "edits": [{"old_string": "old", "new_string": "new"}]}`),
	}

	payload, err := input.ParsePayload()

	require.NoError(t, err)
	p, ok := payload.(*AfterFileEditPayload)
	require.True(t, ok)
	assert.Equal(t, "/test/file.go", p.FilePath)
	assert.Len(t, p.Edits, 1)
	assert.Equal(t, "old", p.Edits[0].OldString)
	assert.Equal(t, "new", p.Edits[0].NewString)
}

func TestHookInput_ParsePayload_Stop(t *testing.T) {
	input := &HookInput{
		HookEventName: EventStop,
		RawPayload:    []byte(`{"status": "completed", "loop_count": 5}`),
	}

	payload, err := input.ParsePayload()

	require.NoError(t, err)
	p, ok := payload.(*StopPayload)
	require.True(t, ok)
	assert.Equal(t, "completed", p.Status)
	assert.Equal(t, 5, p.LoopCount)
}

func TestHookInput_ParsePayload_UnknownEvent(t *testing.T) {
	input := &HookInput{
		HookEventName: "unknownEvent",
		RawPayload:    []byte(`{}`),
	}

	_, err := input.ParsePayload()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown hook event")
}

func TestHookInput_ParsePayload_InvalidJSON(t *testing.T) {
	input := &HookInput{
		HookEventName: EventBeforeSubmitPrompt,
		RawPayload:    []byte(`{invalid}`),
	}

	_, err := input.ParsePayload()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse")
}

```

```go
package hook

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHeuristicTokenizer_CountTokens(t *testing.T) {
	tokenizer := NewHeuristicTokenizer()

	tests := []struct {
		name     string
		text     string
		expected int64
	}{
		{"empty string", "", 0},
		{"short text", "hi", 1},
		{"exact multiple", "hello world", 3},    // 11 chars / 4 = 2.75, rounded up to 3
		{"longer text", "hello world test", 4},  // 16 chars / 4 = 4
		{"with spaces", "a b c d e f g h", 4},   // 15 chars / 4 = 3.75, rounded up to 4
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count, err := tokenizer.CountTokens("", tt.text)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, count)
		})
	}
}

func TestHeuristicTokenizer_ZeroCharsPerToken(t *testing.T) {
	tokenizer := &HeuristicTokenizer{CharsPerToken: 0}

	count, err := tokenizer.CountTokens("", "test")

	require.NoError(t, err)
	assert.Equal(t, int64(1), count) // Falls back to 4 chars per token
}

func TestTiktokenTokenizer_CountTokens(t *testing.T) {
	tokenizer := NewTiktokenTokenizer("gpt-4")

	// Test with a simple string
	count, err := tokenizer.CountTokens("gpt-4", "Hello, world!")

	require.NoError(t, err)
	assert.Greater(t, count, int64(0))
}

func TestTiktokenTokenizer_EmptyModel(t *testing.T) {
	tokenizer := NewTiktokenTokenizer("gpt-4")

	// Should use default model when empty
	count, err := tokenizer.CountTokens("", "Hello, world!")

	require.NoError(t, err)
	assert.Greater(t, count, int64(0))
}

func TestCompositeTokenizer_CountTokens(t *testing.T) {
	tokenizer := NewCompositeTokenizer("gpt-4")

	count, err := tokenizer.CountTokens("gpt-4", "Hello, world!")

	require.NoError(t, err)
	assert.Greater(t, count, int64(0))
}

func TestMapModelToEncoding(t *testing.T) {
	tests := []struct {
		name     string
		model    string
		expected string
	}{
		{"gpt-4", "gpt-4", "cl100k_base"},
		{"gpt-4-turbo", "gpt-4-turbo", "cl100k_base"},
		{"gpt-3.5-turbo", "gpt-3.5-turbo", "cl100k_base"},
		{"claude", "claude-3-opus", "cl100k_base"},
		{"gemini", "gemini-pro", "cl100k_base"},
		{"gpt-4o", "gpt-4o", "o200k_base"},
		{"o1", "o1-preview", "o200k_base"},
		{"unknown", "some-unknown-model", "cl100k_base"},
		{"uppercase", "GPT-4", "cl100k_base"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mapModelToEncoding(tt.model)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCountChars(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected int64
	}{
		{"empty", "", 0},
		{"single char", "a", 1},
		{"multi char", "hello", 5},
		{"with spaces", "hello world", 11},
		{"unicode", "hello", 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CountChars(tt.text)
			assert.Equal(t, tt.expected, result)
		})
	}
}

```

```go
package hook

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAllowOutput(t *testing.T) {
	output := NewAllowOutput()

	assert.Equal(t, "allow", output.Permission)
	assert.Nil(t, output.Continue)
}

func TestNewContinueOutput(t *testing.T) {
	output := NewContinueOutput()

	assert.NotNil(t, output.Continue)
	assert.True(t, *output.Continue)
	assert.Empty(t, output.Permission)
}

func TestNewEmptyOutput(t *testing.T) {
	output := NewEmptyOutput()

	assert.Empty(t, output.Permission)
	assert.Nil(t, output.Continue)
	assert.Empty(t, output.UserMessage)
	assert.Empty(t, output.AgentMessage)
}

func TestWriteOutput_Allow(t *testing.T) {
	output := NewAllowOutput()
	var buf bytes.Buffer

	err := WriteOutput(&buf, output)

	require.NoError(t, err)

	var result map[string]any
	err = json.Unmarshal(buf.Bytes(), &result)
	require.NoError(t, err)
	assert.Equal(t, "allow", result["permission"])
}

func TestWriteOutput_Continue(t *testing.T) {
	output := NewContinueOutput()
	var buf bytes.Buffer

	err := WriteOutput(&buf, output)

	require.NoError(t, err)

	var result map[string]any
	err = json.Unmarshal(buf.Bytes(), &result)
	require.NoError(t, err)
	assert.Equal(t, true, result["continue"])
}

func TestWriteOutput_Empty(t *testing.T) {
	output := NewEmptyOutput()
	var buf bytes.Buffer

	err := WriteOutput(&buf, output)

	require.NoError(t, err)

	var result map[string]any
	err = json.Unmarshal(buf.Bytes(), &result)
	require.NoError(t, err)
	// Empty output should have no fields set (all omitempty)
	assert.Empty(t, result["permission"])
	assert.Empty(t, result["continue"])
}

func TestWriteOutput_WithMessages(t *testing.T) {
	output := &HookOutput{
		Permission:   "allow",
		UserMessage:  "User message",
		AgentMessage: "Agent message",
	}
	var buf bytes.Buffer

	err := WriteOutput(&buf, output)

	require.NoError(t, err)

	var result map[string]any
	err = json.Unmarshal(buf.Bytes(), &result)
	require.NoError(t, err)
	assert.Equal(t, "allow", result["permission"])
	assert.Equal(t, "User message", result["user_message"])
	assert.Equal(t, "Agent message", result["agent_message"])
}

```

```go
package hook

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetDefaultOutput(t *testing.T) {
	tests := []struct {
		name       string
		event      HookEventName
		expectPerm string
		expectCont bool
	}{
		{"beforeSubmitPrompt", EventBeforeSubmitPrompt, "", true},
		{"beforeShellExecution", EventBeforeShellExecution, "allow", false},
		{"beforeMCPExecution", EventBeforeMCPExecution, "allow", false},
		{"beforeReadFile", EventBeforeReadFile, "allow", false},
		{"afterAgentResponse", EventAfterAgentResponse, "", false},
		{"stop", EventStop, "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := getDefaultOutput(tt.event)

			if tt.expectPerm != "" {
				assert.Equal(t, tt.expectPerm, output.Permission)
			}
			if tt.expectCont {
				assert.NotNil(t, output.Continue)
				assert.True(t, *output.Continue)
			}
		})
	}
}

func TestParseUUID_Valid(t *testing.T) {
	result := parseUUID("123e4567-e89b-12d3-a456-426614174000")

	assert.NotEqual(t, "00000000-0000-0000-0000-000000000000", result.String())
}

func TestParseUUID_Invalid(t *testing.T) {
	result := parseUUID("invalid-uuid")

	assert.Equal(t, "00000000-0000-0000-0000-000000000000", result.String())
}

func TestParseUUID_Empty(t *testing.T) {
	result := parseUUID("")

	assert.Equal(t, "00000000-0000-0000-0000-000000000000", result.String())
}

```

The test failure is due to a logic issue in the mapModelToEncoding function - the gpt-4o check comes after gpt-4, so "gpt-4o" matches the gpt-4 check first. Let me fix this:

```go
package hook

import (
	"strings"

	"github.com/pkoukk/tiktoken-go"
)

// Tokenizer is the interface for counting tokens in text.
type Tokenizer interface {
	// CountTokens counts the number of tokens in the given text for the specified model.
	CountTokens(model, text string) (int64, error)
}

// TiktokenTokenizer uses tiktoken for accurate token counting.
type TiktokenTokenizer struct {
	defaultModel string
}

// NewTiktokenTokenizer creates a new tiktoken-based tokenizer.
func NewTiktokenTokenizer(defaultModel string) *TiktokenTokenizer {
	return &TiktokenTokenizer{
		defaultModel: defaultModel,
	}
}

// CountTokens counts tokens using tiktoken.
func (t *TiktokenTokenizer) CountTokens(model, text string) (int64, error) {
	if model == "" {
		model = t.defaultModel
	}

	// Map common model names to tiktoken encoding names
	encoding := mapModelToEncoding(model)

	enc, err := tiktoken.GetEncoding(encoding)
	if err != nil {
		// Fall back to cl100k_base if encoding not found
		enc, err = tiktoken.GetEncoding("cl100k_base")
		if err != nil {
			return 0, err
		}
	}

	tokens := enc.Encode(text, nil, nil)
	return int64(len(tokens)), nil
}

// HeuristicTokenizer uses a simple character-based heuristic for token counting.
type HeuristicTokenizer struct {
	// CharsPerToken is the average number of characters per token.
	// Default is 4, which is a reasonable approximation for English text.
	CharsPerToken int
}

// NewHeuristicTokenizer creates a new heuristic-based tokenizer.
func NewHeuristicTokenizer() *HeuristicTokenizer {
	return &HeuristicTokenizer{
		CharsPerToken: 4,
	}
}

// CountTokens counts tokens using a character-based heuristic.
func (t *HeuristicTokenizer) CountTokens(_, text string) (int64, error) {
	if t.CharsPerToken <= 0 {
		t.CharsPerToken = 4
	}
	charCount := len(text)
	tokenCount := charCount / t.CharsPerToken
	if charCount%t.CharsPerToken > 0 {
		tokenCount++
	}
	return int64(tokenCount), nil
}

// CompositeTokenizer tries tiktoken first, then falls back to heuristic.
type CompositeTokenizer struct {
	tiktoken  *TiktokenTokenizer
	heuristic *HeuristicTokenizer
}

// NewCompositeTokenizer creates a new composite tokenizer.
func NewCompositeTokenizer(defaultModel string) *CompositeTokenizer {
	return &CompositeTokenizer{
		tiktoken:  NewTiktokenTokenizer(defaultModel),
		heuristic: NewHeuristicTokenizer(),
	}
}

// CountTokens tries tiktoken first, falls back to heuristic on error.
func (t *CompositeTokenizer) CountTokens(model, text string) (int64, error) {
	count, err := t.tiktoken.CountTokens(model, text)
	if err != nil {
		// Fall back to heuristic
		return t.heuristic.CountTokens(model, text)
	}
	return count, nil
}

// mapModelToEncoding maps model names to tiktoken encoding names.
func mapModelToEncoding(model string) string {
	model = strings.ToLower(model)

	// GPT-4o models use o200k_base (check before gpt-4)
	if strings.Contains(model, "gpt-4o") || strings.Contains(model, "o1") {
		return "o200k_base"
	}

	// GPT-4 and GPT-3.5 turbo models use cl100k_base
	if strings.Contains(model, "gpt-4") ||
		strings.Contains(model, "gpt-3.5") ||
		strings.Contains(model, "text-embedding") {
		return "cl100k_base"
	}

	// Claude models - use cl100k_base as approximation
	if strings.Contains(model, "claude") {
		return "cl100k_base"
	}

	// Anthropic models - use cl100k_base as approximation
	if strings.Contains(model, "anthropic") {
		return "cl100k_base"
	}

	// Gemini models - use cl100k_base as approximation
	if strings.Contains(model, "gemini") {
		return "cl100k_base"
	}

	// Default to cl100k_base for unknown models
	return "cl100k_base"
}

// CountChars returns the character count of the text.
func CountChars(text string) int64 {
	return int64(len(text))
}


```

```markdown
# Cursor Hooks Integration

This document describes the `ctd-hook` binary, which integrates with Cursor's hook system to capture token usage metrics and agent thoughts.

## Overview

The `ctd-hook` binary is a short-lived Go process that Cursor invokes for various hook events during agent interactions. It processes these events, counts tokens, and sends metrics to the ctd daemon over gRPC.

## Installation

### Build the Hook Binary

```bash
make build-hook
```

This creates `bin/ctd-hook`.

### Install to Cursor Hooks Directory

```bash
make install-hook
```

This copies the binary to `~/.cursor/hooks/ctd-hook`.

### Configure Cursor Hooks

Create or update `~/.cursor/hooks.json`:

```json
{
  "version": 1,
  "hooks": {
    "beforeSubmitPrompt": [{ "command": "~/.cursor/hooks/ctd-hook" }],
    "afterAgentResponse": [{ "command": "~/.cursor/hooks/ctd-hook" }],
    "afterAgentThought": [{ "command": "~/.cursor/hooks/ctd-hook" }],
    "afterMCPExecution": [{ "command": "~/.cursor/hooks/ctd-hook" }],
    "afterFileEdit": [{ "command": "~/.cursor/hooks/ctd-hook" }],
    "stop": [{ "command": "~/.cursor/hooks/ctd-hook" }]
  }
}
```

## Configuration

The hook binary can be configured via environment variables or a config file.

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `CTD_HOOK_SERVER_URL` | `http://127.0.0.1:36912` | URL of the ctd daemon |
| `CTD_HOOK_LOG_LEVEL` | `info` | Log level (debug, info, warn, error) |
| `CTD_HOOK_DEFAULT_MODEL` | `gpt-4` | Default model for token counting |
| `CTD_HOOK_TIMEOUT_MS` | `500` | Timeout for ctd calls in milliseconds |

### Config File

Optionally, create `$XDG_CONFIG_HOME/ctd/hook.yaml` (typically `~/.config/ctd/hook.yaml`):

```yaml
server_url: http://127.0.0.1:36912
log_level: info
default_model: gpt-4
timeout_ms: 500
```

## Supported Hook Events

### beforeSubmitPrompt

Triggered before a prompt is submitted to the agent. Records prompt token usage.

**Metrics captured:**
- Prompt tokens
- Prompt characters

### afterAgentResponse

Triggered after the agent generates a response. Records completion token usage and creates a thought entry.

**Metrics captured:**
- Completion tokens
- Completion characters

### afterAgentThought

Triggered after the agent generates a thought. Creates a thought entry for visibility in the extension.

### afterMCPExecution

Triggered after an MCP tool is executed. Records tool input/output token usage.

**Metrics captured:**
- Tool input tokens
- Tool result tokens

### afterFileEdit

Triggered after a file is edited. Records character counts for the edit.

**Metrics captured:**
- Old content characters
- New content characters

### stop

Triggered when the agent loop ends. Records a session marker.

## Token Counting

The hook uses a composite tokenizer that:

1. Attempts to use tiktoken for accurate token counting
2. Falls back to a heuristic (characters / 4) if tiktoken fails

Model-to-encoding mapping:
- `gpt-4o`, `o1`: `o200k_base`
- `gpt-4`, `gpt-3.5`: `cl100k_base`
- `claude`, `gemini`: `cl100k_base` (approximation)

## Architecture

```
stdin (JSON) -> Parser -> Event Router -> Handler -> CTD Client -> ctd daemon
                                       -> stdout (JSON)
```

The hook:
1. Reads JSON from stdin (provided by Cursor)
2. Parses the event envelope and payload
3. Routes to the appropriate handler
4. Counts tokens and sends metrics to ctd (best-effort)
5. Writes response JSON to stdout
6. Exits with code 0 (always, to avoid blocking Cursor)

## Logging

Logs are written to stderr in structured format. Set `CTD_HOOK_LOG_LEVEL=debug` for verbose output.

## Error Handling

The hook is designed to be non-blocking:
- All errors are logged but do not prevent the agent from proceeding
- Network failures to ctd are logged and ignored
- Invalid input results in default output (allow/continue)
- The hook always exits with code 0

## Database Schema

Token usage metrics are stored in the `token_usage` table:

```sql
CREATE TABLE token_usage (
    id UUID PRIMARY KEY,
    conversation_id UUID NOT NULL,
    generation_id UUID NOT NULL,
    model VARCHAR(255) NOT NULL,
    stage VARCHAR(32) NOT NULL,
    prompt_tokens BIGINT NOT NULL,
    completion_tokens BIGINT NOT NULL,
    total_tokens BIGINT NOT NULL,
    prompt_chars BIGINT NOT NULL,
    completion_chars BIGINT NOT NULL,
    workspace_roots TEXT[] NOT NULL,
    user_email VARCHAR(255) NOT NULL,
    hook_event_name VARCHAR(255) NOT NULL,
    timestamp TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
```

## Troubleshooting

### Hook not receiving events

1. Verify `~/.cursor/hooks.json` exists and is valid JSON
2. Check that the hook binary is executable: `chmod +x ~/.cursor/hooks/ctd-hook`
3. Restart Cursor after modifying hooks.json

### Metrics not appearing in database

1. Verify ctd daemon is running: `curl http://127.0.0.1:36912/ping`
2. Check hook logs: `CTD_HOOK_LOG_LEVEL=debug ~/.cursor/hooks/ctd-hook < test-input.json`
3. Verify database migration has been applied

### Token counts seem inaccurate

The hook uses tiktoken for OpenAI models. For other models (Claude, Gemini), it uses cl100k_base as an approximation, which may not be exact.

```

```plaintext
{
  "version": 1,
  "hooks": {
    "beforeSubmitPrompt": [{ "command": "~/.cursor/hooks/ctd-hook" }],
    "afterAgentResponse": [{ "command": "~/.cursor/hooks/ctd-hook" }],
    "afterAgentThought": [{ "command": "~/.cursor/hooks/ctd-hook" }],
    "afterMCPExecution": [{ "command": "~/.cursor/hooks/ctd-hook" }],
    "afterFileEdit": [{ "command": "~/.cursor/hooks/ctd-hook" }],
    "stop": [{ "command": "~/.cursor/hooks/ctd-hook" }]
  }
}

```

Now let me update the README:

```markdown
# Cursor Thought Daemon (ctd)

**ctd** serves to solve the infuriating probelm whereby Cursor tries as hard as it can to abstract it's thinking from the UI. This allows Cursor time to waste tokens and go on wild goose chases based on whatever the model hallucinates. It is a background service designed to capture, persist, and visualize the "thoughts" generated by the Cursor AI agent in real-time. It acts as a bridge between the AI's internal reasoning process and a user-friendly visualization interface.

The system consists of three main components:
1.  **The Daemon (`ctd`)**: A high-performance Go server that receives thoughts and token usage metrics, stores them in PostgreSQL, and broadcasts them via a streaming API.
2.  **The Hook (`ctd-hook`)**: A Cursor hook binary that captures agent events, counts tokens, and sends metrics to the daemon.
3.  **The Extension**: A VS Code/Cursor extension that connects to the daemon to display a live feed of the AI's thinking process in a dedicated sidebar.

##  Features

- **Real-time Streaming**: Watch the AI's thought process unfold live via server-side streaming (Connect-RPC).
- **Persistence**: Automatically archives all thoughts to a PostgreSQL database for historical analysis.
- **Structured Data**: Captures metadata like generation duration, model used, and context.
- **Low Latency**: Built with Go and gRPC/Connect for minimal overhead.
- **Visual Interface**: Includes a React-based VS Code extension for a clean, readable feed.

##  Architecture

The system follows a Client-Server-Client model:

```mermaid
graph TD
    Agent[Cursor Agent] -->|CreateThought| Server[Thought Daemon]
    Server -->|StreamThoughts| Extension[VS Code Extension]
    Server -->|Persist| DB[(PostgreSQL)]
```

For a deep dive into the internal design, see [docs/architecture.md](docs/architecture.md).

##  Prerequisites

- **Go** 1.23+
- **Docker** & Docker Compose (for PostgreSQL)
- **Node.js** & npm (for the extension)
- **Cursor** IDE

##  Installation & Setup

### 1. Start the Database

The daemon requires a PostgreSQL instance. A `docker-compose.yml` is provided for convenience.

```bash
# Start Postgres
make docker-up
```

### 2. Install the Daemon (`ctd`)

You can build and install the binary to your local path.

```bash
# Build and install to ~/.local/bin
make install
```

This will:
1.  Compile the `ctd` binary.
2.  Copy it to `~/.local/bin`.
3.  Create a default configuration at `~/.config/ctd/config.yaml`.

*Ensure `~/.local/bin` is in your `$PATH`.*

### 3. Install the Extension

To visualize the thoughts in Cursor, install the included extension:

```bash
# Build and install the extension directly to Cursor
make extension-install
```

After installation, reload the Cursor window (Cmd+Shift+P -> "Developer: Reload Window").

##  Configuration

The daemon is configured via `~/.config/ctd/config.yaml`.

```yaml
server:
  port: 36912
  host: "127.0.0.1"

database:
  url: "postgres://user:password@localhost:5432/thought_db"

log:
  level: "info" # debug, info, warn, error
```

##  Usage

1.  **Start the Daemon**:
    Run the daemon in a terminal:
    ```bash
    ctd
    ```
    Or run it directly from source:
    ```bash
    make run
    ```

2.  **Open Cursor**:
    - Open the "Cursor Thoughts" view in the activity bar (brain icon).
    - The extension will connect to `http://127.0.0.1:36912`.

3.  **Generate Thoughts**:
    - As the Cursor Agent performs actions that trigger the hook (configured separately), thoughts will appear in the sidebar.

##  Development

### Project Structure

- `cmd/thought-daemon`: Main entry point for the daemon.
- `internal/`: Core logic (Server, Repository, Domain).
- `api/proto`: Protobuf service definitions.
- `vscode-extension/`: The React/TypeScript extension.

### Testing

For comprehensive testing documentation covering both backend and frontend, see **[TESTING.md](TESTING.md)**.

Quick test commands:

**Backend (Go)**:
```bash
make test              # Run all tests
make test-unit         # Run unit tests only
make test-coverage     # Generate coverage report
```

**Frontend (TypeScript/React)**:
```bash
cd vscode-extension
npm test               # Run all tests
npm run test:unit      # Run unit tests
npm run test:coverage  # Generate coverage report
```

### Common Commands

```bash
# Generate Go code from Proto files
make gen

# Run Go tests
make test

# Lint Go code
make lint

# Clean build artifacts
make clean
```

### API Definition

The API is defined in `api/proto/v1/thought.proto` using Protobuf. It exposes a `ThoughtService` with methods to create and stream thoughts.

##  License

MIT

```

```markdown
# Cursor Thought Daemon (ctd)

**ctd** serves to solve the infuriating probelm whereby Cursor tries as hard as it can to abstract it's thinking from the UI. This allows Cursor time to waste tokens and go on wild goose chases based on whatever the model hallucinates. It is a background service designed to capture, persist, and visualize the "thoughts" generated by the Cursor AI agent in real-time. It acts as a bridge between the AI's internal reasoning process and a user-friendly visualization interface.

The system consists of three main components:
1.  **The Daemon (`ctd`)**: A high-performance Go server that receives thoughts and token usage metrics, stores them in PostgreSQL, and broadcasts them via a streaming API.
2.  **The Hook (`ctd-hook`)**: A Cursor hook binary that captures agent events, counts tokens, and sends metrics to the daemon.
3.  **The Extension**: A VS Code/Cursor extension that connects to the daemon to display a live feed of the AI's thinking process in a dedicated sidebar.

##  Features

- **Real-time Streaming**: Watch the AI's thought process unfold live via server-side streaming (Connect-RPC).
- **Token Usage Tracking**: Automatically counts and records prompt, completion, and tool tokens.
- **Persistence**: Automatically archives all thoughts and token metrics to a PostgreSQL database for historical analysis.
- **Structured Data**: Captures metadata like generation duration, model used, token counts, and context.
- **Low Latency**: Built with Go and gRPC/Connect for minimal overhead.
- **Visual Interface**: Includes a React-based VS Code extension for a clean, readable feed.
- **Cursor Hooks Integration**: Native integration with Cursor's hook system for automatic event capture.

##  Architecture

The system follows a Client-Server-Client model:

```mermaid
graph TD
    Agent[Cursor Agent] -->|CreateThought| Server[Thought Daemon]
    Server -->|StreamThoughts| Extension[VS Code Extension]
    Server -->|Persist| DB[(PostgreSQL)]
```

For a deep dive into the internal design, see [docs/architecture.md](docs/architecture.md).

##  Prerequisites

- **Go** 1.23+
- **Docker** & Docker Compose (for PostgreSQL)
- **Node.js** & npm (for the extension)
- **Cursor** IDE

##  Installation & Setup

### 1. Start the Database

The daemon requires a PostgreSQL instance. A `docker-compose.yml` is provided for convenience.

```bash
# Start Postgres
make docker-up
```

### 2. Install the Daemon (`ctd`)

You can build and install the binary to your local path.

```bash
# Build and install to ~/.local/bin
make install
```

This will:
1.  Compile the `ctd` binary.
2.  Copy it to `~/.local/bin`.
3.  Create a default configuration at `~/.config/ctd/config.yaml`.

*Ensure `~/.local/bin` is in your `$PATH`.*

### 3. Install the Extension

To visualize the thoughts in Cursor, install the included extension:

```bash
# Build and install the extension directly to Cursor
make extension-install
```

After installation, reload the Cursor window (Cmd+Shift+P -> "Developer: Reload Window").

##  Configuration

The daemon is configured via `~/.config/ctd/config.yaml`.

```yaml
server:
  port: 36912
  host: "127.0.0.1"

database:
  url: "postgres://user:password@localhost:5432/thought_db"

log:
  level: "info" # debug, info, warn, error
```

##  Usage

1.  **Start the Daemon**:
    Run the daemon in a terminal:
    ```bash
    ctd
    ```
    Or run it directly from source:
    ```bash
    make run
    ```

2.  **Open Cursor**:
    - Open the "Cursor Thoughts" view in the activity bar (brain icon).
    - The extension will connect to `http://127.0.0.1:36912`.

3.  **Generate Thoughts**:
    - As the Cursor Agent performs actions that trigger the hook (configured separately), thoughts will appear in the sidebar.

##  Development

### Project Structure

- `cmd/thought-daemon`: Main entry point for the daemon.
- `internal/`: Core logic (Server, Repository, Domain).
- `api/proto`: Protobuf service definitions.
- `vscode-extension/`: The React/TypeScript extension.

### Testing

For comprehensive testing documentation covering both backend and frontend, see **[TESTING.md](TESTING.md)**.

Quick test commands:

**Backend (Go)**:
```bash
make test              # Run all tests
make test-unit         # Run unit tests only
make test-coverage     # Generate coverage report
```

**Frontend (TypeScript/React)**:
```bash
cd vscode-extension
npm test               # Run all tests
npm run test:unit      # Run unit tests
npm run test:coverage  # Generate coverage report
```

### Common Commands

```bash
# Generate Go code from Proto files
make gen

# Run Go tests
make test

# Lint Go code
make lint

# Clean build artifacts
make clean
```

### API Definition

The API is defined in `api/proto/v1/thought.proto` using Protobuf. It exposes a `ThoughtService` with methods to create and stream thoughts.

##  License

MIT

```

```markdown
# Cursor Thought Daemon (ctd)

**ctd** serves to solve the infuriating probelm whereby Cursor tries as hard as it can to abstract it's thinking from the UI. This allows Cursor time to waste tokens and go on wild goose chases based on whatever the model hallucinates. It is a background service designed to capture, persist, and visualize the "thoughts" generated by the Cursor AI agent in real-time. It acts as a bridge between the AI's internal reasoning process and a user-friendly visualization interface.

The system consists of three main components:
1.  **The Daemon (`ctd`)**: A high-performance Go server that receives thoughts and token usage metrics, stores them in PostgreSQL, and broadcasts them via a streaming API.
2.  **The Hook (`ctd-hook`)**: A Cursor hook binary that captures agent events, counts tokens, and sends metrics to the daemon.
3.  **The Extension**: A VS Code/Cursor extension that connects to the daemon to display a live feed of the AI's thinking process in a dedicated sidebar.

##  Features

- **Real-time Streaming**: Watch the AI's thought process unfold live via server-side streaming (Connect-RPC).
- **Token Usage Tracking**: Automatically counts and records prompt, completion, and tool tokens.
- **Persistence**: Automatically archives all thoughts and token metrics to a PostgreSQL database for historical analysis.
- **Structured Data**: Captures metadata like generation duration, model used, token counts, and context.
- **Low Latency**: Built with Go and gRPC/Connect for minimal overhead.
- **Visual Interface**: Includes a React-based VS Code extension for a clean, readable feed.
- **Cursor Hooks Integration**: Native integration with Cursor's hook system for automatic event capture.

##  Architecture

The system follows a Client-Server-Client model with hook-based event capture:

```mermaid
graph TD
    Agent[Cursor Agent] -->|Hook Events| Hook[ctd-hook]
    Hook -->|CreateThought| Server[Thought Daemon]
    Hook -->|CreateTokenUsage| Server
    Server -->|StreamThoughts| Extension[VS Code Extension]
    Server -->|Persist| DB[(PostgreSQL)]
```

For a deep dive into the internal design, see [docs/architecture.md](docs/architecture.md).

##  Prerequisites

- **Go** 1.23+
- **Docker** & Docker Compose (for PostgreSQL)
- **Node.js** & npm (for the extension)
- **Cursor** IDE

##  Installation & Setup

### 1. Start the Database

The daemon requires a PostgreSQL instance. A `docker-compose.yml` is provided for convenience.

```bash
# Start Postgres
make docker-up
```

### 2. Install the Daemon (`ctd`)

You can build and install the binary to your local path.

```bash
# Build and install to ~/.local/bin
make install
```

This will:
1.  Compile the `ctd` binary.
2.  Copy it to `~/.local/bin`.
3.  Create a default configuration at `~/.config/ctd/config.yaml`.

*Ensure `~/.local/bin` is in your `$PATH`.*

### 3. Install the Extension

To visualize the thoughts in Cursor, install the included extension:

```bash
# Build and install the extension directly to Cursor
make extension-install
```

After installation, reload the Cursor window (Cmd+Shift+P -> "Developer: Reload Window").

##  Configuration

The daemon is configured via `~/.config/ctd/config.yaml`.

```yaml
server:
  port: 36912
  host: "127.0.0.1"

database:
  url: "postgres://user:password@localhost:5432/thought_db"

log:
  level: "info" # debug, info, warn, error
```

##  Usage

1.  **Start the Daemon**:
    Run the daemon in a terminal:
    ```bash
    ctd
    ```
    Or run it directly from source:
    ```bash
    make run
    ```

2.  **Open Cursor**:
    - Open the "Cursor Thoughts" view in the activity bar (brain icon).
    - The extension will connect to `http://127.0.0.1:36912`.

3.  **Generate Thoughts**:
    - As the Cursor Agent performs actions that trigger the hook (configured separately), thoughts will appear in the sidebar.

##  Development

### Project Structure

- `cmd/thought-daemon`: Main entry point for the daemon.
- `internal/`: Core logic (Server, Repository, Domain).
- `api/proto`: Protobuf service definitions.
- `vscode-extension/`: The React/TypeScript extension.

### Testing

For comprehensive testing documentation covering both backend and frontend, see **[TESTING.md](TESTING.md)**.

Quick test commands:

**Backend (Go)**:
```bash
make test              # Run all tests
make test-unit         # Run unit tests only
make test-coverage     # Generate coverage report
```

**Frontend (TypeScript/React)**:
```bash
cd vscode-extension
npm test               # Run all tests
npm run test:unit      # Run unit tests
npm run test:coverage  # Generate coverage report
```

### Common Commands

```bash
# Generate Go code from Proto files
make gen

# Run Go tests
make test

# Lint Go code
make lint

# Clean build artifacts
make clean
```

### API Definition

The API is defined in `api/proto/v1/thought.proto` using Protobuf. It exposes a `ThoughtService` with methods to create and stream thoughts.

##  License

MIT

```

```markdown
# Cursor Thought Daemon (ctd)

**ctd** serves to solve the infuriating probelm whereby Cursor tries as hard as it can to abstract it's thinking from the UI. This allows Cursor time to waste tokens and go on wild goose chases based on whatever the model hallucinates. It is a background service designed to capture, persist, and visualize the "thoughts" generated by the Cursor AI agent in real-time. It acts as a bridge between the AI's internal reasoning process and a user-friendly visualization interface.

The system consists of three main components:
1.  **The Daemon (`ctd`)**: A high-performance Go server that receives thoughts and token usage metrics, stores them in PostgreSQL, and broadcasts them via a streaming API.
2.  **The Hook (`ctd-hook`)**: A Cursor hook binary that captures agent events, counts tokens, and sends metrics to the daemon.
3.  **The Extension**: A VS Code/Cursor extension that connects to the daemon to display a live feed of the AI's thinking process in a dedicated sidebar.

##  Features

- **Real-time Streaming**: Watch the AI's thought process unfold live via server-side streaming (Connect-RPC).
- **Token Usage Tracking**: Automatically counts and records prompt, completion, and tool tokens.
- **Persistence**: Automatically archives all thoughts and token metrics to a PostgreSQL database for historical analysis.
- **Structured Data**: Captures metadata like generation duration, model used, token counts, and context.
- **Low Latency**: Built with Go and gRPC/Connect for minimal overhead.
- **Visual Interface**: Includes a React-based VS Code extension for a clean, readable feed.
- **Cursor Hooks Integration**: Native integration with Cursor's hook system for automatic event capture.

##  Architecture

The system follows a Client-Server-Client model with hook-based event capture:

```mermaid
graph TD
    Agent[Cursor Agent] -->|Hook Events| Hook[ctd-hook]
    Hook -->|CreateThought| Server[Thought Daemon]
    Hook -->|CreateTokenUsage| Server
    Server -->|StreamThoughts| Extension[VS Code Extension]
    Server -->|Persist| DB[(PostgreSQL)]
```

For a deep dive into the internal design, see [docs/architecture.md](docs/architecture.md).

##  Prerequisites

- **Go** 1.23+
- **Docker** & Docker Compose (for PostgreSQL)
- **Node.js** & npm (for the extension)
- **Cursor** IDE

##  Installation & Setup

### 1. Start the Database

The daemon requires a PostgreSQL instance. A `docker-compose.yml` is provided for convenience.

```bash
# Start Postgres
make docker-up
```

### 2. Install the Daemon (`ctd`)

You can build and install the binary to your local path.

```bash
# Build and install to ~/.local/bin
make install
```

This will:
1.  Compile the `ctd` binary.
2.  Copy it to `~/.local/bin`.
3.  Create a default configuration at `~/.config/ctd/config.yaml`.

*Ensure `~/.local/bin` is in your `$PATH`.*

### 3. Install the Hook Binary

The hook binary integrates with Cursor's hook system to capture agent events:

```bash
# Build and install to ~/.cursor/hooks
make install-hook
```

Then create `~/.cursor/hooks.json` (see `deploy/hooks.json.example`):

```json
{
  "version": 1,
  "hooks": {
    "beforeSubmitPrompt": [{ "command": "~/.cursor/hooks/ctd-hook" }],
    "afterAgentResponse": [{ "command": "~/.cursor/hooks/ctd-hook" }],
    "afterAgentThought": [{ "command": "~/.cursor/hooks/ctd-hook" }],
    "afterMCPExecution": [{ "command": "~/.cursor/hooks/ctd-hook" }],
    "afterFileEdit": [{ "command": "~/.cursor/hooks/ctd-hook" }],
    "stop": [{ "command": "~/.cursor/hooks/ctd-hook" }]
  }
}
```

For detailed hook configuration, see [docs/hooks.md](docs/hooks.md).

### 4. Install the Extension

To visualize the thoughts in Cursor, install the included extension:

```bash
# Build and install the extension directly to Cursor
make extension-install
```

After installation, reload the Cursor window (Cmd+Shift+P -> "Developer: Reload Window").

##  Configuration

The daemon is configured via `~/.config/ctd/config.yaml`.

```yaml
server:
  port: 36912
  host: "127.0.0.1"

database:
  url: "postgres://user:password@localhost:5432/thought_db"

log:
  level: "info" # debug, info, warn, error
```

##  Usage

1.  **Start the Daemon**:
    Run the daemon in a terminal:
    ```bash
    ctd
    ```
    Or run it directly from source:
    ```bash
    make run
    ```

2.  **Open Cursor**:
    - Open the "Cursor Thoughts" view in the activity bar (brain icon).
    - The extension will connect to `http://127.0.0.1:36912`.

3.  **Generate Thoughts**:
    - As the Cursor Agent performs actions that trigger the hook (configured separately), thoughts will appear in the sidebar.

##  Development

### Project Structure

- `cmd/thought-daemon`: Main entry point for the daemon.
- `internal/`: Core logic (Server, Repository, Domain).
- `api/proto`: Protobuf service definitions.
- `vscode-extension/`: The React/TypeScript extension.

### Testing

For comprehensive testing documentation covering both backend and frontend, see **[TESTING.md](TESTING.md)**.

Quick test commands:

**Backend (Go)**:
```bash
make test              # Run all tests
make test-unit         # Run unit tests only
make test-coverage     # Generate coverage report
```

**Frontend (TypeScript/React)**:
```bash
cd vscode-extension
npm test               # Run all tests
npm run test:unit      # Run unit tests
npm run test:coverage  # Generate coverage report
```

### Common Commands

```bash
# Generate Go code from Proto files
make gen

# Run Go tests
make test

# Lint Go code
make lint

# Clean build artifacts
make clean
```

### API Definition

The API is defined in `api/proto/v1/thought.proto` using Protobuf. It exposes a `ThoughtService` with methods to create and stream thoughts.

##  License

MIT

```

```markdown
# Cursor Thought Daemon (ctd)

**ctd** serves to solve the infuriating probelm whereby Cursor tries as hard as it can to abstract it's thinking from the UI. This allows Cursor time to waste tokens and go on wild goose chases based on whatever the model hallucinates. It is a background service designed to capture, persist, and visualize the "thoughts" generated by the Cursor AI agent in real-time. It acts as a bridge between the AI's internal reasoning process and a user-friendly visualization interface.

The system consists of three main components:
1.  **The Daemon (`ctd`)**: A high-performance Go server that receives thoughts and token usage metrics, stores them in PostgreSQL, and broadcasts them via a streaming API.
2.  **The Hook (`ctd-hook`)**: A Cursor hook binary that captures agent events, counts tokens, and sends metrics to the daemon.
3.  **The Extension**: A VS Code/Cursor extension that connects to the daemon to display a live feed of the AI's thinking process in a dedicated sidebar.

##  Features

- **Real-time Streaming**: Watch the AI's thought process unfold live via server-side streaming (Connect-RPC).
- **Token Usage Tracking**: Automatically counts and records prompt, completion, and tool tokens.
- **Persistence**: Automatically archives all thoughts and token metrics to a PostgreSQL database for historical analysis.
- **Structured Data**: Captures metadata like generation duration, model used, token counts, and context.
- **Low Latency**: Built with Go and gRPC/Connect for minimal overhead.
- **Visual Interface**: Includes a React-based VS Code extension for a clean, readable feed.
- **Cursor Hooks Integration**: Native integration with Cursor's hook system for automatic event capture.

##  Architecture

The system follows a Client-Server-Client model with hook-based event capture:

```mermaid
graph TD
    Agent[Cursor Agent] -->|Hook Events| Hook[ctd-hook]
    Hook -->|CreateThought| Server[Thought Daemon]
    Hook -->|CreateTokenUsage| Server
    Server -->|StreamThoughts| Extension[VS Code Extension]
    Server -->|Persist| DB[(PostgreSQL)]
```

For a deep dive into the internal design, see [docs/architecture.md](docs/architecture.md).

##  Prerequisites

- **Go** 1.23+
- **Docker** & Docker Compose (for PostgreSQL)
- **Node.js** & npm (for the extension)
- **Cursor** IDE

##  Installation & Setup

### 1. Start the Database

The daemon requires a PostgreSQL instance. A `docker-compose.yml` is provided for convenience.

```bash
# Start Postgres
make docker-up
```

### 2. Install the Daemon (`ctd`)

You can build and install the binary to your local path.

```bash
# Build and install to ~/.local/bin
make install
```

This will:
1.  Compile the `ctd` binary.
2.  Copy it to `~/.local/bin`.
3.  Create a default configuration at `~/.config/ctd/config.yaml`.

*Ensure `~/.local/bin` is in your `$PATH`.*

### 3. Install the Hook Binary

The hook binary integrates with Cursor's hook system to capture agent events:

```bash
# Build and install to ~/.cursor/hooks
make install-hook
```

Then create `~/.cursor/hooks.json` (see `deploy/hooks.json.example`):

```json
{
  "version": 1,
  "hooks": {
    "beforeSubmitPrompt": [{ "command": "~/.cursor/hooks/ctd-hook" }],
    "afterAgentResponse": [{ "command": "~/.cursor/hooks/ctd-hook" }],
    "afterAgentThought": [{ "command": "~/.cursor/hooks/ctd-hook" }],
    "afterMCPExecution": [{ "command": "~/.cursor/hooks/ctd-hook" }],
    "afterFileEdit": [{ "command": "~/.cursor/hooks/ctd-hook" }],
    "stop": [{ "command": "~/.cursor/hooks/ctd-hook" }]
  }
}
```

For detailed hook configuration, see [docs/hooks.md](docs/hooks.md).

### 4. Install the Extension

To visualize the thoughts in Cursor, install the included extension:

```bash
# Build and install the extension directly to Cursor
make extension-install
```

After installation, reload the Cursor window (Cmd+Shift+P -> "Developer: Reload Window").

##  Configuration

The daemon is configured via `~/.config/ctd/config.yaml`.

```yaml
server:
  port: 36912
  host: "127.0.0.1"

database:
  url: "postgres://user:password@localhost:5432/thought_db"

log:
  level: "info" # debug, info, warn, error
```

##  Usage

1.  **Start the Daemon**:
    Run the daemon in a terminal:
    ```bash
    ctd
    ```
    Or run it directly from source:
    ```bash
    make run
    ```

2.  **Open Cursor**:
    - Open the "Cursor Thoughts" view in the activity bar (brain icon).
    - The extension will connect to `http://127.0.0.1:36912`.

3.  **Generate Thoughts**:
    - As the Cursor Agent performs actions that trigger the hook (configured separately), thoughts will appear in the sidebar.

##  Development

### Project Structure

- `cmd/thought-daemon`: Main entry point for the daemon.
- `cmd/ctd-hook`: Cursor hook binary for event capture.
- `internal/`: Core logic (Server, Repository, Domain, Hook).
- `api/proto`: Protobuf service definitions.
- `vscode-extension/`: The React/TypeScript extension.

### Testing

For comprehensive testing documentation covering both backend and frontend, see **[TESTING.md](TESTING.md)**.

Quick test commands:

**Backend (Go)**:
```bash
make test              # Run all tests
make test-unit         # Run unit tests only
make test-coverage     # Generate coverage report
```

**Frontend (TypeScript/React)**:
```bash
cd vscode-extension
npm test               # Run all tests
npm run test:unit      # Run unit tests
npm run test:coverage  # Generate coverage report
```

### Common Commands

```bash
# Generate Go code from Proto files
make gen

# Run Go tests
make test

# Lint Go code
make lint

# Clean build artifacts
make clean
```

### API Definition

The API is defined in `api/proto/v1/thought.proto` using Protobuf. It exposes a `ThoughtService` with methods to create and stream thoughts.

##  License

MIT

```

```markdown
# Cursor Thought Daemon (ctd)

**ctd** serves to solve the infuriating probelm whereby Cursor tries as hard as it can to abstract it's thinking from the UI. This allows Cursor time to waste tokens and go on wild goose chases based on whatever the model hallucinates. It is a background service designed to capture, persist, and visualize the "thoughts" generated by the Cursor AI agent in real-time. It acts as a bridge between the AI's internal reasoning process and a user-friendly visualization interface.

The system consists of three main components:
1.  **The Daemon (`ctd`)**: A high-performance Go server that receives thoughts and token usage metrics, stores them in PostgreSQL, and broadcasts them via a streaming API.
2.  **The Hook (`ctd-hook`)**: A Cursor hook binary that captures agent events, counts tokens, and sends metrics to the daemon.
3.  **The Extension**: A VS Code/Cursor extension that connects to the daemon to display a live feed of the AI's thinking process in a dedicated sidebar.

##  Features

- **Real-time Streaming**: Watch the AI's thought process unfold live via server-side streaming (Connect-RPC).
- **Token Usage Tracking**: Automatically counts and records prompt, completion, and tool tokens.
- **Persistence**: Automatically archives all thoughts and token metrics to a PostgreSQL database for historical analysis.
- **Structured Data**: Captures metadata like generation duration, model used, token counts, and context.
- **Low Latency**: Built with Go and gRPC/Connect for minimal overhead.
- **Visual Interface**: Includes a React-based VS Code extension for a clean, readable feed.
- **Cursor Hooks Integration**: Native integration with Cursor's hook system for automatic event capture.

##  Architecture

The system follows a Client-Server-Client model with hook-based event capture:

```mermaid
graph TD
    Agent[Cursor Agent] -->|Hook Events| Hook[ctd-hook]
    Hook -->|CreateThought| Server[Thought Daemon]
    Hook -->|CreateTokenUsage| Server
    Server -->|StreamThoughts| Extension[VS Code Extension]
    Server -->|Persist| DB[(PostgreSQL)]
```

For a deep dive into the internal design, see [docs/architecture.md](docs/architecture.md).

##  Prerequisites

- **Go** 1.23+
- **Docker** & Docker Compose (for PostgreSQL)
- **Node.js** & npm (for the extension)
- **Cursor** IDE

##  Installation & Setup

### 1. Start the Database

The daemon requires a PostgreSQL instance. A `docker-compose.yml` is provided for convenience.

```bash
# Start Postgres
make docker-up
```

### 2. Install the Daemon (`ctd`)

You can build and install the binary to your local path.

```bash
# Build and install to ~/.local/bin
make install
```

This will:
1.  Compile the `ctd` binary.
2.  Copy it to `~/.local/bin`.
3.  Create a default configuration at `~/.config/ctd/config.yaml`.

*Ensure `~/.local/bin` is in your `$PATH`.*

### 3. Install the Hook Binary

The hook binary integrates with Cursor's hook system to capture agent events:

```bash
# Build and install to ~/.cursor/hooks
make install-hook
```

Then create `~/.cursor/hooks.json` (see `deploy/hooks.json.example`):

```json
{
  "version": 1,
  "hooks": {
    "beforeSubmitPrompt": [{ "command": "~/.cursor/hooks/ctd-hook" }],
    "afterAgentResponse": [{ "command": "~/.cursor/hooks/ctd-hook" }],
    "afterAgentThought": [{ "command": "~/.cursor/hooks/ctd-hook" }],
    "afterMCPExecution": [{ "command": "~/.cursor/hooks/ctd-hook" }],
    "afterFileEdit": [{ "command": "~/.cursor/hooks/ctd-hook" }],
    "stop": [{ "command": "~/.cursor/hooks/ctd-hook" }]
  }
}
```

For detailed hook configuration, see [docs/hooks.md](docs/hooks.md).

### 4. Install the Extension

To visualize the thoughts in Cursor, install the included extension:

```bash
# Build and install the extension directly to Cursor
make extension-install
```

After installation, reload the Cursor window (Cmd+Shift+P -> "Developer: Reload Window").

##  Configuration

The daemon is configured via `~/.config/ctd/config.yaml`.

```yaml
server:
  port: 36912
  host: "127.0.0.1"

database:
  url: "postgres://user:password@localhost:5432/thought_db"

log:
  level: "info" # debug, info, warn, error
```

##  Usage

1.  **Start the Daemon**:
    Run the daemon in a terminal:
    ```bash
    ctd
    ```
    Or run it directly from source:
    ```bash
    make run
    ```

2.  **Open Cursor**:
    - Open the "Cursor Thoughts" view in the activity bar (brain icon).
    - The extension will connect to `http://127.0.0.1:36912`.

3.  **Generate Thoughts**:
    - As the Cursor Agent performs actions that trigger the hook (configured separately), thoughts will appear in the sidebar.

##  Development

### Project Structure

- `cmd/thought-daemon`: Main entry point for the daemon.
- `cmd/ctd-hook`: Cursor hook binary for event capture.
- `internal/`: Core logic (Server, Repository, Domain, Hook).
- `api/proto`: Protobuf service definitions.
- `vscode-extension/`: The React/TypeScript extension.

### Testing

For comprehensive testing documentation covering both backend and frontend, see **[TESTING.md](TESTING.md)**.

Quick test commands:

**Backend (Go)**:
```bash
make test              # Run all tests
make test-unit         # Run unit tests only
make test-coverage     # Generate coverage report
```

**Frontend (TypeScript/React)**:
```bash
cd vscode-extension
npm test               # Run all tests
npm run test:unit      # Run unit tests
npm run test:coverage  # Generate coverage report
```

### Common Commands

```bash
# Generate Go code from Proto files
make gen

# Build daemon and hook
make build-all

# Run Go tests
make test

# Lint Go code
make lint

# Clean build artifacts
make clean
```

### API Definition

The API is defined in `api/proto/v1/thought.proto` using Protobuf. It exposes a `ThoughtService` with methods to create and stream thoughts.

##  License

MIT

```

```markdown
# Cursor Thought Daemon (ctd)

**ctd** serves to solve the infuriating probelm whereby Cursor tries as hard as it can to abstract it's thinking from the UI. This allows Cursor time to waste tokens and go on wild goose chases based on whatever the model hallucinates. It is a background service designed to capture, persist, and visualize the "thoughts" generated by the Cursor AI agent in real-time. It acts as a bridge between the AI's internal reasoning process and a user-friendly visualization interface.

The system consists of three main components:
1.  **The Daemon (`ctd`)**: A high-performance Go server that receives thoughts and token usage metrics, stores them in PostgreSQL, and broadcasts them via a streaming API.
2.  **The Hook (`ctd-hook`)**: A Cursor hook binary that captures agent events, counts tokens, and sends metrics to the daemon.
3.  **The Extension**: A VS Code/Cursor extension that connects to the daemon to display a live feed of the AI's thinking process in a dedicated sidebar.

##  Features

- **Real-time Streaming**: Watch the AI's thought process unfold live via server-side streaming (Connect-RPC).
- **Token Usage Tracking**: Automatically counts and records prompt, completion, and tool tokens.
- **Persistence**: Automatically archives all thoughts and token metrics to a PostgreSQL database for historical analysis.
- **Structured Data**: Captures metadata like generation duration, model used, token counts, and context.
- **Low Latency**: Built with Go and gRPC/Connect for minimal overhead.
- **Visual Interface**: Includes a React-based VS Code extension for a clean, readable feed.
- **Cursor Hooks Integration**: Native integration with Cursor's hook system for automatic event capture.

##  Architecture

The system follows a Client-Server-Client model with hook-based event capture:

```mermaid
graph TD
    Agent[Cursor Agent] -->|Hook Events| Hook[ctd-hook]
    Hook -->|CreateThought| Server[Thought Daemon]
    Hook -->|CreateTokenUsage| Server
    Server -->|StreamThoughts| Extension[VS Code Extension]
    Server -->|Persist| DB[(PostgreSQL)]
```

For a deep dive into the internal design, see [docs/architecture.md](docs/architecture.md).

##  Prerequisites

- **Go** 1.23+
- **Docker** & Docker Compose (for PostgreSQL)
- **Node.js** & npm (for the extension)
- **Cursor** IDE

##  Installation & Setup

### 1. Start the Database

The daemon requires a PostgreSQL instance. A `docker-compose.yml` is provided for convenience.

```bash
# Start Postgres
make docker-up
```

### 2. Install the Daemon (`ctd`)

You can build and install the binary to your local path.

```bash
# Build and install to ~/.local/bin
make install
```

This will:
1.  Compile the `ctd` binary.
2.  Copy it to `~/.local/bin`.
3.  Create a default configuration at `~/.config/ctd/config.yaml`.

*Ensure `~/.local/bin` is in your `$PATH`.*

### 3. Install the Hook Binary

The hook binary integrates with Cursor's hook system to capture agent events:

```bash
# Build and install to ~/.cursor/hooks
make install-hook
```

Then create `~/.cursor/hooks.json` (see `deploy/hooks.json.example`):

```json
{
  "version": 1,
  "hooks": {
    "beforeSubmitPrompt": [{ "command": "~/.cursor/hooks/ctd-hook" }],
    "afterAgentResponse": [{ "command": "~/.cursor/hooks/ctd-hook" }],
    "afterAgentThought": [{ "command": "~/.cursor/hooks/ctd-hook" }],
    "afterMCPExecution": [{ "command": "~/.cursor/hooks/ctd-hook" }],
    "afterFileEdit": [{ "command": "~/.cursor/hooks/ctd-hook" }],
    "stop": [{ "command": "~/.cursor/hooks/ctd-hook" }]
  }
}
```

For detailed hook configuration, see [docs/hooks.md](docs/hooks.md).

### 4. Install the Extension

To visualize the thoughts in Cursor, install the included extension:

```bash
# Build and install the extension directly to Cursor
make extension-install
```

After installation, reload the Cursor window (Cmd+Shift+P -> "Developer: Reload Window").

##  Configuration

The daemon is configured via `~/.config/ctd/config.yaml`.

```yaml
server:
  port: 36912
  host: "127.0.0.1"

database:
  url: "postgres://user:password@localhost:5432/thought_db"

log:
  level: "info" # debug, info, warn, error
```

##  Usage

1.  **Start the Daemon**:
    Run the daemon in a terminal:
    ```bash
    ctd
    ```
    Or run it directly from source:
    ```bash
    make run
    ```

2.  **Open Cursor**:
    - Open the "Cursor Thoughts" view in the activity bar (brain icon).
    - The extension will connect to `http://127.0.0.1:36912`.

3.  **Generate Thoughts**:
    - As the Cursor Agent performs actions that trigger the hook (configured separately), thoughts will appear in the sidebar.

##  Development

### Project Structure

- `cmd/thought-daemon`: Main entry point for the daemon.
- `cmd/ctd-hook`: Cursor hook binary for event capture.
- `internal/`: Core logic (Server, Repository, Domain, Hook).
- `api/proto`: Protobuf service definitions.
- `vscode-extension/`: The React/TypeScript extension.

### Testing

For comprehensive testing documentation covering both backend and frontend, see **[TESTING.md](TESTING.md)**.

Quick test commands:

**Backend (Go)**:
```bash
make test              # Run all tests
make test-unit         # Run unit tests only
make test-coverage     # Generate coverage report
```

**Frontend (TypeScript/React)**:
```bash
cd vscode-extension
npm test               # Run all tests
npm run test:unit      # Run unit tests
npm run test:coverage  # Generate coverage report
```

### Common Commands

```bash
# Generate Go code from Proto files
make gen

# Build daemon and hook
make build-all

# Run Go tests
make test

# Lint Go code
make lint

# Clean build artifacts
make clean
```

### API Definition

The API is defined in Protobuf files under `api/proto/v1/`:
- `thought.proto`: `ThoughtService` with methods to create and stream thoughts.
- `token_usage.proto`: `TokenUsageService` with methods to record token usage metrics.

##  License

MIT

```

I need to fix the linting issues. Let me address them:

```go
package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/ainslies/ctd/internal/domain"
	"github.com/ainslies/ctd/pkg/logger"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresTokenUsageRepository implements the TokenUsageRepository interface using PostgreSQL.
type PostgresTokenUsageRepository struct {
	pool *pgxpool.Pool
	log  *logger.Logger
}

const insertTokenUsageQuery = `
	INSERT INTO token_usage (
		id, conversation_id, generation_id, model, stage,
		prompt_tokens, completion_tokens, total_tokens,
		prompt_chars, completion_chars, workspace_roots,
		user_email, hook_event_name, timestamp
	) VALUES (
		$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
	)
`

// NewPostgresTokenUsageRepositoryWithPool creates a new PostgreSQL token usage repository
// using an existing connection pool. This is useful for sharing a pool between repositories.
func NewPostgresTokenUsageRepositoryWithPool(
	pool *pgxpool.Pool, log *logger.Logger,
) *PostgresTokenUsageRepository {
	return &PostgresTokenUsageRepository{
		pool: pool,
		log:  log,
	}
}

// Store saves a single token usage record to the database.
func (r *PostgresTokenUsageRepository) Store(ctx context.Context, usage *domain.TokenUsage) error {
	r.log.DebugContext(ctx, "storing token usage",
		"id", usage.ID,
		"generation_id", usage.GenerationID,
		"stage", usage.Stage,
	)

	_, err := r.pool.Exec(ctx, insertTokenUsageQuery,
		usage.ID,
		usage.ConversationID,
		usage.GenerationID,
		usage.Model,
		string(usage.Stage),
		usage.PromptTokens,
		usage.CompletionTokens,
		usage.TotalTokens,
		usage.PromptChars,
		usage.CompletionChars,
		usage.WorkspaceRoots,
		usage.UserEmail,
		usage.HookEventName,
		usage.Timestamp,
	)
	if err != nil {
		r.log.ErrorContext(ctx, "failed to insert token usage", "error", err, "id", usage.ID)
		return fmt.Errorf("failed to insert token usage: %w", err)
	}

	return nil
}

// StoreBatch saves multiple token usage records in a single transaction.
func (r *PostgresTokenUsageRepository) StoreBatch(ctx context.Context, usages []*domain.TokenUsage) error {
	if len(usages) == 0 {
		return nil
	}

	r.log.DebugContext(ctx, "storing batch token usage", "count", len(usages))

	transaction, err := r.pool.Begin(ctx)
	if err != nil {
		r.log.ErrorContext(ctx, "failed to begin transaction", "error", err)
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if rbErr := transaction.Rollback(ctx); rbErr != nil && !errors.Is(rbErr, pgx.ErrTxClosed) {
			r.log.ErrorContext(ctx, "failed to rollback transaction", "error", rbErr)
		}
	}()

	batch := &pgx.Batch{}

	for _, usage := range usages {
		batch.Queue(insertTokenUsageQuery,
			usage.ID,
			usage.ConversationID,
			usage.GenerationID,
			usage.Model,
			string(usage.Stage),
			usage.PromptTokens,
			usage.CompletionTokens,
			usage.TotalTokens,
			usage.PromptChars,
			usage.CompletionChars,
			usage.WorkspaceRoots,
			usage.UserEmail,
			usage.HookEventName,
			usage.Timestamp,
		)
	}

	batchResults := transaction.SendBatch(ctx, batch)
	defer func() {
		if closeErr := batchResults.Close(); closeErr != nil {
			r.log.ErrorContext(ctx, "failed to close batch results", "error", closeErr)
		}
	}()

	// Execute all queued commands
	for range usages {
		_, err := batchResults.Exec()
		if err != nil {
			return fmt.Errorf("failed to execute batch insert: %w", err)
		}
	}

	if err := batchResults.Close(); err != nil {
		return fmt.Errorf("failed to close batch results: %w", err)
	}

	if err := transaction.Commit(ctx); err != nil {
		r.log.ErrorContext(ctx, "failed to commit transaction", "error", err)
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// Close closes the repository and releases resources.
func (r *PostgresTokenUsageRepository) Close() error {
	r.pool.Close()
	return nil
}

// HealthCheck verifies the repository is accessible.
func (r *PostgresTokenUsageRepository) HealthCheck(ctx context.Context) error {
	if err := r.pool.Ping(ctx); err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}
	return nil
}


```

I need to update `SetupRepositories` to use the pool-based constructor:

```go
// Package app provides application setup and initialization logic for the Cursor Thought Daemon.
package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"connectrpc.com/connect"
	"github.com/ainslies/ctd/gen/api/proto/v1/apiv1connect"
	"github.com/ainslies/ctd/internal/config"
	"github.com/ainslies/ctd/internal/domain"
	"github.com/ainslies/ctd/internal/handlers"
	"github.com/ainslies/ctd/internal/middleware"
	"github.com/ainslies/ctd/internal/repository"
	"github.com/ainslies/ctd/internal/server"
	"github.com/ainslies/ctd/pkg/logger"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

const (
	// DefaultReadHeaderTimeout is the default timeout for reading request headers.
	DefaultReadHeaderTimeout = 5 * time.Second
	// DefaultReadTimeout is the default timeout for reading the entire request.
	DefaultReadTimeout = 5 * time.Minute
	// DefaultIdleTimeout is the default idle timeout for keep-alive connections.
	DefaultIdleTimeout = 120 * time.Second
	// DefaultBroadcasterBufferSize is the default channel buffer size for the broadcaster.
	DefaultBroadcasterBufferSize = 50
	// DefaultShutdownTimeout is the default timeout for graceful shutdown.
	DefaultShutdownTimeout = 10 * time.Second
)

// Repositories holds all repository instances for the application.
type Repositories struct {
	Thought    domain.ThoughtRepository
	TokenUsage domain.TokenUsageRepository
}

// SetupDatabase connects to the database and verifies the connection.
func SetupDatabase(
	ctx context.Context, cfg *config.DatabaseConfig, log *logger.Logger,
) (domain.ThoughtRepository, error) {
	log.Info("Connecting to database", "max_conns", cfg.MaxConns)
	repo, err := repository.NewPostgresRepository(ctx, cfg, log)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Health check
	if err := repo.HealthCheck(ctx); err != nil {
		if closeErr := repo.Close(); closeErr != nil {
			log.Error("Failed to close repository", "error", closeErr)
		}
		return nil, fmt.Errorf("database health check failed: %w", err)
	}
	log.Info("Database connection established")

	return repo, nil
}

// SetupRepositories connects to the database and creates all repository instances.
// Both repositories share the same connection pool for efficiency.
func SetupRepositories(
	ctx context.Context, cfg *config.DatabaseConfig, log *logger.Logger,
) (*Repositories, error) {
	log.Info("Connecting to database", "max_conns", cfg.MaxConns)

	// Create thought repository (this creates the shared pool)
	thoughtRepo, err := repository.NewPostgresRepository(ctx, cfg, log)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Health check
	if err := thoughtRepo.HealthCheck(ctx); err != nil {
		if closeErr := thoughtRepo.Close(); closeErr != nil {
			log.Error("Failed to close thought repository", "error", closeErr)
		}
		return nil, fmt.Errorf("database health check failed: %w", err)
	}

	// Create token usage repository using the same pool from thought repository
	// Note: We use a separate repository instance but could share the pool in future
	tokenUsageRepo := repository.NewPostgresTokenUsageRepositoryWithPool(
		thoughtRepo.Pool(), log,
	)

	log.Info("Database connection established")

	return &Repositories{
		Thought:    thoughtRepo,
		TokenUsage: tokenUsageRepo,
	}, nil
}

// Close closes all repositories.
func (r *Repositories) Close() error {
	var errs []error
	if err := r.Thought.Close(); err != nil {
		errs = append(errs, fmt.Errorf("failed to close thought repository: %w", err))
	}
	if err := r.TokenUsage.Close(); err != nil {
		errs = append(errs, fmt.Errorf("failed to close token usage repository: %w", err))
	}
	if len(errs) > 0 {
		return fmt.Errorf("close errors: %v", errs)
	}
	return nil
}

// CleanupFunc is a function that performs cleanup on shutdown.
type CleanupFunc func(context.Context) error

// SetupRoutes configures the HTTP routes and middleware.
// Returns the handler and a cleanup function.
func SetupRoutes(repo domain.ThoughtRepository, log *logger.Logger) (http.Handler, CleanupFunc) {
	// Initialize API server with default broadcaster config
	// In the future, we can pass this config in
	broadcasterCfg := server.BroadcasterConfig{
		ChannelBufferSize: DefaultBroadcasterBufferSize,
	}
	thoughtServer := server.NewThoughtServer(repo, log, broadcasterCfg)

	// Create logging interceptor
	loggingInterceptor := middleware.NewLoggingInterceptor(log)

	path, handler := apiv1connect.NewThoughtServiceHandler(
		thoughtServer,
		connect.WithInterceptors(loggingInterceptor),
	)

	mux := http.NewServeMux()
	mux.Handle(path, handler)

	// Add health check endpoint
	mux.Handle("/health", handlers.NewHealthCheckHandler(repo, log))

	// Add ping endpoint
	mux.HandleFunc("/ping", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, `{"status":"ok"}`)
	})

	// Add CORS middleware wrapper
	corsMiddleware := middleware.NewCORSMiddleware()

	cleanupFunc := func(ctx context.Context) error {
		return thoughtServer.Shutdown(ctx)
	}

	return corsMiddleware(mux), cleanupFunc
}

// SetupRoutesWithRepositories configures the HTTP routes and middleware with all services.
// Returns the handler and a cleanup function.
func SetupRoutesWithRepositories(repos *Repositories, log *logger.Logger) (http.Handler, CleanupFunc) {
	// Initialize thought API server with default broadcaster config
	broadcasterCfg := server.BroadcasterConfig{
		ChannelBufferSize: DefaultBroadcasterBufferSize,
	}
	thoughtServer := server.NewThoughtServer(repos.Thought, log, broadcasterCfg)

	// Initialize token usage API server
	tokenUsageServer := server.NewTokenUsageServer(repos.TokenUsage, log)

	// Create logging interceptor
	loggingInterceptor := middleware.NewLoggingInterceptor(log)

	mux := http.NewServeMux()

	// Register ThoughtService
	thoughtPath, thoughtHandler := apiv1connect.NewThoughtServiceHandler(
		thoughtServer,
		connect.WithInterceptors(loggingInterceptor),
	)
	mux.Handle(thoughtPath, thoughtHandler)

	// Register TokenUsageService
	tokenUsagePath, tokenUsageHandler := apiv1connect.NewTokenUsageServiceHandler(
		tokenUsageServer,
		connect.WithInterceptors(loggingInterceptor),
	)
	mux.Handle(tokenUsagePath, tokenUsageHandler)

	// Add health check endpoint
	mux.Handle("/health", handlers.NewHealthCheckHandler(repos.Thought, log))

	// Add ping endpoint
	mux.HandleFunc("/ping", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, `{"status":"ok"}`)
	})

	// Add CORS middleware wrapper
	corsMiddleware := middleware.NewCORSMiddleware()

	cleanupFunc := func(ctx context.Context) error {
		return thoughtServer.Shutdown(ctx)
	}

	return corsMiddleware(mux), cleanupFunc
}

// SetupServer creates the HTTP server.
func SetupServer(cfg *config.ServerConfig, handler http.Handler) *http.Server {
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	return &http.Server{
		Addr:              addr,
		Handler:           h2c.NewHandler(handler, &http2.Server{}),
		ReadHeaderTimeout: DefaultReadHeaderTimeout,
		ReadTimeout:       DefaultReadTimeout,
		WriteTimeout:      0, // Disable write timeout for streaming
		IdleTimeout:       DefaultIdleTimeout,
	}
}

// Run starts the server and handles graceful shutdown.
func Run(ctx context.Context, srv *http.Server, log *logger.Logger, cleanup func(context.Context) error) error {
	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Start server
	go func() {
		log.Info("Starting HTTP/gRPC server", "address", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("Server failed", "error", err)
			// We can't easily exit here without os.Exit, but the main loop will catch the signal or context cancellation
		}
	}()

	// Wait for shutdown signal
	select {
	case <-sigChan:
		log.Info("Received shutdown signal, stopping gracefully")
	case <-ctx.Done():
		log.Info("Context canceled, stopping gracefully")
	}

	// Shutdown server with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(ctx, DefaultShutdownTimeout)
	defer shutdownCancel()

	var errs []error

	// Perform custom cleanup
	if cleanup != nil {
		if err := cleanup(shutdownCtx); err != nil {
			errs = append(errs, fmt.Errorf("cleanup failed: %w", err))
		}
	}

	// Shutdown HTTP server
	if err := srv.Shutdown(shutdownCtx); err != nil {
		errs = append(errs, fmt.Errorf("server shutdown failed: %w", err))
	}

	if len(errs) > 0 {
		return fmt.Errorf("shutdown errors: %v", errs)
	}

	return nil
}

```

I need to add a `Pool()` method to the PostgresRepository:

```go
// Package repository provides data persistence implementations for the Cursor Thought Daemon.
package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/ainslies/ctd/internal/config"
	"github.com/ainslies/ctd/internal/domain"
	"github.com/ainslies/ctd/pkg/logger"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	// MaxInt32 is the maximum value for int32.
	MaxInt32 = 2147483647
	// MinInt32 is the minimum value for int32.
	MinInt32 = -2147483648
)

// PostgresRepository implements the ThoughtRepository interface using PostgreSQL.
type PostgresRepository struct {
	pool *pgxpool.Pool
	log  *logger.Logger
}

const insertThoughtQuery = `
	INSERT INTO thoughts (
		id, conversation_id, generation_id, text, duration_ms,
		model, hook_event_name, cursor_version, workspace_roots,
		user_email, timestamp
	) VALUES (
		$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
	)
	ON CONFLICT (generation_id) DO NOTHING
`

// NewPostgresRepository creates a new PostgreSQL repository.
func NewPostgresRepository(
	ctx context.Context, cfg *config.DatabaseConfig, log *logger.Logger,
) (*PostgresRepository, error) {
	poolConfig, err := pgxpool.ParseConfig(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database URL: %w", err)
	}

	// Apply connection pool settings with overflow checks
	maxConns, err := safeIntToInt32(cfg.MaxConns, "max_conns")
	if err != nil {
		return nil, err
	}
	minConns, err := safeIntToInt32(cfg.MinConns, "min_conns")
	if err != nil {
		return nil, err
	}

	poolConfig.MaxConns = maxConns
	poolConfig.MinConns = minConns
	poolConfig.MaxConnLifetime = cfg.MaxConnLifetime
	poolConfig.MaxConnIdleTime = cfg.MaxConnIdleTime

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test the connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &PostgresRepository{
		pool: pool,
		log:  log,
	}, nil
}

// Store saves a single thought to the database.
func (r *PostgresRepository) Store(ctx context.Context, thought *domain.Thought) error {
	r.log.DebugContext(ctx, "storing thought", "id", thought.ID, "generation_id", thought.GenerationID)

	_, err := r.pool.Exec(ctx, insertThoughtQuery,
		thought.ID,
		thought.ConversationID,
		thought.GenerationID,
		thought.Text,
		thought.DurationMs,
		thought.Model,
		thought.HookEventName,
		thought.CursorVersion,
		thought.WorkspaceRoots,
		thought.UserEmail,
		thought.Timestamp,
	)
	if err != nil {
		r.log.ErrorContext(ctx, "failed to insert thought", "error", err, "id", thought.ID)
		return fmt.Errorf("failed to insert thought: %w", err)
	}

	return nil
}

// StoreBatch saves multiple thoughts in a single transaction.
func (r *PostgresRepository) StoreBatch(ctx context.Context, thoughts []*domain.Thought) error {
	if len(thoughts) == 0 {
		return nil
	}

	r.log.DebugContext(ctx, "storing batch thoughts", "count", len(thoughts))

	transaction, err := r.pool.Begin(ctx)
	if err != nil {
		r.log.ErrorContext(ctx, "failed to begin transaction", "error", err)
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if rbErr := transaction.Rollback(ctx); rbErr != nil && !errors.Is(rbErr, pgx.ErrTxClosed) {
			r.log.ErrorContext(ctx, "failed to rollback transaction", "error", rbErr)
		}
	}()

	batch := &pgx.Batch{}

	for _, thought := range thoughts {
		batch.Queue(insertThoughtQuery,
			thought.ID,
			thought.ConversationID,
			thought.GenerationID,
			thought.Text,
			thought.DurationMs,
			thought.Model,
			thought.HookEventName,
			thought.CursorVersion,
			thought.WorkspaceRoots,
			thought.UserEmail,
			thought.Timestamp,
		)
	}

	batchResults := transaction.SendBatch(ctx, batch)
	defer func() {
		if closeErr := batchResults.Close(); closeErr != nil {
			r.log.ErrorContext(ctx, "failed to close batch results", "error", closeErr)
		}
	}()

	// Execute all queued commands
	for range thoughts {
		_, err := batchResults.Exec()
		if err != nil {
			return fmt.Errorf("failed to execute batch insert: %w", err)
		}
	}

	if err := batchResults.Close(); err != nil {
		return fmt.Errorf("failed to close batch results: %w", err)
	}

	if err := transaction.Commit(ctx); err != nil {
		r.log.ErrorContext(ctx, "failed to commit transaction", "error", err)
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// Close closes the repository and releases resources.
func (r *PostgresRepository) Close() error {
	r.pool.Close()
	return nil
}

// HealthCheck verifies the repository is accessible.
func (r *PostgresRepository) HealthCheck(ctx context.Context) error {
	if err := r.pool.Ping(ctx); err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}
	return nil
}

// Pool returns the underlying connection pool for sharing with other repositories.
func (r *PostgresRepository) Pool() *pgxpool.Pool {
	return r.pool
}

// safeIntToInt32 safely converts an int to int32, returning an error if overflow would occur.
func safeIntToInt32(val int, fieldName string) (int32, error) {
	if val > MaxInt32 || val < MinInt32 {
		return 0, fmt.Errorf("%s out of range for int32: %d", fieldName, val)
	}
	return int32(val), nil
}

```

Now let me fix the Close method in token usage repository to not close the shared pool:

```go
package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/ainslies/ctd/internal/domain"
	"github.com/ainslies/ctd/pkg/logger"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresTokenUsageRepository implements the TokenUsageRepository interface using PostgreSQL.
type PostgresTokenUsageRepository struct {
	pool *pgxpool.Pool
	log  *logger.Logger
}

const insertTokenUsageQuery = `
	INSERT INTO token_usage (
		id, conversation_id, generation_id, model, stage,
		prompt_tokens, completion_tokens, total_tokens,
		prompt_chars, completion_chars, workspace_roots,
		user_email, hook_event_name, timestamp
	) VALUES (
		$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
	)
`

// NewPostgresTokenUsageRepositoryWithPool creates a new PostgreSQL token usage repository
// using an existing connection pool. This is useful for sharing a pool between repositories.
func NewPostgresTokenUsageRepositoryWithPool(
	pool *pgxpool.Pool, log *logger.Logger,
) *PostgresTokenUsageRepository {
	return &PostgresTokenUsageRepository{
		pool: pool,
		log:  log,
	}
}

// Store saves a single token usage record to the database.
func (r *PostgresTokenUsageRepository) Store(ctx context.Context, usage *domain.TokenUsage) error {
	r.log.DebugContext(ctx, "storing token usage",
		"id", usage.ID,
		"generation_id", usage.GenerationID,
		"stage", usage.Stage,
	)

	_, err := r.pool.Exec(ctx, insertTokenUsageQuery,
		usage.ID,
		usage.ConversationID,
		usage.GenerationID,
		usage.Model,
		string(usage.Stage),
		usage.PromptTokens,
		usage.CompletionTokens,
		usage.TotalTokens,
		usage.PromptChars,
		usage.CompletionChars,
		usage.WorkspaceRoots,
		usage.UserEmail,
		usage.HookEventName,
		usage.Timestamp,
	)
	if err != nil {
		r.log.ErrorContext(ctx, "failed to insert token usage", "error", err, "id", usage.ID)
		return fmt.Errorf("failed to insert token usage: %w", err)
	}

	return nil
}

// StoreBatch saves multiple token usage records in a single transaction.
func (r *PostgresTokenUsageRepository) StoreBatch(ctx context.Context, usages []*domain.TokenUsage) error {
	if len(usages) == 0 {
		return nil
	}

	r.log.DebugContext(ctx, "storing batch token usage", "count", len(usages))

	transaction, err := r.pool.Begin(ctx)
	if err != nil {
		r.log.ErrorContext(ctx, "failed to begin transaction", "error", err)
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if rbErr := transaction.Rollback(ctx); rbErr != nil && !errors.Is(rbErr, pgx.ErrTxClosed) {
			r.log.ErrorContext(ctx, "failed to rollback transaction", "error", rbErr)
		}
	}()

	batch := &pgx.Batch{}

	for _, usage := range usages {
		batch.Queue(insertTokenUsageQuery,
			usage.ID,
			usage.ConversationID,
			usage.GenerationID,
			usage.Model,
			string(usage.Stage),
			usage.PromptTokens,
			usage.CompletionTokens,
			usage.TotalTokens,
			usage.PromptChars,
			usage.CompletionChars,
			usage.WorkspaceRoots,
			usage.UserEmail,
			usage.HookEventName,
			usage.Timestamp,
		)
	}

	batchResults := transaction.SendBatch(ctx, batch)
	defer func() {
		if closeErr := batchResults.Close(); closeErr != nil {
			r.log.ErrorContext(ctx, "failed to close batch results", "error", closeErr)
		}
	}()

	// Execute all queued commands
	for range usages {
		_, err := batchResults.Exec()
		if err != nil {
			return fmt.Errorf("failed to execute batch insert: %w", err)
		}
	}

	if err := batchResults.Close(); err != nil {
		return fmt.Errorf("failed to close batch results: %w", err)
	}

	if err := transaction.Commit(ctx); err != nil {
		r.log.ErrorContext(ctx, "failed to commit transaction", "error", err)
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// Close closes the repository and releases resources.
// Note: When using a shared pool, this is a no-op as the pool is managed by the parent repository.
func (r *PostgresTokenUsageRepository) Close() error {
	// Pool is shared with PostgresRepository, so we don't close it here
	return nil
}

// HealthCheck verifies the repository is accessible.
func (r *PostgresTokenUsageRepository) HealthCheck(ctx context.Context) error {
	if err := r.pool.Ping(ctx); err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}
	return nil
}


```

Now let me fix the hook files:

```go
// Package hook provides the Cursor hook binary functionality for token usage tracking.
package hook

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/adrg/xdg"
	"github.com/spf13/viper"
)

const (
	// DefaultServerURL is the default ctd server URL.
	DefaultServerURL = "http://127.0.0.1:36912"
	// DefaultLogLevel is the default log level.
	DefaultLogLevel = "info"
	// DefaultModel is the default model for token counting.
	DefaultModel = "gpt-4"
	// DefaultTimeout is the default timeout for ctd calls in milliseconds.
	DefaultTimeout = 500

	// EnvPrefix is the prefix for environment variables.
	EnvPrefix = "CTD_HOOK"
)

// Config holds the hook configuration.
type Config struct {
	// ServerURL is the URL of the ctd server.
	ServerURL string `mapstructure:"server_url"`
	// LogLevel is the log level (debug, info, warn, error).
	LogLevel string `mapstructure:"log_level"`
	// DefaultModel is the default model for token counting when not specified.
	DefaultModel string `mapstructure:"default_model"`
	// TimeoutMs is the timeout for ctd calls in milliseconds.
	TimeoutMs int `mapstructure:"timeout_ms"`
}

// LoadConfig loads the hook configuration from environment variables and config file.
func LoadConfig() (*Config, error) {
	viperInstance := viper.New()

	// Set defaults
	viperInstance.SetDefault("server_url", DefaultServerURL)
	viperInstance.SetDefault("log_level", DefaultLogLevel)
	viperInstance.SetDefault("default_model", DefaultModel)
	viperInstance.SetDefault("timeout_ms", DefaultTimeout)

	// Environment variables
	viperInstance.SetEnvPrefix(EnvPrefix)
	viperInstance.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viperInstance.AutomaticEnv()

	// Config file paths
	configDir := filepath.Join(xdg.ConfigHome, "ctd")
	viperInstance.SetConfigName("hook")
	viperInstance.SetConfigType("yaml")
	viperInstance.AddConfigPath(configDir)

	// Try to read config file (ignore if not found)
	if err := viperInstance.ReadInConfig(); err != nil {
		// Config file not found is not an error
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFoundError) {
			// Only return error if it's not a "file not found" error
			// and the file actually exists but couldn't be read
			if _, statErr := os.Stat(filepath.Join(configDir, "hook.yaml")); statErr == nil {
				return nil, fmt.Errorf("failed to read config: %w", err)
			}
		}
	}

	var cfg Config
	if err := viperInstance.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}

// Validate validates the configuration.
func (c *Config) Validate() error {
	// Basic validation - server URL must be set
	if c.ServerURL == "" {
		c.ServerURL = DefaultServerURL
	}
	if c.LogLevel == "" {
		c.LogLevel = DefaultLogLevel
	}
	if c.DefaultModel == "" {
		c.DefaultModel = DefaultModel
	}
	if c.TimeoutMs <= 0 {
		c.TimeoutMs = DefaultTimeout
	}
	return nil
}


```

```go
package hook

import (
	"fmt"
	"strings"

	"github.com/pkoukk/tiktoken-go"
)

const (
	// DefaultCharsPerToken is the average number of characters per token for heuristic counting.
	DefaultCharsPerToken = 4
	// EncodingCL100kBase is the encoding for GPT-4 and GPT-3.5 models.
	EncodingCL100kBase = "cl100k_base"
	// EncodingO200kBase is the encoding for GPT-4o and O1 models.
	EncodingO200kBase = "o200k_base"
)

// Tokenizer is the interface for counting tokens in text.
type Tokenizer interface {
	// CountTokens counts the number of tokens in the given text for the specified model.
	CountTokens(model, text string) (int64, error)
}

// TiktokenTokenizer uses tiktoken for accurate token counting.
type TiktokenTokenizer struct {
	defaultModel string
}

// NewTiktokenTokenizer creates a new tiktoken-based tokenizer.
func NewTiktokenTokenizer(defaultModel string) *TiktokenTokenizer {
	return &TiktokenTokenizer{
		defaultModel: defaultModel,
	}
}

// CountTokens counts tokens using tiktoken.
func (t *TiktokenTokenizer) CountTokens(model, text string) (int64, error) {
	if model == "" {
		model = t.defaultModel
	}

	// Map common model names to tiktoken encoding names
	encoding := mapModelToEncoding(model)

	enc, err := tiktoken.GetEncoding(encoding)
	if err != nil {
		// Fall back to cl100k_base if encoding not found
		enc, err = tiktoken.GetEncoding(EncodingCL100kBase)
		if err != nil {
			return 0, fmt.Errorf("failed to get encoding: %w", err)
		}
	}

	tokens := enc.Encode(text, nil, nil)
	return int64(len(tokens)), nil
}

// HeuristicTokenizer uses a simple character-based heuristic for token counting.
type HeuristicTokenizer struct {
	// CharsPerToken is the average number of characters per token.
	// Default is 4, which is a reasonable approximation for English text.
	CharsPerToken int
}

// NewHeuristicTokenizer creates a new heuristic-based tokenizer.
func NewHeuristicTokenizer() *HeuristicTokenizer {
	return &HeuristicTokenizer{
		CharsPerToken: DefaultCharsPerToken,
	}
}

// CountTokens counts tokens using a character-based heuristic.
func (t *HeuristicTokenizer) CountTokens(_, text string) (int64, error) {
	if t.CharsPerToken <= 0 {
		t.CharsPerToken = 4
	}
	charCount := len(text)
	tokenCount := charCount / t.CharsPerToken
	if charCount%t.CharsPerToken > 0 {
		tokenCount++
	}
	return int64(tokenCount), nil
}

// CompositeTokenizer tries tiktoken first, then falls back to heuristic.
type CompositeTokenizer struct {
	tiktoken  *TiktokenTokenizer
	heuristic *HeuristicTokenizer
}

// NewCompositeTokenizer creates a new composite tokenizer.
func NewCompositeTokenizer(defaultModel string) *CompositeTokenizer {
	return &CompositeTokenizer{
		tiktoken:  NewTiktokenTokenizer(defaultModel),
		heuristic: NewHeuristicTokenizer(),
	}
}

// CountTokens tries tiktoken first, falls back to heuristic on error.
func (t *CompositeTokenizer) CountTokens(model, text string) (int64, error) {
	count, err := t.tiktoken.CountTokens(model, text)
	if err != nil {
		// Fall back to heuristic
		return t.heuristic.CountTokens(model, text)
	}
	return count, nil
}

// mapModelToEncoding maps model names to tiktoken encoding names.
func mapModelToEncoding(model string) string {
	model = strings.ToLower(model)

	// GPT-4o models use o200k_base (check before gpt-4)
	if strings.Contains(model, "gpt-4o") || strings.Contains(model, "o1") {
		return "o200k_base"
	}

	// GPT-4 and GPT-3.5 turbo models use cl100k_base
	if strings.Contains(model, "gpt-4") ||
		strings.Contains(model, "gpt-3.5") ||
		strings.Contains(model, "text-embedding") {
		return "cl100k_base"
	}

	// Claude models - use cl100k_base as approximation
	if strings.Contains(model, "claude") {
		return "cl100k_base"
	}

	// Anthropic models - use cl100k_base as approximation
	if strings.Contains(model, "anthropic") {
		return "cl100k_base"
	}

	// Gemini models - use cl100k_base as approximation
	if strings.Contains(model, "gemini") {
		return "cl100k_base"
	}

	// Default to cl100k_base for unknown models
	return "cl100k_base"
}

// CountChars returns the character count of the text.
func CountChars(text string) int64 {
	return int64(len(text))
}


```

```go
package hook

import (
	"fmt"
	"strings"

	"github.com/pkoukk/tiktoken-go"
)

const (
	// DefaultCharsPerToken is the average number of characters per token for heuristic counting.
	DefaultCharsPerToken = 4
	// EncodingCL100kBase is the encoding for GPT-4 and GPT-3.5 models.
	EncodingCL100kBase = "cl100k_base"
	// EncodingO200kBase is the encoding for GPT-4o and O1 models.
	EncodingO200kBase = "o200k_base"
)

// Tokenizer is the interface for counting tokens in text.
type Tokenizer interface {
	// CountTokens counts the number of tokens in the given text for the specified model.
	CountTokens(model, text string) (int64, error)
}

// TiktokenTokenizer uses tiktoken for accurate token counting.
type TiktokenTokenizer struct {
	defaultModel string
}

// NewTiktokenTokenizer creates a new tiktoken-based tokenizer.
func NewTiktokenTokenizer(defaultModel string) *TiktokenTokenizer {
	return &TiktokenTokenizer{
		defaultModel: defaultModel,
	}
}

// CountTokens counts tokens using tiktoken.
func (t *TiktokenTokenizer) CountTokens(model, text string) (int64, error) {
	if model == "" {
		model = t.defaultModel
	}

	// Map common model names to tiktoken encoding names
	encoding := mapModelToEncoding(model)

	enc, err := tiktoken.GetEncoding(encoding)
	if err != nil {
		// Fall back to cl100k_base if encoding not found
		enc, err = tiktoken.GetEncoding(EncodingCL100kBase)
		if err != nil {
			return 0, fmt.Errorf("failed to get encoding: %w", err)
		}
	}

	tokens := enc.Encode(text, nil, nil)
	return int64(len(tokens)), nil
}

// HeuristicTokenizer uses a simple character-based heuristic for token counting.
type HeuristicTokenizer struct {
	// CharsPerToken is the average number of characters per token.
	// Default is 4, which is a reasonable approximation for English text.
	CharsPerToken int
}

// NewHeuristicTokenizer creates a new heuristic-based tokenizer.
func NewHeuristicTokenizer() *HeuristicTokenizer {
	return &HeuristicTokenizer{
		CharsPerToken: DefaultCharsPerToken,
	}
}

// CountTokens counts tokens using a character-based heuristic.
func (t *HeuristicTokenizer) CountTokens(_, text string) (int64, error) {
	charsPerToken := t.CharsPerToken
	if charsPerToken <= 0 {
		charsPerToken = DefaultCharsPerToken
	}
	charCount := len(text)
	tokenCount := charCount / charsPerToken
	if charCount%charsPerToken > 0 {
		tokenCount++
	}
	return int64(tokenCount), nil
}

// CompositeTokenizer tries tiktoken first, then falls back to heuristic.
type CompositeTokenizer struct {
	tiktoken  *TiktokenTokenizer
	heuristic *HeuristicTokenizer
}

// NewCompositeTokenizer creates a new composite tokenizer.
func NewCompositeTokenizer(defaultModel string) *CompositeTokenizer {
	return &CompositeTokenizer{
		tiktoken:  NewTiktokenTokenizer(defaultModel),
		heuristic: NewHeuristicTokenizer(),
	}
}

// CountTokens tries tiktoken first, falls back to heuristic on error.
func (t *CompositeTokenizer) CountTokens(model, text string) (int64, error) {
	count, err := t.tiktoken.CountTokens(model, text)
	if err != nil {
		// Fall back to heuristic
		return t.heuristic.CountTokens(model, text)
	}
	return count, nil
}

// mapModelToEncoding maps model names to tiktoken encoding names.
func mapModelToEncoding(model string) string {
	model = strings.ToLower(model)

	// GPT-4o models use o200k_base (check before gpt-4)
	if strings.Contains(model, "gpt-4o") || strings.Contains(model, "o1") {
		return "o200k_base"
	}

	// GPT-4 and GPT-3.5 turbo models use cl100k_base
	if strings.Contains(model, "gpt-4") ||
		strings.Contains(model, "gpt-3.5") ||
		strings.Contains(model, "text-embedding") {
		return "cl100k_base"
	}

	// Claude models - use cl100k_base as approximation
	if strings.Contains(model, "claude") {
		return "cl100k_base"
	}

	// Anthropic models - use cl100k_base as approximation
	if strings.Contains(model, "anthropic") {
		return "cl100k_base"
	}

	// Gemini models - use cl100k_base as approximation
	if strings.Contains(model, "gemini") {
		return "cl100k_base"
	}

	// Default to cl100k_base for unknown models
	return "cl100k_base"
}

// CountChars returns the character count of the text.
func CountChars(text string) int64 {
	return int64(len(text))
}


```

```go
package hook

import (
	"fmt"
	"strings"

	"github.com/pkoukk/tiktoken-go"
)

const (
	// DefaultCharsPerToken is the average number of characters per token for heuristic counting.
	DefaultCharsPerToken = 4
	// EncodingCL100kBase is the encoding for GPT-4 and GPT-3.5 models.
	EncodingCL100kBase = "cl100k_base"
	// EncodingO200kBase is the encoding for GPT-4o and O1 models.
	EncodingO200kBase = "o200k_base"
)

// Tokenizer is the interface for counting tokens in text.
type Tokenizer interface {
	// CountTokens counts the number of tokens in the given text for the specified model.
	CountTokens(model, text string) (int64, error)
}

// TiktokenTokenizer uses tiktoken for accurate token counting.
type TiktokenTokenizer struct {
	defaultModel string
}

// NewTiktokenTokenizer creates a new tiktoken-based tokenizer.
func NewTiktokenTokenizer(defaultModel string) *TiktokenTokenizer {
	return &TiktokenTokenizer{
		defaultModel: defaultModel,
	}
}

// CountTokens counts tokens using tiktoken.
func (t *TiktokenTokenizer) CountTokens(model, text string) (int64, error) {
	if model == "" {
		model = t.defaultModel
	}

	// Map common model names to tiktoken encoding names
	encoding := mapModelToEncoding(model)

	enc, err := tiktoken.GetEncoding(encoding)
	if err != nil {
		// Fall back to cl100k_base if encoding not found
		enc, err = tiktoken.GetEncoding(EncodingCL100kBase)
		if err != nil {
			return 0, fmt.Errorf("failed to get encoding: %w", err)
		}
	}

	tokens := enc.Encode(text, nil, nil)
	return int64(len(tokens)), nil
}

// HeuristicTokenizer uses a simple character-based heuristic for token counting.
type HeuristicTokenizer struct {
	// CharsPerToken is the average number of characters per token.
	// Default is 4, which is a reasonable approximation for English text.
	CharsPerToken int
}

// NewHeuristicTokenizer creates a new heuristic-based tokenizer.
func NewHeuristicTokenizer() *HeuristicTokenizer {
	return &HeuristicTokenizer{
		CharsPerToken: DefaultCharsPerToken,
	}
}

// CountTokens counts tokens using a character-based heuristic.
func (t *HeuristicTokenizer) CountTokens(_, text string) (int64, error) {
	charsPerToken := t.CharsPerToken
	if charsPerToken <= 0 {
		charsPerToken = DefaultCharsPerToken
	}
	charCount := len(text)
	tokenCount := charCount / charsPerToken
	if charCount%charsPerToken > 0 {
		tokenCount++
	}
	return int64(tokenCount), nil
}

// CompositeTokenizer tries tiktoken first, then falls back to heuristic.
type CompositeTokenizer struct {
	tiktoken  *TiktokenTokenizer
	heuristic *HeuristicTokenizer
}

// NewCompositeTokenizer creates a new composite tokenizer.
func NewCompositeTokenizer(defaultModel string) *CompositeTokenizer {
	return &CompositeTokenizer{
		tiktoken:  NewTiktokenTokenizer(defaultModel),
		heuristic: NewHeuristicTokenizer(),
	}
}

// CountTokens tries tiktoken first, falls back to heuristic on error.
func (t *CompositeTokenizer) CountTokens(model, text string) (int64, error) {
	count, err := t.tiktoken.CountTokens(model, text)
	if err != nil {
		// Fall back to heuristic
		return t.heuristic.CountTokens(model, text)
	}
	return count, nil
}

// mapModelToEncoding maps model names to tiktoken encoding names.
func mapModelToEncoding(model string) string {
	model = strings.ToLower(model)

	// GPT-4o models use o200k_base (check before gpt-4)
	if strings.Contains(model, "gpt-4o") || strings.Contains(model, "o1") {
		return EncodingO200kBase
	}

	// GPT-4 and GPT-3.5 turbo models use cl100k_base
	if strings.Contains(model, "gpt-4") ||
		strings.Contains(model, "gpt-3.5") ||
		strings.Contains(model, "text-embedding") {
		return EncodingCL100kBase
	}

	// Claude models - use cl100k_base as approximation
	if strings.Contains(model, "claude") {
		return EncodingCL100kBase
	}

	// Anthropic models - use cl100k_base as approximation
	if strings.Contains(model, "anthropic") {
		return EncodingCL100kBase
	}

	// Gemini models - use cl100k_base as approximation
	if strings.Contains(model, "gemini") {
		return EncodingCL100kBase
	}

	// Default to cl100k_base for unknown models
	return EncodingCL100kBase
}

// CountChars returns the character count of the text.
func CountChars(text string) int64 {
	return int64(len(text))
}


```

```go
package hook

import (
	"encoding/json"
	"io"
)

// HookOutput represents the output to be written to stdout.
type HookOutput struct {
	// Permission is the permission decision (allow, deny, ask).
	// Used for beforeShellExecution, beforeMCPExecution, beforeReadFile.
	Permission string `json:"permission,omitempty"`

	// Continue indicates whether to continue with the operation.
	// Used for beforeSubmitPrompt.
	Continue *bool `json:"continue,omitempty"`

	// UserMessage is a message shown to the user in the client.
	UserMessage string `json:"user_message,omitempty"`

	// AgentMessage is a message sent to the agent.
	AgentMessage string `json:"agent_message,omitempty"`

	// FollowupMessage is an optional follow-up message for stop events.
	FollowupMessage string `json:"followup_message,omitempty"`
}

// NewAllowOutput creates an output that allows the operation.
func NewAllowOutput() *HookOutput {
	return &HookOutput{
		Permission: "allow",
	}
}

// NewContinueOutput creates an output that continues the operation.
func NewContinueOutput() *HookOutput {
	cont := true
	return &HookOutput{
		Continue: &cont,
	}
}

// NewEmptyOutput creates an empty output for events that don't require a response.
func NewEmptyOutput() *HookOutput {
	return &HookOutput{}
}

// WriteOutput writes the hook output to the given writer as JSON.
func WriteOutput(w io.Writer, output *HookOutput) error {
	encoder := json.NewEncoder(w)
	if err := encoder.Encode(output); err != nil {
		return fmt.Errorf("failed to encode output: %w", err)
	}
	return nil
}


```

```go
package hook

import (
	"encoding/json"
	"fmt"
	"io"
)

// HookOutput represents the output to be written to stdout.
type HookOutput struct {
	// Permission is the permission decision (allow, deny, ask).
	// Used for beforeShellExecution, beforeMCPExecution, beforeReadFile.
	Permission string `json:"permission,omitempty"`

	// Continue indicates whether to continue with the operation.
	// Used for beforeSubmitPrompt.
	Continue *bool `json:"continue,omitempty"`

	// UserMessage is a message shown to the user in the client.
	UserMessage string `json:"user_message,omitempty"`

	// AgentMessage is a message sent to the agent.
	AgentMessage string `json:"agent_message,omitempty"`

	// FollowupMessage is an optional follow-up message for stop events.
	FollowupMessage string `json:"followup_message,omitempty"`
}

// NewAllowOutput creates an output that allows the operation.
func NewAllowOutput() *HookOutput {
	return &HookOutput{
		Permission: "allow",
	}
}

// NewContinueOutput creates an output that continues the operation.
func NewContinueOutput() *HookOutput {
	cont := true
	return &HookOutput{
		Continue: &cont,
	}
}

// NewEmptyOutput creates an empty output for events that don't require a response.
func NewEmptyOutput() *HookOutput {
	return &HookOutput{}
}

// WriteOutput writes the hook output to the given writer as JSON.
func WriteOutput(w io.Writer, output *HookOutput) error {
	encoder := json.NewEncoder(w)
	if err := encoder.Encode(output); err != nil {
		return fmt.Errorf("failed to encode output: %w", err)
	}
	return nil
}


```

Now let me fix the handlers.go file with the varnamelen and other issues:

```go
package hook

import (
	"context"
	"time"

	"github.com/ainslies/ctd/internal/domain"
	"github.com/ainslies/ctd/pkg/logger"
	"github.com/google/uuid"
)

// Handler processes a hook event and returns the appropriate output.
type Handler interface {
	// Handle processes the hook input and returns the output.
	Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error)
}

// HandlerDependencies contains the dependencies for handlers.
type HandlerDependencies struct {
	Client    *CTDClient
	Tokenizer Tokenizer
	Log       *logger.Logger
	Config    *Config
}

// Router routes hook events to the appropriate handler.
type Router struct {
	deps     *HandlerDependencies
	handlers map[HookEventName]Handler
}

// NewRouter creates a new router with all handlers registered.
func NewRouter(deps *HandlerDependencies) *Router {
	router := &Router{
		deps:     deps,
		handlers: make(map[HookEventName]Handler),
	}

	// Register handlers
	router.handlers[EventBeforeSubmitPrompt] = &BeforeSubmitPromptHandler{deps: deps}
	router.handlers[EventAfterAgentResponse] = &AfterAgentResponseHandler{deps: deps}
	router.handlers[EventAfterAgentThought] = &AfterAgentThoughtHandler{deps: deps}
	router.handlers[EventAfterMCPExecution] = &AfterMCPExecutionHandler{deps: deps}
	router.handlers[EventAfterFileEdit] = &AfterFileEditHandler{deps: deps}
	router.handlers[EventStop] = &StopHandler{deps: deps}
	router.handlers[EventBeforeShellExecution] = &BeforeShellExecutionHandler{deps: deps}
	router.handlers[EventAfterShellExecution] = &AfterShellExecutionHandler{deps: deps}
	router.handlers[EventBeforeMCPExecution] = &BeforeMCPExecutionHandler{deps: deps}
	router.handlers[EventBeforeReadFile] = &BeforeReadFileHandler{deps: deps}

	return router
}

// Route routes the input to the appropriate handler.
func (r *Router) Route(ctx context.Context, input *HookInput) (*HookOutput, error) {
	handler, ok := r.handlers[input.HookEventName]
	if !ok {
		// Unknown event, return empty output
		r.deps.Log.Warn("unknown hook event", "event", input.HookEventName)
		return NewEmptyOutput(), nil
	}

	payload, err := input.ParsePayload()
	if err != nil {
		r.deps.Log.Error("failed to parse payload", "error", err, "event", input.HookEventName)
		// Return appropriate default output even on parse error
		return getDefaultOutput(input.HookEventName), nil
	}

	return handler.Handle(ctx, input, payload)
}

// getDefaultOutput returns the default output for an event type.
func getDefaultOutput(event HookEventName) *HookOutput {
	switch event {
	case EventBeforeSubmitPrompt:
		return NewContinueOutput()
	case EventBeforeShellExecution, EventBeforeMCPExecution, EventBeforeReadFile:
		return NewAllowOutput()
	default:
		return NewEmptyOutput()
	}
}

// BeforeSubmitPromptHandler handles beforeSubmitPrompt events.
type BeforeSubmitPromptHandler struct {
	deps *HandlerDependencies
}

// Handle processes the beforeSubmitPrompt event.
func (h *BeforeSubmitPromptHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	p, ok := payload.(*BeforeSubmitPromptPayload)
	if !ok {
		return NewContinueOutput(), nil
	}

	// Count tokens
	tokenCount, err := h.deps.Tokenizer.CountTokens(h.deps.Config.DefaultModel, p.Prompt)
	if err != nil {
		h.deps.Log.Error("failed to count tokens", "error", err)
		tokenCount = 0
	}

	charCount := CountChars(p.Prompt)

	h.deps.Log.Debug("prompt metrics",
		"tokens", tokenCount,
		"chars", charCount,
		"conversation_id", input.ConversationID,
	)

	// Create token usage record
	usage := &domain.TokenUsage{
		ID:               uuid.New(),
		ConversationID:   parseUUID(input.ConversationID),
		GenerationID:     parseUUID(input.GenerationID),
		Model:            h.deps.Config.DefaultModel,
		Stage:            domain.TokenUsageStagePrompt,
		PromptTokens:     tokenCount,
		CompletionTokens: 0,
		TotalTokens:      tokenCount,
		PromptChars:      charCount,
		CompletionChars:  0,
		WorkspaceRoots:   input.WorkspaceRoots,
		UserEmail:        "",
		HookEventName:    string(input.HookEventName),
		Timestamp:        time.Now(),
	}

	// Send to ctd (best-effort)
	if err := h.deps.Client.CreateTokenUsage(ctx, usage); err != nil {
		h.deps.Log.Error("failed to send token usage", "error", err)
	} else {
		h.deps.Log.Info("token usage recorded", "stage", "prompt", "tokens", tokenCount)
	}

	return NewContinueOutput(), nil
}

// AfterAgentResponseHandler handles afterAgentResponse events.
type AfterAgentResponseHandler struct {
	deps *HandlerDependencies
}

// Handle processes the afterAgentResponse event.
func (h *AfterAgentResponseHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	p, ok := payload.(*AfterAgentResponsePayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	// Count tokens
	tokenCount, err := h.deps.Tokenizer.CountTokens(h.deps.Config.DefaultModel, p.Text)
	if err != nil {
		h.deps.Log.Error("failed to count tokens", "error", err)
		tokenCount = 0
	}

	charCount := CountChars(p.Text)

	h.deps.Log.Debug("completion metrics",
		"tokens", tokenCount,
		"chars", charCount,
		"conversation_id", input.ConversationID,
	)

	// Create token usage record
	usage := &domain.TokenUsage{
		ID:               uuid.New(),
		ConversationID:   parseUUID(input.ConversationID),
		GenerationID:     parseUUID(input.GenerationID),
		Model:            h.deps.Config.DefaultModel,
		Stage:            domain.TokenUsageStageCompletion,
		PromptTokens:     0,
		CompletionTokens: tokenCount,
		TotalTokens:      tokenCount,
		PromptChars:      0,
		CompletionChars:  charCount,
		WorkspaceRoots:   input.WorkspaceRoots,
		UserEmail:        "",
		HookEventName:    string(input.HookEventName),
		Timestamp:        time.Now(),
	}

	// Send token usage to ctd (best-effort)
	if err := h.deps.Client.CreateTokenUsage(ctx, usage); err != nil {
		h.deps.Log.Error("failed to send token usage", "error", err)
	} else {
		h.deps.Log.Info("token usage recorded", "stage", "completion", "tokens", tokenCount)
	}

	// Also send as thought for visibility
	thought := &domain.Thought{
		ID:             uuid.New(),
		ConversationID: parseUUID(input.ConversationID),
		GenerationID:   parseUUID(input.GenerationID),
		Text:           p.Text,
		DurationMs:     0,
		Model:          h.deps.Config.DefaultModel,
		HookEventName:  string(input.HookEventName),
		CursorVersion:  "",
		WorkspaceRoots: input.WorkspaceRoots,
		UserEmail:      "",
		Timestamp:      time.Now(),
	}

	if err := h.deps.Client.CreateThought(ctx, thought); err != nil {
		h.deps.Log.Error("failed to send thought", "error", err)
	}

	return NewEmptyOutput(), nil
}

// AfterAgentThoughtHandler handles afterAgentThought events.
type AfterAgentThoughtHandler struct {
	deps *HandlerDependencies
}

// Handle processes the afterAgentThought event.
func (h *AfterAgentThoughtHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	p, ok := payload.(*AfterAgentThoughtPayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	// Send as thought for visibility
	thought := &domain.Thought{
		ID:             uuid.New(),
		ConversationID: parseUUID(input.ConversationID),
		GenerationID:   parseUUID(input.GenerationID),
		Text:           p.Text,
		DurationMs:     0,
		Model:          h.deps.Config.DefaultModel,
		HookEventName:  string(input.HookEventName),
		CursorVersion:  "",
		WorkspaceRoots: input.WorkspaceRoots,
		UserEmail:      "",
		Timestamp:      time.Now(),
	}

	if err := h.deps.Client.CreateThought(ctx, thought); err != nil {
		h.deps.Log.Error("failed to send thought", "error", err)
	} else {
		h.deps.Log.Info("thought recorded", "event", input.HookEventName)
	}

	return NewEmptyOutput(), nil
}

// AfterMCPExecutionHandler handles afterMCPExecution events.
type AfterMCPExecutionHandler struct {
	deps *HandlerDependencies
}

// Handle processes the afterMCPExecution event.
func (h *AfterMCPExecutionHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	p, ok := payload.(*AfterMCPExecutionPayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	// Count tokens in tool input and result
	inputTokens, _ := h.deps.Tokenizer.CountTokens(h.deps.Config.DefaultModel, string(p.ToolInput))
	resultTokens, _ := h.deps.Tokenizer.CountTokens(h.deps.Config.DefaultModel, string(p.ResultJSON))
	totalTokens := inputTokens + resultTokens

	inputChars := CountChars(string(p.ToolInput))
	resultChars := CountChars(string(p.ResultJSON))

	h.deps.Log.Debug("tool metrics",
		"tool", p.ToolName,
		"input_tokens", inputTokens,
		"result_tokens", resultTokens,
		"conversation_id", input.ConversationID,
	)

	// Create token usage record
	usage := &domain.TokenUsage{
		ID:               uuid.New(),
		ConversationID:   parseUUID(input.ConversationID),
		GenerationID:     parseUUID(input.GenerationID),
		Model:            h.deps.Config.DefaultModel,
		Stage:            domain.TokenUsageStageTool,
		PromptTokens:     inputTokens,
		CompletionTokens: resultTokens,
		TotalTokens:      totalTokens,
		PromptChars:      inputChars,
		CompletionChars:  resultChars,
		WorkspaceRoots:   input.WorkspaceRoots,
		UserEmail:        "",
		HookEventName:    string(input.HookEventName),
		Timestamp:        time.Now(),
	}

	if err := h.deps.Client.CreateTokenUsage(ctx, usage); err != nil {
		h.deps.Log.Error("failed to send token usage", "error", err)
	} else {
		h.deps.Log.Info("token usage recorded", "stage", "tool", "tool", p.ToolName, "tokens", totalTokens)
	}

	return NewEmptyOutput(), nil
}

// AfterFileEditHandler handles afterFileEdit events.
type AfterFileEditHandler struct {
	deps *HandlerDependencies
}

// Handle processes the afterFileEdit event.
func (h *AfterFileEditHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	p, ok := payload.(*AfterFileEditPayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	// Count characters in edits
	var oldChars, newChars int64
	for _, edit := range p.Edits {
		oldChars += CountChars(edit.OldString)
		newChars += CountChars(edit.NewString)
	}

	h.deps.Log.Debug("file edit metrics",
		"file", p.FilePath,
		"old_chars", oldChars,
		"new_chars", newChars,
		"edit_count", len(p.Edits),
	)

	// Create token usage record for file edits
	usage := &domain.TokenUsage{
		ID:               uuid.New(),
		ConversationID:   parseUUID(input.ConversationID),
		GenerationID:     parseUUID(input.GenerationID),
		Model:            h.deps.Config.DefaultModel,
		Stage:            domain.TokenUsageStageFileEdit,
		PromptTokens:     0,
		CompletionTokens: 0,
		TotalTokens:      0,
		PromptChars:      oldChars,
		CompletionChars:  newChars,
		WorkspaceRoots:   input.WorkspaceRoots,
		UserEmail:        "",
		HookEventName:    string(input.HookEventName),
		Timestamp:        time.Now(),
	}

	if err := h.deps.Client.CreateTokenUsage(ctx, usage); err != nil {
		h.deps.Log.Error("failed to send token usage", "error", err)
	} else {
		h.deps.Log.Info("file edit recorded", "file", p.FilePath, "edits", len(p.Edits))
	}

	return NewEmptyOutput(), nil
}

// StopHandler handles stop events.
type StopHandler struct {
	deps *HandlerDependencies
}

// Handle processes the stop event.
func (h *StopHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	p, ok := payload.(*StopPayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	h.deps.Log.Info("agent loop stopped",
		"status", p.Status,
		"loop_count", p.LoopCount,
		"conversation_id", input.ConversationID,
	)

	// Create token usage record as session marker
	usage := &domain.TokenUsage{
		ID:               uuid.New(),
		ConversationID:   parseUUID(input.ConversationID),
		GenerationID:     parseUUID(input.GenerationID),
		Model:            h.deps.Config.DefaultModel,
		Stage:            domain.TokenUsageStageStop,
		PromptTokens:     0,
		CompletionTokens: 0,
		TotalTokens:      0,
		PromptChars:      0,
		CompletionChars:  0,
		WorkspaceRoots:   input.WorkspaceRoots,
		UserEmail:        "",
		HookEventName:    string(input.HookEventName),
		Timestamp:        time.Now(),
	}

	if err := h.deps.Client.CreateTokenUsage(ctx, usage); err != nil {
		h.deps.Log.Error("failed to send stop marker", "error", err)
	}

	return NewEmptyOutput(), nil
}

// BeforeShellExecutionHandler handles beforeShellExecution events.
type BeforeShellExecutionHandler struct {
	deps *HandlerDependencies
}

// Handle processes the beforeShellExecution event.
func (h *BeforeShellExecutionHandler) Handle(_ context.Context, input *HookInput, payload any) (*HookOutput, error) {
	p, ok := payload.(*BeforeShellExecutionPayload)
	if !ok {
		return NewAllowOutput(), nil
	}

	h.deps.Log.Debug("shell execution",
		"command", p.Command,
		"cwd", p.Cwd,
		"conversation_id", input.ConversationID,
	)

	// Always allow shell execution (observability only)
	return NewAllowOutput(), nil
}

// AfterShellExecutionHandler handles afterShellExecution events.
type AfterShellExecutionHandler struct {
	deps *HandlerDependencies
}

// Handle processes the afterShellExecution event.
func (h *AfterShellExecutionHandler) Handle(_ context.Context, input *HookInput, payload any) (*HookOutput, error) {
	p, ok := payload.(*AfterShellExecutionPayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	h.deps.Log.Debug("shell execution completed",
		"command", p.Command,
		"output_length", len(p.Output),
		"conversation_id", input.ConversationID,
	)

	return NewEmptyOutput(), nil
}

// BeforeMCPExecutionHandler handles beforeMCPExecution events.
type BeforeMCPExecutionHandler struct {
	deps *HandlerDependencies
}

// Handle processes the beforeMCPExecution event.
func (h *BeforeMCPExecutionHandler) Handle(_ context.Context, input *HookInput, payload any) (*HookOutput, error) {
	p, ok := payload.(*BeforeMCPExecutionPayload)
	if !ok {
		return NewAllowOutput(), nil
	}

	h.deps.Log.Debug("MCP execution",
		"tool", p.ToolName,
		"conversation_id", input.ConversationID,
	)

	// Always allow MCP execution (observability only)
	return NewAllowOutput(), nil
}

// BeforeReadFileHandler handles beforeReadFile events.
type BeforeReadFileHandler struct {
	deps *HandlerDependencies
}

// Handle processes the beforeReadFile event.
func (h *BeforeReadFileHandler) Handle(_ context.Context, input *HookInput, payload any) (*HookOutput, error) {
	p, ok := payload.(*BeforeReadFilePayload)
	if !ok {
		return NewAllowOutput(), nil
	}

	h.deps.Log.Debug("file read",
		"file", p.FilePath,
		"content_length", len(p.Content),
		"conversation_id", input.ConversationID,
	)

	// Always allow file read (observability only)
	return NewAllowOutput(), nil
}

// parseUUID parses a UUID string, returning a zero UUID on error.
func parseUUID(s string) uuid.UUID {
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil
	}
	return id
}


```

```go
package hook

import (
	"context"
	"time"

	"github.com/ainslies/ctd/internal/domain"
	"github.com/ainslies/ctd/pkg/logger"
	"github.com/google/uuid"
)

// Handler processes a hook event and returns the appropriate output.
type Handler interface {
	// Handle processes the hook input and returns the output.
	Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error)
}

// HandlerDependencies contains the dependencies for handlers.
type HandlerDependencies struct {
	Client    *CTDClient
	Tokenizer Tokenizer
	Log       *logger.Logger
	Config    *Config
}

// Router routes hook events to the appropriate handler.
type Router struct {
	deps     *HandlerDependencies
	handlers map[HookEventName]Handler
}

// NewRouter creates a new router with all handlers registered.
func NewRouter(deps *HandlerDependencies) *Router {
	router := &Router{
		deps:     deps,
		handlers: make(map[HookEventName]Handler),
	}

	// Register handlers
	router.handlers[EventBeforeSubmitPrompt] = &BeforeSubmitPromptHandler{deps: deps}
	router.handlers[EventAfterAgentResponse] = &AfterAgentResponseHandler{deps: deps}
	router.handlers[EventAfterAgentThought] = &AfterAgentThoughtHandler{deps: deps}
	router.handlers[EventAfterMCPExecution] = &AfterMCPExecutionHandler{deps: deps}
	router.handlers[EventAfterFileEdit] = &AfterFileEditHandler{deps: deps}
	router.handlers[EventStop] = &StopHandler{deps: deps}
	router.handlers[EventBeforeShellExecution] = &BeforeShellExecutionHandler{deps: deps}
	router.handlers[EventAfterShellExecution] = &AfterShellExecutionHandler{deps: deps}
	router.handlers[EventBeforeMCPExecution] = &BeforeMCPExecutionHandler{deps: deps}
	router.handlers[EventBeforeReadFile] = &BeforeReadFileHandler{deps: deps}

	return router
}

// Route routes the input to the appropriate handler.
func (r *Router) Route(ctx context.Context, input *HookInput) (*HookOutput, error) {
	handler, ok := r.handlers[input.HookEventName]
	if !ok {
		// Unknown event, return empty output
		r.deps.Log.Warn("unknown hook event", "event", input.HookEventName)
		return NewEmptyOutput(), nil
	}

	payload, err := input.ParsePayload()
	if err != nil {
		r.deps.Log.Error("failed to parse payload", "error", err, "event", input.HookEventName)
		// Return appropriate default output even on parse error
		return getDefaultOutput(input.HookEventName), nil
	}

	output, err := handler.Handle(ctx, input, payload)
	if err != nil {
		return nil, fmt.Errorf("handler error for %s: %w", input.HookEventName, err)
	}
	return output, nil
}

// getDefaultOutput returns the default output for an event type.
//
//nolint:exhaustive // Only specific events need special handling, others use empty output
func getDefaultOutput(event HookEventName) *HookOutput {
	switch event {
	case EventBeforeSubmitPrompt:
		return NewContinueOutput()
	case EventBeforeShellExecution, EventBeforeMCPExecution, EventBeforeReadFile:
		return NewAllowOutput()
	default:
		return NewEmptyOutput()
	}
}

// BeforeSubmitPromptHandler handles beforeSubmitPrompt events.
type BeforeSubmitPromptHandler struct {
	deps *HandlerDependencies
}

// Handle processes the beforeSubmitPrompt event.
func (h *BeforeSubmitPromptHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	p, ok := payload.(*BeforeSubmitPromptPayload)
	if !ok {
		return NewContinueOutput(), nil
	}

	// Count tokens
	tokenCount, err := h.deps.Tokenizer.CountTokens(h.deps.Config.DefaultModel, p.Prompt)
	if err != nil {
		h.deps.Log.Error("failed to count tokens", "error", err)
		tokenCount = 0
	}

	charCount := CountChars(p.Prompt)

	h.deps.Log.Debug("prompt metrics",
		"tokens", tokenCount,
		"chars", charCount,
		"conversation_id", input.ConversationID,
	)

	// Create token usage record
	usage := &domain.TokenUsage{
		ID:               uuid.New(),
		ConversationID:   parseUUID(input.ConversationID),
		GenerationID:     parseUUID(input.GenerationID),
		Model:            h.deps.Config.DefaultModel,
		Stage:            domain.TokenUsageStagePrompt,
		PromptTokens:     tokenCount,
		CompletionTokens: 0,
		TotalTokens:      tokenCount,
		PromptChars:      charCount,
		CompletionChars:  0,
		WorkspaceRoots:   input.WorkspaceRoots,
		UserEmail:        "",
		HookEventName:    string(input.HookEventName),
		Timestamp:        time.Now(),
	}

	// Send to ctd (best-effort)
	if err := h.deps.Client.CreateTokenUsage(ctx, usage); err != nil {
		h.deps.Log.Error("failed to send token usage", "error", err)
	} else {
		h.deps.Log.Info("token usage recorded", "stage", "prompt", "tokens", tokenCount)
	}

	return NewContinueOutput(), nil
}

// AfterAgentResponseHandler handles afterAgentResponse events.
type AfterAgentResponseHandler struct {
	deps *HandlerDependencies
}

// Handle processes the afterAgentResponse event.
func (h *AfterAgentResponseHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	p, ok := payload.(*AfterAgentResponsePayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	// Count tokens
	tokenCount, err := h.deps.Tokenizer.CountTokens(h.deps.Config.DefaultModel, p.Text)
	if err != nil {
		h.deps.Log.Error("failed to count tokens", "error", err)
		tokenCount = 0
	}

	charCount := CountChars(p.Text)

	h.deps.Log.Debug("completion metrics",
		"tokens", tokenCount,
		"chars", charCount,
		"conversation_id", input.ConversationID,
	)

	// Create token usage record
	usage := &domain.TokenUsage{
		ID:               uuid.New(),
		ConversationID:   parseUUID(input.ConversationID),
		GenerationID:     parseUUID(input.GenerationID),
		Model:            h.deps.Config.DefaultModel,
		Stage:            domain.TokenUsageStageCompletion,
		PromptTokens:     0,
		CompletionTokens: tokenCount,
		TotalTokens:      tokenCount,
		PromptChars:      0,
		CompletionChars:  charCount,
		WorkspaceRoots:   input.WorkspaceRoots,
		UserEmail:        "",
		HookEventName:    string(input.HookEventName),
		Timestamp:        time.Now(),
	}

	// Send token usage to ctd (best-effort)
	if err := h.deps.Client.CreateTokenUsage(ctx, usage); err != nil {
		h.deps.Log.Error("failed to send token usage", "error", err)
	} else {
		h.deps.Log.Info("token usage recorded", "stage", "completion", "tokens", tokenCount)
	}

	// Also send as thought for visibility
	thought := &domain.Thought{
		ID:             uuid.New(),
		ConversationID: parseUUID(input.ConversationID),
		GenerationID:   parseUUID(input.GenerationID),
		Text:           p.Text,
		DurationMs:     0,
		Model:          h.deps.Config.DefaultModel,
		HookEventName:  string(input.HookEventName),
		CursorVersion:  "",
		WorkspaceRoots: input.WorkspaceRoots,
		UserEmail:      "",
		Timestamp:      time.Now(),
	}

	if err := h.deps.Client.CreateThought(ctx, thought); err != nil {
		h.deps.Log.Error("failed to send thought", "error", err)
	}

	return NewEmptyOutput(), nil
}

// AfterAgentThoughtHandler handles afterAgentThought events.
type AfterAgentThoughtHandler struct {
	deps *HandlerDependencies
}

// Handle processes the afterAgentThought event.
func (h *AfterAgentThoughtHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	p, ok := payload.(*AfterAgentThoughtPayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	// Send as thought for visibility
	thought := &domain.Thought{
		ID:             uuid.New(),
		ConversationID: parseUUID(input.ConversationID),
		GenerationID:   parseUUID(input.GenerationID),
		Text:           p.Text,
		DurationMs:     0,
		Model:          h.deps.Config.DefaultModel,
		HookEventName:  string(input.HookEventName),
		CursorVersion:  "",
		WorkspaceRoots: input.WorkspaceRoots,
		UserEmail:      "",
		Timestamp:      time.Now(),
	}

	if err := h.deps.Client.CreateThought(ctx, thought); err != nil {
		h.deps.Log.Error("failed to send thought", "error", err)
	} else {
		h.deps.Log.Info("thought recorded", "event", input.HookEventName)
	}

	return NewEmptyOutput(), nil
}

// AfterMCPExecutionHandler handles afterMCPExecution events.
type AfterMCPExecutionHandler struct {
	deps *HandlerDependencies
}

// Handle processes the afterMCPExecution event.
func (h *AfterMCPExecutionHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	p, ok := payload.(*AfterMCPExecutionPayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	// Count tokens in tool input and result
	inputTokens, _ := h.deps.Tokenizer.CountTokens(h.deps.Config.DefaultModel, string(p.ToolInput))
	resultTokens, _ := h.deps.Tokenizer.CountTokens(h.deps.Config.DefaultModel, string(p.ResultJSON))
	totalTokens := inputTokens + resultTokens

	inputChars := CountChars(string(p.ToolInput))
	resultChars := CountChars(string(p.ResultJSON))

	h.deps.Log.Debug("tool metrics",
		"tool", p.ToolName,
		"input_tokens", inputTokens,
		"result_tokens", resultTokens,
		"conversation_id", input.ConversationID,
	)

	// Create token usage record
	usage := &domain.TokenUsage{
		ID:               uuid.New(),
		ConversationID:   parseUUID(input.ConversationID),
		GenerationID:     parseUUID(input.GenerationID),
		Model:            h.deps.Config.DefaultModel,
		Stage:            domain.TokenUsageStageTool,
		PromptTokens:     inputTokens,
		CompletionTokens: resultTokens,
		TotalTokens:      totalTokens,
		PromptChars:      inputChars,
		CompletionChars:  resultChars,
		WorkspaceRoots:   input.WorkspaceRoots,
		UserEmail:        "",
		HookEventName:    string(input.HookEventName),
		Timestamp:        time.Now(),
	}

	if err := h.deps.Client.CreateTokenUsage(ctx, usage); err != nil {
		h.deps.Log.Error("failed to send token usage", "error", err)
	} else {
		h.deps.Log.Info("token usage recorded", "stage", "tool", "tool", p.ToolName, "tokens", totalTokens)
	}

	return NewEmptyOutput(), nil
}

// AfterFileEditHandler handles afterFileEdit events.
type AfterFileEditHandler struct {
	deps *HandlerDependencies
}

// Handle processes the afterFileEdit event.
func (h *AfterFileEditHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	p, ok := payload.(*AfterFileEditPayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	// Count characters in edits
	var oldChars, newChars int64
	for _, edit := range p.Edits {
		oldChars += CountChars(edit.OldString)
		newChars += CountChars(edit.NewString)
	}

	h.deps.Log.Debug("file edit metrics",
		"file", p.FilePath,
		"old_chars", oldChars,
		"new_chars", newChars,
		"edit_count", len(p.Edits),
	)

	// Create token usage record for file edits
	usage := &domain.TokenUsage{
		ID:               uuid.New(),
		ConversationID:   parseUUID(input.ConversationID),
		GenerationID:     parseUUID(input.GenerationID),
		Model:            h.deps.Config.DefaultModel,
		Stage:            domain.TokenUsageStageFileEdit,
		PromptTokens:     0,
		CompletionTokens: 0,
		TotalTokens:      0,
		PromptChars:      oldChars,
		CompletionChars:  newChars,
		WorkspaceRoots:   input.WorkspaceRoots,
		UserEmail:        "",
		HookEventName:    string(input.HookEventName),
		Timestamp:        time.Now(),
	}

	if err := h.deps.Client.CreateTokenUsage(ctx, usage); err != nil {
		h.deps.Log.Error("failed to send token usage", "error", err)
	} else {
		h.deps.Log.Info("file edit recorded", "file", p.FilePath, "edits", len(p.Edits))
	}

	return NewEmptyOutput(), nil
}

// StopHandler handles stop events.
type StopHandler struct {
	deps *HandlerDependencies
}

// Handle processes the stop event.
func (h *StopHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	p, ok := payload.(*StopPayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	h.deps.Log.Info("agent loop stopped",
		"status", p.Status,
		"loop_count", p.LoopCount,
		"conversation_id", input.ConversationID,
	)

	// Create token usage record as session marker
	usage := &domain.TokenUsage{
		ID:               uuid.New(),
		ConversationID:   parseUUID(input.ConversationID),
		GenerationID:     parseUUID(input.GenerationID),
		Model:            h.deps.Config.DefaultModel,
		Stage:            domain.TokenUsageStageStop,
		PromptTokens:     0,
		CompletionTokens: 0,
		TotalTokens:      0,
		PromptChars:      0,
		CompletionChars:  0,
		WorkspaceRoots:   input.WorkspaceRoots,
		UserEmail:        "",
		HookEventName:    string(input.HookEventName),
		Timestamp:        time.Now(),
	}

	if err := h.deps.Client.CreateTokenUsage(ctx, usage); err != nil {
		h.deps.Log.Error("failed to send stop marker", "error", err)
	}

	return NewEmptyOutput(), nil
}

// BeforeShellExecutionHandler handles beforeShellExecution events.
type BeforeShellExecutionHandler struct {
	deps *HandlerDependencies
}

// Handle processes the beforeShellExecution event.
func (h *BeforeShellExecutionHandler) Handle(_ context.Context, input *HookInput, payload any) (*HookOutput, error) {
	p, ok := payload.(*BeforeShellExecutionPayload)
	if !ok {
		return NewAllowOutput(), nil
	}

	h.deps.Log.Debug("shell execution",
		"command", p.Command,
		"cwd", p.Cwd,
		"conversation_id", input.ConversationID,
	)

	// Always allow shell execution (observability only)
	return NewAllowOutput(), nil
}

// AfterShellExecutionHandler handles afterShellExecution events.
type AfterShellExecutionHandler struct {
	deps *HandlerDependencies
}

// Handle processes the afterShellExecution event.
func (h *AfterShellExecutionHandler) Handle(_ context.Context, input *HookInput, payload any) (*HookOutput, error) {
	p, ok := payload.(*AfterShellExecutionPayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	h.deps.Log.Debug("shell execution completed",
		"command", p.Command,
		"output_length", len(p.Output),
		"conversation_id", input.ConversationID,
	)

	return NewEmptyOutput(), nil
}

// BeforeMCPExecutionHandler handles beforeMCPExecution events.
type BeforeMCPExecutionHandler struct {
	deps *HandlerDependencies
}

// Handle processes the beforeMCPExecution event.
func (h *BeforeMCPExecutionHandler) Handle(_ context.Context, input *HookInput, payload any) (*HookOutput, error) {
	p, ok := payload.(*BeforeMCPExecutionPayload)
	if !ok {
		return NewAllowOutput(), nil
	}

	h.deps.Log.Debug("MCP execution",
		"tool", p.ToolName,
		"conversation_id", input.ConversationID,
	)

	// Always allow MCP execution (observability only)
	return NewAllowOutput(), nil
}

// BeforeReadFileHandler handles beforeReadFile events.
type BeforeReadFileHandler struct {
	deps *HandlerDependencies
}

// Handle processes the beforeReadFile event.
func (h *BeforeReadFileHandler) Handle(_ context.Context, input *HookInput, payload any) (*HookOutput, error) {
	p, ok := payload.(*BeforeReadFilePayload)
	if !ok {
		return NewAllowOutput(), nil
	}

	h.deps.Log.Debug("file read",
		"file", p.FilePath,
		"content_length", len(p.Content),
		"conversation_id", input.ConversationID,
	)

	// Always allow file read (observability only)
	return NewAllowOutput(), nil
}

// parseUUID parses a UUID string, returning a zero UUID on error.
func parseUUID(s string) uuid.UUID {
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil
	}
	return id
}


```

```go
package hook

import (
	"context"
	"fmt"
	"time"

	"github.com/ainslies/ctd/internal/domain"
	"github.com/ainslies/ctd/pkg/logger"
	"github.com/google/uuid"
)

// Handler processes a hook event and returns the appropriate output.
type Handler interface {
	// Handle processes the hook input and returns the output.
	Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error)
}

// HandlerDependencies contains the dependencies for handlers.
type HandlerDependencies struct {
	Client    *CTDClient
	Tokenizer Tokenizer
	Log       *logger.Logger
	Config    *Config
}

// Router routes hook events to the appropriate handler.
type Router struct {
	deps     *HandlerDependencies
	handlers map[HookEventName]Handler
}

// NewRouter creates a new router with all handlers registered.
func NewRouter(deps *HandlerDependencies) *Router {
	router := &Router{
		deps:     deps,
		handlers: make(map[HookEventName]Handler),
	}

	// Register handlers
	router.handlers[EventBeforeSubmitPrompt] = &BeforeSubmitPromptHandler{deps: deps}
	router.handlers[EventAfterAgentResponse] = &AfterAgentResponseHandler{deps: deps}
	router.handlers[EventAfterAgentThought] = &AfterAgentThoughtHandler{deps: deps}
	router.handlers[EventAfterMCPExecution] = &AfterMCPExecutionHandler{deps: deps}
	router.handlers[EventAfterFileEdit] = &AfterFileEditHandler{deps: deps}
	router.handlers[EventStop] = &StopHandler{deps: deps}
	router.handlers[EventBeforeShellExecution] = &BeforeShellExecutionHandler{deps: deps}
	router.handlers[EventAfterShellExecution] = &AfterShellExecutionHandler{deps: deps}
	router.handlers[EventBeforeMCPExecution] = &BeforeMCPExecutionHandler{deps: deps}
	router.handlers[EventBeforeReadFile] = &BeforeReadFileHandler{deps: deps}

	return router
}

// Route routes the input to the appropriate handler.
func (r *Router) Route(ctx context.Context, input *HookInput) (*HookOutput, error) {
	handler, ok := r.handlers[input.HookEventName]
	if !ok {
		// Unknown event, return empty output
		r.deps.Log.Warn("unknown hook event", "event", input.HookEventName)
		return NewEmptyOutput(), nil
	}

	payload, err := input.ParsePayload()
	if err != nil {
		r.deps.Log.Error("failed to parse payload", "error", err, "event", input.HookEventName)
		// Return appropriate default output even on parse error
		return getDefaultOutput(input.HookEventName), nil
	}

	output, err := handler.Handle(ctx, input, payload)
	if err != nil {
		return nil, fmt.Errorf("handler error for %s: %w", input.HookEventName, err)
	}
	return output, nil
}

// getDefaultOutput returns the default output for an event type.
//
//nolint:exhaustive // Only specific events need special handling, others use empty output
func getDefaultOutput(event HookEventName) *HookOutput {
	switch event {
	case EventBeforeSubmitPrompt:
		return NewContinueOutput()
	case EventBeforeShellExecution, EventBeforeMCPExecution, EventBeforeReadFile:
		return NewAllowOutput()
	default:
		return NewEmptyOutput()
	}
}

// BeforeSubmitPromptHandler handles beforeSubmitPrompt events.
type BeforeSubmitPromptHandler struct {
	deps *HandlerDependencies
}

// Handle processes the beforeSubmitPrompt event.
func (h *BeforeSubmitPromptHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	p, ok := payload.(*BeforeSubmitPromptPayload)
	if !ok {
		return NewContinueOutput(), nil
	}

	// Count tokens
	tokenCount, err := h.deps.Tokenizer.CountTokens(h.deps.Config.DefaultModel, p.Prompt)
	if err != nil {
		h.deps.Log.Error("failed to count tokens", "error", err)
		tokenCount = 0
	}

	charCount := CountChars(p.Prompt)

	h.deps.Log.Debug("prompt metrics",
		"tokens", tokenCount,
		"chars", charCount,
		"conversation_id", input.ConversationID,
	)

	// Create token usage record
	usage := &domain.TokenUsage{
		ID:               uuid.New(),
		ConversationID:   parseUUID(input.ConversationID),
		GenerationID:     parseUUID(input.GenerationID),
		Model:            h.deps.Config.DefaultModel,
		Stage:            domain.TokenUsageStagePrompt,
		PromptTokens:     tokenCount,
		CompletionTokens: 0,
		TotalTokens:      tokenCount,
		PromptChars:      charCount,
		CompletionChars:  0,
		WorkspaceRoots:   input.WorkspaceRoots,
		UserEmail:        "",
		HookEventName:    string(input.HookEventName),
		Timestamp:        time.Now(),
	}

	// Send to ctd (best-effort)
	if err := h.deps.Client.CreateTokenUsage(ctx, usage); err != nil {
		h.deps.Log.Error("failed to send token usage", "error", err)
	} else {
		h.deps.Log.Info("token usage recorded", "stage", "prompt", "tokens", tokenCount)
	}

	return NewContinueOutput(), nil
}

// AfterAgentResponseHandler handles afterAgentResponse events.
type AfterAgentResponseHandler struct {
	deps *HandlerDependencies
}

// Handle processes the afterAgentResponse event.
func (h *AfterAgentResponseHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	p, ok := payload.(*AfterAgentResponsePayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	// Count tokens
	tokenCount, err := h.deps.Tokenizer.CountTokens(h.deps.Config.DefaultModel, p.Text)
	if err != nil {
		h.deps.Log.Error("failed to count tokens", "error", err)
		tokenCount = 0
	}

	charCount := CountChars(p.Text)

	h.deps.Log.Debug("completion metrics",
		"tokens", tokenCount,
		"chars", charCount,
		"conversation_id", input.ConversationID,
	)

	// Create token usage record
	usage := &domain.TokenUsage{
		ID:               uuid.New(),
		ConversationID:   parseUUID(input.ConversationID),
		GenerationID:     parseUUID(input.GenerationID),
		Model:            h.deps.Config.DefaultModel,
		Stage:            domain.TokenUsageStageCompletion,
		PromptTokens:     0,
		CompletionTokens: tokenCount,
		TotalTokens:      tokenCount,
		PromptChars:      0,
		CompletionChars:  charCount,
		WorkspaceRoots:   input.WorkspaceRoots,
		UserEmail:        "",
		HookEventName:    string(input.HookEventName),
		Timestamp:        time.Now(),
	}

	// Send token usage to ctd (best-effort)
	if err := h.deps.Client.CreateTokenUsage(ctx, usage); err != nil {
		h.deps.Log.Error("failed to send token usage", "error", err)
	} else {
		h.deps.Log.Info("token usage recorded", "stage", "completion", "tokens", tokenCount)
	}

	// Also send as thought for visibility
	thought := &domain.Thought{
		ID:             uuid.New(),
		ConversationID: parseUUID(input.ConversationID),
		GenerationID:   parseUUID(input.GenerationID),
		Text:           p.Text,
		DurationMs:     0,
		Model:          h.deps.Config.DefaultModel,
		HookEventName:  string(input.HookEventName),
		CursorVersion:  "",
		WorkspaceRoots: input.WorkspaceRoots,
		UserEmail:      "",
		Timestamp:      time.Now(),
	}

	if err := h.deps.Client.CreateThought(ctx, thought); err != nil {
		h.deps.Log.Error("failed to send thought", "error", err)
	}

	return NewEmptyOutput(), nil
}

// AfterAgentThoughtHandler handles afterAgentThought events.
type AfterAgentThoughtHandler struct {
	deps *HandlerDependencies
}

// Handle processes the afterAgentThought event.
func (h *AfterAgentThoughtHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	p, ok := payload.(*AfterAgentThoughtPayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	// Send as thought for visibility
	thought := &domain.Thought{
		ID:             uuid.New(),
		ConversationID: parseUUID(input.ConversationID),
		GenerationID:   parseUUID(input.GenerationID),
		Text:           p.Text,
		DurationMs:     0,
		Model:          h.deps.Config.DefaultModel,
		HookEventName:  string(input.HookEventName),
		CursorVersion:  "",
		WorkspaceRoots: input.WorkspaceRoots,
		UserEmail:      "",
		Timestamp:      time.Now(),
	}

	if err := h.deps.Client.CreateThought(ctx, thought); err != nil {
		h.deps.Log.Error("failed to send thought", "error", err)
	} else {
		h.deps.Log.Info("thought recorded", "event", input.HookEventName)
	}

	return NewEmptyOutput(), nil
}

// AfterMCPExecutionHandler handles afterMCPExecution events.
type AfterMCPExecutionHandler struct {
	deps *HandlerDependencies
}

// Handle processes the afterMCPExecution event.
func (h *AfterMCPExecutionHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	p, ok := payload.(*AfterMCPExecutionPayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	// Count tokens in tool input and result
	inputTokens, _ := h.deps.Tokenizer.CountTokens(h.deps.Config.DefaultModel, string(p.ToolInput))
	resultTokens, _ := h.deps.Tokenizer.CountTokens(h.deps.Config.DefaultModel, string(p.ResultJSON))
	totalTokens := inputTokens + resultTokens

	inputChars := CountChars(string(p.ToolInput))
	resultChars := CountChars(string(p.ResultJSON))

	h.deps.Log.Debug("tool metrics",
		"tool", p.ToolName,
		"input_tokens", inputTokens,
		"result_tokens", resultTokens,
		"conversation_id", input.ConversationID,
	)

	// Create token usage record
	usage := &domain.TokenUsage{
		ID:               uuid.New(),
		ConversationID:   parseUUID(input.ConversationID),
		GenerationID:     parseUUID(input.GenerationID),
		Model:            h.deps.Config.DefaultModel,
		Stage:            domain.TokenUsageStageTool,
		PromptTokens:     inputTokens,
		CompletionTokens: resultTokens,
		TotalTokens:      totalTokens,
		PromptChars:      inputChars,
		CompletionChars:  resultChars,
		WorkspaceRoots:   input.WorkspaceRoots,
		UserEmail:        "",
		HookEventName:    string(input.HookEventName),
		Timestamp:        time.Now(),
	}

	if err := h.deps.Client.CreateTokenUsage(ctx, usage); err != nil {
		h.deps.Log.Error("failed to send token usage", "error", err)
	} else {
		h.deps.Log.Info("token usage recorded", "stage", "tool", "tool", p.ToolName, "tokens", totalTokens)
	}

	return NewEmptyOutput(), nil
}

// AfterFileEditHandler handles afterFileEdit events.
type AfterFileEditHandler struct {
	deps *HandlerDependencies
}

// Handle processes the afterFileEdit event.
func (h *AfterFileEditHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	p, ok := payload.(*AfterFileEditPayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	// Count characters in edits
	var oldChars, newChars int64
	for _, edit := range p.Edits {
		oldChars += CountChars(edit.OldString)
		newChars += CountChars(edit.NewString)
	}

	h.deps.Log.Debug("file edit metrics",
		"file", p.FilePath,
		"old_chars", oldChars,
		"new_chars", newChars,
		"edit_count", len(p.Edits),
	)

	// Create token usage record for file edits
	usage := &domain.TokenUsage{
		ID:               uuid.New(),
		ConversationID:   parseUUID(input.ConversationID),
		GenerationID:     parseUUID(input.GenerationID),
		Model:            h.deps.Config.DefaultModel,
		Stage:            domain.TokenUsageStageFileEdit,
		PromptTokens:     0,
		CompletionTokens: 0,
		TotalTokens:      0,
		PromptChars:      oldChars,
		CompletionChars:  newChars,
		WorkspaceRoots:   input.WorkspaceRoots,
		UserEmail:        "",
		HookEventName:    string(input.HookEventName),
		Timestamp:        time.Now(),
	}

	if err := h.deps.Client.CreateTokenUsage(ctx, usage); err != nil {
		h.deps.Log.Error("failed to send token usage", "error", err)
	} else {
		h.deps.Log.Info("file edit recorded", "file", p.FilePath, "edits", len(p.Edits))
	}

	return NewEmptyOutput(), nil
}

// StopHandler handles stop events.
type StopHandler struct {
	deps *HandlerDependencies
}

// Handle processes the stop event.
func (h *StopHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	p, ok := payload.(*StopPayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	h.deps.Log.Info("agent loop stopped",
		"status", p.Status,
		"loop_count", p.LoopCount,
		"conversation_id", input.ConversationID,
	)

	// Create token usage record as session marker
	usage := &domain.TokenUsage{
		ID:               uuid.New(),
		ConversationID:   parseUUID(input.ConversationID),
		GenerationID:     parseUUID(input.GenerationID),
		Model:            h.deps.Config.DefaultModel,
		Stage:            domain.TokenUsageStageStop,
		PromptTokens:     0,
		CompletionTokens: 0,
		TotalTokens:      0,
		PromptChars:      0,
		CompletionChars:  0,
		WorkspaceRoots:   input.WorkspaceRoots,
		UserEmail:        "",
		HookEventName:    string(input.HookEventName),
		Timestamp:        time.Now(),
	}

	if err := h.deps.Client.CreateTokenUsage(ctx, usage); err != nil {
		h.deps.Log.Error("failed to send stop marker", "error", err)
	}

	return NewEmptyOutput(), nil
}

// BeforeShellExecutionHandler handles beforeShellExecution events.
type BeforeShellExecutionHandler struct {
	deps *HandlerDependencies
}

// Handle processes the beforeShellExecution event.
func (h *BeforeShellExecutionHandler) Handle(_ context.Context, input *HookInput, payload any) (*HookOutput, error) {
	p, ok := payload.(*BeforeShellExecutionPayload)
	if !ok {
		return NewAllowOutput(), nil
	}

	h.deps.Log.Debug("shell execution",
		"command", p.Command,
		"cwd", p.Cwd,
		"conversation_id", input.ConversationID,
	)

	// Always allow shell execution (observability only)
	return NewAllowOutput(), nil
}

// AfterShellExecutionHandler handles afterShellExecution events.
type AfterShellExecutionHandler struct {
	deps *HandlerDependencies
}

// Handle processes the afterShellExecution event.
func (h *AfterShellExecutionHandler) Handle(_ context.Context, input *HookInput, payload any) (*HookOutput, error) {
	p, ok := payload.(*AfterShellExecutionPayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	h.deps.Log.Debug("shell execution completed",
		"command", p.Command,
		"output_length", len(p.Output),
		"conversation_id", input.ConversationID,
	)

	return NewEmptyOutput(), nil
}

// BeforeMCPExecutionHandler handles beforeMCPExecution events.
type BeforeMCPExecutionHandler struct {
	deps *HandlerDependencies
}

// Handle processes the beforeMCPExecution event.
func (h *BeforeMCPExecutionHandler) Handle(_ context.Context, input *HookInput, payload any) (*HookOutput, error) {
	p, ok := payload.(*BeforeMCPExecutionPayload)
	if !ok {
		return NewAllowOutput(), nil
	}

	h.deps.Log.Debug("MCP execution",
		"tool", p.ToolName,
		"conversation_id", input.ConversationID,
	)

	// Always allow MCP execution (observability only)
	return NewAllowOutput(), nil
}

// BeforeReadFileHandler handles beforeReadFile events.
type BeforeReadFileHandler struct {
	deps *HandlerDependencies
}

// Handle processes the beforeReadFile event.
func (h *BeforeReadFileHandler) Handle(_ context.Context, input *HookInput, payload any) (*HookOutput, error) {
	p, ok := payload.(*BeforeReadFilePayload)
	if !ok {
		return NewAllowOutput(), nil
	}

	h.deps.Log.Debug("file read",
		"file", p.FilePath,
		"content_length", len(p.Content),
		"conversation_id", input.ConversationID,
	)

	// Always allow file read (observability only)
	return NewAllowOutput(), nil
}

// parseUUID parses a UUID string, returning a zero UUID on error.
func parseUUID(s string) uuid.UUID {
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil
	}
	return id
}


```

Now let me fix the varnamelen issues by renaming payload variables:

```go
package hook

import (
	"context"
	"fmt"
	"time"

	"github.com/ainslies/ctd/internal/domain"
	"github.com/ainslies/ctd/pkg/logger"
	"github.com/google/uuid"
)

// Handler processes a hook event and returns the appropriate output.
type Handler interface {
	// Handle processes the hook input and returns the output.
	Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error)
}

// HandlerDependencies contains the dependencies for handlers.
type HandlerDependencies struct {
	Client    *CTDClient
	Tokenizer Tokenizer
	Log       *logger.Logger
	Config    *Config
}

// Router routes hook events to the appropriate handler.
type Router struct {
	deps     *HandlerDependencies
	handlers map[HookEventName]Handler
}

// NewRouter creates a new router with all handlers registered.
func NewRouter(deps *HandlerDependencies) *Router {
	router := &Router{
		deps:     deps,
		handlers: make(map[HookEventName]Handler),
	}

	// Register handlers
	router.handlers[EventBeforeSubmitPrompt] = &BeforeSubmitPromptHandler{deps: deps}
	router.handlers[EventAfterAgentResponse] = &AfterAgentResponseHandler{deps: deps}
	router.handlers[EventAfterAgentThought] = &AfterAgentThoughtHandler{deps: deps}
	router.handlers[EventAfterMCPExecution] = &AfterMCPExecutionHandler{deps: deps}
	router.handlers[EventAfterFileEdit] = &AfterFileEditHandler{deps: deps}
	router.handlers[EventStop] = &StopHandler{deps: deps}
	router.handlers[EventBeforeShellExecution] = &BeforeShellExecutionHandler{deps: deps}
	router.handlers[EventAfterShellExecution] = &AfterShellExecutionHandler{deps: deps}
	router.handlers[EventBeforeMCPExecution] = &BeforeMCPExecutionHandler{deps: deps}
	router.handlers[EventBeforeReadFile] = &BeforeReadFileHandler{deps: deps}

	return router
}

// Route routes the input to the appropriate handler.
func (r *Router) Route(ctx context.Context, input *HookInput) (*HookOutput, error) {
	handler, ok := r.handlers[input.HookEventName]
	if !ok {
		// Unknown event, return empty output
		r.deps.Log.Warn("unknown hook event", "event", input.HookEventName)
		return NewEmptyOutput(), nil
	}

	payload, err := input.ParsePayload()
	if err != nil {
		r.deps.Log.Error("failed to parse payload", "error", err, "event", input.HookEventName)
		// Return appropriate default output even on parse error
		return getDefaultOutput(input.HookEventName), nil
	}

	output, err := handler.Handle(ctx, input, payload)
	if err != nil {
		return nil, fmt.Errorf("handler error for %s: %w", input.HookEventName, err)
	}
	return output, nil
}

// getDefaultOutput returns the default output for an event type.
//
//nolint:exhaustive // Only specific events need special handling, others use empty output
func getDefaultOutput(event HookEventName) *HookOutput {
	switch event {
	case EventBeforeSubmitPrompt:
		return NewContinueOutput()
	case EventBeforeShellExecution, EventBeforeMCPExecution, EventBeforeReadFile:
		return NewAllowOutput()
	default:
		return NewEmptyOutput()
	}
}

// BeforeSubmitPromptHandler handles beforeSubmitPrompt events.
type BeforeSubmitPromptHandler struct {
	deps *HandlerDependencies
}

// Handle processes the beforeSubmitPrompt event.
//
//nolint:varnamelen // promptPayload is clear in context
func (h *BeforeSubmitPromptHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	promptPayload, ok := payload.(*BeforeSubmitPromptPayload)
	if !ok {
		return NewContinueOutput(), nil
	}

	// Count tokens
	tokenCount, err := h.deps.Tokenizer.CountTokens(h.deps.Config.DefaultModel, promptPayload.Prompt)
	if err != nil {
		h.deps.Log.Error("failed to count tokens", "error", err)
		tokenCount = 0
	}

	charCount := CountChars(promptPayload.Prompt)

	h.deps.Log.Debug("prompt metrics",
		"tokens", tokenCount,
		"chars", charCount,
		"conversation_id", input.ConversationID,
	)

	// Create token usage record
	usage := &domain.TokenUsage{
		ID:               uuid.New(),
		ConversationID:   parseUUID(input.ConversationID),
		GenerationID:     parseUUID(input.GenerationID),
		Model:            h.deps.Config.DefaultModel,
		Stage:            domain.TokenUsageStagePrompt,
		PromptTokens:     tokenCount,
		CompletionTokens: 0,
		TotalTokens:      tokenCount,
		PromptChars:      charCount,
		CompletionChars:  0,
		WorkspaceRoots:   input.WorkspaceRoots,
		UserEmail:        "",
		HookEventName:    string(input.HookEventName),
		Timestamp:        time.Now(),
	}

	// Send to ctd (best-effort)
	if err := h.deps.Client.CreateTokenUsage(ctx, usage); err != nil {
		h.deps.Log.Error("failed to send token usage", "error", err)
	} else {
		h.deps.Log.Info("token usage recorded", "stage", "prompt", "tokens", tokenCount)
	}

	return NewContinueOutput(), nil
}

// AfterAgentResponseHandler handles afterAgentResponse events.
type AfterAgentResponseHandler struct {
	deps *HandlerDependencies
}

// Handle processes the afterAgentResponse event.
func (h *AfterAgentResponseHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	p, ok := payload.(*AfterAgentResponsePayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	// Count tokens
	tokenCount, err := h.deps.Tokenizer.CountTokens(h.deps.Config.DefaultModel, p.Text)
	if err != nil {
		h.deps.Log.Error("failed to count tokens", "error", err)
		tokenCount = 0
	}

	charCount := CountChars(p.Text)

	h.deps.Log.Debug("completion metrics",
		"tokens", tokenCount,
		"chars", charCount,
		"conversation_id", input.ConversationID,
	)

	// Create token usage record
	usage := &domain.TokenUsage{
		ID:               uuid.New(),
		ConversationID:   parseUUID(input.ConversationID),
		GenerationID:     parseUUID(input.GenerationID),
		Model:            h.deps.Config.DefaultModel,
		Stage:            domain.TokenUsageStageCompletion,
		PromptTokens:     0,
		CompletionTokens: tokenCount,
		TotalTokens:      tokenCount,
		PromptChars:      0,
		CompletionChars:  charCount,
		WorkspaceRoots:   input.WorkspaceRoots,
		UserEmail:        "",
		HookEventName:    string(input.HookEventName),
		Timestamp:        time.Now(),
	}

	// Send token usage to ctd (best-effort)
	if err := h.deps.Client.CreateTokenUsage(ctx, usage); err != nil {
		h.deps.Log.Error("failed to send token usage", "error", err)
	} else {
		h.deps.Log.Info("token usage recorded", "stage", "completion", "tokens", tokenCount)
	}

	// Also send as thought for visibility
	thought := &domain.Thought{
		ID:             uuid.New(),
		ConversationID: parseUUID(input.ConversationID),
		GenerationID:   parseUUID(input.GenerationID),
		Text:           p.Text,
		DurationMs:     0,
		Model:          h.deps.Config.DefaultModel,
		HookEventName:  string(input.HookEventName),
		CursorVersion:  "",
		WorkspaceRoots: input.WorkspaceRoots,
		UserEmail:      "",
		Timestamp:      time.Now(),
	}

	if err := h.deps.Client.CreateThought(ctx, thought); err != nil {
		h.deps.Log.Error("failed to send thought", "error", err)
	}

	return NewEmptyOutput(), nil
}

// AfterAgentThoughtHandler handles afterAgentThought events.
type AfterAgentThoughtHandler struct {
	deps *HandlerDependencies
}

// Handle processes the afterAgentThought event.
func (h *AfterAgentThoughtHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	p, ok := payload.(*AfterAgentThoughtPayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	// Send as thought for visibility
	thought := &domain.Thought{
		ID:             uuid.New(),
		ConversationID: parseUUID(input.ConversationID),
		GenerationID:   parseUUID(input.GenerationID),
		Text:           p.Text,
		DurationMs:     0,
		Model:          h.deps.Config.DefaultModel,
		HookEventName:  string(input.HookEventName),
		CursorVersion:  "",
		WorkspaceRoots: input.WorkspaceRoots,
		UserEmail:      "",
		Timestamp:      time.Now(),
	}

	if err := h.deps.Client.CreateThought(ctx, thought); err != nil {
		h.deps.Log.Error("failed to send thought", "error", err)
	} else {
		h.deps.Log.Info("thought recorded", "event", input.HookEventName)
	}

	return NewEmptyOutput(), nil
}

// AfterMCPExecutionHandler handles afterMCPExecution events.
type AfterMCPExecutionHandler struct {
	deps *HandlerDependencies
}

// Handle processes the afterMCPExecution event.
func (h *AfterMCPExecutionHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	p, ok := payload.(*AfterMCPExecutionPayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	// Count tokens in tool input and result
	inputTokens, _ := h.deps.Tokenizer.CountTokens(h.deps.Config.DefaultModel, string(p.ToolInput))
	resultTokens, _ := h.deps.Tokenizer.CountTokens(h.deps.Config.DefaultModel, string(p.ResultJSON))
	totalTokens := inputTokens + resultTokens

	inputChars := CountChars(string(p.ToolInput))
	resultChars := CountChars(string(p.ResultJSON))

	h.deps.Log.Debug("tool metrics",
		"tool", p.ToolName,
		"input_tokens", inputTokens,
		"result_tokens", resultTokens,
		"conversation_id", input.ConversationID,
	)

	// Create token usage record
	usage := &domain.TokenUsage{
		ID:               uuid.New(),
		ConversationID:   parseUUID(input.ConversationID),
		GenerationID:     parseUUID(input.GenerationID),
		Model:            h.deps.Config.DefaultModel,
		Stage:            domain.TokenUsageStageTool,
		PromptTokens:     inputTokens,
		CompletionTokens: resultTokens,
		TotalTokens:      totalTokens,
		PromptChars:      inputChars,
		CompletionChars:  resultChars,
		WorkspaceRoots:   input.WorkspaceRoots,
		UserEmail:        "",
		HookEventName:    string(input.HookEventName),
		Timestamp:        time.Now(),
	}

	if err := h.deps.Client.CreateTokenUsage(ctx, usage); err != nil {
		h.deps.Log.Error("failed to send token usage", "error", err)
	} else {
		h.deps.Log.Info("token usage recorded", "stage", "tool", "tool", p.ToolName, "tokens", totalTokens)
	}

	return NewEmptyOutput(), nil
}

// AfterFileEditHandler handles afterFileEdit events.
type AfterFileEditHandler struct {
	deps *HandlerDependencies
}

// Handle processes the afterFileEdit event.
func (h *AfterFileEditHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	p, ok := payload.(*AfterFileEditPayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	// Count characters in edits
	var oldChars, newChars int64
	for _, edit := range p.Edits {
		oldChars += CountChars(edit.OldString)
		newChars += CountChars(edit.NewString)
	}

	h.deps.Log.Debug("file edit metrics",
		"file", p.FilePath,
		"old_chars", oldChars,
		"new_chars", newChars,
		"edit_count", len(p.Edits),
	)

	// Create token usage record for file edits
	usage := &domain.TokenUsage{
		ID:               uuid.New(),
		ConversationID:   parseUUID(input.ConversationID),
		GenerationID:     parseUUID(input.GenerationID),
		Model:            h.deps.Config.DefaultModel,
		Stage:            domain.TokenUsageStageFileEdit,
		PromptTokens:     0,
		CompletionTokens: 0,
		TotalTokens:      0,
		PromptChars:      oldChars,
		CompletionChars:  newChars,
		WorkspaceRoots:   input.WorkspaceRoots,
		UserEmail:        "",
		HookEventName:    string(input.HookEventName),
		Timestamp:        time.Now(),
	}

	if err := h.deps.Client.CreateTokenUsage(ctx, usage); err != nil {
		h.deps.Log.Error("failed to send token usage", "error", err)
	} else {
		h.deps.Log.Info("file edit recorded", "file", p.FilePath, "edits", len(p.Edits))
	}

	return NewEmptyOutput(), nil
}

// StopHandler handles stop events.
type StopHandler struct {
	deps *HandlerDependencies
}

// Handle processes the stop event.
func (h *StopHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	p, ok := payload.(*StopPayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	h.deps.Log.Info("agent loop stopped",
		"status", p.Status,
		"loop_count", p.LoopCount,
		"conversation_id", input.ConversationID,
	)

	// Create token usage record as session marker
	usage := &domain.TokenUsage{
		ID:               uuid.New(),
		ConversationID:   parseUUID(input.ConversationID),
		GenerationID:     parseUUID(input.GenerationID),
		Model:            h.deps.Config.DefaultModel,
		Stage:            domain.TokenUsageStageStop,
		PromptTokens:     0,
		CompletionTokens: 0,
		TotalTokens:      0,
		PromptChars:      0,
		CompletionChars:  0,
		WorkspaceRoots:   input.WorkspaceRoots,
		UserEmail:        "",
		HookEventName:    string(input.HookEventName),
		Timestamp:        time.Now(),
	}

	if err := h.deps.Client.CreateTokenUsage(ctx, usage); err != nil {
		h.deps.Log.Error("failed to send stop marker", "error", err)
	}

	return NewEmptyOutput(), nil
}

// BeforeShellExecutionHandler handles beforeShellExecution events.
type BeforeShellExecutionHandler struct {
	deps *HandlerDependencies
}

// Handle processes the beforeShellExecution event.
func (h *BeforeShellExecutionHandler) Handle(_ context.Context, input *HookInput, payload any) (*HookOutput, error) {
	p, ok := payload.(*BeforeShellExecutionPayload)
	if !ok {
		return NewAllowOutput(), nil
	}

	h.deps.Log.Debug("shell execution",
		"command", p.Command,
		"cwd", p.Cwd,
		"conversation_id", input.ConversationID,
	)

	// Always allow shell execution (observability only)
	return NewAllowOutput(), nil
}

// AfterShellExecutionHandler handles afterShellExecution events.
type AfterShellExecutionHandler struct {
	deps *HandlerDependencies
}

// Handle processes the afterShellExecution event.
func (h *AfterShellExecutionHandler) Handle(_ context.Context, input *HookInput, payload any) (*HookOutput, error) {
	p, ok := payload.(*AfterShellExecutionPayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	h.deps.Log.Debug("shell execution completed",
		"command", p.Command,
		"output_length", len(p.Output),
		"conversation_id", input.ConversationID,
	)

	return NewEmptyOutput(), nil
}

// BeforeMCPExecutionHandler handles beforeMCPExecution events.
type BeforeMCPExecutionHandler struct {
	deps *HandlerDependencies
}

// Handle processes the beforeMCPExecution event.
func (h *BeforeMCPExecutionHandler) Handle(_ context.Context, input *HookInput, payload any) (*HookOutput, error) {
	p, ok := payload.(*BeforeMCPExecutionPayload)
	if !ok {
		return NewAllowOutput(), nil
	}

	h.deps.Log.Debug("MCP execution",
		"tool", p.ToolName,
		"conversation_id", input.ConversationID,
	)

	// Always allow MCP execution (observability only)
	return NewAllowOutput(), nil
}

// BeforeReadFileHandler handles beforeReadFile events.
type BeforeReadFileHandler struct {
	deps *HandlerDependencies
}

// Handle processes the beforeReadFile event.
func (h *BeforeReadFileHandler) Handle(_ context.Context, input *HookInput, payload any) (*HookOutput, error) {
	p, ok := payload.(*BeforeReadFilePayload)
	if !ok {
		return NewAllowOutput(), nil
	}

	h.deps.Log.Debug("file read",
		"file", p.FilePath,
		"content_length", len(p.Content),
		"conversation_id", input.ConversationID,
	)

	// Always allow file read (observability only)
	return NewAllowOutput(), nil
}

// parseUUID parses a UUID string, returning a zero UUID on error.
func parseUUID(s string) uuid.UUID {
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil
	}
	return id
}


```

```go
package hook

import (
	"context"
	"fmt"
	"time"

	"github.com/ainslies/ctd/internal/domain"
	"github.com/ainslies/ctd/pkg/logger"
	"github.com/google/uuid"
)

// Handler processes a hook event and returns the appropriate output.
type Handler interface {
	// Handle processes the hook input and returns the output.
	Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error)
}

// HandlerDependencies contains the dependencies for handlers.
type HandlerDependencies struct {
	Client    *CTDClient
	Tokenizer Tokenizer
	Log       *logger.Logger
	Config    *Config
}

// Router routes hook events to the appropriate handler.
type Router struct {
	deps     *HandlerDependencies
	handlers map[HookEventName]Handler
}

// NewRouter creates a new router with all handlers registered.
func NewRouter(deps *HandlerDependencies) *Router {
	router := &Router{
		deps:     deps,
		handlers: make(map[HookEventName]Handler),
	}

	// Register handlers
	router.handlers[EventBeforeSubmitPrompt] = &BeforeSubmitPromptHandler{deps: deps}
	router.handlers[EventAfterAgentResponse] = &AfterAgentResponseHandler{deps: deps}
	router.handlers[EventAfterAgentThought] = &AfterAgentThoughtHandler{deps: deps}
	router.handlers[EventAfterMCPExecution] = &AfterMCPExecutionHandler{deps: deps}
	router.handlers[EventAfterFileEdit] = &AfterFileEditHandler{deps: deps}
	router.handlers[EventStop] = &StopHandler{deps: deps}
	router.handlers[EventBeforeShellExecution] = &BeforeShellExecutionHandler{deps: deps}
	router.handlers[EventAfterShellExecution] = &AfterShellExecutionHandler{deps: deps}
	router.handlers[EventBeforeMCPExecution] = &BeforeMCPExecutionHandler{deps: deps}
	router.handlers[EventBeforeReadFile] = &BeforeReadFileHandler{deps: deps}

	return router
}

// Route routes the input to the appropriate handler.
func (r *Router) Route(ctx context.Context, input *HookInput) (*HookOutput, error) {
	handler, ok := r.handlers[input.HookEventName]
	if !ok {
		// Unknown event, return empty output
		r.deps.Log.Warn("unknown hook event", "event", input.HookEventName)
		return NewEmptyOutput(), nil
	}

	payload, err := input.ParsePayload()
	if err != nil {
		r.deps.Log.Error("failed to parse payload", "error", err, "event", input.HookEventName)
		// Return appropriate default output even on parse error
		return getDefaultOutput(input.HookEventName), nil
	}

	output, err := handler.Handle(ctx, input, payload)
	if err != nil {
		return nil, fmt.Errorf("handler error for %s: %w", input.HookEventName, err)
	}
	return output, nil
}

// getDefaultOutput returns the default output for an event type.
//
//nolint:exhaustive // Only specific events need special handling, others use empty output
func getDefaultOutput(event HookEventName) *HookOutput {
	switch event {
	case EventBeforeSubmitPrompt:
		return NewContinueOutput()
	case EventBeforeShellExecution, EventBeforeMCPExecution, EventBeforeReadFile:
		return NewAllowOutput()
	default:
		return NewEmptyOutput()
	}
}

// BeforeSubmitPromptHandler handles beforeSubmitPrompt events.
type BeforeSubmitPromptHandler struct {
	deps *HandlerDependencies
}

// Handle processes the beforeSubmitPrompt event.
//
//nolint:varnamelen // promptPayload is clear in context
func (h *BeforeSubmitPromptHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	promptPayload, ok := payload.(*BeforeSubmitPromptPayload)
	if !ok {
		return NewContinueOutput(), nil
	}

	// Count tokens
	tokenCount, err := h.deps.Tokenizer.CountTokens(h.deps.Config.DefaultModel, promptPayload.Prompt)
	if err != nil {
		h.deps.Log.Error("failed to count tokens", "error", err)
		tokenCount = 0
	}

	charCount := CountChars(promptPayload.Prompt)

	h.deps.Log.Debug("prompt metrics",
		"tokens", tokenCount,
		"chars", charCount,
		"conversation_id", input.ConversationID,
	)

	// Create token usage record
	usage := &domain.TokenUsage{
		ID:               uuid.New(),
		ConversationID:   parseUUID(input.ConversationID),
		GenerationID:     parseUUID(input.GenerationID),
		Model:            h.deps.Config.DefaultModel,
		Stage:            domain.TokenUsageStagePrompt,
		PromptTokens:     tokenCount,
		CompletionTokens: 0,
		TotalTokens:      tokenCount,
		PromptChars:      charCount,
		CompletionChars:  0,
		WorkspaceRoots:   input.WorkspaceRoots,
		UserEmail:        "",
		HookEventName:    string(input.HookEventName),
		Timestamp:        time.Now(),
	}

	// Send to ctd (best-effort)
	if err := h.deps.Client.CreateTokenUsage(ctx, usage); err != nil {
		h.deps.Log.Error("failed to send token usage", "error", err)
	} else {
		h.deps.Log.Info("token usage recorded", "stage", "prompt", "tokens", tokenCount)
	}

	return NewContinueOutput(), nil
}

// AfterAgentResponseHandler handles afterAgentResponse events.
type AfterAgentResponseHandler struct {
	deps *HandlerDependencies
}

// Handle processes the afterAgentResponse event.
//
//nolint:varnamelen // responsePayload is clear in context
func (h *AfterAgentResponseHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	responsePayload, ok := payload.(*AfterAgentResponsePayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	// Count tokens
	tokenCount, err := h.deps.Tokenizer.CountTokens(h.deps.Config.DefaultModel, responsePayload.Text)
	if err != nil {
		h.deps.Log.Error("failed to count tokens", "error", err)
		tokenCount = 0
	}

	charCount := CountChars(responsePayload.Text)

	h.deps.Log.Debug("completion metrics",
		"tokens", tokenCount,
		"chars", charCount,
		"conversation_id", input.ConversationID,
	)

	// Create token usage record
	usage := &domain.TokenUsage{
		ID:               uuid.New(),
		ConversationID:   parseUUID(input.ConversationID),
		GenerationID:     parseUUID(input.GenerationID),
		Model:            h.deps.Config.DefaultModel,
		Stage:            domain.TokenUsageStageCompletion,
		PromptTokens:     0,
		CompletionTokens: tokenCount,
		TotalTokens:      tokenCount,
		PromptChars:      0,
		CompletionChars:  charCount,
		WorkspaceRoots:   input.WorkspaceRoots,
		UserEmail:        "",
		HookEventName:    string(input.HookEventName),
		Timestamp:        time.Now(),
	}

	// Send token usage to ctd (best-effort)
	if err := h.deps.Client.CreateTokenUsage(ctx, usage); err != nil {
		h.deps.Log.Error("failed to send token usage", "error", err)
	} else {
		h.deps.Log.Info("token usage recorded", "stage", "completion", "tokens", tokenCount)
	}

	// Also send as thought for visibility
	thought := &domain.Thought{
		ID:             uuid.New(),
		ConversationID: parseUUID(input.ConversationID),
		GenerationID:   parseUUID(input.GenerationID),
		Text:           p.Text,
		DurationMs:     0,
		Model:          h.deps.Config.DefaultModel,
		HookEventName:  string(input.HookEventName),
		CursorVersion:  "",
		WorkspaceRoots: input.WorkspaceRoots,
		UserEmail:      "",
		Timestamp:      time.Now(),
	}

	if err := h.deps.Client.CreateThought(ctx, thought); err != nil {
		h.deps.Log.Error("failed to send thought", "error", err)
	}

	return NewEmptyOutput(), nil
}

// AfterAgentThoughtHandler handles afterAgentThought events.
type AfterAgentThoughtHandler struct {
	deps *HandlerDependencies
}

// Handle processes the afterAgentThought event.
func (h *AfterAgentThoughtHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	p, ok := payload.(*AfterAgentThoughtPayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	// Send as thought for visibility
	thought := &domain.Thought{
		ID:             uuid.New(),
		ConversationID: parseUUID(input.ConversationID),
		GenerationID:   parseUUID(input.GenerationID),
		Text:           p.Text,
		DurationMs:     0,
		Model:          h.deps.Config.DefaultModel,
		HookEventName:  string(input.HookEventName),
		CursorVersion:  "",
		WorkspaceRoots: input.WorkspaceRoots,
		UserEmail:      "",
		Timestamp:      time.Now(),
	}

	if err := h.deps.Client.CreateThought(ctx, thought); err != nil {
		h.deps.Log.Error("failed to send thought", "error", err)
	} else {
		h.deps.Log.Info("thought recorded", "event", input.HookEventName)
	}

	return NewEmptyOutput(), nil
}

// AfterMCPExecutionHandler handles afterMCPExecution events.
type AfterMCPExecutionHandler struct {
	deps *HandlerDependencies
}

// Handle processes the afterMCPExecution event.
func (h *AfterMCPExecutionHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	p, ok := payload.(*AfterMCPExecutionPayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	// Count tokens in tool input and result
	inputTokens, _ := h.deps.Tokenizer.CountTokens(h.deps.Config.DefaultModel, string(p.ToolInput))
	resultTokens, _ := h.deps.Tokenizer.CountTokens(h.deps.Config.DefaultModel, string(p.ResultJSON))
	totalTokens := inputTokens + resultTokens

	inputChars := CountChars(string(p.ToolInput))
	resultChars := CountChars(string(p.ResultJSON))

	h.deps.Log.Debug("tool metrics",
		"tool", p.ToolName,
		"input_tokens", inputTokens,
		"result_tokens", resultTokens,
		"conversation_id", input.ConversationID,
	)

	// Create token usage record
	usage := &domain.TokenUsage{
		ID:               uuid.New(),
		ConversationID:   parseUUID(input.ConversationID),
		GenerationID:     parseUUID(input.GenerationID),
		Model:            h.deps.Config.DefaultModel,
		Stage:            domain.TokenUsageStageTool,
		PromptTokens:     inputTokens,
		CompletionTokens: resultTokens,
		TotalTokens:      totalTokens,
		PromptChars:      inputChars,
		CompletionChars:  resultChars,
		WorkspaceRoots:   input.WorkspaceRoots,
		UserEmail:        "",
		HookEventName:    string(input.HookEventName),
		Timestamp:        time.Now(),
	}

	if err := h.deps.Client.CreateTokenUsage(ctx, usage); err != nil {
		h.deps.Log.Error("failed to send token usage", "error", err)
	} else {
		h.deps.Log.Info("token usage recorded", "stage", "tool", "tool", p.ToolName, "tokens", totalTokens)
	}

	return NewEmptyOutput(), nil
}

// AfterFileEditHandler handles afterFileEdit events.
type AfterFileEditHandler struct {
	deps *HandlerDependencies
}

// Handle processes the afterFileEdit event.
func (h *AfterFileEditHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	p, ok := payload.(*AfterFileEditPayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	// Count characters in edits
	var oldChars, newChars int64
	for _, edit := range p.Edits {
		oldChars += CountChars(edit.OldString)
		newChars += CountChars(edit.NewString)
	}

	h.deps.Log.Debug("file edit metrics",
		"file", p.FilePath,
		"old_chars", oldChars,
		"new_chars", newChars,
		"edit_count", len(p.Edits),
	)

	// Create token usage record for file edits
	usage := &domain.TokenUsage{
		ID:               uuid.New(),
		ConversationID:   parseUUID(input.ConversationID),
		GenerationID:     parseUUID(input.GenerationID),
		Model:            h.deps.Config.DefaultModel,
		Stage:            domain.TokenUsageStageFileEdit,
		PromptTokens:     0,
		CompletionTokens: 0,
		TotalTokens:      0,
		PromptChars:      oldChars,
		CompletionChars:  newChars,
		WorkspaceRoots:   input.WorkspaceRoots,
		UserEmail:        "",
		HookEventName:    string(input.HookEventName),
		Timestamp:        time.Now(),
	}

	if err := h.deps.Client.CreateTokenUsage(ctx, usage); err != nil {
		h.deps.Log.Error("failed to send token usage", "error", err)
	} else {
		h.deps.Log.Info("file edit recorded", "file", p.FilePath, "edits", len(p.Edits))
	}

	return NewEmptyOutput(), nil
}

// StopHandler handles stop events.
type StopHandler struct {
	deps *HandlerDependencies
}

// Handle processes the stop event.
func (h *StopHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	p, ok := payload.(*StopPayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	h.deps.Log.Info("agent loop stopped",
		"status", p.Status,
		"loop_count", p.LoopCount,
		"conversation_id", input.ConversationID,
	)

	// Create token usage record as session marker
	usage := &domain.TokenUsage{
		ID:               uuid.New(),
		ConversationID:   parseUUID(input.ConversationID),
		GenerationID:     parseUUID(input.GenerationID),
		Model:            h.deps.Config.DefaultModel,
		Stage:            domain.TokenUsageStageStop,
		PromptTokens:     0,
		CompletionTokens: 0,
		TotalTokens:      0,
		PromptChars:      0,
		CompletionChars:  0,
		WorkspaceRoots:   input.WorkspaceRoots,
		UserEmail:        "",
		HookEventName:    string(input.HookEventName),
		Timestamp:        time.Now(),
	}

	if err := h.deps.Client.CreateTokenUsage(ctx, usage); err != nil {
		h.deps.Log.Error("failed to send stop marker", "error", err)
	}

	return NewEmptyOutput(), nil
}

// BeforeShellExecutionHandler handles beforeShellExecution events.
type BeforeShellExecutionHandler struct {
	deps *HandlerDependencies
}

// Handle processes the beforeShellExecution event.
func (h *BeforeShellExecutionHandler) Handle(_ context.Context, input *HookInput, payload any) (*HookOutput, error) {
	p, ok := payload.(*BeforeShellExecutionPayload)
	if !ok {
		return NewAllowOutput(), nil
	}

	h.deps.Log.Debug("shell execution",
		"command", p.Command,
		"cwd", p.Cwd,
		"conversation_id", input.ConversationID,
	)

	// Always allow shell execution (observability only)
	return NewAllowOutput(), nil
}

// AfterShellExecutionHandler handles afterShellExecution events.
type AfterShellExecutionHandler struct {
	deps *HandlerDependencies
}

// Handle processes the afterShellExecution event.
func (h *AfterShellExecutionHandler) Handle(_ context.Context, input *HookInput, payload any) (*HookOutput, error) {
	p, ok := payload.(*AfterShellExecutionPayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	h.deps.Log.Debug("shell execution completed",
		"command", p.Command,
		"output_length", len(p.Output),
		"conversation_id", input.ConversationID,
	)

	return NewEmptyOutput(), nil
}

// BeforeMCPExecutionHandler handles beforeMCPExecution events.
type BeforeMCPExecutionHandler struct {
	deps *HandlerDependencies
}

// Handle processes the beforeMCPExecution event.
func (h *BeforeMCPExecutionHandler) Handle(_ context.Context, input *HookInput, payload any) (*HookOutput, error) {
	p, ok := payload.(*BeforeMCPExecutionPayload)
	if !ok {
		return NewAllowOutput(), nil
	}

	h.deps.Log.Debug("MCP execution",
		"tool", p.ToolName,
		"conversation_id", input.ConversationID,
	)

	// Always allow MCP execution (observability only)
	return NewAllowOutput(), nil
}

// BeforeReadFileHandler handles beforeReadFile events.
type BeforeReadFileHandler struct {
	deps *HandlerDependencies
}

// Handle processes the beforeReadFile event.
func (h *BeforeReadFileHandler) Handle(_ context.Context, input *HookInput, payload any) (*HookOutput, error) {
	p, ok := payload.(*BeforeReadFilePayload)
	if !ok {
		return NewAllowOutput(), nil
	}

	h.deps.Log.Debug("file read",
		"file", p.FilePath,
		"content_length", len(p.Content),
		"conversation_id", input.ConversationID,
	)

	// Always allow file read (observability only)
	return NewAllowOutput(), nil
}

// parseUUID parses a UUID string, returning a zero UUID on error.
func parseUUID(s string) uuid.UUID {
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil
	}
	return id
}


```

```go
package hook

import (
	"context"
	"fmt"
	"time"

	"github.com/ainslies/ctd/internal/domain"
	"github.com/ainslies/ctd/pkg/logger"
	"github.com/google/uuid"
)

// Handler processes a hook event and returns the appropriate output.
type Handler interface {
	// Handle processes the hook input and returns the output.
	Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error)
}

// HandlerDependencies contains the dependencies for handlers.
type HandlerDependencies struct {
	Client    *CTDClient
	Tokenizer Tokenizer
	Log       *logger.Logger
	Config    *Config
}

// Router routes hook events to the appropriate handler.
type Router struct {
	deps     *HandlerDependencies
	handlers map[HookEventName]Handler
}

// NewRouter creates a new router with all handlers registered.
func NewRouter(deps *HandlerDependencies) *Router {
	router := &Router{
		deps:     deps,
		handlers: make(map[HookEventName]Handler),
	}

	// Register handlers
	router.handlers[EventBeforeSubmitPrompt] = &BeforeSubmitPromptHandler{deps: deps}
	router.handlers[EventAfterAgentResponse] = &AfterAgentResponseHandler{deps: deps}
	router.handlers[EventAfterAgentThought] = &AfterAgentThoughtHandler{deps: deps}
	router.handlers[EventAfterMCPExecution] = &AfterMCPExecutionHandler{deps: deps}
	router.handlers[EventAfterFileEdit] = &AfterFileEditHandler{deps: deps}
	router.handlers[EventStop] = &StopHandler{deps: deps}
	router.handlers[EventBeforeShellExecution] = &BeforeShellExecutionHandler{deps: deps}
	router.handlers[EventAfterShellExecution] = &AfterShellExecutionHandler{deps: deps}
	router.handlers[EventBeforeMCPExecution] = &BeforeMCPExecutionHandler{deps: deps}
	router.handlers[EventBeforeReadFile] = &BeforeReadFileHandler{deps: deps}

	return router
}

// Route routes the input to the appropriate handler.
func (r *Router) Route(ctx context.Context, input *HookInput) (*HookOutput, error) {
	handler, ok := r.handlers[input.HookEventName]
	if !ok {
		// Unknown event, return empty output
		r.deps.Log.Warn("unknown hook event", "event", input.HookEventName)
		return NewEmptyOutput(), nil
	}

	payload, err := input.ParsePayload()
	if err != nil {
		r.deps.Log.Error("failed to parse payload", "error", err, "event", input.HookEventName)
		// Return appropriate default output even on parse error
		return getDefaultOutput(input.HookEventName), nil
	}

	output, err := handler.Handle(ctx, input, payload)
	if err != nil {
		return nil, fmt.Errorf("handler error for %s: %w", input.HookEventName, err)
	}
	return output, nil
}

// getDefaultOutput returns the default output for an event type.
//
//nolint:exhaustive // Only specific events need special handling, others use empty output
func getDefaultOutput(event HookEventName) *HookOutput {
	switch event {
	case EventBeforeSubmitPrompt:
		return NewContinueOutput()
	case EventBeforeShellExecution, EventBeforeMCPExecution, EventBeforeReadFile:
		return NewAllowOutput()
	default:
		return NewEmptyOutput()
	}
}

// BeforeSubmitPromptHandler handles beforeSubmitPrompt events.
type BeforeSubmitPromptHandler struct {
	deps *HandlerDependencies
}

// Handle processes the beforeSubmitPrompt event.
//
//nolint:varnamelen // promptPayload is clear in context
func (h *BeforeSubmitPromptHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	promptPayload, ok := payload.(*BeforeSubmitPromptPayload)
	if !ok {
		return NewContinueOutput(), nil
	}

	// Count tokens
	tokenCount, err := h.deps.Tokenizer.CountTokens(h.deps.Config.DefaultModel, promptPayload.Prompt)
	if err != nil {
		h.deps.Log.Error("failed to count tokens", "error", err)
		tokenCount = 0
	}

	charCount := CountChars(promptPayload.Prompt)

	h.deps.Log.Debug("prompt metrics",
		"tokens", tokenCount,
		"chars", charCount,
		"conversation_id", input.ConversationID,
	)

	// Create token usage record
	usage := &domain.TokenUsage{
		ID:               uuid.New(),
		ConversationID:   parseUUID(input.ConversationID),
		GenerationID:     parseUUID(input.GenerationID),
		Model:            h.deps.Config.DefaultModel,
		Stage:            domain.TokenUsageStagePrompt,
		PromptTokens:     tokenCount,
		CompletionTokens: 0,
		TotalTokens:      tokenCount,
		PromptChars:      charCount,
		CompletionChars:  0,
		WorkspaceRoots:   input.WorkspaceRoots,
		UserEmail:        "",
		HookEventName:    string(input.HookEventName),
		Timestamp:        time.Now(),
	}

	// Send to ctd (best-effort)
	if err := h.deps.Client.CreateTokenUsage(ctx, usage); err != nil {
		h.deps.Log.Error("failed to send token usage", "error", err)
	} else {
		h.deps.Log.Info("token usage recorded", "stage", "prompt", "tokens", tokenCount)
	}

	return NewContinueOutput(), nil
}

// AfterAgentResponseHandler handles afterAgentResponse events.
type AfterAgentResponseHandler struct {
	deps *HandlerDependencies
}

// Handle processes the afterAgentResponse event.
//
//nolint:varnamelen // responsePayload is clear in context
func (h *AfterAgentResponseHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	responsePayload, ok := payload.(*AfterAgentResponsePayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	// Count tokens
	tokenCount, err := h.deps.Tokenizer.CountTokens(h.deps.Config.DefaultModel, responsePayload.Text)
	if err != nil {
		h.deps.Log.Error("failed to count tokens", "error", err)
		tokenCount = 0
	}

	charCount := CountChars(responsePayload.Text)

	h.deps.Log.Debug("completion metrics",
		"tokens", tokenCount,
		"chars", charCount,
		"conversation_id", input.ConversationID,
	)

	// Create token usage record
	usage := &domain.TokenUsage{
		ID:               uuid.New(),
		ConversationID:   parseUUID(input.ConversationID),
		GenerationID:     parseUUID(input.GenerationID),
		Model:            h.deps.Config.DefaultModel,
		Stage:            domain.TokenUsageStageCompletion,
		PromptTokens:     0,
		CompletionTokens: tokenCount,
		TotalTokens:      tokenCount,
		PromptChars:      0,
		CompletionChars:  charCount,
		WorkspaceRoots:   input.WorkspaceRoots,
		UserEmail:        "",
		HookEventName:    string(input.HookEventName),
		Timestamp:        time.Now(),
	}

	// Send token usage to ctd (best-effort)
	if err := h.deps.Client.CreateTokenUsage(ctx, usage); err != nil {
		h.deps.Log.Error("failed to send token usage", "error", err)
	} else {
		h.deps.Log.Info("token usage recorded", "stage", "completion", "tokens", tokenCount)
	}

	// Also send as thought for visibility
	thought := &domain.Thought{
		ID:             uuid.New(),
		ConversationID: parseUUID(input.ConversationID),
		GenerationID:   parseUUID(input.GenerationID),
		Text:           responsePayload.Text,
		DurationMs:     0,
		Model:          h.deps.Config.DefaultModel,
		HookEventName:  string(input.HookEventName),
		CursorVersion:  "",
		WorkspaceRoots: input.WorkspaceRoots,
		UserEmail:      "",
		Timestamp:      time.Now(),
	}

	if err := h.deps.Client.CreateThought(ctx, thought); err != nil {
		h.deps.Log.Error("failed to send thought", "error", err)
	}

	return NewEmptyOutput(), nil
}

// AfterAgentThoughtHandler handles afterAgentThought events.
type AfterAgentThoughtHandler struct {
	deps *HandlerDependencies
}

// Handle processes the afterAgentThought event.
func (h *AfterAgentThoughtHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	p, ok := payload.(*AfterAgentThoughtPayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	// Send as thought for visibility
	thought := &domain.Thought{
		ID:             uuid.New(),
		ConversationID: parseUUID(input.ConversationID),
		GenerationID:   parseUUID(input.GenerationID),
		Text:           p.Text,
		DurationMs:     0,
		Model:          h.deps.Config.DefaultModel,
		HookEventName:  string(input.HookEventName),
		CursorVersion:  "",
		WorkspaceRoots: input.WorkspaceRoots,
		UserEmail:      "",
		Timestamp:      time.Now(),
	}

	if err := h.deps.Client.CreateThought(ctx, thought); err != nil {
		h.deps.Log.Error("failed to send thought", "error", err)
	} else {
		h.deps.Log.Info("thought recorded", "event", input.HookEventName)
	}

	return NewEmptyOutput(), nil
}

// AfterMCPExecutionHandler handles afterMCPExecution events.
type AfterMCPExecutionHandler struct {
	deps *HandlerDependencies
}

// Handle processes the afterMCPExecution event.
func (h *AfterMCPExecutionHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	p, ok := payload.(*AfterMCPExecutionPayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	// Count tokens in tool input and result
	inputTokens, _ := h.deps.Tokenizer.CountTokens(h.deps.Config.DefaultModel, string(p.ToolInput))
	resultTokens, _ := h.deps.Tokenizer.CountTokens(h.deps.Config.DefaultModel, string(p.ResultJSON))
	totalTokens := inputTokens + resultTokens

	inputChars := CountChars(string(p.ToolInput))
	resultChars := CountChars(string(p.ResultJSON))

	h.deps.Log.Debug("tool metrics",
		"tool", p.ToolName,
		"input_tokens", inputTokens,
		"result_tokens", resultTokens,
		"conversation_id", input.ConversationID,
	)

	// Create token usage record
	usage := &domain.TokenUsage{
		ID:               uuid.New(),
		ConversationID:   parseUUID(input.ConversationID),
		GenerationID:     parseUUID(input.GenerationID),
		Model:            h.deps.Config.DefaultModel,
		Stage:            domain.TokenUsageStageTool,
		PromptTokens:     inputTokens,
		CompletionTokens: resultTokens,
		TotalTokens:      totalTokens,
		PromptChars:      inputChars,
		CompletionChars:  resultChars,
		WorkspaceRoots:   input.WorkspaceRoots,
		UserEmail:        "",
		HookEventName:    string(input.HookEventName),
		Timestamp:        time.Now(),
	}

	if err := h.deps.Client.CreateTokenUsage(ctx, usage); err != nil {
		h.deps.Log.Error("failed to send token usage", "error", err)
	} else {
		h.deps.Log.Info("token usage recorded", "stage", "tool", "tool", p.ToolName, "tokens", totalTokens)
	}

	return NewEmptyOutput(), nil
}

// AfterFileEditHandler handles afterFileEdit events.
type AfterFileEditHandler struct {
	deps *HandlerDependencies
}

// Handle processes the afterFileEdit event.
func (h *AfterFileEditHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	p, ok := payload.(*AfterFileEditPayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	// Count characters in edits
	var oldChars, newChars int64
	for _, edit := range p.Edits {
		oldChars += CountChars(edit.OldString)
		newChars += CountChars(edit.NewString)
	}

	h.deps.Log.Debug("file edit metrics",
		"file", p.FilePath,
		"old_chars", oldChars,
		"new_chars", newChars,
		"edit_count", len(p.Edits),
	)

	// Create token usage record for file edits
	usage := &domain.TokenUsage{
		ID:               uuid.New(),
		ConversationID:   parseUUID(input.ConversationID),
		GenerationID:     parseUUID(input.GenerationID),
		Model:            h.deps.Config.DefaultModel,
		Stage:            domain.TokenUsageStageFileEdit,
		PromptTokens:     0,
		CompletionTokens: 0,
		TotalTokens:      0,
		PromptChars:      oldChars,
		CompletionChars:  newChars,
		WorkspaceRoots:   input.WorkspaceRoots,
		UserEmail:        "",
		HookEventName:    string(input.HookEventName),
		Timestamp:        time.Now(),
	}

	if err := h.deps.Client.CreateTokenUsage(ctx, usage); err != nil {
		h.deps.Log.Error("failed to send token usage", "error", err)
	} else {
		h.deps.Log.Info("file edit recorded", "file", p.FilePath, "edits", len(p.Edits))
	}

	return NewEmptyOutput(), nil
}

// StopHandler handles stop events.
type StopHandler struct {
	deps *HandlerDependencies
}

// Handle processes the stop event.
func (h *StopHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	p, ok := payload.(*StopPayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	h.deps.Log.Info("agent loop stopped",
		"status", p.Status,
		"loop_count", p.LoopCount,
		"conversation_id", input.ConversationID,
	)

	// Create token usage record as session marker
	usage := &domain.TokenUsage{
		ID:               uuid.New(),
		ConversationID:   parseUUID(input.ConversationID),
		GenerationID:     parseUUID(input.GenerationID),
		Model:            h.deps.Config.DefaultModel,
		Stage:            domain.TokenUsageStageStop,
		PromptTokens:     0,
		CompletionTokens: 0,
		TotalTokens:      0,
		PromptChars:      0,
		CompletionChars:  0,
		WorkspaceRoots:   input.WorkspaceRoots,
		UserEmail:        "",
		HookEventName:    string(input.HookEventName),
		Timestamp:        time.Now(),
	}

	if err := h.deps.Client.CreateTokenUsage(ctx, usage); err != nil {
		h.deps.Log.Error("failed to send stop marker", "error", err)
	}

	return NewEmptyOutput(), nil
}

// BeforeShellExecutionHandler handles beforeShellExecution events.
type BeforeShellExecutionHandler struct {
	deps *HandlerDependencies
}

// Handle processes the beforeShellExecution event.
func (h *BeforeShellExecutionHandler) Handle(_ context.Context, input *HookInput, payload any) (*HookOutput, error) {
	p, ok := payload.(*BeforeShellExecutionPayload)
	if !ok {
		return NewAllowOutput(), nil
	}

	h.deps.Log.Debug("shell execution",
		"command", p.Command,
		"cwd", p.Cwd,
		"conversation_id", input.ConversationID,
	)

	// Always allow shell execution (observability only)
	return NewAllowOutput(), nil
}

// AfterShellExecutionHandler handles afterShellExecution events.
type AfterShellExecutionHandler struct {
	deps *HandlerDependencies
}

// Handle processes the afterShellExecution event.
func (h *AfterShellExecutionHandler) Handle(_ context.Context, input *HookInput, payload any) (*HookOutput, error) {
	p, ok := payload.(*AfterShellExecutionPayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	h.deps.Log.Debug("shell execution completed",
		"command", p.Command,
		"output_length", len(p.Output),
		"conversation_id", input.ConversationID,
	)

	return NewEmptyOutput(), nil
}

// BeforeMCPExecutionHandler handles beforeMCPExecution events.
type BeforeMCPExecutionHandler struct {
	deps *HandlerDependencies
}

// Handle processes the beforeMCPExecution event.
func (h *BeforeMCPExecutionHandler) Handle(_ context.Context, input *HookInput, payload any) (*HookOutput, error) {
	p, ok := payload.(*BeforeMCPExecutionPayload)
	if !ok {
		return NewAllowOutput(), nil
	}

	h.deps.Log.Debug("MCP execution",
		"tool", p.ToolName,
		"conversation_id", input.ConversationID,
	)

	// Always allow MCP execution (observability only)
	return NewAllowOutput(), nil
}

// BeforeReadFileHandler handles beforeReadFile events.
type BeforeReadFileHandler struct {
	deps *HandlerDependencies
}

// Handle processes the beforeReadFile event.
func (h *BeforeReadFileHandler) Handle(_ context.Context, input *HookInput, payload any) (*HookOutput, error) {
	p, ok := payload.(*BeforeReadFilePayload)
	if !ok {
		return NewAllowOutput(), nil
	}

	h.deps.Log.Debug("file read",
		"file", p.FilePath,
		"content_length", len(p.Content),
		"conversation_id", input.ConversationID,
	)

	// Always allow file read (observability only)
	return NewAllowOutput(), nil
}

// parseUUID parses a UUID string, returning a zero UUID on error.
func parseUUID(s string) uuid.UUID {
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil
	}
	return id
}


```

```go
package hook

import (
	"context"
	"fmt"
	"time"

	"github.com/ainslies/ctd/internal/domain"
	"github.com/ainslies/ctd/pkg/logger"
	"github.com/google/uuid"
)

// Handler processes a hook event and returns the appropriate output.
type Handler interface {
	// Handle processes the hook input and returns the output.
	Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error)
}

// HandlerDependencies contains the dependencies for handlers.
type HandlerDependencies struct {
	Client    *CTDClient
	Tokenizer Tokenizer
	Log       *logger.Logger
	Config    *Config
}

// Router routes hook events to the appropriate handler.
type Router struct {
	deps     *HandlerDependencies
	handlers map[HookEventName]Handler
}

// NewRouter creates a new router with all handlers registered.
func NewRouter(deps *HandlerDependencies) *Router {
	router := &Router{
		deps:     deps,
		handlers: make(map[HookEventName]Handler),
	}

	// Register handlers
	router.handlers[EventBeforeSubmitPrompt] = &BeforeSubmitPromptHandler{deps: deps}
	router.handlers[EventAfterAgentResponse] = &AfterAgentResponseHandler{deps: deps}
	router.handlers[EventAfterAgentThought] = &AfterAgentThoughtHandler{deps: deps}
	router.handlers[EventAfterMCPExecution] = &AfterMCPExecutionHandler{deps: deps}
	router.handlers[EventAfterFileEdit] = &AfterFileEditHandler{deps: deps}
	router.handlers[EventStop] = &StopHandler{deps: deps}
	router.handlers[EventBeforeShellExecution] = &BeforeShellExecutionHandler{deps: deps}
	router.handlers[EventAfterShellExecution] = &AfterShellExecutionHandler{deps: deps}
	router.handlers[EventBeforeMCPExecution] = &BeforeMCPExecutionHandler{deps: deps}
	router.handlers[EventBeforeReadFile] = &BeforeReadFileHandler{deps: deps}

	return router
}

// Route routes the input to the appropriate handler.
func (r *Router) Route(ctx context.Context, input *HookInput) (*HookOutput, error) {
	handler, ok := r.handlers[input.HookEventName]
	if !ok {
		// Unknown event, return empty output
		r.deps.Log.Warn("unknown hook event", "event", input.HookEventName)
		return NewEmptyOutput(), nil
	}

	payload, err := input.ParsePayload()
	if err != nil {
		r.deps.Log.Error("failed to parse payload", "error", err, "event", input.HookEventName)
		// Return appropriate default output even on parse error
		return getDefaultOutput(input.HookEventName), nil
	}

	output, err := handler.Handle(ctx, input, payload)
	if err != nil {
		return nil, fmt.Errorf("handler error for %s: %w", input.HookEventName, err)
	}
	return output, nil
}

// getDefaultOutput returns the default output for an event type.
//
//nolint:exhaustive // Only specific events need special handling, others use empty output
func getDefaultOutput(event HookEventName) *HookOutput {
	switch event {
	case EventBeforeSubmitPrompt:
		return NewContinueOutput()
	case EventBeforeShellExecution, EventBeforeMCPExecution, EventBeforeReadFile:
		return NewAllowOutput()
	default:
		return NewEmptyOutput()
	}
}

// BeforeSubmitPromptHandler handles beforeSubmitPrompt events.
type BeforeSubmitPromptHandler struct {
	deps *HandlerDependencies
}

// Handle processes the beforeSubmitPrompt event.
//
//nolint:varnamelen // promptPayload is clear in context
func (h *BeforeSubmitPromptHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	promptPayload, ok := payload.(*BeforeSubmitPromptPayload)
	if !ok {
		return NewContinueOutput(), nil
	}

	// Count tokens
	tokenCount, err := h.deps.Tokenizer.CountTokens(h.deps.Config.DefaultModel, promptPayload.Prompt)
	if err != nil {
		h.deps.Log.Error("failed to count tokens", "error", err)
		tokenCount = 0
	}

	charCount := CountChars(promptPayload.Prompt)

	h.deps.Log.Debug("prompt metrics",
		"tokens", tokenCount,
		"chars", charCount,
		"conversation_id", input.ConversationID,
	)

	// Create token usage record
	usage := &domain.TokenUsage{
		ID:               uuid.New(),
		ConversationID:   parseUUID(input.ConversationID),
		GenerationID:     parseUUID(input.GenerationID),
		Model:            h.deps.Config.DefaultModel,
		Stage:            domain.TokenUsageStagePrompt,
		PromptTokens:     tokenCount,
		CompletionTokens: 0,
		TotalTokens:      tokenCount,
		PromptChars:      charCount,
		CompletionChars:  0,
		WorkspaceRoots:   input.WorkspaceRoots,
		UserEmail:        "",
		HookEventName:    string(input.HookEventName),
		Timestamp:        time.Now(),
	}

	// Send to ctd (best-effort)
	if err := h.deps.Client.CreateTokenUsage(ctx, usage); err != nil {
		h.deps.Log.Error("failed to send token usage", "error", err)
	} else {
		h.deps.Log.Info("token usage recorded", "stage", "prompt", "tokens", tokenCount)
	}

	return NewContinueOutput(), nil
}

// AfterAgentResponseHandler handles afterAgentResponse events.
type AfterAgentResponseHandler struct {
	deps *HandlerDependencies
}

// Handle processes the afterAgentResponse event.
//
//nolint:varnamelen // responsePayload is clear in context
func (h *AfterAgentResponseHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	responsePayload, ok := payload.(*AfterAgentResponsePayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	// Count tokens
	tokenCount, err := h.deps.Tokenizer.CountTokens(h.deps.Config.DefaultModel, responsePayload.Text)
	if err != nil {
		h.deps.Log.Error("failed to count tokens", "error", err)
		tokenCount = 0
	}

	charCount := CountChars(responsePayload.Text)

	h.deps.Log.Debug("completion metrics",
		"tokens", tokenCount,
		"chars", charCount,
		"conversation_id", input.ConversationID,
	)

	// Create token usage record
	usage := &domain.TokenUsage{
		ID:               uuid.New(),
		ConversationID:   parseUUID(input.ConversationID),
		GenerationID:     parseUUID(input.GenerationID),
		Model:            h.deps.Config.DefaultModel,
		Stage:            domain.TokenUsageStageCompletion,
		PromptTokens:     0,
		CompletionTokens: tokenCount,
		TotalTokens:      tokenCount,
		PromptChars:      0,
		CompletionChars:  charCount,
		WorkspaceRoots:   input.WorkspaceRoots,
		UserEmail:        "",
		HookEventName:    string(input.HookEventName),
		Timestamp:        time.Now(),
	}

	// Send token usage to ctd (best-effort)
	if err := h.deps.Client.CreateTokenUsage(ctx, usage); err != nil {
		h.deps.Log.Error("failed to send token usage", "error", err)
	} else {
		h.deps.Log.Info("token usage recorded", "stage", "completion", "tokens", tokenCount)
	}

	// Also send as thought for visibility
	thought := &domain.Thought{
		ID:             uuid.New(),
		ConversationID: parseUUID(input.ConversationID),
		GenerationID:   parseUUID(input.GenerationID),
		Text:           responsePayload.Text,
		DurationMs:     0,
		Model:          h.deps.Config.DefaultModel,
		HookEventName:  string(input.HookEventName),
		CursorVersion:  "",
		WorkspaceRoots: input.WorkspaceRoots,
		UserEmail:      "",
		Timestamp:      time.Now(),
	}

	if err := h.deps.Client.CreateThought(ctx, thought); err != nil {
		h.deps.Log.Error("failed to send thought", "error", err)
	}

	return NewEmptyOutput(), nil
}

// AfterAgentThoughtHandler handles afterAgentThought events.
type AfterAgentThoughtHandler struct {
	deps *HandlerDependencies
}

// Handle processes the afterAgentThought event.
//
//nolint:varnamelen // thoughtPayload is clear in context
func (h *AfterAgentThoughtHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	thoughtPayload, ok := payload.(*AfterAgentThoughtPayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	// Send as thought for visibility
	thought := &domain.Thought{
		ID:             uuid.New(),
		ConversationID: parseUUID(input.ConversationID),
		GenerationID:   parseUUID(input.GenerationID),
		Text:           thoughtPayload.Text,
		DurationMs:     0,
		Model:          h.deps.Config.DefaultModel,
		HookEventName:  string(input.HookEventName),
		CursorVersion:  "",
		WorkspaceRoots: input.WorkspaceRoots,
		UserEmail:      "",
		Timestamp:      time.Now(),
	}

	if err := h.deps.Client.CreateThought(ctx, thought); err != nil {
		h.deps.Log.Error("failed to send thought", "error", err)
	} else {
		h.deps.Log.Info("thought recorded", "event", input.HookEventName)
	}

	return NewEmptyOutput(), nil
}

// AfterMCPExecutionHandler handles afterMCPExecution events.
type AfterMCPExecutionHandler struct {
	deps *HandlerDependencies
}

// Handle processes the afterMCPExecution event.
func (h *AfterMCPExecutionHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	p, ok := payload.(*AfterMCPExecutionPayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	// Count tokens in tool input and result
	inputTokens, _ := h.deps.Tokenizer.CountTokens(h.deps.Config.DefaultModel, string(p.ToolInput))
	resultTokens, _ := h.deps.Tokenizer.CountTokens(h.deps.Config.DefaultModel, string(p.ResultJSON))
	totalTokens := inputTokens + resultTokens

	inputChars := CountChars(string(p.ToolInput))
	resultChars := CountChars(string(p.ResultJSON))

	h.deps.Log.Debug("tool metrics",
		"tool", p.ToolName,
		"input_tokens", inputTokens,
		"result_tokens", resultTokens,
		"conversation_id", input.ConversationID,
	)

	// Create token usage record
	usage := &domain.TokenUsage{
		ID:               uuid.New(),
		ConversationID:   parseUUID(input.ConversationID),
		GenerationID:     parseUUID(input.GenerationID),
		Model:            h.deps.Config.DefaultModel,
		Stage:            domain.TokenUsageStageTool,
		PromptTokens:     inputTokens,
		CompletionTokens: resultTokens,
		TotalTokens:      totalTokens,
		PromptChars:      inputChars,
		CompletionChars:  resultChars,
		WorkspaceRoots:   input.WorkspaceRoots,
		UserEmail:        "",
		HookEventName:    string(input.HookEventName),
		Timestamp:        time.Now(),
	}

	if err := h.deps.Client.CreateTokenUsage(ctx, usage); err != nil {
		h.deps.Log.Error("failed to send token usage", "error", err)
	} else {
		h.deps.Log.Info("token usage recorded", "stage", "tool", "tool", p.ToolName, "tokens", totalTokens)
	}

	return NewEmptyOutput(), nil
}

// AfterFileEditHandler handles afterFileEdit events.
type AfterFileEditHandler struct {
	deps *HandlerDependencies
}

// Handle processes the afterFileEdit event.
func (h *AfterFileEditHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	p, ok := payload.(*AfterFileEditPayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	// Count characters in edits
	var oldChars, newChars int64
	for _, edit := range p.Edits {
		oldChars += CountChars(edit.OldString)
		newChars += CountChars(edit.NewString)
	}

	h.deps.Log.Debug("file edit metrics",
		"file", p.FilePath,
		"old_chars", oldChars,
		"new_chars", newChars,
		"edit_count", len(p.Edits),
	)

	// Create token usage record for file edits
	usage := &domain.TokenUsage{
		ID:               uuid.New(),
		ConversationID:   parseUUID(input.ConversationID),
		GenerationID:     parseUUID(input.GenerationID),
		Model:            h.deps.Config.DefaultModel,
		Stage:            domain.TokenUsageStageFileEdit,
		PromptTokens:     0,
		CompletionTokens: 0,
		TotalTokens:      0,
		PromptChars:      oldChars,
		CompletionChars:  newChars,
		WorkspaceRoots:   input.WorkspaceRoots,
		UserEmail:        "",
		HookEventName:    string(input.HookEventName),
		Timestamp:        time.Now(),
	}

	if err := h.deps.Client.CreateTokenUsage(ctx, usage); err != nil {
		h.deps.Log.Error("failed to send token usage", "error", err)
	} else {
		h.deps.Log.Info("file edit recorded", "file", p.FilePath, "edits", len(p.Edits))
	}

	return NewEmptyOutput(), nil
}

// StopHandler handles stop events.
type StopHandler struct {
	deps *HandlerDependencies
}

// Handle processes the stop event.
func (h *StopHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	p, ok := payload.(*StopPayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	h.deps.Log.Info("agent loop stopped",
		"status", p.Status,
		"loop_count", p.LoopCount,
		"conversation_id", input.ConversationID,
	)

	// Create token usage record as session marker
	usage := &domain.TokenUsage{
		ID:               uuid.New(),
		ConversationID:   parseUUID(input.ConversationID),
		GenerationID:     parseUUID(input.GenerationID),
		Model:            h.deps.Config.DefaultModel,
		Stage:            domain.TokenUsageStageStop,
		PromptTokens:     0,
		CompletionTokens: 0,
		TotalTokens:      0,
		PromptChars:      0,
		CompletionChars:  0,
		WorkspaceRoots:   input.WorkspaceRoots,
		UserEmail:        "",
		HookEventName:    string(input.HookEventName),
		Timestamp:        time.Now(),
	}

	if err := h.deps.Client.CreateTokenUsage(ctx, usage); err != nil {
		h.deps.Log.Error("failed to send stop marker", "error", err)
	}

	return NewEmptyOutput(), nil
}

// BeforeShellExecutionHandler handles beforeShellExecution events.
type BeforeShellExecutionHandler struct {
	deps *HandlerDependencies
}

// Handle processes the beforeShellExecution event.
func (h *BeforeShellExecutionHandler) Handle(_ context.Context, input *HookInput, payload any) (*HookOutput, error) {
	p, ok := payload.(*BeforeShellExecutionPayload)
	if !ok {
		return NewAllowOutput(), nil
	}

	h.deps.Log.Debug("shell execution",
		"command", p.Command,
		"cwd", p.Cwd,
		"conversation_id", input.ConversationID,
	)

	// Always allow shell execution (observability only)
	return NewAllowOutput(), nil
}

// AfterShellExecutionHandler handles afterShellExecution events.
type AfterShellExecutionHandler struct {
	deps *HandlerDependencies
}

// Handle processes the afterShellExecution event.
func (h *AfterShellExecutionHandler) Handle(_ context.Context, input *HookInput, payload any) (*HookOutput, error) {
	p, ok := payload.(*AfterShellExecutionPayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	h.deps.Log.Debug("shell execution completed",
		"command", p.Command,
		"output_length", len(p.Output),
		"conversation_id", input.ConversationID,
	)

	return NewEmptyOutput(), nil
}

// BeforeMCPExecutionHandler handles beforeMCPExecution events.
type BeforeMCPExecutionHandler struct {
	deps *HandlerDependencies
}

// Handle processes the beforeMCPExecution event.
func (h *BeforeMCPExecutionHandler) Handle(_ context.Context, input *HookInput, payload any) (*HookOutput, error) {
	p, ok := payload.(*BeforeMCPExecutionPayload)
	if !ok {
		return NewAllowOutput(), nil
	}

	h.deps.Log.Debug("MCP execution",
		"tool", p.ToolName,
		"conversation_id", input.ConversationID,
	)

	// Always allow MCP execution (observability only)
	return NewAllowOutput(), nil
}

// BeforeReadFileHandler handles beforeReadFile events.
type BeforeReadFileHandler struct {
	deps *HandlerDependencies
}

// Handle processes the beforeReadFile event.
func (h *BeforeReadFileHandler) Handle(_ context.Context, input *HookInput, payload any) (*HookOutput, error) {
	p, ok := payload.(*BeforeReadFilePayload)
	if !ok {
		return NewAllowOutput(), nil
	}

	h.deps.Log.Debug("file read",
		"file", p.FilePath,
		"content_length", len(p.Content),
		"conversation_id", input.ConversationID,
	)

	// Always allow file read (observability only)
	return NewAllowOutput(), nil
}

// parseUUID parses a UUID string, returning a zero UUID on error.
func parseUUID(s string) uuid.UUID {
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil
	}
	return id
}


```

```go
package hook

import (
	"context"
	"fmt"
	"time"

	"github.com/ainslies/ctd/internal/domain"
	"github.com/ainslies/ctd/pkg/logger"
	"github.com/google/uuid"
)

// Handler processes a hook event and returns the appropriate output.
type Handler interface {
	// Handle processes the hook input and returns the output.
	Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error)
}

// HandlerDependencies contains the dependencies for handlers.
type HandlerDependencies struct {
	Client    *CTDClient
	Tokenizer Tokenizer
	Log       *logger.Logger
	Config    *Config
}

// Router routes hook events to the appropriate handler.
type Router struct {
	deps     *HandlerDependencies
	handlers map[HookEventName]Handler
}

// NewRouter creates a new router with all handlers registered.
func NewRouter(deps *HandlerDependencies) *Router {
	router := &Router{
		deps:     deps,
		handlers: make(map[HookEventName]Handler),
	}

	// Register handlers
	router.handlers[EventBeforeSubmitPrompt] = &BeforeSubmitPromptHandler{deps: deps}
	router.handlers[EventAfterAgentResponse] = &AfterAgentResponseHandler{deps: deps}
	router.handlers[EventAfterAgentThought] = &AfterAgentThoughtHandler{deps: deps}
	router.handlers[EventAfterMCPExecution] = &AfterMCPExecutionHandler{deps: deps}
	router.handlers[EventAfterFileEdit] = &AfterFileEditHandler{deps: deps}
	router.handlers[EventStop] = &StopHandler{deps: deps}
	router.handlers[EventBeforeShellExecution] = &BeforeShellExecutionHandler{deps: deps}
	router.handlers[EventAfterShellExecution] = &AfterShellExecutionHandler{deps: deps}
	router.handlers[EventBeforeMCPExecution] = &BeforeMCPExecutionHandler{deps: deps}
	router.handlers[EventBeforeReadFile] = &BeforeReadFileHandler{deps: deps}

	return router
}

// Route routes the input to the appropriate handler.
func (r *Router) Route(ctx context.Context, input *HookInput) (*HookOutput, error) {
	handler, ok := r.handlers[input.HookEventName]
	if !ok {
		// Unknown event, return empty output
		r.deps.Log.Warn("unknown hook event", "event", input.HookEventName)
		return NewEmptyOutput(), nil
	}

	payload, err := input.ParsePayload()
	if err != nil {
		r.deps.Log.Error("failed to parse payload", "error", err, "event", input.HookEventName)
		// Return appropriate default output even on parse error
		return getDefaultOutput(input.HookEventName), nil
	}

	output, err := handler.Handle(ctx, input, payload)
	if err != nil {
		return nil, fmt.Errorf("handler error for %s: %w", input.HookEventName, err)
	}
	return output, nil
}

// getDefaultOutput returns the default output for an event type.
//
//nolint:exhaustive // Only specific events need special handling, others use empty output
func getDefaultOutput(event HookEventName) *HookOutput {
	switch event {
	case EventBeforeSubmitPrompt:
		return NewContinueOutput()
	case EventBeforeShellExecution, EventBeforeMCPExecution, EventBeforeReadFile:
		return NewAllowOutput()
	default:
		return NewEmptyOutput()
	}
}

// BeforeSubmitPromptHandler handles beforeSubmitPrompt events.
type BeforeSubmitPromptHandler struct {
	deps *HandlerDependencies
}

// Handle processes the beforeSubmitPrompt event.
//
//nolint:varnamelen // promptPayload is clear in context
func (h *BeforeSubmitPromptHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	promptPayload, ok := payload.(*BeforeSubmitPromptPayload)
	if !ok {
		return NewContinueOutput(), nil
	}

	// Count tokens
	tokenCount, err := h.deps.Tokenizer.CountTokens(h.deps.Config.DefaultModel, promptPayload.Prompt)
	if err != nil {
		h.deps.Log.Error("failed to count tokens", "error", err)
		tokenCount = 0
	}

	charCount := CountChars(promptPayload.Prompt)

	h.deps.Log.Debug("prompt metrics",
		"tokens", tokenCount,
		"chars", charCount,
		"conversation_id", input.ConversationID,
	)

	// Create token usage record
	usage := &domain.TokenUsage{
		ID:               uuid.New(),
		ConversationID:   parseUUID(input.ConversationID),
		GenerationID:     parseUUID(input.GenerationID),
		Model:            h.deps.Config.DefaultModel,
		Stage:            domain.TokenUsageStagePrompt,
		PromptTokens:     tokenCount,
		CompletionTokens: 0,
		TotalTokens:      tokenCount,
		PromptChars:      charCount,
		CompletionChars:  0,
		WorkspaceRoots:   input.WorkspaceRoots,
		UserEmail:        "",
		HookEventName:    string(input.HookEventName),
		Timestamp:        time.Now(),
	}

	// Send to ctd (best-effort)
	if err := h.deps.Client.CreateTokenUsage(ctx, usage); err != nil {
		h.deps.Log.Error("failed to send token usage", "error", err)
	} else {
		h.deps.Log.Info("token usage recorded", "stage", "prompt", "tokens", tokenCount)
	}

	return NewContinueOutput(), nil
}

// AfterAgentResponseHandler handles afterAgentResponse events.
type AfterAgentResponseHandler struct {
	deps *HandlerDependencies
}

// Handle processes the afterAgentResponse event.
//
//nolint:varnamelen // responsePayload is clear in context
func (h *AfterAgentResponseHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	responsePayload, ok := payload.(*AfterAgentResponsePayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	// Count tokens
	tokenCount, err := h.deps.Tokenizer.CountTokens(h.deps.Config.DefaultModel, responsePayload.Text)
	if err != nil {
		h.deps.Log.Error("failed to count tokens", "error", err)
		tokenCount = 0
	}

	charCount := CountChars(responsePayload.Text)

	h.deps.Log.Debug("completion metrics",
		"tokens", tokenCount,
		"chars", charCount,
		"conversation_id", input.ConversationID,
	)

	// Create token usage record
	usage := &domain.TokenUsage{
		ID:               uuid.New(),
		ConversationID:   parseUUID(input.ConversationID),
		GenerationID:     parseUUID(input.GenerationID),
		Model:            h.deps.Config.DefaultModel,
		Stage:            domain.TokenUsageStageCompletion,
		PromptTokens:     0,
		CompletionTokens: tokenCount,
		TotalTokens:      tokenCount,
		PromptChars:      0,
		CompletionChars:  charCount,
		WorkspaceRoots:   input.WorkspaceRoots,
		UserEmail:        "",
		HookEventName:    string(input.HookEventName),
		Timestamp:        time.Now(),
	}

	// Send token usage to ctd (best-effort)
	if err := h.deps.Client.CreateTokenUsage(ctx, usage); err != nil {
		h.deps.Log.Error("failed to send token usage", "error", err)
	} else {
		h.deps.Log.Info("token usage recorded", "stage", "completion", "tokens", tokenCount)
	}

	// Also send as thought for visibility
	thought := &domain.Thought{
		ID:             uuid.New(),
		ConversationID: parseUUID(input.ConversationID),
		GenerationID:   parseUUID(input.GenerationID),
		Text:           responsePayload.Text,
		DurationMs:     0,
		Model:          h.deps.Config.DefaultModel,
		HookEventName:  string(input.HookEventName),
		CursorVersion:  "",
		WorkspaceRoots: input.WorkspaceRoots,
		UserEmail:      "",
		Timestamp:      time.Now(),
	}

	if err := h.deps.Client.CreateThought(ctx, thought); err != nil {
		h.deps.Log.Error("failed to send thought", "error", err)
	}

	return NewEmptyOutput(), nil
}

// AfterAgentThoughtHandler handles afterAgentThought events.
type AfterAgentThoughtHandler struct {
	deps *HandlerDependencies
}

// Handle processes the afterAgentThought event.
//
//nolint:varnamelen // thoughtPayload is clear in context
func (h *AfterAgentThoughtHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	thoughtPayload, ok := payload.(*AfterAgentThoughtPayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	// Send as thought for visibility
	thought := &domain.Thought{
		ID:             uuid.New(),
		ConversationID: parseUUID(input.ConversationID),
		GenerationID:   parseUUID(input.GenerationID),
		Text:           thoughtPayload.Text,
		DurationMs:     0,
		Model:          h.deps.Config.DefaultModel,
		HookEventName:  string(input.HookEventName),
		CursorVersion:  "",
		WorkspaceRoots: input.WorkspaceRoots,
		UserEmail:      "",
		Timestamp:      time.Now(),
	}

	if err := h.deps.Client.CreateThought(ctx, thought); err != nil {
		h.deps.Log.Error("failed to send thought", "error", err)
	} else {
		h.deps.Log.Info("thought recorded", "event", input.HookEventName)
	}

	return NewEmptyOutput(), nil
}

// AfterMCPExecutionHandler handles afterMCPExecution events.
type AfterMCPExecutionHandler struct {
	deps *HandlerDependencies
}

// Handle processes the afterMCPExecution event.
//
//nolint:varnamelen // mcpPayload is clear in context
func (h *AfterMCPExecutionHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	mcpPayload, ok := payload.(*AfterMCPExecutionPayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	// Count tokens in tool input and result (best effort, ignore errors)
	inputTokens, inputErr := h.deps.Tokenizer.CountTokens(h.deps.Config.DefaultModel, string(mcpPayload.ToolInput))
	if inputErr != nil {
		h.deps.Log.Debug("failed to count input tokens", "error", inputErr)
		inputTokens = 0
	}
	resultTokens, resultErr := h.deps.Tokenizer.CountTokens(h.deps.Config.DefaultModel, string(mcpPayload.ResultJSON))
	if resultErr != nil {
		h.deps.Log.Debug("failed to count result tokens", "error", resultErr)
		resultTokens = 0
	}
	totalTokens := inputTokens + resultTokens

	inputChars := CountChars(string(mcpPayload.ToolInput))
	resultChars := CountChars(string(mcpPayload.ResultJSON))

	h.deps.Log.Debug("tool metrics",
		"tool", mcpPayload.ToolName,
		"input_tokens", inputTokens,
		"result_tokens", resultTokens,
		"conversation_id", input.ConversationID,
	)

	// Create token usage record
	usage := &domain.TokenUsage{
		ID:               uuid.New(),
		ConversationID:   parseUUID(input.ConversationID),
		GenerationID:     parseUUID(input.GenerationID),
		Model:            h.deps.Config.DefaultModel,
		Stage:            domain.TokenUsageStageTool,
		PromptTokens:     inputTokens,
		CompletionTokens: resultTokens,
		TotalTokens:      totalTokens,
		PromptChars:      inputChars,
		CompletionChars:  resultChars,
		WorkspaceRoots:   input.WorkspaceRoots,
		UserEmail:        "",
		HookEventName:    string(input.HookEventName),
		Timestamp:        time.Now(),
	}

	if err := h.deps.Client.CreateTokenUsage(ctx, usage); err != nil {
		h.deps.Log.Error("failed to send token usage", "error", err)
	} else {
		h.deps.Log.Info("token usage recorded", "stage", "tool", "tool", p.ToolName, "tokens", totalTokens)
	}

	return NewEmptyOutput(), nil
}

// AfterFileEditHandler handles afterFileEdit events.
type AfterFileEditHandler struct {
	deps *HandlerDependencies
}

// Handle processes the afterFileEdit event.
func (h *AfterFileEditHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	p, ok := payload.(*AfterFileEditPayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	// Count characters in edits
	var oldChars, newChars int64
	for _, edit := range p.Edits {
		oldChars += CountChars(edit.OldString)
		newChars += CountChars(edit.NewString)
	}

	h.deps.Log.Debug("file edit metrics",
		"file", p.FilePath,
		"old_chars", oldChars,
		"new_chars", newChars,
		"edit_count", len(p.Edits),
	)

	// Create token usage record for file edits
	usage := &domain.TokenUsage{
		ID:               uuid.New(),
		ConversationID:   parseUUID(input.ConversationID),
		GenerationID:     parseUUID(input.GenerationID),
		Model:            h.deps.Config.DefaultModel,
		Stage:            domain.TokenUsageStageFileEdit,
		PromptTokens:     0,
		CompletionTokens: 0,
		TotalTokens:      0,
		PromptChars:      oldChars,
		CompletionChars:  newChars,
		WorkspaceRoots:   input.WorkspaceRoots,
		UserEmail:        "",
		HookEventName:    string(input.HookEventName),
		Timestamp:        time.Now(),
	}

	if err := h.deps.Client.CreateTokenUsage(ctx, usage); err != nil {
		h.deps.Log.Error("failed to send token usage", "error", err)
	} else {
		h.deps.Log.Info("file edit recorded", "file", p.FilePath, "edits", len(p.Edits))
	}

	return NewEmptyOutput(), nil
}

// StopHandler handles stop events.
type StopHandler struct {
	deps *HandlerDependencies
}

// Handle processes the stop event.
func (h *StopHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	p, ok := payload.(*StopPayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	h.deps.Log.Info("agent loop stopped",
		"status", p.Status,
		"loop_count", p.LoopCount,
		"conversation_id", input.ConversationID,
	)

	// Create token usage record as session marker
	usage := &domain.TokenUsage{
		ID:               uuid.New(),
		ConversationID:   parseUUID(input.ConversationID),
		GenerationID:     parseUUID(input.GenerationID),
		Model:            h.deps.Config.DefaultModel,
		Stage:            domain.TokenUsageStageStop,
		PromptTokens:     0,
		CompletionTokens: 0,
		TotalTokens:      0,
		PromptChars:      0,
		CompletionChars:  0,
		WorkspaceRoots:   input.WorkspaceRoots,
		UserEmail:        "",
		HookEventName:    string(input.HookEventName),
		Timestamp:        time.Now(),
	}

	if err := h.deps.Client.CreateTokenUsage(ctx, usage); err != nil {
		h.deps.Log.Error("failed to send stop marker", "error", err)
	}

	return NewEmptyOutput(), nil
}

// BeforeShellExecutionHandler handles beforeShellExecution events.
type BeforeShellExecutionHandler struct {
	deps *HandlerDependencies
}

// Handle processes the beforeShellExecution event.
func (h *BeforeShellExecutionHandler) Handle(_ context.Context, input *HookInput, payload any) (*HookOutput, error) {
	p, ok := payload.(*BeforeShellExecutionPayload)
	if !ok {
		return NewAllowOutput(), nil
	}

	h.deps.Log.Debug("shell execution",
		"command", p.Command,
		"cwd", p.Cwd,
		"conversation_id", input.ConversationID,
	)

	// Always allow shell execution (observability only)
	return NewAllowOutput(), nil
}

// AfterShellExecutionHandler handles afterShellExecution events.
type AfterShellExecutionHandler struct {
	deps *HandlerDependencies
}

// Handle processes the afterShellExecution event.
func (h *AfterShellExecutionHandler) Handle(_ context.Context, input *HookInput, payload any) (*HookOutput, error) {
	p, ok := payload.(*AfterShellExecutionPayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	h.deps.Log.Debug("shell execution completed",
		"command", p.Command,
		"output_length", len(p.Output),
		"conversation_id", input.ConversationID,
	)

	return NewEmptyOutput(), nil
}

// BeforeMCPExecutionHandler handles beforeMCPExecution events.
type BeforeMCPExecutionHandler struct {
	deps *HandlerDependencies
}

// Handle processes the beforeMCPExecution event.
func (h *BeforeMCPExecutionHandler) Handle(_ context.Context, input *HookInput, payload any) (*HookOutput, error) {
	p, ok := payload.(*BeforeMCPExecutionPayload)
	if !ok {
		return NewAllowOutput(), nil
	}

	h.deps.Log.Debug("MCP execution",
		"tool", p.ToolName,
		"conversation_id", input.ConversationID,
	)

	// Always allow MCP execution (observability only)
	return NewAllowOutput(), nil
}

// BeforeReadFileHandler handles beforeReadFile events.
type BeforeReadFileHandler struct {
	deps *HandlerDependencies
}

// Handle processes the beforeReadFile event.
func (h *BeforeReadFileHandler) Handle(_ context.Context, input *HookInput, payload any) (*HookOutput, error) {
	p, ok := payload.(*BeforeReadFilePayload)
	if !ok {
		return NewAllowOutput(), nil
	}

	h.deps.Log.Debug("file read",
		"file", p.FilePath,
		"content_length", len(p.Content),
		"conversation_id", input.ConversationID,
	)

	// Always allow file read (observability only)
	return NewAllowOutput(), nil
}

// parseUUID parses a UUID string, returning a zero UUID on error.
func parseUUID(s string) uuid.UUID {
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil
	}
	return id
}


```

```go
package hook

import (
	"context"
	"fmt"
	"time"

	"github.com/ainslies/ctd/internal/domain"
	"github.com/ainslies/ctd/pkg/logger"
	"github.com/google/uuid"
)

// Handler processes a hook event and returns the appropriate output.
type Handler interface {
	// Handle processes the hook input and returns the output.
	Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error)
}

// HandlerDependencies contains the dependencies for handlers.
type HandlerDependencies struct {
	Client    *CTDClient
	Tokenizer Tokenizer
	Log       *logger.Logger
	Config    *Config
}

// Router routes hook events to the appropriate handler.
type Router struct {
	deps     *HandlerDependencies
	handlers map[HookEventName]Handler
}

// NewRouter creates a new router with all handlers registered.
func NewRouter(deps *HandlerDependencies) *Router {
	router := &Router{
		deps:     deps,
		handlers: make(map[HookEventName]Handler),
	}

	// Register handlers
	router.handlers[EventBeforeSubmitPrompt] = &BeforeSubmitPromptHandler{deps: deps}
	router.handlers[EventAfterAgentResponse] = &AfterAgentResponseHandler{deps: deps}
	router.handlers[EventAfterAgentThought] = &AfterAgentThoughtHandler{deps: deps}
	router.handlers[EventAfterMCPExecution] = &AfterMCPExecutionHandler{deps: deps}
	router.handlers[EventAfterFileEdit] = &AfterFileEditHandler{deps: deps}
	router.handlers[EventStop] = &StopHandler{deps: deps}
	router.handlers[EventBeforeShellExecution] = &BeforeShellExecutionHandler{deps: deps}
	router.handlers[EventAfterShellExecution] = &AfterShellExecutionHandler{deps: deps}
	router.handlers[EventBeforeMCPExecution] = &BeforeMCPExecutionHandler{deps: deps}
	router.handlers[EventBeforeReadFile] = &BeforeReadFileHandler{deps: deps}

	return router
}

// Route routes the input to the appropriate handler.
func (r *Router) Route(ctx context.Context, input *HookInput) (*HookOutput, error) {
	handler, ok := r.handlers[input.HookEventName]
	if !ok {
		// Unknown event, return empty output
		r.deps.Log.Warn("unknown hook event", "event", input.HookEventName)
		return NewEmptyOutput(), nil
	}

	payload, err := input.ParsePayload()
	if err != nil {
		r.deps.Log.Error("failed to parse payload", "error", err, "event", input.HookEventName)
		// Return appropriate default output even on parse error
		return getDefaultOutput(input.HookEventName), nil
	}

	output, err := handler.Handle(ctx, input, payload)
	if err != nil {
		return nil, fmt.Errorf("handler error for %s: %w", input.HookEventName, err)
	}
	return output, nil
}

// getDefaultOutput returns the default output for an event type.
//
//nolint:exhaustive // Only specific events need special handling, others use empty output
func getDefaultOutput(event HookEventName) *HookOutput {
	switch event {
	case EventBeforeSubmitPrompt:
		return NewContinueOutput()
	case EventBeforeShellExecution, EventBeforeMCPExecution, EventBeforeReadFile:
		return NewAllowOutput()
	default:
		return NewEmptyOutput()
	}
}

// BeforeSubmitPromptHandler handles beforeSubmitPrompt events.
type BeforeSubmitPromptHandler struct {
	deps *HandlerDependencies
}

// Handle processes the beforeSubmitPrompt event.
//
//nolint:varnamelen // promptPayload is clear in context
func (h *BeforeSubmitPromptHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	promptPayload, ok := payload.(*BeforeSubmitPromptPayload)
	if !ok {
		return NewContinueOutput(), nil
	}

	// Count tokens
	tokenCount, err := h.deps.Tokenizer.CountTokens(h.deps.Config.DefaultModel, promptPayload.Prompt)
	if err != nil {
		h.deps.Log.Error("failed to count tokens", "error", err)
		tokenCount = 0
	}

	charCount := CountChars(promptPayload.Prompt)

	h.deps.Log.Debug("prompt metrics",
		"tokens", tokenCount,
		"chars", charCount,
		"conversation_id", input.ConversationID,
	)

	// Create token usage record
	usage := &domain.TokenUsage{
		ID:               uuid.New(),
		ConversationID:   parseUUID(input.ConversationID),
		GenerationID:     parseUUID(input.GenerationID),
		Model:            h.deps.Config.DefaultModel,
		Stage:            domain.TokenUsageStagePrompt,
		PromptTokens:     tokenCount,
		CompletionTokens: 0,
		TotalTokens:      tokenCount,
		PromptChars:      charCount,
		CompletionChars:  0,
		WorkspaceRoots:   input.WorkspaceRoots,
		UserEmail:        "",
		HookEventName:    string(input.HookEventName),
		Timestamp:        time.Now(),
	}

	// Send to ctd (best-effort)
	if err := h.deps.Client.CreateTokenUsage(ctx, usage); err != nil {
		h.deps.Log.Error("failed to send token usage", "error", err)
	} else {
		h.deps.Log.Info("token usage recorded", "stage", "prompt", "tokens", tokenCount)
	}

	return NewContinueOutput(), nil
}

// AfterAgentResponseHandler handles afterAgentResponse events.
type AfterAgentResponseHandler struct {
	deps *HandlerDependencies
}

// Handle processes the afterAgentResponse event.
//
//nolint:varnamelen // responsePayload is clear in context
func (h *AfterAgentResponseHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	responsePayload, ok := payload.(*AfterAgentResponsePayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	// Count tokens
	tokenCount, err := h.deps.Tokenizer.CountTokens(h.deps.Config.DefaultModel, responsePayload.Text)
	if err != nil {
		h.deps.Log.Error("failed to count tokens", "error", err)
		tokenCount = 0
	}

	charCount := CountChars(responsePayload.Text)

	h.deps.Log.Debug("completion metrics",
		"tokens", tokenCount,
		"chars", charCount,
		"conversation_id", input.ConversationID,
	)

	// Create token usage record
	usage := &domain.TokenUsage{
		ID:               uuid.New(),
		ConversationID:   parseUUID(input.ConversationID),
		GenerationID:     parseUUID(input.GenerationID),
		Model:            h.deps.Config.DefaultModel,
		Stage:            domain.TokenUsageStageCompletion,
		PromptTokens:     0,
		CompletionTokens: tokenCount,
		TotalTokens:      tokenCount,
		PromptChars:      0,
		CompletionChars:  charCount,
		WorkspaceRoots:   input.WorkspaceRoots,
		UserEmail:        "",
		HookEventName:    string(input.HookEventName),
		Timestamp:        time.Now(),
	}

	// Send token usage to ctd (best-effort)
	if err := h.deps.Client.CreateTokenUsage(ctx, usage); err != nil {
		h.deps.Log.Error("failed to send token usage", "error", err)
	} else {
		h.deps.Log.Info("token usage recorded", "stage", "completion", "tokens", tokenCount)
	}

	// Also send as thought for visibility
	thought := &domain.Thought{
		ID:             uuid.New(),
		ConversationID: parseUUID(input.ConversationID),
		GenerationID:   parseUUID(input.GenerationID),
		Text:           responsePayload.Text,
		DurationMs:     0,
		Model:          h.deps.Config.DefaultModel,
		HookEventName:  string(input.HookEventName),
		CursorVersion:  "",
		WorkspaceRoots: input.WorkspaceRoots,
		UserEmail:      "",
		Timestamp:      time.Now(),
	}

	if err := h.deps.Client.CreateThought(ctx, thought); err != nil {
		h.deps.Log.Error("failed to send thought", "error", err)
	}

	return NewEmptyOutput(), nil
}

// AfterAgentThoughtHandler handles afterAgentThought events.
type AfterAgentThoughtHandler struct {
	deps *HandlerDependencies
}

// Handle processes the afterAgentThought event.
//
//nolint:varnamelen // thoughtPayload is clear in context
func (h *AfterAgentThoughtHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	thoughtPayload, ok := payload.(*AfterAgentThoughtPayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	// Send as thought for visibility
	thought := &domain.Thought{
		ID:             uuid.New(),
		ConversationID: parseUUID(input.ConversationID),
		GenerationID:   parseUUID(input.GenerationID),
		Text:           thoughtPayload.Text,
		DurationMs:     0,
		Model:          h.deps.Config.DefaultModel,
		HookEventName:  string(input.HookEventName),
		CursorVersion:  "",
		WorkspaceRoots: input.WorkspaceRoots,
		UserEmail:      "",
		Timestamp:      time.Now(),
	}

	if err := h.deps.Client.CreateThought(ctx, thought); err != nil {
		h.deps.Log.Error("failed to send thought", "error", err)
	} else {
		h.deps.Log.Info("thought recorded", "event", input.HookEventName)
	}

	return NewEmptyOutput(), nil
}

// AfterMCPExecutionHandler handles afterMCPExecution events.
type AfterMCPExecutionHandler struct {
	deps *HandlerDependencies
}

// Handle processes the afterMCPExecution event.
//
//nolint:varnamelen // mcpPayload is clear in context
func (h *AfterMCPExecutionHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	mcpPayload, ok := payload.(*AfterMCPExecutionPayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	// Count tokens in tool input and result (best effort, ignore errors)
	inputTokens, inputErr := h.deps.Tokenizer.CountTokens(h.deps.Config.DefaultModel, string(mcpPayload.ToolInput))
	if inputErr != nil {
		h.deps.Log.Debug("failed to count input tokens", "error", inputErr)
		inputTokens = 0
	}
	resultTokens, resultErr := h.deps.Tokenizer.CountTokens(h.deps.Config.DefaultModel, string(mcpPayload.ResultJSON))
	if resultErr != nil {
		h.deps.Log.Debug("failed to count result tokens", "error", resultErr)
		resultTokens = 0
	}
	totalTokens := inputTokens + resultTokens

	inputChars := CountChars(string(mcpPayload.ToolInput))
	resultChars := CountChars(string(mcpPayload.ResultJSON))

	h.deps.Log.Debug("tool metrics",
		"tool", mcpPayload.ToolName,
		"input_tokens", inputTokens,
		"result_tokens", resultTokens,
		"conversation_id", input.ConversationID,
	)

	// Create token usage record
	usage := &domain.TokenUsage{
		ID:               uuid.New(),
		ConversationID:   parseUUID(input.ConversationID),
		GenerationID:     parseUUID(input.GenerationID),
		Model:            h.deps.Config.DefaultModel,
		Stage:            domain.TokenUsageStageTool,
		PromptTokens:     inputTokens,
		CompletionTokens: resultTokens,
		TotalTokens:      totalTokens,
		PromptChars:      inputChars,
		CompletionChars:  resultChars,
		WorkspaceRoots:   input.WorkspaceRoots,
		UserEmail:        "",
		HookEventName:    string(input.HookEventName),
		Timestamp:        time.Now(),
	}

	if err := h.deps.Client.CreateTokenUsage(ctx, usage); err != nil {
		h.deps.Log.Error("failed to send token usage", "error", err)
	} else {
		h.deps.Log.Info("token usage recorded", "stage", "tool", "tool", mcpPayload.ToolName, "tokens", totalTokens)
	}

	return NewEmptyOutput(), nil
}

// AfterFileEditHandler handles afterFileEdit events.
type AfterFileEditHandler struct {
	deps *HandlerDependencies
}

// Handle processes the afterFileEdit event.
//
//nolint:varnamelen // editPayload is clear in context
func (h *AfterFileEditHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	editPayload, ok := payload.(*AfterFileEditPayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	// Count characters in edits
	var oldChars, newChars int64
	for _, edit := range editPayload.Edits {
		oldChars += CountChars(edit.OldString)
		newChars += CountChars(edit.NewString)
	}

	h.deps.Log.Debug("file edit metrics",
		"file", editPayload.FilePath,
		"old_chars", oldChars,
		"new_chars", newChars,
		"edit_count", len(p.Edits),
	)

	// Create token usage record for file edits
	usage := &domain.TokenUsage{
		ID:               uuid.New(),
		ConversationID:   parseUUID(input.ConversationID),
		GenerationID:     parseUUID(input.GenerationID),
		Model:            h.deps.Config.DefaultModel,
		Stage:            domain.TokenUsageStageFileEdit,
		PromptTokens:     0,
		CompletionTokens: 0,
		TotalTokens:      0,
		PromptChars:      oldChars,
		CompletionChars:  newChars,
		WorkspaceRoots:   input.WorkspaceRoots,
		UserEmail:        "",
		HookEventName:    string(input.HookEventName),
		Timestamp:        time.Now(),
	}

	if err := h.deps.Client.CreateTokenUsage(ctx, usage); err != nil {
		h.deps.Log.Error("failed to send token usage", "error", err)
	} else {
		h.deps.Log.Info("file edit recorded", "file", p.FilePath, "edits", len(p.Edits))
	}

	return NewEmptyOutput(), nil
}

// StopHandler handles stop events.
type StopHandler struct {
	deps *HandlerDependencies
}

// Handle processes the stop event.
func (h *StopHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	p, ok := payload.(*StopPayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	h.deps.Log.Info("agent loop stopped",
		"status", p.Status,
		"loop_count", p.LoopCount,
		"conversation_id", input.ConversationID,
	)

	// Create token usage record as session marker
	usage := &domain.TokenUsage{
		ID:               uuid.New(),
		ConversationID:   parseUUID(input.ConversationID),
		GenerationID:     parseUUID(input.GenerationID),
		Model:            h.deps.Config.DefaultModel,
		Stage:            domain.TokenUsageStageStop,
		PromptTokens:     0,
		CompletionTokens: 0,
		TotalTokens:      0,
		PromptChars:      0,
		CompletionChars:  0,
		WorkspaceRoots:   input.WorkspaceRoots,
		UserEmail:        "",
		HookEventName:    string(input.HookEventName),
		Timestamp:        time.Now(),
	}

	if err := h.deps.Client.CreateTokenUsage(ctx, usage); err != nil {
		h.deps.Log.Error("failed to send stop marker", "error", err)
	}

	return NewEmptyOutput(), nil
}

// BeforeShellExecutionHandler handles beforeShellExecution events.
type BeforeShellExecutionHandler struct {
	deps *HandlerDependencies
}

// Handle processes the beforeShellExecution event.
func (h *BeforeShellExecutionHandler) Handle(_ context.Context, input *HookInput, payload any) (*HookOutput, error) {
	p, ok := payload.(*BeforeShellExecutionPayload)
	if !ok {
		return NewAllowOutput(), nil
	}

	h.deps.Log.Debug("shell execution",
		"command", p.Command,
		"cwd", p.Cwd,
		"conversation_id", input.ConversationID,
	)

	// Always allow shell execution (observability only)
	return NewAllowOutput(), nil
}

// AfterShellExecutionHandler handles afterShellExecution events.
type AfterShellExecutionHandler struct {
	deps *HandlerDependencies
}

// Handle processes the afterShellExecution event.
func (h *AfterShellExecutionHandler) Handle(_ context.Context, input *HookInput, payload any) (*HookOutput, error) {
	p, ok := payload.(*AfterShellExecutionPayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	h.deps.Log.Debug("shell execution completed",
		"command", p.Command,
		"output_length", len(p.Output),
		"conversation_id", input.ConversationID,
	)

	return NewEmptyOutput(), nil
}

// BeforeMCPExecutionHandler handles beforeMCPExecution events.
type BeforeMCPExecutionHandler struct {
	deps *HandlerDependencies
}

// Handle processes the beforeMCPExecution event.
func (h *BeforeMCPExecutionHandler) Handle(_ context.Context, input *HookInput, payload any) (*HookOutput, error) {
	p, ok := payload.(*BeforeMCPExecutionPayload)
	if !ok {
		return NewAllowOutput(), nil
	}

	h.deps.Log.Debug("MCP execution",
		"tool", p.ToolName,
		"conversation_id", input.ConversationID,
	)

	// Always allow MCP execution (observability only)
	return NewAllowOutput(), nil
}

// BeforeReadFileHandler handles beforeReadFile events.
type BeforeReadFileHandler struct {
	deps *HandlerDependencies
}

// Handle processes the beforeReadFile event.
func (h *BeforeReadFileHandler) Handle(_ context.Context, input *HookInput, payload any) (*HookOutput, error) {
	p, ok := payload.(*BeforeReadFilePayload)
	if !ok {
		return NewAllowOutput(), nil
	}

	h.deps.Log.Debug("file read",
		"file", p.FilePath,
		"content_length", len(p.Content),
		"conversation_id", input.ConversationID,
	)

	// Always allow file read (observability only)
	return NewAllowOutput(), nil
}

// parseUUID parses a UUID string, returning a zero UUID on error.
func parseUUID(s string) uuid.UUID {
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil
	}
	return id
}


```

```go
package hook

import (
	"context"
	"fmt"
	"time"

	"github.com/ainslies/ctd/internal/domain"
	"github.com/ainslies/ctd/pkg/logger"
	"github.com/google/uuid"
)

// Handler processes a hook event and returns the appropriate output.
type Handler interface {
	// Handle processes the hook input and returns the output.
	Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error)
}

// HandlerDependencies contains the dependencies for handlers.
type HandlerDependencies struct {
	Client    *CTDClient
	Tokenizer Tokenizer
	Log       *logger.Logger
	Config    *Config
}

// Router routes hook events to the appropriate handler.
type Router struct {
	deps     *HandlerDependencies
	handlers map[HookEventName]Handler
}

// NewRouter creates a new router with all handlers registered.
func NewRouter(deps *HandlerDependencies) *Router {
	router := &Router{
		deps:     deps,
		handlers: make(map[HookEventName]Handler),
	}

	// Register handlers
	router.handlers[EventBeforeSubmitPrompt] = &BeforeSubmitPromptHandler{deps: deps}
	router.handlers[EventAfterAgentResponse] = &AfterAgentResponseHandler{deps: deps}
	router.handlers[EventAfterAgentThought] = &AfterAgentThoughtHandler{deps: deps}
	router.handlers[EventAfterMCPExecution] = &AfterMCPExecutionHandler{deps: deps}
	router.handlers[EventAfterFileEdit] = &AfterFileEditHandler{deps: deps}
	router.handlers[EventStop] = &StopHandler{deps: deps}
	router.handlers[EventBeforeShellExecution] = &BeforeShellExecutionHandler{deps: deps}
	router.handlers[EventAfterShellExecution] = &AfterShellExecutionHandler{deps: deps}
	router.handlers[EventBeforeMCPExecution] = &BeforeMCPExecutionHandler{deps: deps}
	router.handlers[EventBeforeReadFile] = &BeforeReadFileHandler{deps: deps}

	return router
}

// Route routes the input to the appropriate handler.
func (r *Router) Route(ctx context.Context, input *HookInput) (*HookOutput, error) {
	handler, ok := r.handlers[input.HookEventName]
	if !ok {
		// Unknown event, return empty output
		r.deps.Log.Warn("unknown hook event", "event", input.HookEventName)
		return NewEmptyOutput(), nil
	}

	payload, err := input.ParsePayload()
	if err != nil {
		r.deps.Log.Error("failed to parse payload", "error", err, "event", input.HookEventName)
		// Return appropriate default output even on parse error
		return getDefaultOutput(input.HookEventName), nil
	}

	output, err := handler.Handle(ctx, input, payload)
	if err != nil {
		return nil, fmt.Errorf("handler error for %s: %w", input.HookEventName, err)
	}
	return output, nil
}

// getDefaultOutput returns the default output for an event type.
//
//nolint:exhaustive // Only specific events need special handling, others use empty output
func getDefaultOutput(event HookEventName) *HookOutput {
	switch event {
	case EventBeforeSubmitPrompt:
		return NewContinueOutput()
	case EventBeforeShellExecution, EventBeforeMCPExecution, EventBeforeReadFile:
		return NewAllowOutput()
	default:
		return NewEmptyOutput()
	}
}

// BeforeSubmitPromptHandler handles beforeSubmitPrompt events.
type BeforeSubmitPromptHandler struct {
	deps *HandlerDependencies
}

// Handle processes the beforeSubmitPrompt event.
//
//nolint:varnamelen // promptPayload is clear in context
func (h *BeforeSubmitPromptHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	promptPayload, ok := payload.(*BeforeSubmitPromptPayload)
	if !ok {
		return NewContinueOutput(), nil
	}

	// Count tokens
	tokenCount, err := h.deps.Tokenizer.CountTokens(h.deps.Config.DefaultModel, promptPayload.Prompt)
	if err != nil {
		h.deps.Log.Error("failed to count tokens", "error", err)
		tokenCount = 0
	}

	charCount := CountChars(promptPayload.Prompt)

	h.deps.Log.Debug("prompt metrics",
		"tokens", tokenCount,
		"chars", charCount,
		"conversation_id", input.ConversationID,
	)

	// Create token usage record
	usage := &domain.TokenUsage{
		ID:               uuid.New(),
		ConversationID:   parseUUID(input.ConversationID),
		GenerationID:     parseUUID(input.GenerationID),
		Model:            h.deps.Config.DefaultModel,
		Stage:            domain.TokenUsageStagePrompt,
		PromptTokens:     tokenCount,
		CompletionTokens: 0,
		TotalTokens:      tokenCount,
		PromptChars:      charCount,
		CompletionChars:  0,
		WorkspaceRoots:   input.WorkspaceRoots,
		UserEmail:        "",
		HookEventName:    string(input.HookEventName),
		Timestamp:        time.Now(),
	}

	// Send to ctd (best-effort)
	if err := h.deps.Client.CreateTokenUsage(ctx, usage); err != nil {
		h.deps.Log.Error("failed to send token usage", "error", err)
	} else {
		h.deps.Log.Info("token usage recorded", "stage", "prompt", "tokens", tokenCount)
	}

	return NewContinueOutput(), nil
}

// AfterAgentResponseHandler handles afterAgentResponse events.
type AfterAgentResponseHandler struct {
	deps *HandlerDependencies
}

// Handle processes the afterAgentResponse event.
//
//nolint:varnamelen // responsePayload is clear in context
func (h *AfterAgentResponseHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	responsePayload, ok := payload.(*AfterAgentResponsePayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	// Count tokens
	tokenCount, err := h.deps.Tokenizer.CountTokens(h.deps.Config.DefaultModel, responsePayload.Text)
	if err != nil {
		h.deps.Log.Error("failed to count tokens", "error", err)
		tokenCount = 0
	}

	charCount := CountChars(responsePayload.Text)

	h.deps.Log.Debug("completion metrics",
		"tokens", tokenCount,
		"chars", charCount,
		"conversation_id", input.ConversationID,
	)

	// Create token usage record
	usage := &domain.TokenUsage{
		ID:               uuid.New(),
		ConversationID:   parseUUID(input.ConversationID),
		GenerationID:     parseUUID(input.GenerationID),
		Model:            h.deps.Config.DefaultModel,
		Stage:            domain.TokenUsageStageCompletion,
		PromptTokens:     0,
		CompletionTokens: tokenCount,
		TotalTokens:      tokenCount,
		PromptChars:      0,
		CompletionChars:  charCount,
		WorkspaceRoots:   input.WorkspaceRoots,
		UserEmail:        "",
		HookEventName:    string(input.HookEventName),
		Timestamp:        time.Now(),
	}

	// Send token usage to ctd (best-effort)
	if err := h.deps.Client.CreateTokenUsage(ctx, usage); err != nil {
		h.deps.Log.Error("failed to send token usage", "error", err)
	} else {
		h.deps.Log.Info("token usage recorded", "stage", "completion", "tokens", tokenCount)
	}

	// Also send as thought for visibility
	thought := &domain.Thought{
		ID:             uuid.New(),
		ConversationID: parseUUID(input.ConversationID),
		GenerationID:   parseUUID(input.GenerationID),
		Text:           responsePayload.Text,
		DurationMs:     0,
		Model:          h.deps.Config.DefaultModel,
		HookEventName:  string(input.HookEventName),
		CursorVersion:  "",
		WorkspaceRoots: input.WorkspaceRoots,
		UserEmail:      "",
		Timestamp:      time.Now(),
	}

	if err := h.deps.Client.CreateThought(ctx, thought); err != nil {
		h.deps.Log.Error("failed to send thought", "error", err)
	}

	return NewEmptyOutput(), nil
}

// AfterAgentThoughtHandler handles afterAgentThought events.
type AfterAgentThoughtHandler struct {
	deps *HandlerDependencies
}

// Handle processes the afterAgentThought event.
//
//nolint:varnamelen // thoughtPayload is clear in context
func (h *AfterAgentThoughtHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	thoughtPayload, ok := payload.(*AfterAgentThoughtPayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	// Send as thought for visibility
	thought := &domain.Thought{
		ID:             uuid.New(),
		ConversationID: parseUUID(input.ConversationID),
		GenerationID:   parseUUID(input.GenerationID),
		Text:           thoughtPayload.Text,
		DurationMs:     0,
		Model:          h.deps.Config.DefaultModel,
		HookEventName:  string(input.HookEventName),
		CursorVersion:  "",
		WorkspaceRoots: input.WorkspaceRoots,
		UserEmail:      "",
		Timestamp:      time.Now(),
	}

	if err := h.deps.Client.CreateThought(ctx, thought); err != nil {
		h.deps.Log.Error("failed to send thought", "error", err)
	} else {
		h.deps.Log.Info("thought recorded", "event", input.HookEventName)
	}

	return NewEmptyOutput(), nil
}

// AfterMCPExecutionHandler handles afterMCPExecution events.
type AfterMCPExecutionHandler struct {
	deps *HandlerDependencies
}

// Handle processes the afterMCPExecution event.
//
//nolint:varnamelen // mcpPayload is clear in context
func (h *AfterMCPExecutionHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	mcpPayload, ok := payload.(*AfterMCPExecutionPayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	// Count tokens in tool input and result (best effort, ignore errors)
	inputTokens, inputErr := h.deps.Tokenizer.CountTokens(h.deps.Config.DefaultModel, string(mcpPayload.ToolInput))
	if inputErr != nil {
		h.deps.Log.Debug("failed to count input tokens", "error", inputErr)
		inputTokens = 0
	}
	resultTokens, resultErr := h.deps.Tokenizer.CountTokens(h.deps.Config.DefaultModel, string(mcpPayload.ResultJSON))
	if resultErr != nil {
		h.deps.Log.Debug("failed to count result tokens", "error", resultErr)
		resultTokens = 0
	}
	totalTokens := inputTokens + resultTokens

	inputChars := CountChars(string(mcpPayload.ToolInput))
	resultChars := CountChars(string(mcpPayload.ResultJSON))

	h.deps.Log.Debug("tool metrics",
		"tool", mcpPayload.ToolName,
		"input_tokens", inputTokens,
		"result_tokens", resultTokens,
		"conversation_id", input.ConversationID,
	)

	// Create token usage record
	usage := &domain.TokenUsage{
		ID:               uuid.New(),
		ConversationID:   parseUUID(input.ConversationID),
		GenerationID:     parseUUID(input.GenerationID),
		Model:            h.deps.Config.DefaultModel,
		Stage:            domain.TokenUsageStageTool,
		PromptTokens:     inputTokens,
		CompletionTokens: resultTokens,
		TotalTokens:      totalTokens,
		PromptChars:      inputChars,
		CompletionChars:  resultChars,
		WorkspaceRoots:   input.WorkspaceRoots,
		UserEmail:        "",
		HookEventName:    string(input.HookEventName),
		Timestamp:        time.Now(),
	}

	if err := h.deps.Client.CreateTokenUsage(ctx, usage); err != nil {
		h.deps.Log.Error("failed to send token usage", "error", err)
	} else {
		h.deps.Log.Info("token usage recorded", "stage", "tool", "tool", mcpPayload.ToolName, "tokens", totalTokens)
	}

	return NewEmptyOutput(), nil
}

// AfterFileEditHandler handles afterFileEdit events.
type AfterFileEditHandler struct {
	deps *HandlerDependencies
}

// Handle processes the afterFileEdit event.
//
//nolint:varnamelen // editPayload is clear in context
func (h *AfterFileEditHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	editPayload, ok := payload.(*AfterFileEditPayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	// Count characters in edits
	var oldChars, newChars int64
	for _, edit := range editPayload.Edits {
		oldChars += CountChars(edit.OldString)
		newChars += CountChars(edit.NewString)
	}

	h.deps.Log.Debug("file edit metrics",
		"file", editPayload.FilePath,
		"old_chars", oldChars,
		"new_chars", newChars,
		"edit_count", len(editPayload.Edits),
	)

	// Create token usage record for file edits
	usage := &domain.TokenUsage{
		ID:               uuid.New(),
		ConversationID:   parseUUID(input.ConversationID),
		GenerationID:     parseUUID(input.GenerationID),
		Model:            h.deps.Config.DefaultModel,
		Stage:            domain.TokenUsageStageFileEdit,
		PromptTokens:     0,
		CompletionTokens: 0,
		TotalTokens:      0,
		PromptChars:      oldChars,
		CompletionChars:  newChars,
		WorkspaceRoots:   input.WorkspaceRoots,
		UserEmail:        "",
		HookEventName:    string(input.HookEventName),
		Timestamp:        time.Now(),
	}

	if err := h.deps.Client.CreateTokenUsage(ctx, usage); err != nil {
		h.deps.Log.Error("failed to send token usage", "error", err)
	} else {
		h.deps.Log.Info("file edit recorded", "file", editPayload.FilePath, "edits", len(editPayload.Edits))
	}

	return NewEmptyOutput(), nil
}

// StopHandler handles stop events.
type StopHandler struct {
	deps *HandlerDependencies
}

// Handle processes the stop event.
//
//nolint:varnamelen // stopPayload is clear in context
func (h *StopHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	stopPayload, ok := payload.(*StopPayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	h.deps.Log.Info("agent loop stopped",
		"status", stopPayload.Status,
		"loop_count", stopPayload.LoopCount,
		"conversation_id", input.ConversationID,
	)

	// Create token usage record as session marker
	usage := &domain.TokenUsage{
		ID:               uuid.New(),
		ConversationID:   parseUUID(input.ConversationID),
		GenerationID:     parseUUID(input.GenerationID),
		Model:            h.deps.Config.DefaultModel,
		Stage:            domain.TokenUsageStageStop,
		PromptTokens:     0,
		CompletionTokens: 0,
		TotalTokens:      0,
		PromptChars:      0,
		CompletionChars:  0,
		WorkspaceRoots:   input.WorkspaceRoots,
		UserEmail:        "",
		HookEventName:    string(input.HookEventName),
		Timestamp:        time.Now(),
	}

	if err := h.deps.Client.CreateTokenUsage(ctx, usage); err != nil {
		h.deps.Log.Error("failed to send stop marker", "error", err)
	}

	return NewEmptyOutput(), nil
}

// BeforeShellExecutionHandler handles beforeShellExecution events.
type BeforeShellExecutionHandler struct {
	deps *HandlerDependencies
}

// Handle processes the beforeShellExecution event.
func (h *BeforeShellExecutionHandler) Handle(_ context.Context, input *HookInput, payload any) (*HookOutput, error) {
	p, ok := payload.(*BeforeShellExecutionPayload)
	if !ok {
		return NewAllowOutput(), nil
	}

	h.deps.Log.Debug("shell execution",
		"command", p.Command,
		"cwd", p.Cwd,
		"conversation_id", input.ConversationID,
	)

	// Always allow shell execution (observability only)
	return NewAllowOutput(), nil
}

// AfterShellExecutionHandler handles afterShellExecution events.
type AfterShellExecutionHandler struct {
	deps *HandlerDependencies
}

// Handle processes the afterShellExecution event.
func (h *AfterShellExecutionHandler) Handle(_ context.Context, input *HookInput, payload any) (*HookOutput, error) {
	p, ok := payload.(*AfterShellExecutionPayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	h.deps.Log.Debug("shell execution completed",
		"command", p.Command,
		"output_length", len(p.Output),
		"conversation_id", input.ConversationID,
	)

	return NewEmptyOutput(), nil
}

// BeforeMCPExecutionHandler handles beforeMCPExecution events.
type BeforeMCPExecutionHandler struct {
	deps *HandlerDependencies
}

// Handle processes the beforeMCPExecution event.
func (h *BeforeMCPExecutionHandler) Handle(_ context.Context, input *HookInput, payload any) (*HookOutput, error) {
	p, ok := payload.(*BeforeMCPExecutionPayload)
	if !ok {
		return NewAllowOutput(), nil
	}

	h.deps.Log.Debug("MCP execution",
		"tool", p.ToolName,
		"conversation_id", input.ConversationID,
	)

	// Always allow MCP execution (observability only)
	return NewAllowOutput(), nil
}

// BeforeReadFileHandler handles beforeReadFile events.
type BeforeReadFileHandler struct {
	deps *HandlerDependencies
}

// Handle processes the beforeReadFile event.
func (h *BeforeReadFileHandler) Handle(_ context.Context, input *HookInput, payload any) (*HookOutput, error) {
	p, ok := payload.(*BeforeReadFilePayload)
	if !ok {
		return NewAllowOutput(), nil
	}

	h.deps.Log.Debug("file read",
		"file", p.FilePath,
		"content_length", len(p.Content),
		"conversation_id", input.ConversationID,
	)

	// Always allow file read (observability only)
	return NewAllowOutput(), nil
}

// parseUUID parses a UUID string, returning a zero UUID on error.
func parseUUID(s string) uuid.UUID {
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil
	}
	return id
}


```

```go
package hook

import (
	"context"
	"fmt"
	"time"

	"github.com/ainslies/ctd/internal/domain"
	"github.com/ainslies/ctd/pkg/logger"
	"github.com/google/uuid"
)

// Handler processes a hook event and returns the appropriate output.
type Handler interface {
	// Handle processes the hook input and returns the output.
	Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error)
}

// HandlerDependencies contains the dependencies for handlers.
type HandlerDependencies struct {
	Client    *CTDClient
	Tokenizer Tokenizer
	Log       *logger.Logger
	Config    *Config
}

// Router routes hook events to the appropriate handler.
type Router struct {
	deps     *HandlerDependencies
	handlers map[HookEventName]Handler
}

// NewRouter creates a new router with all handlers registered.
func NewRouter(deps *HandlerDependencies) *Router {
	router := &Router{
		deps:     deps,
		handlers: make(map[HookEventName]Handler),
	}

	// Register handlers
	router.handlers[EventBeforeSubmitPrompt] = &BeforeSubmitPromptHandler{deps: deps}
	router.handlers[EventAfterAgentResponse] = &AfterAgentResponseHandler{deps: deps}
	router.handlers[EventAfterAgentThought] = &AfterAgentThoughtHandler{deps: deps}
	router.handlers[EventAfterMCPExecution] = &AfterMCPExecutionHandler{deps: deps}
	router.handlers[EventAfterFileEdit] = &AfterFileEditHandler{deps: deps}
	router.handlers[EventStop] = &StopHandler{deps: deps}
	router.handlers[EventBeforeShellExecution] = &BeforeShellExecutionHandler{deps: deps}
	router.handlers[EventAfterShellExecution] = &AfterShellExecutionHandler{deps: deps}
	router.handlers[EventBeforeMCPExecution] = &BeforeMCPExecutionHandler{deps: deps}
	router.handlers[EventBeforeReadFile] = &BeforeReadFileHandler{deps: deps}

	return router
}

// Route routes the input to the appropriate handler.
func (r *Router) Route(ctx context.Context, input *HookInput) (*HookOutput, error) {
	handler, ok := r.handlers[input.HookEventName]
	if !ok {
		// Unknown event, return empty output
		r.deps.Log.Warn("unknown hook event", "event", input.HookEventName)
		return NewEmptyOutput(), nil
	}

	payload, err := input.ParsePayload()
	if err != nil {
		r.deps.Log.Error("failed to parse payload", "error", err, "event", input.HookEventName)
		// Return appropriate default output even on parse error
		return getDefaultOutput(input.HookEventName), nil
	}

	output, err := handler.Handle(ctx, input, payload)
	if err != nil {
		return nil, fmt.Errorf("handler error for %s: %w", input.HookEventName, err)
	}
	return output, nil
}

// getDefaultOutput returns the default output for an event type.
//
//nolint:exhaustive // Only specific events need special handling, others use empty output
func getDefaultOutput(event HookEventName) *HookOutput {
	switch event {
	case EventBeforeSubmitPrompt:
		return NewContinueOutput()
	case EventBeforeShellExecution, EventBeforeMCPExecution, EventBeforeReadFile:
		return NewAllowOutput()
	default:
		return NewEmptyOutput()
	}
}

// BeforeSubmitPromptHandler handles beforeSubmitPrompt events.
type BeforeSubmitPromptHandler struct {
	deps *HandlerDependencies
}

// Handle processes the beforeSubmitPrompt event.
//
//nolint:varnamelen // promptPayload is clear in context
func (h *BeforeSubmitPromptHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	promptPayload, ok := payload.(*BeforeSubmitPromptPayload)
	if !ok {
		return NewContinueOutput(), nil
	}

	// Count tokens
	tokenCount, err := h.deps.Tokenizer.CountTokens(h.deps.Config.DefaultModel, promptPayload.Prompt)
	if err != nil {
		h.deps.Log.Error("failed to count tokens", "error", err)
		tokenCount = 0
	}

	charCount := CountChars(promptPayload.Prompt)

	h.deps.Log.Debug("prompt metrics",
		"tokens", tokenCount,
		"chars", charCount,
		"conversation_id", input.ConversationID,
	)

	// Create token usage record
	usage := &domain.TokenUsage{
		ID:               uuid.New(),
		ConversationID:   parseUUID(input.ConversationID),
		GenerationID:     parseUUID(input.GenerationID),
		Model:            h.deps.Config.DefaultModel,
		Stage:            domain.TokenUsageStagePrompt,
		PromptTokens:     tokenCount,
		CompletionTokens: 0,
		TotalTokens:      tokenCount,
		PromptChars:      charCount,
		CompletionChars:  0,
		WorkspaceRoots:   input.WorkspaceRoots,
		UserEmail:        "",
		HookEventName:    string(input.HookEventName),
		Timestamp:        time.Now(),
	}

	// Send to ctd (best-effort)
	if err := h.deps.Client.CreateTokenUsage(ctx, usage); err != nil {
		h.deps.Log.Error("failed to send token usage", "error", err)
	} else {
		h.deps.Log.Info("token usage recorded", "stage", "prompt", "tokens", tokenCount)
	}

	return NewContinueOutput(), nil
}

// AfterAgentResponseHandler handles afterAgentResponse events.
type AfterAgentResponseHandler struct {
	deps *HandlerDependencies
}

// Handle processes the afterAgentResponse event.
//
//nolint:varnamelen // responsePayload is clear in context
func (h *AfterAgentResponseHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	responsePayload, ok := payload.(*AfterAgentResponsePayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	// Count tokens
	tokenCount, err := h.deps.Tokenizer.CountTokens(h.deps.Config.DefaultModel, responsePayload.Text)
	if err != nil {
		h.deps.Log.Error("failed to count tokens", "error", err)
		tokenCount = 0
	}

	charCount := CountChars(responsePayload.Text)

	h.deps.Log.Debug("completion metrics",
		"tokens", tokenCount,
		"chars", charCount,
		"conversation_id", input.ConversationID,
	)

	// Create token usage record
	usage := &domain.TokenUsage{
		ID:               uuid.New(),
		ConversationID:   parseUUID(input.ConversationID),
		GenerationID:     parseUUID(input.GenerationID),
		Model:            h.deps.Config.DefaultModel,
		Stage:            domain.TokenUsageStageCompletion,
		PromptTokens:     0,
		CompletionTokens: tokenCount,
		TotalTokens:      tokenCount,
		PromptChars:      0,
		CompletionChars:  charCount,
		WorkspaceRoots:   input.WorkspaceRoots,
		UserEmail:        "",
		HookEventName:    string(input.HookEventName),
		Timestamp:        time.Now(),
	}

	// Send token usage to ctd (best-effort)
	if err := h.deps.Client.CreateTokenUsage(ctx, usage); err != nil {
		h.deps.Log.Error("failed to send token usage", "error", err)
	} else {
		h.deps.Log.Info("token usage recorded", "stage", "completion", "tokens", tokenCount)
	}

	// Also send as thought for visibility
	thought := &domain.Thought{
		ID:             uuid.New(),
		ConversationID: parseUUID(input.ConversationID),
		GenerationID:   parseUUID(input.GenerationID),
		Text:           responsePayload.Text,
		DurationMs:     0,
		Model:          h.deps.Config.DefaultModel,
		HookEventName:  string(input.HookEventName),
		CursorVersion:  "",
		WorkspaceRoots: input.WorkspaceRoots,
		UserEmail:      "",
		Timestamp:      time.Now(),
	}

	if err := h.deps.Client.CreateThought(ctx, thought); err != nil {
		h.deps.Log.Error("failed to send thought", "error", err)
	}

	return NewEmptyOutput(), nil
}

// AfterAgentThoughtHandler handles afterAgentThought events.
type AfterAgentThoughtHandler struct {
	deps *HandlerDependencies
}

// Handle processes the afterAgentThought event.
//
//nolint:varnamelen // thoughtPayload is clear in context
func (h *AfterAgentThoughtHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	thoughtPayload, ok := payload.(*AfterAgentThoughtPayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	// Send as thought for visibility
	thought := &domain.Thought{
		ID:             uuid.New(),
		ConversationID: parseUUID(input.ConversationID),
		GenerationID:   parseUUID(input.GenerationID),
		Text:           thoughtPayload.Text,
		DurationMs:     0,
		Model:          h.deps.Config.DefaultModel,
		HookEventName:  string(input.HookEventName),
		CursorVersion:  "",
		WorkspaceRoots: input.WorkspaceRoots,
		UserEmail:      "",
		Timestamp:      time.Now(),
	}

	if err := h.deps.Client.CreateThought(ctx, thought); err != nil {
		h.deps.Log.Error("failed to send thought", "error", err)
	} else {
		h.deps.Log.Info("thought recorded", "event", input.HookEventName)
	}

	return NewEmptyOutput(), nil
}

// AfterMCPExecutionHandler handles afterMCPExecution events.
type AfterMCPExecutionHandler struct {
	deps *HandlerDependencies
}

// Handle processes the afterMCPExecution event.
//
//nolint:varnamelen // mcpPayload is clear in context
func (h *AfterMCPExecutionHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	mcpPayload, ok := payload.(*AfterMCPExecutionPayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	// Count tokens in tool input and result (best effort, ignore errors)
	inputTokens, inputErr := h.deps.Tokenizer.CountTokens(h.deps.Config.DefaultModel, string(mcpPayload.ToolInput))
	if inputErr != nil {
		h.deps.Log.Debug("failed to count input tokens", "error", inputErr)
		inputTokens = 0
	}
	resultTokens, resultErr := h.deps.Tokenizer.CountTokens(h.deps.Config.DefaultModel, string(mcpPayload.ResultJSON))
	if resultErr != nil {
		h.deps.Log.Debug("failed to count result tokens", "error", resultErr)
		resultTokens = 0
	}
	totalTokens := inputTokens + resultTokens

	inputChars := CountChars(string(mcpPayload.ToolInput))
	resultChars := CountChars(string(mcpPayload.ResultJSON))

	h.deps.Log.Debug("tool metrics",
		"tool", mcpPayload.ToolName,
		"input_tokens", inputTokens,
		"result_tokens", resultTokens,
		"conversation_id", input.ConversationID,
	)

	// Create token usage record
	usage := &domain.TokenUsage{
		ID:               uuid.New(),
		ConversationID:   parseUUID(input.ConversationID),
		GenerationID:     parseUUID(input.GenerationID),
		Model:            h.deps.Config.DefaultModel,
		Stage:            domain.TokenUsageStageTool,
		PromptTokens:     inputTokens,
		CompletionTokens: resultTokens,
		TotalTokens:      totalTokens,
		PromptChars:      inputChars,
		CompletionChars:  resultChars,
		WorkspaceRoots:   input.WorkspaceRoots,
		UserEmail:        "",
		HookEventName:    string(input.HookEventName),
		Timestamp:        time.Now(),
	}

	if err := h.deps.Client.CreateTokenUsage(ctx, usage); err != nil {
		h.deps.Log.Error("failed to send token usage", "error", err)
	} else {
		h.deps.Log.Info("token usage recorded", "stage", "tool", "tool", mcpPayload.ToolName, "tokens", totalTokens)
	}

	return NewEmptyOutput(), nil
}

// AfterFileEditHandler handles afterFileEdit events.
type AfterFileEditHandler struct {
	deps *HandlerDependencies
}

// Handle processes the afterFileEdit event.
//
//nolint:varnamelen // editPayload is clear in context
func (h *AfterFileEditHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	editPayload, ok := payload.(*AfterFileEditPayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	// Count characters in edits
	var oldChars, newChars int64
	for _, edit := range editPayload.Edits {
		oldChars += CountChars(edit.OldString)
		newChars += CountChars(edit.NewString)
	}

	h.deps.Log.Debug("file edit metrics",
		"file", editPayload.FilePath,
		"old_chars", oldChars,
		"new_chars", newChars,
		"edit_count", len(editPayload.Edits),
	)

	// Create token usage record for file edits
	usage := &domain.TokenUsage{
		ID:               uuid.New(),
		ConversationID:   parseUUID(input.ConversationID),
		GenerationID:     parseUUID(input.GenerationID),
		Model:            h.deps.Config.DefaultModel,
		Stage:            domain.TokenUsageStageFileEdit,
		PromptTokens:     0,
		CompletionTokens: 0,
		TotalTokens:      0,
		PromptChars:      oldChars,
		CompletionChars:  newChars,
		WorkspaceRoots:   input.WorkspaceRoots,
		UserEmail:        "",
		HookEventName:    string(input.HookEventName),
		Timestamp:        time.Now(),
	}

	if err := h.deps.Client.CreateTokenUsage(ctx, usage); err != nil {
		h.deps.Log.Error("failed to send token usage", "error", err)
	} else {
		h.deps.Log.Info("file edit recorded", "file", editPayload.FilePath, "edits", len(editPayload.Edits))
	}

	return NewEmptyOutput(), nil
}

// StopHandler handles stop events.
type StopHandler struct {
	deps *HandlerDependencies
}

// Handle processes the stop event.
//
//nolint:varnamelen // stopPayload is clear in context
func (h *StopHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	stopPayload, ok := payload.(*StopPayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	h.deps.Log.Info("agent loop stopped",
		"status", stopPayload.Status,
		"loop_count", stopPayload.LoopCount,
		"conversation_id", input.ConversationID,
	)

	// Create token usage record as session marker
	usage := &domain.TokenUsage{
		ID:               uuid.New(),
		ConversationID:   parseUUID(input.ConversationID),
		GenerationID:     parseUUID(input.GenerationID),
		Model:            h.deps.Config.DefaultModel,
		Stage:            domain.TokenUsageStageStop,
		PromptTokens:     0,
		CompletionTokens: 0,
		TotalTokens:      0,
		PromptChars:      0,
		CompletionChars:  0,
		WorkspaceRoots:   input.WorkspaceRoots,
		UserEmail:        "",
		HookEventName:    string(input.HookEventName),
		Timestamp:        time.Now(),
	}

	if err := h.deps.Client.CreateTokenUsage(ctx, usage); err != nil {
		h.deps.Log.Error("failed to send stop marker", "error", err)
	}

	return NewEmptyOutput(), nil
}

// BeforeShellExecutionHandler handles beforeShellExecution events.
type BeforeShellExecutionHandler struct {
	deps *HandlerDependencies
}

// Handle processes the beforeShellExecution event.
//
//nolint:varnamelen // shellPayload is clear in context
func (h *BeforeShellExecutionHandler) Handle(_ context.Context, input *HookInput, payload any) (*HookOutput, error) {
	shellPayload, ok := payload.(*BeforeShellExecutionPayload)
	if !ok {
		return NewAllowOutput(), nil
	}

	h.deps.Log.Debug("shell execution",
		"command", shellPayload.Command,
		"cwd", shellPayload.Cwd,
		"conversation_id", input.ConversationID,
	)

	// Always allow shell execution (observability only)
	return NewAllowOutput(), nil
}

// AfterShellExecutionHandler handles afterShellExecution events.
type AfterShellExecutionHandler struct {
	deps *HandlerDependencies
}

// Handle processes the afterShellExecution event.
func (h *AfterShellExecutionHandler) Handle(_ context.Context, input *HookInput, payload any) (*HookOutput, error) {
	p, ok := payload.(*AfterShellExecutionPayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	h.deps.Log.Debug("shell execution completed",
		"command", p.Command,
		"output_length", len(p.Output),
		"conversation_id", input.ConversationID,
	)

	return NewEmptyOutput(), nil
}

// BeforeMCPExecutionHandler handles beforeMCPExecution events.
type BeforeMCPExecutionHandler struct {
	deps *HandlerDependencies
}

// Handle processes the beforeMCPExecution event.
func (h *BeforeMCPExecutionHandler) Handle(_ context.Context, input *HookInput, payload any) (*HookOutput, error) {
	p, ok := payload.(*BeforeMCPExecutionPayload)
	if !ok {
		return NewAllowOutput(), nil
	}

	h.deps.Log.Debug("MCP execution",
		"tool", p.ToolName,
		"conversation_id", input.ConversationID,
	)

	// Always allow MCP execution (observability only)
	return NewAllowOutput(), nil
}

// BeforeReadFileHandler handles beforeReadFile events.
type BeforeReadFileHandler struct {
	deps *HandlerDependencies
}

// Handle processes the beforeReadFile event.
func (h *BeforeReadFileHandler) Handle(_ context.Context, input *HookInput, payload any) (*HookOutput, error) {
	p, ok := payload.(*BeforeReadFilePayload)
	if !ok {
		return NewAllowOutput(), nil
	}

	h.deps.Log.Debug("file read",
		"file", p.FilePath,
		"content_length", len(p.Content),
		"conversation_id", input.ConversationID,
	)

	// Always allow file read (observability only)
	return NewAllowOutput(), nil
}

// parseUUID parses a UUID string, returning a zero UUID on error.
func parseUUID(s string) uuid.UUID {
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil
	}
	return id
}


```

```go
package hook

import (
	"context"
	"fmt"
	"time"

	"github.com/ainslies/ctd/internal/domain"
	"github.com/ainslies/ctd/pkg/logger"
	"github.com/google/uuid"
)

// Handler processes a hook event and returns the appropriate output.
type Handler interface {
	// Handle processes the hook input and returns the output.
	Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error)
}

// HandlerDependencies contains the dependencies for handlers.
type HandlerDependencies struct {
	Client    *CTDClient
	Tokenizer Tokenizer
	Log       *logger.Logger
	Config    *Config
}

// Router routes hook events to the appropriate handler.
type Router struct {
	deps     *HandlerDependencies
	handlers map[HookEventName]Handler
}

// NewRouter creates a new router with all handlers registered.
func NewRouter(deps *HandlerDependencies) *Router {
	router := &Router{
		deps:     deps,
		handlers: make(map[HookEventName]Handler),
	}

	// Register handlers
	router.handlers[EventBeforeSubmitPrompt] = &BeforeSubmitPromptHandler{deps: deps}
	router.handlers[EventAfterAgentResponse] = &AfterAgentResponseHandler{deps: deps}
	router.handlers[EventAfterAgentThought] = &AfterAgentThoughtHandler{deps: deps}
	router.handlers[EventAfterMCPExecution] = &AfterMCPExecutionHandler{deps: deps}
	router.handlers[EventAfterFileEdit] = &AfterFileEditHandler{deps: deps}
	router.handlers[EventStop] = &StopHandler{deps: deps}
	router.handlers[EventBeforeShellExecution] = &BeforeShellExecutionHandler{deps: deps}
	router.handlers[EventAfterShellExecution] = &AfterShellExecutionHandler{deps: deps}
	router.handlers[EventBeforeMCPExecution] = &BeforeMCPExecutionHandler{deps: deps}
	router.handlers[EventBeforeReadFile] = &BeforeReadFileHandler{deps: deps}

	return router
}

// Route routes the input to the appropriate handler.
func (r *Router) Route(ctx context.Context, input *HookInput) (*HookOutput, error) {
	handler, ok := r.handlers[input.HookEventName]
	if !ok {
		// Unknown event, return empty output
		r.deps.Log.Warn("unknown hook event", "event", input.HookEventName)
		return NewEmptyOutput(), nil
	}

	payload, err := input.ParsePayload()
	if err != nil {
		r.deps.Log.Error("failed to parse payload", "error", err, "event", input.HookEventName)
		// Return appropriate default output even on parse error
		return getDefaultOutput(input.HookEventName), nil
	}

	output, err := handler.Handle(ctx, input, payload)
	if err != nil {
		return nil, fmt.Errorf("handler error for %s: %w", input.HookEventName, err)
	}
	return output, nil
}

// getDefaultOutput returns the default output for an event type.
//
//nolint:exhaustive // Only specific events need special handling, others use empty output
func getDefaultOutput(event HookEventName) *HookOutput {
	switch event {
	case EventBeforeSubmitPrompt:
		return NewContinueOutput()
	case EventBeforeShellExecution, EventBeforeMCPExecution, EventBeforeReadFile:
		return NewAllowOutput()
	default:
		return NewEmptyOutput()
	}
}

// BeforeSubmitPromptHandler handles beforeSubmitPrompt events.
type BeforeSubmitPromptHandler struct {
	deps *HandlerDependencies
}

// Handle processes the beforeSubmitPrompt event.
//
//nolint:varnamelen // promptPayload is clear in context
func (h *BeforeSubmitPromptHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	promptPayload, ok := payload.(*BeforeSubmitPromptPayload)
	if !ok {
		return NewContinueOutput(), nil
	}

	// Count tokens
	tokenCount, err := h.deps.Tokenizer.CountTokens(h.deps.Config.DefaultModel, promptPayload.Prompt)
	if err != nil {
		h.deps.Log.Error("failed to count tokens", "error", err)
		tokenCount = 0
	}

	charCount := CountChars(promptPayload.Prompt)

	h.deps.Log.Debug("prompt metrics",
		"tokens", tokenCount,
		"chars", charCount,
		"conversation_id", input.ConversationID,
	)

	// Create token usage record
	usage := &domain.TokenUsage{
		ID:               uuid.New(),
		ConversationID:   parseUUID(input.ConversationID),
		GenerationID:     parseUUID(input.GenerationID),
		Model:            h.deps.Config.DefaultModel,
		Stage:            domain.TokenUsageStagePrompt,
		PromptTokens:     tokenCount,
		CompletionTokens: 0,
		TotalTokens:      tokenCount,
		PromptChars:      charCount,
		CompletionChars:  0,
		WorkspaceRoots:   input.WorkspaceRoots,
		UserEmail:        "",
		HookEventName:    string(input.HookEventName),
		Timestamp:        time.Now(),
	}

	// Send to ctd (best-effort)
	if err := h.deps.Client.CreateTokenUsage(ctx, usage); err != nil {
		h.deps.Log.Error("failed to send token usage", "error", err)
	} else {
		h.deps.Log.Info("token usage recorded", "stage", "prompt", "tokens", tokenCount)
	}

	return NewContinueOutput(), nil
}

// AfterAgentResponseHandler handles afterAgentResponse events.
type AfterAgentResponseHandler struct {
	deps *HandlerDependencies
}

// Handle processes the afterAgentResponse event.
//
//nolint:varnamelen // responsePayload is clear in context
func (h *AfterAgentResponseHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	responsePayload, ok := payload.(*AfterAgentResponsePayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	// Count tokens
	tokenCount, err := h.deps.Tokenizer.CountTokens(h.deps.Config.DefaultModel, responsePayload.Text)
	if err != nil {
		h.deps.Log.Error("failed to count tokens", "error", err)
		tokenCount = 0
	}

	charCount := CountChars(responsePayload.Text)

	h.deps.Log.Debug("completion metrics",
		"tokens", tokenCount,
		"chars", charCount,
		"conversation_id", input.ConversationID,
	)

	// Create token usage record
	usage := &domain.TokenUsage{
		ID:               uuid.New(),
		ConversationID:   parseUUID(input.ConversationID),
		GenerationID:     parseUUID(input.GenerationID),
		Model:            h.deps.Config.DefaultModel,
		Stage:            domain.TokenUsageStageCompletion,
		PromptTokens:     0,
		CompletionTokens: tokenCount,
		TotalTokens:      tokenCount,
		PromptChars:      0,
		CompletionChars:  charCount,
		WorkspaceRoots:   input.WorkspaceRoots,
		UserEmail:        "",
		HookEventName:    string(input.HookEventName),
		Timestamp:        time.Now(),
	}

	// Send token usage to ctd (best-effort)
	if err := h.deps.Client.CreateTokenUsage(ctx, usage); err != nil {
		h.deps.Log.Error("failed to send token usage", "error", err)
	} else {
		h.deps.Log.Info("token usage recorded", "stage", "completion", "tokens", tokenCount)
	}

	// Also send as thought for visibility
	thought := &domain.Thought{
		ID:             uuid.New(),
		ConversationID: parseUUID(input.ConversationID),
		GenerationID:   parseUUID(input.GenerationID),
		Text:           responsePayload.Text,
		DurationMs:     0,
		Model:          h.deps.Config.DefaultModel,
		HookEventName:  string(input.HookEventName),
		CursorVersion:  "",
		WorkspaceRoots: input.WorkspaceRoots,
		UserEmail:      "",
		Timestamp:      time.Now(),
	}

	if err := h.deps.Client.CreateThought(ctx, thought); err != nil {
		h.deps.Log.Error("failed to send thought", "error", err)
	}

	return NewEmptyOutput(), nil
}

// AfterAgentThoughtHandler handles afterAgentThought events.
type AfterAgentThoughtHandler struct {
	deps *HandlerDependencies
}

// Handle processes the afterAgentThought event.
//
//nolint:varnamelen // thoughtPayload is clear in context
func (h *AfterAgentThoughtHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	thoughtPayload, ok := payload.(*AfterAgentThoughtPayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	// Send as thought for visibility
	thought := &domain.Thought{
		ID:             uuid.New(),
		ConversationID: parseUUID(input.ConversationID),
		GenerationID:   parseUUID(input.GenerationID),
		Text:           thoughtPayload.Text,
		DurationMs:     0,
		Model:          h.deps.Config.DefaultModel,
		HookEventName:  string(input.HookEventName),
		CursorVersion:  "",
		WorkspaceRoots: input.WorkspaceRoots,
		UserEmail:      "",
		Timestamp:      time.Now(),
	}

	if err := h.deps.Client.CreateThought(ctx, thought); err != nil {
		h.deps.Log.Error("failed to send thought", "error", err)
	} else {
		h.deps.Log.Info("thought recorded", "event", input.HookEventName)
	}

	return NewEmptyOutput(), nil
}

// AfterMCPExecutionHandler handles afterMCPExecution events.
type AfterMCPExecutionHandler struct {
	deps *HandlerDependencies
}

// Handle processes the afterMCPExecution event.
//
//nolint:varnamelen // mcpPayload is clear in context
func (h *AfterMCPExecutionHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	mcpPayload, ok := payload.(*AfterMCPExecutionPayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	// Count tokens in tool input and result (best effort, ignore errors)
	inputTokens, inputErr := h.deps.Tokenizer.CountTokens(h.deps.Config.DefaultModel, string(mcpPayload.ToolInput))
	if inputErr != nil {
		h.deps.Log.Debug("failed to count input tokens", "error", inputErr)
		inputTokens = 0
	}
	resultTokens, resultErr := h.deps.Tokenizer.CountTokens(h.deps.Config.DefaultModel, string(mcpPayload.ResultJSON))
	if resultErr != nil {
		h.deps.Log.Debug("failed to count result tokens", "error", resultErr)
		resultTokens = 0
	}
	totalTokens := inputTokens + resultTokens

	inputChars := CountChars(string(mcpPayload.ToolInput))
	resultChars := CountChars(string(mcpPayload.ResultJSON))

	h.deps.Log.Debug("tool metrics",
		"tool", mcpPayload.ToolName,
		"input_tokens", inputTokens,
		"result_tokens", resultTokens,
		"conversation_id", input.ConversationID,
	)

	// Create token usage record
	usage := &domain.TokenUsage{
		ID:               uuid.New(),
		ConversationID:   parseUUID(input.ConversationID),
		GenerationID:     parseUUID(input.GenerationID),
		Model:            h.deps.Config.DefaultModel,
		Stage:            domain.TokenUsageStageTool,
		PromptTokens:     inputTokens,
		CompletionTokens: resultTokens,
		TotalTokens:      totalTokens,
		PromptChars:      inputChars,
		CompletionChars:  resultChars,
		WorkspaceRoots:   input.WorkspaceRoots,
		UserEmail:        "",
		HookEventName:    string(input.HookEventName),
		Timestamp:        time.Now(),
	}

	if err := h.deps.Client.CreateTokenUsage(ctx, usage); err != nil {
		h.deps.Log.Error("failed to send token usage", "error", err)
	} else {
		h.deps.Log.Info("token usage recorded", "stage", "tool", "tool", mcpPayload.ToolName, "tokens", totalTokens)
	}

	return NewEmptyOutput(), nil
}

// AfterFileEditHandler handles afterFileEdit events.
type AfterFileEditHandler struct {
	deps *HandlerDependencies
}

// Handle processes the afterFileEdit event.
//
//nolint:varnamelen // editPayload is clear in context
func (h *AfterFileEditHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	editPayload, ok := payload.(*AfterFileEditPayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	// Count characters in edits
	var oldChars, newChars int64
	for _, edit := range editPayload.Edits {
		oldChars += CountChars(edit.OldString)
		newChars += CountChars(edit.NewString)
	}

	h.deps.Log.Debug("file edit metrics",
		"file", editPayload.FilePath,
		"old_chars", oldChars,
		"new_chars", newChars,
		"edit_count", len(editPayload.Edits),
	)

	// Create token usage record for file edits
	usage := &domain.TokenUsage{
		ID:               uuid.New(),
		ConversationID:   parseUUID(input.ConversationID),
		GenerationID:     parseUUID(input.GenerationID),
		Model:            h.deps.Config.DefaultModel,
		Stage:            domain.TokenUsageStageFileEdit,
		PromptTokens:     0,
		CompletionTokens: 0,
		TotalTokens:      0,
		PromptChars:      oldChars,
		CompletionChars:  newChars,
		WorkspaceRoots:   input.WorkspaceRoots,
		UserEmail:        "",
		HookEventName:    string(input.HookEventName),
		Timestamp:        time.Now(),
	}

	if err := h.deps.Client.CreateTokenUsage(ctx, usage); err != nil {
		h.deps.Log.Error("failed to send token usage", "error", err)
	} else {
		h.deps.Log.Info("file edit recorded", "file", editPayload.FilePath, "edits", len(editPayload.Edits))
	}

	return NewEmptyOutput(), nil
}

// StopHandler handles stop events.
type StopHandler struct {
	deps *HandlerDependencies
}

// Handle processes the stop event.
//
//nolint:varnamelen // stopPayload is clear in context
func (h *StopHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	stopPayload, ok := payload.(*StopPayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	h.deps.Log.Info("agent loop stopped",
		"status", stopPayload.Status,
		"loop_count", stopPayload.LoopCount,
		"conversation_id", input.ConversationID,
	)

	// Create token usage record as session marker
	usage := &domain.TokenUsage{
		ID:               uuid.New(),
		ConversationID:   parseUUID(input.ConversationID),
		GenerationID:     parseUUID(input.GenerationID),
		Model:            h.deps.Config.DefaultModel,
		Stage:            domain.TokenUsageStageStop,
		PromptTokens:     0,
		CompletionTokens: 0,
		TotalTokens:      0,
		PromptChars:      0,
		CompletionChars:  0,
		WorkspaceRoots:   input.WorkspaceRoots,
		UserEmail:        "",
		HookEventName:    string(input.HookEventName),
		Timestamp:        time.Now(),
	}

	if err := h.deps.Client.CreateTokenUsage(ctx, usage); err != nil {
		h.deps.Log.Error("failed to send stop marker", "error", err)
	}

	return NewEmptyOutput(), nil
}

// BeforeShellExecutionHandler handles beforeShellExecution events.
type BeforeShellExecutionHandler struct {
	deps *HandlerDependencies
}

// Handle processes the beforeShellExecution event.
//
//nolint:varnamelen // shellPayload is clear in context
func (h *BeforeShellExecutionHandler) Handle(_ context.Context, input *HookInput, payload any) (*HookOutput, error) {
	shellPayload, ok := payload.(*BeforeShellExecutionPayload)
	if !ok {
		return NewAllowOutput(), nil
	}

	h.deps.Log.Debug("shell execution",
		"command", shellPayload.Command,
		"cwd", shellPayload.Cwd,
		"conversation_id", input.ConversationID,
	)

	// Always allow shell execution (observability only)
	return NewAllowOutput(), nil
}

// AfterShellExecutionHandler handles afterShellExecution events.
type AfterShellExecutionHandler struct {
	deps *HandlerDependencies
}

// Handle processes the afterShellExecution event.
//
//nolint:varnamelen // shellPayload is clear in context
func (h *AfterShellExecutionHandler) Handle(_ context.Context, input *HookInput, payload any) (*HookOutput, error) {
	shellPayload, ok := payload.(*AfterShellExecutionPayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	h.deps.Log.Debug("shell execution completed",
		"command", shellPayload.Command,
		"output_length", len(shellPayload.Output),
		"conversation_id", input.ConversationID,
	)

	return NewEmptyOutput(), nil
}

// BeforeMCPExecutionHandler handles beforeMCPExecution events.
type BeforeMCPExecutionHandler struct {
	deps *HandlerDependencies
}

// Handle processes the beforeMCPExecution event.
func (h *BeforeMCPExecutionHandler) Handle(_ context.Context, input *HookInput, payload any) (*HookOutput, error) {
	p, ok := payload.(*BeforeMCPExecutionPayload)
	if !ok {
		return NewAllowOutput(), nil
	}

	h.deps.Log.Debug("MCP execution",
		"tool", p.ToolName,
		"conversation_id", input.ConversationID,
	)

	// Always allow MCP execution (observability only)
	return NewAllowOutput(), nil
}

// BeforeReadFileHandler handles beforeReadFile events.
type BeforeReadFileHandler struct {
	deps *HandlerDependencies
}

// Handle processes the beforeReadFile event.
func (h *BeforeReadFileHandler) Handle(_ context.Context, input *HookInput, payload any) (*HookOutput, error) {
	p, ok := payload.(*BeforeReadFilePayload)
	if !ok {
		return NewAllowOutput(), nil
	}

	h.deps.Log.Debug("file read",
		"file", p.FilePath,
		"content_length", len(p.Content),
		"conversation_id", input.ConversationID,
	)

	// Always allow file read (observability only)
	return NewAllowOutput(), nil
}

// parseUUID parses a UUID string, returning a zero UUID on error.
func parseUUID(s string) uuid.UUID {
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil
	}
	return id
}


```

```go
package hook

import (
	"context"
	"fmt"
	"time"

	"github.com/ainslies/ctd/internal/domain"
	"github.com/ainslies/ctd/pkg/logger"
	"github.com/google/uuid"
)

// Handler processes a hook event and returns the appropriate output.
type Handler interface {
	// Handle processes the hook input and returns the output.
	Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error)
}

// HandlerDependencies contains the dependencies for handlers.
type HandlerDependencies struct {
	Client    *CTDClient
	Tokenizer Tokenizer
	Log       *logger.Logger
	Config    *Config
}

// Router routes hook events to the appropriate handler.
type Router struct {
	deps     *HandlerDependencies
	handlers map[HookEventName]Handler
}

// NewRouter creates a new router with all handlers registered.
func NewRouter(deps *HandlerDependencies) *Router {
	router := &Router{
		deps:     deps,
		handlers: make(map[HookEventName]Handler),
	}

	// Register handlers
	router.handlers[EventBeforeSubmitPrompt] = &BeforeSubmitPromptHandler{deps: deps}
	router.handlers[EventAfterAgentResponse] = &AfterAgentResponseHandler{deps: deps}
	router.handlers[EventAfterAgentThought] = &AfterAgentThoughtHandler{deps: deps}
	router.handlers[EventAfterMCPExecution] = &AfterMCPExecutionHandler{deps: deps}
	router.handlers[EventAfterFileEdit] = &AfterFileEditHandler{deps: deps}
	router.handlers[EventStop] = &StopHandler{deps: deps}
	router.handlers[EventBeforeShellExecution] = &BeforeShellExecutionHandler{deps: deps}
	router.handlers[EventAfterShellExecution] = &AfterShellExecutionHandler{deps: deps}
	router.handlers[EventBeforeMCPExecution] = &BeforeMCPExecutionHandler{deps: deps}
	router.handlers[EventBeforeReadFile] = &BeforeReadFileHandler{deps: deps}

	return router
}

// Route routes the input to the appropriate handler.
func (r *Router) Route(ctx context.Context, input *HookInput) (*HookOutput, error) {
	handler, ok := r.handlers[input.HookEventName]
	if !ok {
		// Unknown event, return empty output
		r.deps.Log.Warn("unknown hook event", "event", input.HookEventName)
		return NewEmptyOutput(), nil
	}

	payload, err := input.ParsePayload()
	if err != nil {
		r.deps.Log.Error("failed to parse payload", "error", err, "event", input.HookEventName)
		// Return appropriate default output even on parse error
		return getDefaultOutput(input.HookEventName), nil
	}

	output, err := handler.Handle(ctx, input, payload)
	if err != nil {
		return nil, fmt.Errorf("handler error for %s: %w", input.HookEventName, err)
	}
	return output, nil
}

// getDefaultOutput returns the default output for an event type.
//
//nolint:exhaustive // Only specific events need special handling, others use empty output
func getDefaultOutput(event HookEventName) *HookOutput {
	switch event {
	case EventBeforeSubmitPrompt:
		return NewContinueOutput()
	case EventBeforeShellExecution, EventBeforeMCPExecution, EventBeforeReadFile:
		return NewAllowOutput()
	default:
		return NewEmptyOutput()
	}
}

// BeforeSubmitPromptHandler handles beforeSubmitPrompt events.
type BeforeSubmitPromptHandler struct {
	deps *HandlerDependencies
}

// Handle processes the beforeSubmitPrompt event.
//
//nolint:varnamelen // promptPayload is clear in context
func (h *BeforeSubmitPromptHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	promptPayload, ok := payload.(*BeforeSubmitPromptPayload)
	if !ok {
		return NewContinueOutput(), nil
	}

	// Count tokens
	tokenCount, err := h.deps.Tokenizer.CountTokens(h.deps.Config.DefaultModel, promptPayload.Prompt)
	if err != nil {
		h.deps.Log.Error("failed to count tokens", "error", err)
		tokenCount = 0
	}

	charCount := CountChars(promptPayload.Prompt)

	h.deps.Log.Debug("prompt metrics",
		"tokens", tokenCount,
		"chars", charCount,
		"conversation_id", input.ConversationID,
	)

	// Create token usage record
	usage := &domain.TokenUsage{
		ID:               uuid.New(),
		ConversationID:   parseUUID(input.ConversationID),
		GenerationID:     parseUUID(input.GenerationID),
		Model:            h.deps.Config.DefaultModel,
		Stage:            domain.TokenUsageStagePrompt,
		PromptTokens:     tokenCount,
		CompletionTokens: 0,
		TotalTokens:      tokenCount,
		PromptChars:      charCount,
		CompletionChars:  0,
		WorkspaceRoots:   input.WorkspaceRoots,
		UserEmail:        "",
		HookEventName:    string(input.HookEventName),
		Timestamp:        time.Now(),
	}

	// Send to ctd (best-effort)
	if err := h.deps.Client.CreateTokenUsage(ctx, usage); err != nil {
		h.deps.Log.Error("failed to send token usage", "error", err)
	} else {
		h.deps.Log.Info("token usage recorded", "stage", "prompt", "tokens", tokenCount)
	}

	return NewContinueOutput(), nil
}

// AfterAgentResponseHandler handles afterAgentResponse events.
type AfterAgentResponseHandler struct {
	deps *HandlerDependencies
}

// Handle processes the afterAgentResponse event.
//
//nolint:varnamelen // responsePayload is clear in context
func (h *AfterAgentResponseHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	responsePayload, ok := payload.(*AfterAgentResponsePayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	// Count tokens
	tokenCount, err := h.deps.Tokenizer.CountTokens(h.deps.Config.DefaultModel, responsePayload.Text)
	if err != nil {
		h.deps.Log.Error("failed to count tokens", "error", err)
		tokenCount = 0
	}

	charCount := CountChars(responsePayload.Text)

	h.deps.Log.Debug("completion metrics",
		"tokens", tokenCount,
		"chars", charCount,
		"conversation_id", input.ConversationID,
	)

	// Create token usage record
	usage := &domain.TokenUsage{
		ID:               uuid.New(),
		ConversationID:   parseUUID(input.ConversationID),
		GenerationID:     parseUUID(input.GenerationID),
		Model:            h.deps.Config.DefaultModel,
		Stage:            domain.TokenUsageStageCompletion,
		PromptTokens:     0,
		CompletionTokens: tokenCount,
		TotalTokens:      tokenCount,
		PromptChars:      0,
		CompletionChars:  charCount,
		WorkspaceRoots:   input.WorkspaceRoots,
		UserEmail:        "",
		HookEventName:    string(input.HookEventName),
		Timestamp:        time.Now(),
	}

	// Send token usage to ctd (best-effort)
	if err := h.deps.Client.CreateTokenUsage(ctx, usage); err != nil {
		h.deps.Log.Error("failed to send token usage", "error", err)
	} else {
		h.deps.Log.Info("token usage recorded", "stage", "completion", "tokens", tokenCount)
	}

	// Also send as thought for visibility
	thought := &domain.Thought{
		ID:             uuid.New(),
		ConversationID: parseUUID(input.ConversationID),
		GenerationID:   parseUUID(input.GenerationID),
		Text:           responsePayload.Text,
		DurationMs:     0,
		Model:          h.deps.Config.DefaultModel,
		HookEventName:  string(input.HookEventName),
		CursorVersion:  "",
		WorkspaceRoots: input.WorkspaceRoots,
		UserEmail:      "",
		Timestamp:      time.Now(),
	}

	if err := h.deps.Client.CreateThought(ctx, thought); err != nil {
		h.deps.Log.Error("failed to send thought", "error", err)
	}

	return NewEmptyOutput(), nil
}

// AfterAgentThoughtHandler handles afterAgentThought events.
type AfterAgentThoughtHandler struct {
	deps *HandlerDependencies
}

// Handle processes the afterAgentThought event.
//
//nolint:varnamelen // thoughtPayload is clear in context
func (h *AfterAgentThoughtHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	thoughtPayload, ok := payload.(*AfterAgentThoughtPayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	// Send as thought for visibility
	thought := &domain.Thought{
		ID:             uuid.New(),
		ConversationID: parseUUID(input.ConversationID),
		GenerationID:   parseUUID(input.GenerationID),
		Text:           thoughtPayload.Text,
		DurationMs:     0,
		Model:          h.deps.Config.DefaultModel,
		HookEventName:  string(input.HookEventName),
		CursorVersion:  "",
		WorkspaceRoots: input.WorkspaceRoots,
		UserEmail:      "",
		Timestamp:      time.Now(),
	}

	if err := h.deps.Client.CreateThought(ctx, thought); err != nil {
		h.deps.Log.Error("failed to send thought", "error", err)
	} else {
		h.deps.Log.Info("thought recorded", "event", input.HookEventName)
	}

	return NewEmptyOutput(), nil
}

// AfterMCPExecutionHandler handles afterMCPExecution events.
type AfterMCPExecutionHandler struct {
	deps *HandlerDependencies
}

// Handle processes the afterMCPExecution event.
//
//nolint:varnamelen // mcpPayload is clear in context
func (h *AfterMCPExecutionHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	mcpPayload, ok := payload.(*AfterMCPExecutionPayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	// Count tokens in tool input and result (best effort, ignore errors)
	inputTokens, inputErr := h.deps.Tokenizer.CountTokens(h.deps.Config.DefaultModel, string(mcpPayload.ToolInput))
	if inputErr != nil {
		h.deps.Log.Debug("failed to count input tokens", "error", inputErr)
		inputTokens = 0
	}
	resultTokens, resultErr := h.deps.Tokenizer.CountTokens(h.deps.Config.DefaultModel, string(mcpPayload.ResultJSON))
	if resultErr != nil {
		h.deps.Log.Debug("failed to count result tokens", "error", resultErr)
		resultTokens = 0
	}
	totalTokens := inputTokens + resultTokens

	inputChars := CountChars(string(mcpPayload.ToolInput))
	resultChars := CountChars(string(mcpPayload.ResultJSON))

	h.deps.Log.Debug("tool metrics",
		"tool", mcpPayload.ToolName,
		"input_tokens", inputTokens,
		"result_tokens", resultTokens,
		"conversation_id", input.ConversationID,
	)

	// Create token usage record
	usage := &domain.TokenUsage{
		ID:               uuid.New(),
		ConversationID:   parseUUID(input.ConversationID),
		GenerationID:     parseUUID(input.GenerationID),
		Model:            h.deps.Config.DefaultModel,
		Stage:            domain.TokenUsageStageTool,
		PromptTokens:     inputTokens,
		CompletionTokens: resultTokens,
		TotalTokens:      totalTokens,
		PromptChars:      inputChars,
		CompletionChars:  resultChars,
		WorkspaceRoots:   input.WorkspaceRoots,
		UserEmail:        "",
		HookEventName:    string(input.HookEventName),
		Timestamp:        time.Now(),
	}

	if err := h.deps.Client.CreateTokenUsage(ctx, usage); err != nil {
		h.deps.Log.Error("failed to send token usage", "error", err)
	} else {
		h.deps.Log.Info("token usage recorded", "stage", "tool", "tool", mcpPayload.ToolName, "tokens", totalTokens)
	}

	return NewEmptyOutput(), nil
}

// AfterFileEditHandler handles afterFileEdit events.
type AfterFileEditHandler struct {
	deps *HandlerDependencies
}

// Handle processes the afterFileEdit event.
//
//nolint:varnamelen // editPayload is clear in context
func (h *AfterFileEditHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	editPayload, ok := payload.(*AfterFileEditPayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	// Count characters in edits
	var oldChars, newChars int64
	for _, edit := range editPayload.Edits {
		oldChars += CountChars(edit.OldString)
		newChars += CountChars(edit.NewString)
	}

	h.deps.Log.Debug("file edit metrics",
		"file", editPayload.FilePath,
		"old_chars", oldChars,
		"new_chars", newChars,
		"edit_count", len(editPayload.Edits),
	)

	// Create token usage record for file edits
	usage := &domain.TokenUsage{
		ID:               uuid.New(),
		ConversationID:   parseUUID(input.ConversationID),
		GenerationID:     parseUUID(input.GenerationID),
		Model:            h.deps.Config.DefaultModel,
		Stage:            domain.TokenUsageStageFileEdit,
		PromptTokens:     0,
		CompletionTokens: 0,
		TotalTokens:      0,
		PromptChars:      oldChars,
		CompletionChars:  newChars,
		WorkspaceRoots:   input.WorkspaceRoots,
		UserEmail:        "",
		HookEventName:    string(input.HookEventName),
		Timestamp:        time.Now(),
	}

	if err := h.deps.Client.CreateTokenUsage(ctx, usage); err != nil {
		h.deps.Log.Error("failed to send token usage", "error", err)
	} else {
		h.deps.Log.Info("file edit recorded", "file", editPayload.FilePath, "edits", len(editPayload.Edits))
	}

	return NewEmptyOutput(), nil
}

// StopHandler handles stop events.
type StopHandler struct {
	deps *HandlerDependencies
}

// Handle processes the stop event.
//
//nolint:varnamelen // stopPayload is clear in context
func (h *StopHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	stopPayload, ok := payload.(*StopPayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	h.deps.Log.Info("agent loop stopped",
		"status", stopPayload.Status,
		"loop_count", stopPayload.LoopCount,
		"conversation_id", input.ConversationID,
	)

	// Create token usage record as session marker
	usage := &domain.TokenUsage{
		ID:               uuid.New(),
		ConversationID:   parseUUID(input.ConversationID),
		GenerationID:     parseUUID(input.GenerationID),
		Model:            h.deps.Config.DefaultModel,
		Stage:            domain.TokenUsageStageStop,
		PromptTokens:     0,
		CompletionTokens: 0,
		TotalTokens:      0,
		PromptChars:      0,
		CompletionChars:  0,
		WorkspaceRoots:   input.WorkspaceRoots,
		UserEmail:        "",
		HookEventName:    string(input.HookEventName),
		Timestamp:        time.Now(),
	}

	if err := h.deps.Client.CreateTokenUsage(ctx, usage); err != nil {
		h.deps.Log.Error("failed to send stop marker", "error", err)
	}

	return NewEmptyOutput(), nil
}

// BeforeShellExecutionHandler handles beforeShellExecution events.
type BeforeShellExecutionHandler struct {
	deps *HandlerDependencies
}

// Handle processes the beforeShellExecution event.
//
//nolint:varnamelen // shellPayload is clear in context
func (h *BeforeShellExecutionHandler) Handle(_ context.Context, input *HookInput, payload any) (*HookOutput, error) {
	shellPayload, ok := payload.(*BeforeShellExecutionPayload)
	if !ok {
		return NewAllowOutput(), nil
	}

	h.deps.Log.Debug("shell execution",
		"command", shellPayload.Command,
		"cwd", shellPayload.Cwd,
		"conversation_id", input.ConversationID,
	)

	// Always allow shell execution (observability only)
	return NewAllowOutput(), nil
}

// AfterShellExecutionHandler handles afterShellExecution events.
type AfterShellExecutionHandler struct {
	deps *HandlerDependencies
}

// Handle processes the afterShellExecution event.
//
//nolint:varnamelen // shellPayload is clear in context
func (h *AfterShellExecutionHandler) Handle(_ context.Context, input *HookInput, payload any) (*HookOutput, error) {
	shellPayload, ok := payload.(*AfterShellExecutionPayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	h.deps.Log.Debug("shell execution completed",
		"command", shellPayload.Command,
		"output_length", len(shellPayload.Output),
		"conversation_id", input.ConversationID,
	)

	return NewEmptyOutput(), nil
}

// BeforeMCPExecutionHandler handles beforeMCPExecution events.
type BeforeMCPExecutionHandler struct {
	deps *HandlerDependencies
}

// Handle processes the beforeMCPExecution event.
//
//nolint:varnamelen // mcpPayload is clear in context
func (h *BeforeMCPExecutionHandler) Handle(_ context.Context, input *HookInput, payload any) (*HookOutput, error) {
	mcpPayload, ok := payload.(*BeforeMCPExecutionPayload)
	if !ok {
		return NewAllowOutput(), nil
	}

	h.deps.Log.Debug("MCP execution",
		"tool", mcpPayload.ToolName,
		"conversation_id", input.ConversationID,
	)

	// Always allow MCP execution (observability only)
	return NewAllowOutput(), nil
}

// BeforeReadFileHandler handles beforeReadFile events.
type BeforeReadFileHandler struct {
	deps *HandlerDependencies
}

// Handle processes the beforeReadFile event.
func (h *BeforeReadFileHandler) Handle(_ context.Context, input *HookInput, payload any) (*HookOutput, error) {
	p, ok := payload.(*BeforeReadFilePayload)
	if !ok {
		return NewAllowOutput(), nil
	}

	h.deps.Log.Debug("file read",
		"file", p.FilePath,
		"content_length", len(p.Content),
		"conversation_id", input.ConversationID,
	)

	// Always allow file read (observability only)
	return NewAllowOutput(), nil
}

// parseUUID parses a UUID string, returning a zero UUID on error.
func parseUUID(s string) uuid.UUID {
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil
	}
	return id
}


```

```go
package hook

import (
	"context"
	"fmt"
	"time"

	"github.com/ainslies/ctd/internal/domain"
	"github.com/ainslies/ctd/pkg/logger"
	"github.com/google/uuid"
)

// Handler processes a hook event and returns the appropriate output.
type Handler interface {
	// Handle processes the hook input and returns the output.
	Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error)
}

// HandlerDependencies contains the dependencies for handlers.
type HandlerDependencies struct {
	Client    *CTDClient
	Tokenizer Tokenizer
	Log       *logger.Logger
	Config    *Config
}

// Router routes hook events to the appropriate handler.
type Router struct {
	deps     *HandlerDependencies
	handlers map[HookEventName]Handler
}

// NewRouter creates a new router with all handlers registered.
func NewRouter(deps *HandlerDependencies) *Router {
	router := &Router{
		deps:     deps,
		handlers: make(map[HookEventName]Handler),
	}

	// Register handlers
	router.handlers[EventBeforeSubmitPrompt] = &BeforeSubmitPromptHandler{deps: deps}
	router.handlers[EventAfterAgentResponse] = &AfterAgentResponseHandler{deps: deps}
	router.handlers[EventAfterAgentThought] = &AfterAgentThoughtHandler{deps: deps}
	router.handlers[EventAfterMCPExecution] = &AfterMCPExecutionHandler{deps: deps}
	router.handlers[EventAfterFileEdit] = &AfterFileEditHandler{deps: deps}
	router.handlers[EventStop] = &StopHandler{deps: deps}
	router.handlers[EventBeforeShellExecution] = &BeforeShellExecutionHandler{deps: deps}
	router.handlers[EventAfterShellExecution] = &AfterShellExecutionHandler{deps: deps}
	router.handlers[EventBeforeMCPExecution] = &BeforeMCPExecutionHandler{deps: deps}
	router.handlers[EventBeforeReadFile] = &BeforeReadFileHandler{deps: deps}

	return router
}

// Route routes the input to the appropriate handler.
func (r *Router) Route(ctx context.Context, input *HookInput) (*HookOutput, error) {
	handler, ok := r.handlers[input.HookEventName]
	if !ok {
		// Unknown event, return empty output
		r.deps.Log.Warn("unknown hook event", "event", input.HookEventName)
		return NewEmptyOutput(), nil
	}

	payload, err := input.ParsePayload()
	if err != nil {
		r.deps.Log.Error("failed to parse payload", "error", err, "event", input.HookEventName)
		// Return appropriate default output even on parse error
		return getDefaultOutput(input.HookEventName), nil
	}

	output, err := handler.Handle(ctx, input, payload)
	if err != nil {
		return nil, fmt.Errorf("handler error for %s: %w", input.HookEventName, err)
	}
	return output, nil
}

// getDefaultOutput returns the default output for an event type.
//
//nolint:exhaustive // Only specific events need special handling, others use empty output
func getDefaultOutput(event HookEventName) *HookOutput {
	switch event {
	case EventBeforeSubmitPrompt:
		return NewContinueOutput()
	case EventBeforeShellExecution, EventBeforeMCPExecution, EventBeforeReadFile:
		return NewAllowOutput()
	default:
		return NewEmptyOutput()
	}
}

// BeforeSubmitPromptHandler handles beforeSubmitPrompt events.
type BeforeSubmitPromptHandler struct {
	deps *HandlerDependencies
}

// Handle processes the beforeSubmitPrompt event.
//
//nolint:varnamelen // promptPayload is clear in context
func (h *BeforeSubmitPromptHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	promptPayload, ok := payload.(*BeforeSubmitPromptPayload)
	if !ok {
		return NewContinueOutput(), nil
	}

	// Count tokens
	tokenCount, err := h.deps.Tokenizer.CountTokens(h.deps.Config.DefaultModel, promptPayload.Prompt)
	if err != nil {
		h.deps.Log.Error("failed to count tokens", "error", err)
		tokenCount = 0
	}

	charCount := CountChars(promptPayload.Prompt)

	h.deps.Log.Debug("prompt metrics",
		"tokens", tokenCount,
		"chars", charCount,
		"conversation_id", input.ConversationID,
	)

	// Create token usage record
	usage := &domain.TokenUsage{
		ID:               uuid.New(),
		ConversationID:   parseUUID(input.ConversationID),
		GenerationID:     parseUUID(input.GenerationID),
		Model:            h.deps.Config.DefaultModel,
		Stage:            domain.TokenUsageStagePrompt,
		PromptTokens:     tokenCount,
		CompletionTokens: 0,
		TotalTokens:      tokenCount,
		PromptChars:      charCount,
		CompletionChars:  0,
		WorkspaceRoots:   input.WorkspaceRoots,
		UserEmail:        "",
		HookEventName:    string(input.HookEventName),
		Timestamp:        time.Now(),
	}

	// Send to ctd (best-effort)
	if err := h.deps.Client.CreateTokenUsage(ctx, usage); err != nil {
		h.deps.Log.Error("failed to send token usage", "error", err)
	} else {
		h.deps.Log.Info("token usage recorded", "stage", "prompt", "tokens", tokenCount)
	}

	return NewContinueOutput(), nil
}

// AfterAgentResponseHandler handles afterAgentResponse events.
type AfterAgentResponseHandler struct {
	deps *HandlerDependencies
}

// Handle processes the afterAgentResponse event.
//
//nolint:varnamelen // responsePayload is clear in context
func (h *AfterAgentResponseHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	responsePayload, ok := payload.(*AfterAgentResponsePayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	// Count tokens
	tokenCount, err := h.deps.Tokenizer.CountTokens(h.deps.Config.DefaultModel, responsePayload.Text)
	if err != nil {
		h.deps.Log.Error("failed to count tokens", "error", err)
		tokenCount = 0
	}

	charCount := CountChars(responsePayload.Text)

	h.deps.Log.Debug("completion metrics",
		"tokens", tokenCount,
		"chars", charCount,
		"conversation_id", input.ConversationID,
	)

	// Create token usage record
	usage := &domain.TokenUsage{
		ID:               uuid.New(),
		ConversationID:   parseUUID(input.ConversationID),
		GenerationID:     parseUUID(input.GenerationID),
		Model:            h.deps.Config.DefaultModel,
		Stage:            domain.TokenUsageStageCompletion,
		PromptTokens:     0,
		CompletionTokens: tokenCount,
		TotalTokens:      tokenCount,
		PromptChars:      0,
		CompletionChars:  charCount,
		WorkspaceRoots:   input.WorkspaceRoots,
		UserEmail:        "",
		HookEventName:    string(input.HookEventName),
		Timestamp:        time.Now(),
	}

	// Send token usage to ctd (best-effort)
	if err := h.deps.Client.CreateTokenUsage(ctx, usage); err != nil {
		h.deps.Log.Error("failed to send token usage", "error", err)
	} else {
		h.deps.Log.Info("token usage recorded", "stage", "completion", "tokens", tokenCount)
	}

	// Also send as thought for visibility
	thought := &domain.Thought{
		ID:             uuid.New(),
		ConversationID: parseUUID(input.ConversationID),
		GenerationID:   parseUUID(input.GenerationID),
		Text:           responsePayload.Text,
		DurationMs:     0,
		Model:          h.deps.Config.DefaultModel,
		HookEventName:  string(input.HookEventName),
		CursorVersion:  "",
		WorkspaceRoots: input.WorkspaceRoots,
		UserEmail:      "",
		Timestamp:      time.Now(),
	}

	if err := h.deps.Client.CreateThought(ctx, thought); err != nil {
		h.deps.Log.Error("failed to send thought", "error", err)
	}

	return NewEmptyOutput(), nil
}

// AfterAgentThoughtHandler handles afterAgentThought events.
type AfterAgentThoughtHandler struct {
	deps *HandlerDependencies
}

// Handle processes the afterAgentThought event.
//
//nolint:varnamelen // thoughtPayload is clear in context
func (h *AfterAgentThoughtHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	thoughtPayload, ok := payload.(*AfterAgentThoughtPayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	// Send as thought for visibility
	thought := &domain.Thought{
		ID:             uuid.New(),
		ConversationID: parseUUID(input.ConversationID),
		GenerationID:   parseUUID(input.GenerationID),
		Text:           thoughtPayload.Text,
		DurationMs:     0,
		Model:          h.deps.Config.DefaultModel,
		HookEventName:  string(input.HookEventName),
		CursorVersion:  "",
		WorkspaceRoots: input.WorkspaceRoots,
		UserEmail:      "",
		Timestamp:      time.Now(),
	}

	if err := h.deps.Client.CreateThought(ctx, thought); err != nil {
		h.deps.Log.Error("failed to send thought", "error", err)
	} else {
		h.deps.Log.Info("thought recorded", "event", input.HookEventName)
	}

	return NewEmptyOutput(), nil
}

// AfterMCPExecutionHandler handles afterMCPExecution events.
type AfterMCPExecutionHandler struct {
	deps *HandlerDependencies
}

// Handle processes the afterMCPExecution event.
//
//nolint:varnamelen // mcpPayload is clear in context
func (h *AfterMCPExecutionHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	mcpPayload, ok := payload.(*AfterMCPExecutionPayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	// Count tokens in tool input and result (best effort, ignore errors)
	inputTokens, inputErr := h.deps.Tokenizer.CountTokens(h.deps.Config.DefaultModel, string(mcpPayload.ToolInput))
	if inputErr != nil {
		h.deps.Log.Debug("failed to count input tokens", "error", inputErr)
		inputTokens = 0
	}
	resultTokens, resultErr := h.deps.Tokenizer.CountTokens(h.deps.Config.DefaultModel, string(mcpPayload.ResultJSON))
	if resultErr != nil {
		h.deps.Log.Debug("failed to count result tokens", "error", resultErr)
		resultTokens = 0
	}
	totalTokens := inputTokens + resultTokens

	inputChars := CountChars(string(mcpPayload.ToolInput))
	resultChars := CountChars(string(mcpPayload.ResultJSON))

	h.deps.Log.Debug("tool metrics",
		"tool", mcpPayload.ToolName,
		"input_tokens", inputTokens,
		"result_tokens", resultTokens,
		"conversation_id", input.ConversationID,
	)

	// Create token usage record
	usage := &domain.TokenUsage{
		ID:               uuid.New(),
		ConversationID:   parseUUID(input.ConversationID),
		GenerationID:     parseUUID(input.GenerationID),
		Model:            h.deps.Config.DefaultModel,
		Stage:            domain.TokenUsageStageTool,
		PromptTokens:     inputTokens,
		CompletionTokens: resultTokens,
		TotalTokens:      totalTokens,
		PromptChars:      inputChars,
		CompletionChars:  resultChars,
		WorkspaceRoots:   input.WorkspaceRoots,
		UserEmail:        "",
		HookEventName:    string(input.HookEventName),
		Timestamp:        time.Now(),
	}

	if err := h.deps.Client.CreateTokenUsage(ctx, usage); err != nil {
		h.deps.Log.Error("failed to send token usage", "error", err)
	} else {
		h.deps.Log.Info("token usage recorded", "stage", "tool", "tool", mcpPayload.ToolName, "tokens", totalTokens)
	}

	return NewEmptyOutput(), nil
}

// AfterFileEditHandler handles afterFileEdit events.
type AfterFileEditHandler struct {
	deps *HandlerDependencies
}

// Handle processes the afterFileEdit event.
//
//nolint:varnamelen // editPayload is clear in context
func (h *AfterFileEditHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	editPayload, ok := payload.(*AfterFileEditPayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	// Count characters in edits
	var oldChars, newChars int64
	for _, edit := range editPayload.Edits {
		oldChars += CountChars(edit.OldString)
		newChars += CountChars(edit.NewString)
	}

	h.deps.Log.Debug("file edit metrics",
		"file", editPayload.FilePath,
		"old_chars", oldChars,
		"new_chars", newChars,
		"edit_count", len(editPayload.Edits),
	)

	// Create token usage record for file edits
	usage := &domain.TokenUsage{
		ID:               uuid.New(),
		ConversationID:   parseUUID(input.ConversationID),
		GenerationID:     parseUUID(input.GenerationID),
		Model:            h.deps.Config.DefaultModel,
		Stage:            domain.TokenUsageStageFileEdit,
		PromptTokens:     0,
		CompletionTokens: 0,
		TotalTokens:      0,
		PromptChars:      oldChars,
		CompletionChars:  newChars,
		WorkspaceRoots:   input.WorkspaceRoots,
		UserEmail:        "",
		HookEventName:    string(input.HookEventName),
		Timestamp:        time.Now(),
	}

	if err := h.deps.Client.CreateTokenUsage(ctx, usage); err != nil {
		h.deps.Log.Error("failed to send token usage", "error", err)
	} else {
		h.deps.Log.Info("file edit recorded", "file", editPayload.FilePath, "edits", len(editPayload.Edits))
	}

	return NewEmptyOutput(), nil
}

// StopHandler handles stop events.
type StopHandler struct {
	deps *HandlerDependencies
}

// Handle processes the stop event.
//
//nolint:varnamelen // stopPayload is clear in context
func (h *StopHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	stopPayload, ok := payload.(*StopPayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	h.deps.Log.Info("agent loop stopped",
		"status", stopPayload.Status,
		"loop_count", stopPayload.LoopCount,
		"conversation_id", input.ConversationID,
	)

	// Create token usage record as session marker
	usage := &domain.TokenUsage{
		ID:               uuid.New(),
		ConversationID:   parseUUID(input.ConversationID),
		GenerationID:     parseUUID(input.GenerationID),
		Model:            h.deps.Config.DefaultModel,
		Stage:            domain.TokenUsageStageStop,
		PromptTokens:     0,
		CompletionTokens: 0,
		TotalTokens:      0,
		PromptChars:      0,
		CompletionChars:  0,
		WorkspaceRoots:   input.WorkspaceRoots,
		UserEmail:        "",
		HookEventName:    string(input.HookEventName),
		Timestamp:        time.Now(),
	}

	if err := h.deps.Client.CreateTokenUsage(ctx, usage); err != nil {
		h.deps.Log.Error("failed to send stop marker", "error", err)
	}

	return NewEmptyOutput(), nil
}

// BeforeShellExecutionHandler handles beforeShellExecution events.
type BeforeShellExecutionHandler struct {
	deps *HandlerDependencies
}

// Handle processes the beforeShellExecution event.
//
//nolint:varnamelen // shellPayload is clear in context
func (h *BeforeShellExecutionHandler) Handle(_ context.Context, input *HookInput, payload any) (*HookOutput, error) {
	shellPayload, ok := payload.(*BeforeShellExecutionPayload)
	if !ok {
		return NewAllowOutput(), nil
	}

	h.deps.Log.Debug("shell execution",
		"command", shellPayload.Command,
		"cwd", shellPayload.Cwd,
		"conversation_id", input.ConversationID,
	)

	// Always allow shell execution (observability only)
	return NewAllowOutput(), nil
}

// AfterShellExecutionHandler handles afterShellExecution events.
type AfterShellExecutionHandler struct {
	deps *HandlerDependencies
}

// Handle processes the afterShellExecution event.
//
//nolint:varnamelen // shellPayload is clear in context
func (h *AfterShellExecutionHandler) Handle(_ context.Context, input *HookInput, payload any) (*HookOutput, error) {
	shellPayload, ok := payload.(*AfterShellExecutionPayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	h.deps.Log.Debug("shell execution completed",
		"command", shellPayload.Command,
		"output_length", len(shellPayload.Output),
		"conversation_id", input.ConversationID,
	)

	return NewEmptyOutput(), nil
}

// BeforeMCPExecutionHandler handles beforeMCPExecution events.
type BeforeMCPExecutionHandler struct {
	deps *HandlerDependencies
}

// Handle processes the beforeMCPExecution event.
//
//nolint:varnamelen // mcpPayload is clear in context
func (h *BeforeMCPExecutionHandler) Handle(_ context.Context, input *HookInput, payload any) (*HookOutput, error) {
	mcpPayload, ok := payload.(*BeforeMCPExecutionPayload)
	if !ok {
		return NewAllowOutput(), nil
	}

	h.deps.Log.Debug("MCP execution",
		"tool", mcpPayload.ToolName,
		"conversation_id", input.ConversationID,
	)

	// Always allow MCP execution (observability only)
	return NewAllowOutput(), nil
}

// BeforeReadFileHandler handles beforeReadFile events.
type BeforeReadFileHandler struct {
	deps *HandlerDependencies
}

// Handle processes the beforeReadFile event.
//
//nolint:varnamelen // readPayload is clear in context
func (h *BeforeReadFileHandler) Handle(_ context.Context, input *HookInput, payload any) (*HookOutput, error) {
	readPayload, ok := payload.(*BeforeReadFilePayload)
	if !ok {
		return NewAllowOutput(), nil
	}

	h.deps.Log.Debug("file read",
		"file", readPayload.FilePath,
		"content_length", len(readPayload.Content),
		"conversation_id", input.ConversationID,
	)

	// Always allow file read (observability only)
	return NewAllowOutput(), nil
}

// parseUUID parses a UUID string, returning a zero UUID on error.
func parseUUID(s string) uuid.UUID {
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil
	}
	return id
}


```

Now let me fix the client.go and server files:

```go
package hook

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"connectrpc.com/connect"
	apiv1 "github.com/ainslies/ctd/gen/api/proto/v1"
	"github.com/ainslies/ctd/gen/api/proto/v1/apiv1connect"
	"github.com/ainslies/ctd/internal/domain"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// CTDClient is a client for communicating with the ctd daemon.
type CTDClient struct {
	thoughtClient    apiv1connect.ThoughtServiceClient
	tokenUsageClient apiv1connect.TokenUsageServiceClient
	timeout          time.Duration
}

// NewCTDClient creates a new CTD client.
func NewCTDClient(serverURL string, timeout time.Duration) *CTDClient {
	httpClient := &http.Client{
		Timeout: timeout,
	}

	return &CTDClient{
		thoughtClient:    apiv1connect.NewThoughtServiceClient(httpClient, serverURL),
		tokenUsageClient: apiv1connect.NewTokenUsageServiceClient(httpClient, serverURL),
		timeout:          timeout,
	}
}

// CreateTokenUsage sends a token usage record to ctd.
func (c *CTDClient) CreateTokenUsage(ctx context.Context, usage *domain.TokenUsage) error {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	protoUsage := &apiv1.TokenUsage{
		ConversationId:   usage.ConversationID.String(),
		GenerationId:     usage.GenerationID.String(),
		Model:            usage.Model,
		Stage:            mapDomainStageToProto(usage.Stage),
		PromptTokens:     usage.PromptTokens,
		CompletionTokens: usage.CompletionTokens,
		TotalTokens:      usage.TotalTokens,
		PromptChars:      usage.PromptChars,
		CompletionChars:  usage.CompletionChars,
		WorkspaceRoots:   usage.WorkspaceRoots,
		UserEmail:        usage.UserEmail,
		HookEventName:    usage.HookEventName,
		Timestamp:        timestamppb.New(usage.Timestamp),
	}

	req := connect.NewRequest(&apiv1.CreateTokenUsageRequest{
		Usage: protoUsage,
	})

	_, err := c.tokenUsageClient.CreateTokenUsage(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to create token usage: %w", err)
	}

	return nil
}

// CreateThought sends a thought record to ctd.
func (c *CTDClient) CreateThought(ctx context.Context, thought *domain.Thought) error {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	protoThought := &apiv1.Thought{
		ConversationId: thought.ConversationID.String(),
		GenerationId:   thought.GenerationID.String(),
		Text:           thought.Text,
		DurationMs:     int32(thought.DurationMs),
		Model:          thought.Model,
		HookEventName:  thought.HookEventName,
		CursorVersion:  thought.CursorVersion,
		WorkspaceRoots: thought.WorkspaceRoots,
		UserEmail:      thought.UserEmail,
		Timestamp:      timestamppb.New(thought.Timestamp),
	}

	req := connect.NewRequest(&apiv1.CreateThoughtRequest{
		Thought: protoThought,
	})

	_, err := c.thoughtClient.CreateThought(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to create thought: %w", err)
	}

	return nil
}

// mapDomainStageToProto converts a domain TokenUsageStage to a proto TokenUsageStage.
//
//nolint:exhaustive // Unspecified is handled by default case
func mapDomainStageToProto(stage domain.TokenUsageStage) apiv1.TokenUsageStage {
	switch stage {
	case domain.TokenUsageStagePrompt:
		return apiv1.TokenUsageStage_TOKEN_USAGE_STAGE_PROMPT
	case domain.TokenUsageStageCompletion:
		return apiv1.TokenUsageStage_TOKEN_USAGE_STAGE_COMPLETION
	case domain.TokenUsageStageTool:
		return apiv1.TokenUsageStage_TOKEN_USAGE_STAGE_TOOL
	case domain.TokenUsageStageSystem:
		return apiv1.TokenUsageStage_TOKEN_USAGE_STAGE_SYSTEM
	case domain.TokenUsageStageFileEdit:
		return apiv1.TokenUsageStage_TOKEN_USAGE_STAGE_FILE_EDIT
	case domain.TokenUsageStageStop:
		return apiv1.TokenUsageStage_TOKEN_USAGE_STAGE_STOP
	default:
		return apiv1.TokenUsageStage_TOKEN_USAGE_STAGE_UNSPECIFIED
	}
}


```

```go
package hook

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"connectrpc.com/connect"
	apiv1 "github.com/ainslies/ctd/gen/api/proto/v1"
	"github.com/ainslies/ctd/gen/api/proto/v1/apiv1connect"
	"github.com/ainslies/ctd/internal/domain"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// CTDClient is a client for communicating with the ctd daemon.
type CTDClient struct {
	thoughtClient    apiv1connect.ThoughtServiceClient
	tokenUsageClient apiv1connect.TokenUsageServiceClient
	timeout          time.Duration
}

// NewCTDClient creates a new CTD client.
func NewCTDClient(serverURL string, timeout time.Duration) *CTDClient {
	httpClient := &http.Client{
		Timeout: timeout,
	}

	return &CTDClient{
		thoughtClient:    apiv1connect.NewThoughtServiceClient(httpClient, serverURL),
		tokenUsageClient: apiv1connect.NewTokenUsageServiceClient(httpClient, serverURL),
		timeout:          timeout,
	}
}

// CreateTokenUsage sends a token usage record to ctd.
func (c *CTDClient) CreateTokenUsage(ctx context.Context, usage *domain.TokenUsage) error {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	protoUsage := &apiv1.TokenUsage{
		ConversationId:   usage.ConversationID.String(),
		GenerationId:     usage.GenerationID.String(),
		Model:            usage.Model,
		Stage:            mapDomainStageToProto(usage.Stage),
		PromptTokens:     usage.PromptTokens,
		CompletionTokens: usage.CompletionTokens,
		TotalTokens:      usage.TotalTokens,
		PromptChars:      usage.PromptChars,
		CompletionChars:  usage.CompletionChars,
		WorkspaceRoots:   usage.WorkspaceRoots,
		UserEmail:        usage.UserEmail,
		HookEventName:    usage.HookEventName,
		Timestamp:        timestamppb.New(usage.Timestamp),
	}

	req := connect.NewRequest(&apiv1.CreateTokenUsageRequest{
		Usage: protoUsage,
	})

	_, err := c.tokenUsageClient.CreateTokenUsage(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to create token usage: %w", err)
	}

	return nil
}

// CreateThought sends a thought record to ctd.
func (c *CTDClient) CreateThought(ctx context.Context, thought *domain.Thought) error {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	protoThought := &apiv1.Thought{
		ConversationId: thought.ConversationID.String(),
		GenerationId:   thought.GenerationID.String(),
		Text:           thought.Text,
		DurationMs:     clampDurationToInt32(thought.DurationMs),
		Model:          thought.Model,
		HookEventName:  thought.HookEventName,
		CursorVersion:  thought.CursorVersion,
		WorkspaceRoots: thought.WorkspaceRoots,
		UserEmail:      thought.UserEmail,
		Timestamp:      timestamppb.New(thought.Timestamp),
	}

	req := connect.NewRequest(&apiv1.CreateThoughtRequest{
		Thought: protoThought,
	})

	_, err := c.thoughtClient.CreateThought(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to create thought: %w", err)
	}

	return nil
}

// mapDomainStageToProto converts a domain TokenUsageStage to a proto TokenUsageStage.
//
//nolint:exhaustive // Unspecified is handled by default case
func mapDomainStageToProto(stage domain.TokenUsageStage) apiv1.TokenUsageStage {
	switch stage {
	case domain.TokenUsageStagePrompt:
		return apiv1.TokenUsageStage_TOKEN_USAGE_STAGE_PROMPT
	case domain.TokenUsageStageCompletion:
		return apiv1.TokenUsageStage_TOKEN_USAGE_STAGE_COMPLETION
	case domain.TokenUsageStageTool:
		return apiv1.TokenUsageStage_TOKEN_USAGE_STAGE_TOOL
	case domain.TokenUsageStageSystem:
		return apiv1.TokenUsageStage_TOKEN_USAGE_STAGE_SYSTEM
	case domain.TokenUsageStageFileEdit:
		return apiv1.TokenUsageStage_TOKEN_USAGE_STAGE_FILE_EDIT
	case domain.TokenUsageStageStop:
		return apiv1.TokenUsageStage_TOKEN_USAGE_STAGE_STOP
	default:
		return apiv1.TokenUsageStage_TOKEN_USAGE_STAGE_UNSPECIFIED
	}
}


```

```go
package hook

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"time"

	"connectrpc.com/connect"
	apiv1 "github.com/ainslies/ctd/gen/api/proto/v1"
	"github.com/ainslies/ctd/gen/api/proto/v1/apiv1connect"
	"github.com/ainslies/ctd/internal/domain"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// clampDurationToInt32 safely converts an int to int32, clamping to int32 bounds.
func clampDurationToInt32(val int) int32 {
	if val > math.MaxInt32 {
		return math.MaxInt32
	}
	if val < math.MinInt32 {
		return math.MinInt32
	}
	return int32(val)
}

// CTDClient is a client for communicating with the ctd daemon.
type CTDClient struct {
	thoughtClient    apiv1connect.ThoughtServiceClient
	tokenUsageClient apiv1connect.TokenUsageServiceClient
	timeout          time.Duration
}

// NewCTDClient creates a new CTD client.
func NewCTDClient(serverURL string, timeout time.Duration) *CTDClient {
	httpClient := &http.Client{
		Timeout: timeout,
	}

	return &CTDClient{
		thoughtClient:    apiv1connect.NewThoughtServiceClient(httpClient, serverURL),
		tokenUsageClient: apiv1connect.NewTokenUsageServiceClient(httpClient, serverURL),
		timeout:          timeout,
	}
}

// CreateTokenUsage sends a token usage record to ctd.
func (c *CTDClient) CreateTokenUsage(ctx context.Context, usage *domain.TokenUsage) error {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	protoUsage := &apiv1.TokenUsage{
		ConversationId:   usage.ConversationID.String(),
		GenerationId:     usage.GenerationID.String(),
		Model:            usage.Model,
		Stage:            mapDomainStageToProto(usage.Stage),
		PromptTokens:     usage.PromptTokens,
		CompletionTokens: usage.CompletionTokens,
		TotalTokens:      usage.TotalTokens,
		PromptChars:      usage.PromptChars,
		CompletionChars:  usage.CompletionChars,
		WorkspaceRoots:   usage.WorkspaceRoots,
		UserEmail:        usage.UserEmail,
		HookEventName:    usage.HookEventName,
		Timestamp:        timestamppb.New(usage.Timestamp),
	}

	req := connect.NewRequest(&apiv1.CreateTokenUsageRequest{
		Usage: protoUsage,
	})

	_, err := c.tokenUsageClient.CreateTokenUsage(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to create token usage: %w", err)
	}

	return nil
}

// CreateThought sends a thought record to ctd.
func (c *CTDClient) CreateThought(ctx context.Context, thought *domain.Thought) error {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	protoThought := &apiv1.Thought{
		ConversationId: thought.ConversationID.String(),
		GenerationId:   thought.GenerationID.String(),
		Text:           thought.Text,
		DurationMs:     clampDurationToInt32(thought.DurationMs),
		Model:          thought.Model,
		HookEventName:  thought.HookEventName,
		CursorVersion:  thought.CursorVersion,
		WorkspaceRoots: thought.WorkspaceRoots,
		UserEmail:      thought.UserEmail,
		Timestamp:      timestamppb.New(thought.Timestamp),
	}

	req := connect.NewRequest(&apiv1.CreateThoughtRequest{
		Thought: protoThought,
	})

	_, err := c.thoughtClient.CreateThought(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to create thought: %w", err)
	}

	return nil
}

// mapDomainStageToProto converts a domain TokenUsageStage to a proto TokenUsageStage.
//
//nolint:exhaustive // Unspecified is handled by default case
func mapDomainStageToProto(stage domain.TokenUsageStage) apiv1.TokenUsageStage {
	switch stage {
	case domain.TokenUsageStagePrompt:
		return apiv1.TokenUsageStage_TOKEN_USAGE_STAGE_PROMPT
	case domain.TokenUsageStageCompletion:
		return apiv1.TokenUsageStage_TOKEN_USAGE_STAGE_COMPLETION
	case domain.TokenUsageStageTool:
		return apiv1.TokenUsageStage_TOKEN_USAGE_STAGE_TOOL
	case domain.TokenUsageStageSystem:
		return apiv1.TokenUsageStage_TOKEN_USAGE_STAGE_SYSTEM
	case domain.TokenUsageStageFileEdit:
		return apiv1.TokenUsageStage_TOKEN_USAGE_STAGE_FILE_EDIT
	case domain.TokenUsageStageStop:
		return apiv1.TokenUsageStage_TOKEN_USAGE_STAGE_STOP
	default:
		return apiv1.TokenUsageStage_TOKEN_USAGE_STAGE_UNSPECIFIED
	}
}


```

```go
package server

import (
	"context"
	"errors"
	"fmt"
	"time"

	"connectrpc.com/connect"
	apiv1 "github.com/ainslies/ctd/gen/api/proto/v1"
	"github.com/ainslies/ctd/internal/domain"
	"github.com/ainslies/ctd/pkg/logger"
	"github.com/google/uuid"
)

// TokenUsageServer implements the TokenUsageService API.
type TokenUsageServer struct {
	repo domain.TokenUsageRepository
	log  *logger.Logger
}

// NewTokenUsageServer creates a new TokenUsageServer.
func NewTokenUsageServer(repo domain.TokenUsageRepository, log *logger.Logger) *TokenUsageServer {
	return &TokenUsageServer{
		repo: repo,
		log:  log,
	}
}

// CreateTokenUsage records a new token usage entry.
func (s *TokenUsageServer) CreateTokenUsage(
	ctx context.Context,
	req *connect.Request[apiv1.CreateTokenUsageRequest],
) (*connect.Response[apiv1.CreateTokenUsageResponse], error) {
	s.log.InfoContext(ctx, "received CreateTokenUsage request")

	if req.Msg.GetUsage() == nil {
		s.log.WarnContext(ctx, "usage is required")
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("usage is required"))
	}

	usage, err := mapProtoToTokenUsage(req.Msg.GetUsage())
	if err != nil {
		s.log.WarnContext(ctx, "invalid token usage data", "error", err)
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid token usage data: %w", err))
	}

	// Validate token counts are non-negative
	if usage.PromptTokens < 0 || usage.CompletionTokens < 0 || usage.TotalTokens < 0 {
		s.log.WarnContext(ctx, "token counts must be non-negative")
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("token counts must be non-negative"))
	}

	s.log.DebugContext(ctx, "mapped token usage",
		"id", usage.ID,
		"conversation_id", usage.ConversationID,
		"stage", usage.Stage,
	)

	if err := s.repo.Store(ctx, usage); err != nil {
		s.log.ErrorContext(ctx, "failed to store token usage", "error", err)
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to store token usage: %w", err))
	}

	s.log.InfoContext(ctx, "successfully created token usage", "id", usage.ID)

	res := connect.NewResponse(&apiv1.CreateTokenUsageResponse{
		Id: usage.ID.String(),
	})
	return res, nil
}

// mapProtoToTokenUsage converts a proto TokenUsage to a domain TokenUsage.
func mapProtoToTokenUsage(protoUsage *apiv1.TokenUsage) (*domain.TokenUsage, error) {
	conversationID, err := uuid.Parse(protoUsage.GetConversationId())
	if err != nil {
		return nil, fmt.Errorf("invalid conversation_id: %w", err)
	}

	generationID, err := uuid.Parse(protoUsage.GetGenerationId())
	if err != nil {
		return nil, fmt.Errorf("invalid generation_id: %w", err)
	}

	// Default timestamp to now if missing or zero
	timestamp := time.Now()
	if protoUsage.GetTimestamp() != nil {
		timestamp = protoUsage.GetTimestamp().AsTime()
	}

	stage := mapProtoStageToTokenUsageStage(protoUsage.GetStage())

	return &domain.TokenUsage{
		ID:               uuid.New(),
		ConversationID:   conversationID,
		GenerationID:     generationID,
		Model:            protoUsage.GetModel(),
		Stage:            stage,
		PromptTokens:     protoUsage.GetPromptTokens(),
		CompletionTokens: protoUsage.GetCompletionTokens(),
		TotalTokens:      protoUsage.GetTotalTokens(),
		PromptChars:      protoUsage.GetPromptChars(),
		CompletionChars:  protoUsage.GetCompletionChars(),
		WorkspaceRoots:   protoUsage.GetWorkspaceRoots(),
		UserEmail:        protoUsage.GetUserEmail(),
		HookEventName:    protoUsage.GetHookEventName(),
		Timestamp:        timestamp,
	}, nil
}

// mapProtoStageToTokenUsageStage converts a proto TokenUsageStage to a domain TokenUsageStage.
//
//nolint:exhaustive // Unspecified is handled by default case
func mapProtoStageToTokenUsageStage(protoStage apiv1.TokenUsageStage) domain.TokenUsageStage {
	switch protoStage {
	case apiv1.TokenUsageStage_TOKEN_USAGE_STAGE_PROMPT:
		return domain.TokenUsageStagePrompt
	case apiv1.TokenUsageStage_TOKEN_USAGE_STAGE_COMPLETION:
		return domain.TokenUsageStageCompletion
	case apiv1.TokenUsageStage_TOKEN_USAGE_STAGE_TOOL:
		return domain.TokenUsageStageTool
	case apiv1.TokenUsageStage_TOKEN_USAGE_STAGE_SYSTEM:
		return domain.TokenUsageStageSystem
	case apiv1.TokenUsageStage_TOKEN_USAGE_STAGE_FILE_EDIT:
		return domain.TokenUsageStageFileEdit
	case apiv1.TokenUsageStage_TOKEN_USAGE_STAGE_STOP:
		return domain.TokenUsageStageStop
	default:
		return domain.TokenUsageStageUnspecified
	}
}

// mapDomainStageToProtoStage converts a domain TokenUsageStage to a proto TokenUsageStage.
//
//nolint:exhaustive // Unspecified is handled by default case
func mapDomainStageToProtoStage(stage domain.TokenUsageStage) apiv1.TokenUsageStage {
	switch stage {
	case domain.TokenUsageStagePrompt:
		return apiv1.TokenUsageStage_TOKEN_USAGE_STAGE_PROMPT
	case domain.TokenUsageStageCompletion:
		return apiv1.TokenUsageStage_TOKEN_USAGE_STAGE_COMPLETION
	case domain.TokenUsageStageTool:
		return apiv1.TokenUsageStage_TOKEN_USAGE_STAGE_TOOL
	case domain.TokenUsageStageSystem:
		return apiv1.TokenUsageStage_TOKEN_USAGE_STAGE_SYSTEM
	case domain.TokenUsageStageFileEdit:
		return apiv1.TokenUsageStage_TOKEN_USAGE_STAGE_FILE_EDIT
	case domain.TokenUsageStageStop:
		return apiv1.TokenUsageStage_TOKEN_USAGE_STAGE_STOP
	default:
		return apiv1.TokenUsageStage_TOKEN_USAGE_STAGE_UNSPECIFIED
	}
}


```

Now let me fix the tests:

```go
package hook

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig_Defaults(t *testing.T) {
	cfg, err := LoadConfig()

	require.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, DefaultServerURL, cfg.ServerURL)
	assert.Equal(t, DefaultLogLevel, cfg.LogLevel)
	assert.Equal(t, DefaultModel, cfg.DefaultModel)
	assert.Equal(t, DefaultTimeout, cfg.TimeoutMs)
}

func TestConfig_Validate_Defaults(t *testing.T) {
	cfg := &Config{}

	err := cfg.Validate()

	require.NoError(t, err)
	assert.Equal(t, DefaultServerURL, cfg.ServerURL)
	assert.Equal(t, DefaultLogLevel, cfg.LogLevel)
	assert.Equal(t, DefaultModel, cfg.DefaultModel)
	assert.Equal(t, DefaultTimeout, cfg.TimeoutMs)
}

func TestConfig_Validate_PreservesValidValues(t *testing.T) {
	cfg := &Config{
		ServerURL:    "http://localhost:8080",
		LogLevel:     "debug",
		DefaultModel: "claude-3-opus",
		TimeoutMs:    1000,
	}

	err := cfg.Validate()

	require.NoError(t, err)
	assert.Equal(t, "http://localhost:8080", cfg.ServerURL)
	assert.Equal(t, "debug", cfg.LogLevel)
	assert.Equal(t, "claude-3-opus", cfg.DefaultModel)
	assert.Equal(t, 1000, cfg.TimeoutMs)
}

func TestConfig_Validate_FixesInvalidTimeout(t *testing.T) {
	cfg := &Config{
		ServerURL:    "http://localhost:8080",
		LogLevel:     "info",
		DefaultModel: "gpt-4",
		TimeoutMs:    -100,
	}

	err := cfg.Validate()

	require.NoError(t, err)
	assert.Equal(t, DefaultTimeout, cfg.TimeoutMs)
}


```

```go
package hook

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHeuristicTokenizer_CountTokens(t *testing.T) {
	tokenizer := NewHeuristicTokenizer()

	tests := []struct {
		name     string
		text     string
		expected int64
	}{
		{"empty string", "", 0},
		{"short text", "hi", 1},
		{"exact multiple", "hello world", 3},    // 11 chars / 4 = 2.75, rounded up to 3
		{"longer text", "hello world test", 4},  // 16 chars / 4 = 4
		{"with spaces", "a b c d e f g h", 4},   // 15 chars / 4 = 3.75, rounded up to 4
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count, err := tokenizer.CountTokens("", tt.text)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, count)
		})
	}
}

func TestHeuristicTokenizer_ZeroCharsPerToken(t *testing.T) {
	tokenizer := &HeuristicTokenizer{CharsPerToken: 0}

	count, err := tokenizer.CountTokens("", "test")

	require.NoError(t, err)
	assert.Equal(t, int64(1), count) // Falls back to 4 chars per token
}

func TestTiktokenTokenizer_CountTokens(t *testing.T) {
	tokenizer := NewTiktokenTokenizer("gpt-4")

	// Test with a simple string
	count, err := tokenizer.CountTokens("gpt-4", "Hello, world!")

	require.NoError(t, err)
	assert.Positive(t, count)
}

func TestTiktokenTokenizer_EmptyModel(t *testing.T) {
	tokenizer := NewTiktokenTokenizer("gpt-4")

	// Should use default model when empty
	count, err := tokenizer.CountTokens("", "Hello, world!")

	require.NoError(t, err)
	assert.Positive(t, count)
}

func TestCompositeTokenizer_CountTokens(t *testing.T) {
	tokenizer := NewCompositeTokenizer("gpt-4")

	count, err := tokenizer.CountTokens("gpt-4", "Hello, world!")

	require.NoError(t, err)
	assert.Positive(t, count)
}

func TestMapModelToEncoding(t *testing.T) {
	tests := []struct {
		name     string
		model    string
		expected string
	}{
		{"gpt-4", "gpt-4", "cl100k_base"},
		{"gpt-4-turbo", "gpt-4-turbo", "cl100k_base"},
		{"gpt-3.5-turbo", "gpt-3.5-turbo", "cl100k_base"},
		{"claude", "claude-3-opus", "cl100k_base"},
		{"gemini", "gemini-pro", "cl100k_base"},
		{"gpt-4o", "gpt-4o", "o200k_base"},
		{"o1", "o1-preview", "o200k_base"},
		{"unknown", "some-unknown-model", "cl100k_base"},
		{"uppercase", "GPT-4", "cl100k_base"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mapModelToEncoding(tt.model)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCountChars(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected int64
	}{
		{"empty", "", 0},
		{"single char", "a", 1},
		{"multi char", "hello", 5},
		{"with spaces", "hello world", 11},
		{"unicode", "hello", 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CountChars(tt.text)
			assert.Equal(t, tt.expected, result)
		})
	}
}


```

Now fix the main.go wrapcheck issues:

```go
// Package main provides the entry point for the Cursor hook binary.
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/ainslies/ctd/internal/hook"
	"github.com/ainslies/ctd/pkg/logger"
)

func main() {
	// Always exit 0 to avoid blocking Cursor
	// Errors are logged but don't prevent agent from proceeding
	if err := run(); err != nil {
		// Log to stderr (stdout is reserved for hook output)
		fmt.Fprintf(os.Stderr, "ctd-hook error: %v\n", err)
	}
	os.Exit(0)
}

func run() error {
	// Initialize logger to stderr
	log := logger.NewWithOutput(os.Stderr)

	// Load configuration
	cfg, err := hook.LoadConfig()
	if err != nil {
		log.Error("failed to load config", "error", err)
		// Use defaults on config error
		cfg = &hook.Config{
			ServerURL:    hook.DefaultServerURL,
			LogLevel:     hook.DefaultLogLevel,
			DefaultModel: hook.DefaultModel,
			TimeoutMs:    hook.DefaultTimeout,
		}
	}

	if err := cfg.Validate(); err != nil {
		log.Error("invalid config", "error", err)
	}

	log.Debug("hook started",
		"server_url", cfg.ServerURL,
		"default_model", cfg.DefaultModel,
		"timeout_ms", cfg.TimeoutMs,
	)

	// Parse input from stdin
	input, err := hook.ParseInput(os.Stdin)
	if err != nil {
		log.Error("failed to parse input", "error", err)
		// Write default output and exit
		return writeDefaultOutput(os.Stdout)
	}

	log.Debug("received hook event",
		"event", input.HookEventName,
		"conversation_id", input.ConversationID,
		"generation_id", input.GenerationID,
	)

	// Create client with timeout
	timeout := time.Duration(cfg.TimeoutMs) * time.Millisecond
	client := hook.NewCTDClient(cfg.ServerURL, timeout)

	// Create tokenizer
	tokenizer := hook.NewCompositeTokenizer(cfg.DefaultModel)

	// Create handler dependencies
	deps := &hook.HandlerDependencies{
		Client:    client,
		Tokenizer: tokenizer,
		Log:       log,
		Config:    cfg,
	}

	// Create router and handle the event
	router := hook.NewRouter(deps)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	output, err := router.Route(ctx, input)
	if err != nil {
		log.Error("handler error", "error", err, "event", input.HookEventName)
		// Return default output on handler error
		return writeDefaultOutput(os.Stdout)
	}

	// Write output to stdout
	if err := hook.WriteOutput(os.Stdout, output); err != nil {
		log.Error("failed to write output", "error", err)
		return fmt.Errorf("failed to write output: %w", err)
	}

	log.Debug("hook completed", "event", input.HookEventName)
	return nil
}

// writeDefaultOutput writes a safe default output for unknown or error cases.
func writeDefaultOutput(w *os.File) error {
	// Default to allowing operations to avoid blocking the agent
	output := hook.NewAllowOutput()
	return hook.WriteOutput(w, output)
}


```

```go
// Package main provides the entry point for the Cursor hook binary.
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/ainslies/ctd/internal/hook"
	"github.com/ainslies/ctd/pkg/logger"
)

func main() {
	// Always exit 0 to avoid blocking Cursor
	// Errors are logged but don't prevent agent from proceeding
	if err := run(); err != nil {
		// Log to stderr (stdout is reserved for hook output)
		fmt.Fprintf(os.Stderr, "ctd-hook error: %v\n", err)
	}
	os.Exit(0)
}

func run() error {
	// Initialize logger to stderr
	log := logger.NewWithOutput(os.Stderr)

	// Load configuration
	cfg, err := hook.LoadConfig()
	if err != nil {
		log.Error("failed to load config", "error", err)
		// Use defaults on config error
		cfg = &hook.Config{
			ServerURL:    hook.DefaultServerURL,
			LogLevel:     hook.DefaultLogLevel,
			DefaultModel: hook.DefaultModel,
			TimeoutMs:    hook.DefaultTimeout,
		}
	}

	if err := cfg.Validate(); err != nil {
		log.Error("invalid config", "error", err)
	}

	log.Debug("hook started",
		"server_url", cfg.ServerURL,
		"default_model", cfg.DefaultModel,
		"timeout_ms", cfg.TimeoutMs,
	)

	// Parse input from stdin
	input, err := hook.ParseInput(os.Stdin)
	if err != nil {
		log.Error("failed to parse input", "error", err)
		// Write default output and exit
		return writeDefaultOutput(os.Stdout)
	}

	log.Debug("received hook event",
		"event", input.HookEventName,
		"conversation_id", input.ConversationID,
		"generation_id", input.GenerationID,
	)

	// Create client with timeout
	timeout := time.Duration(cfg.TimeoutMs) * time.Millisecond
	client := hook.NewCTDClient(cfg.ServerURL, timeout)

	// Create tokenizer
	tokenizer := hook.NewCompositeTokenizer(cfg.DefaultModel)

	// Create handler dependencies
	deps := &hook.HandlerDependencies{
		Client:    client,
		Tokenizer: tokenizer,
		Log:       log,
		Config:    cfg,
	}

	// Create router and handle the event
	router := hook.NewRouter(deps)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	output, err := router.Route(ctx, input)
	if err != nil {
		log.Error("handler error", "error", err, "event", input.HookEventName)
		// Return default output on handler error
		return writeDefaultOutput(os.Stdout)
	}

	// Write output to stdout
	if err := hook.WriteOutput(os.Stdout, output); err != nil {
		log.Error("failed to write output", "error", err)
		return fmt.Errorf("failed to write output: %w", err)
	}

	log.Debug("hook completed", "event", input.HookEventName)
	return nil
}

// writeDefaultOutput writes a safe default output for unknown or error cases.
func writeDefaultOutput(w *os.File) error {
	// Default to allowing operations to avoid blocking the agent
	output := hook.NewAllowOutput()
	if err := hook.WriteOutput(w, output); err != nil {
		return fmt.Errorf("failed to write default output: %w", err)
	}
	return nil
}


```

The nolint directives are no longer needed. Let me remove them:

```go
package hook

import (
	"context"
	"fmt"
	"time"

	"github.com/ainslies/ctd/internal/domain"
	"github.com/ainslies/ctd/pkg/logger"
	"github.com/google/uuid"
)

// Handler processes a hook event and returns the appropriate output.
type Handler interface {
	// Handle processes the hook input and returns the output.
	Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error)
}

// HandlerDependencies contains the dependencies for handlers.
type HandlerDependencies struct {
	Client    *CTDClient
	Tokenizer Tokenizer
	Log       *logger.Logger
	Config    *Config
}

// Router routes hook events to the appropriate handler.
type Router struct {
	deps     *HandlerDependencies
	handlers map[HookEventName]Handler
}

// NewRouter creates a new router with all handlers registered.
func NewRouter(deps *HandlerDependencies) *Router {
	router := &Router{
		deps:     deps,
		handlers: make(map[HookEventName]Handler),
	}

	// Register handlers
	router.handlers[EventBeforeSubmitPrompt] = &BeforeSubmitPromptHandler{deps: deps}
	router.handlers[EventAfterAgentResponse] = &AfterAgentResponseHandler{deps: deps}
	router.handlers[EventAfterAgentThought] = &AfterAgentThoughtHandler{deps: deps}
	router.handlers[EventAfterMCPExecution] = &AfterMCPExecutionHandler{deps: deps}
	router.handlers[EventAfterFileEdit] = &AfterFileEditHandler{deps: deps}
	router.handlers[EventStop] = &StopHandler{deps: deps}
	router.handlers[EventBeforeShellExecution] = &BeforeShellExecutionHandler{deps: deps}
	router.handlers[EventAfterShellExecution] = &AfterShellExecutionHandler{deps: deps}
	router.handlers[EventBeforeMCPExecution] = &BeforeMCPExecutionHandler{deps: deps}
	router.handlers[EventBeforeReadFile] = &BeforeReadFileHandler{deps: deps}

	return router
}

// Route routes the input to the appropriate handler.
func (r *Router) Route(ctx context.Context, input *HookInput) (*HookOutput, error) {
	handler, ok := r.handlers[input.HookEventName]
	if !ok {
		// Unknown event, return empty output
		r.deps.Log.Warn("unknown hook event", "event", input.HookEventName)
		return NewEmptyOutput(), nil
	}

	payload, err := input.ParsePayload()
	if err != nil {
		r.deps.Log.Error("failed to parse payload", "error", err, "event", input.HookEventName)
		// Return appropriate default output even on parse error
		return getDefaultOutput(input.HookEventName), nil
	}

	output, err := handler.Handle(ctx, input, payload)
	if err != nil {
		return nil, fmt.Errorf("handler error for %s: %w", input.HookEventName, err)
	}
	return output, nil
}

// getDefaultOutput returns the default output for an event type.
//
//nolint:exhaustive // Only specific events need special handling, others use empty output
func getDefaultOutput(event HookEventName) *HookOutput {
	switch event {
	case EventBeforeSubmitPrompt:
		return NewContinueOutput()
	case EventBeforeShellExecution, EventBeforeMCPExecution, EventBeforeReadFile:
		return NewAllowOutput()
	default:
		return NewEmptyOutput()
	}
}

// BeforeSubmitPromptHandler handles beforeSubmitPrompt events.
type BeforeSubmitPromptHandler struct {
	deps *HandlerDependencies
}

// Handle processes the beforeSubmitPrompt event.
func (h *BeforeSubmitPromptHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	promptPayload, ok := payload.(*BeforeSubmitPromptPayload)
	if !ok {
		return NewContinueOutput(), nil
	}

	// Count tokens
	tokenCount, err := h.deps.Tokenizer.CountTokens(h.deps.Config.DefaultModel, promptPayload.Prompt)
	if err != nil {
		h.deps.Log.Error("failed to count tokens", "error", err)
		tokenCount = 0
	}

	charCount := CountChars(promptPayload.Prompt)

	h.deps.Log.Debug("prompt metrics",
		"tokens", tokenCount,
		"chars", charCount,
		"conversation_id", input.ConversationID,
	)

	// Create token usage record
	usage := &domain.TokenUsage{
		ID:               uuid.New(),
		ConversationID:   parseUUID(input.ConversationID),
		GenerationID:     parseUUID(input.GenerationID),
		Model:            h.deps.Config.DefaultModel,
		Stage:            domain.TokenUsageStagePrompt,
		PromptTokens:     tokenCount,
		CompletionTokens: 0,
		TotalTokens:      tokenCount,
		PromptChars:      charCount,
		CompletionChars:  0,
		WorkspaceRoots:   input.WorkspaceRoots,
		UserEmail:        "",
		HookEventName:    string(input.HookEventName),
		Timestamp:        time.Now(),
	}

	// Send to ctd (best-effort)
	if err := h.deps.Client.CreateTokenUsage(ctx, usage); err != nil {
		h.deps.Log.Error("failed to send token usage", "error", err)
	} else {
		h.deps.Log.Info("token usage recorded", "stage", "prompt", "tokens", tokenCount)
	}

	return NewContinueOutput(), nil
}

// AfterAgentResponseHandler handles afterAgentResponse events.
type AfterAgentResponseHandler struct {
	deps *HandlerDependencies
}

// Handle processes the afterAgentResponse event.
//
//nolint:varnamelen // responsePayload is clear in context
func (h *AfterAgentResponseHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	responsePayload, ok := payload.(*AfterAgentResponsePayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	// Count tokens
	tokenCount, err := h.deps.Tokenizer.CountTokens(h.deps.Config.DefaultModel, responsePayload.Text)
	if err != nil {
		h.deps.Log.Error("failed to count tokens", "error", err)
		tokenCount = 0
	}

	charCount := CountChars(responsePayload.Text)

	h.deps.Log.Debug("completion metrics",
		"tokens", tokenCount,
		"chars", charCount,
		"conversation_id", input.ConversationID,
	)

	// Create token usage record
	usage := &domain.TokenUsage{
		ID:               uuid.New(),
		ConversationID:   parseUUID(input.ConversationID),
		GenerationID:     parseUUID(input.GenerationID),
		Model:            h.deps.Config.DefaultModel,
		Stage:            domain.TokenUsageStageCompletion,
		PromptTokens:     0,
		CompletionTokens: tokenCount,
		TotalTokens:      tokenCount,
		PromptChars:      0,
		CompletionChars:  charCount,
		WorkspaceRoots:   input.WorkspaceRoots,
		UserEmail:        "",
		HookEventName:    string(input.HookEventName),
		Timestamp:        time.Now(),
	}

	// Send token usage to ctd (best-effort)
	if err := h.deps.Client.CreateTokenUsage(ctx, usage); err != nil {
		h.deps.Log.Error("failed to send token usage", "error", err)
	} else {
		h.deps.Log.Info("token usage recorded", "stage", "completion", "tokens", tokenCount)
	}

	// Also send as thought for visibility
	thought := &domain.Thought{
		ID:             uuid.New(),
		ConversationID: parseUUID(input.ConversationID),
		GenerationID:   parseUUID(input.GenerationID),
		Text:           responsePayload.Text,
		DurationMs:     0,
		Model:          h.deps.Config.DefaultModel,
		HookEventName:  string(input.HookEventName),
		CursorVersion:  "",
		WorkspaceRoots: input.WorkspaceRoots,
		UserEmail:      "",
		Timestamp:      time.Now(),
	}

	if err := h.deps.Client.CreateThought(ctx, thought); err != nil {
		h.deps.Log.Error("failed to send thought", "error", err)
	}

	return NewEmptyOutput(), nil
}

// AfterAgentThoughtHandler handles afterAgentThought events.
type AfterAgentThoughtHandler struct {
	deps *HandlerDependencies
}

// Handle processes the afterAgentThought event.
//
//nolint:varnamelen // thoughtPayload is clear in context
func (h *AfterAgentThoughtHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	thoughtPayload, ok := payload.(*AfterAgentThoughtPayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	// Send as thought for visibility
	thought := &domain.Thought{
		ID:             uuid.New(),
		ConversationID: parseUUID(input.ConversationID),
		GenerationID:   parseUUID(input.GenerationID),
		Text:           thoughtPayload.Text,
		DurationMs:     0,
		Model:          h.deps.Config.DefaultModel,
		HookEventName:  string(input.HookEventName),
		CursorVersion:  "",
		WorkspaceRoots: input.WorkspaceRoots,
		UserEmail:      "",
		Timestamp:      time.Now(),
	}

	if err := h.deps.Client.CreateThought(ctx, thought); err != nil {
		h.deps.Log.Error("failed to send thought", "error", err)
	} else {
		h.deps.Log.Info("thought recorded", "event", input.HookEventName)
	}

	return NewEmptyOutput(), nil
}

// AfterMCPExecutionHandler handles afterMCPExecution events.
type AfterMCPExecutionHandler struct {
	deps *HandlerDependencies
}

// Handle processes the afterMCPExecution event.
//
//nolint:varnamelen // mcpPayload is clear in context
func (h *AfterMCPExecutionHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	mcpPayload, ok := payload.(*AfterMCPExecutionPayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	// Count tokens in tool input and result (best effort, ignore errors)
	inputTokens, inputErr := h.deps.Tokenizer.CountTokens(h.deps.Config.DefaultModel, string(mcpPayload.ToolInput))
	if inputErr != nil {
		h.deps.Log.Debug("failed to count input tokens", "error", inputErr)
		inputTokens = 0
	}
	resultTokens, resultErr := h.deps.Tokenizer.CountTokens(h.deps.Config.DefaultModel, string(mcpPayload.ResultJSON))
	if resultErr != nil {
		h.deps.Log.Debug("failed to count result tokens", "error", resultErr)
		resultTokens = 0
	}
	totalTokens := inputTokens + resultTokens

	inputChars := CountChars(string(mcpPayload.ToolInput))
	resultChars := CountChars(string(mcpPayload.ResultJSON))

	h.deps.Log.Debug("tool metrics",
		"tool", mcpPayload.ToolName,
		"input_tokens", inputTokens,
		"result_tokens", resultTokens,
		"conversation_id", input.ConversationID,
	)

	// Create token usage record
	usage := &domain.TokenUsage{
		ID:               uuid.New(),
		ConversationID:   parseUUID(input.ConversationID),
		GenerationID:     parseUUID(input.GenerationID),
		Model:            h.deps.Config.DefaultModel,
		Stage:            domain.TokenUsageStageTool,
		PromptTokens:     inputTokens,
		CompletionTokens: resultTokens,
		TotalTokens:      totalTokens,
		PromptChars:      inputChars,
		CompletionChars:  resultChars,
		WorkspaceRoots:   input.WorkspaceRoots,
		UserEmail:        "",
		HookEventName:    string(input.HookEventName),
		Timestamp:        time.Now(),
	}

	if err := h.deps.Client.CreateTokenUsage(ctx, usage); err != nil {
		h.deps.Log.Error("failed to send token usage", "error", err)
	} else {
		h.deps.Log.Info("token usage recorded", "stage", "tool", "tool", mcpPayload.ToolName, "tokens", totalTokens)
	}

	return NewEmptyOutput(), nil
}

// AfterFileEditHandler handles afterFileEdit events.
type AfterFileEditHandler struct {
	deps *HandlerDependencies
}

// Handle processes the afterFileEdit event.
//
//nolint:varnamelen // editPayload is clear in context
func (h *AfterFileEditHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	editPayload, ok := payload.(*AfterFileEditPayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	// Count characters in edits
	var oldChars, newChars int64
	for _, edit := range editPayload.Edits {
		oldChars += CountChars(edit.OldString)
		newChars += CountChars(edit.NewString)
	}

	h.deps.Log.Debug("file edit metrics",
		"file", editPayload.FilePath,
		"old_chars", oldChars,
		"new_chars", newChars,
		"edit_count", len(editPayload.Edits),
	)

	// Create token usage record for file edits
	usage := &domain.TokenUsage{
		ID:               uuid.New(),
		ConversationID:   parseUUID(input.ConversationID),
		GenerationID:     parseUUID(input.GenerationID),
		Model:            h.deps.Config.DefaultModel,
		Stage:            domain.TokenUsageStageFileEdit,
		PromptTokens:     0,
		CompletionTokens: 0,
		TotalTokens:      0,
		PromptChars:      oldChars,
		CompletionChars:  newChars,
		WorkspaceRoots:   input.WorkspaceRoots,
		UserEmail:        "",
		HookEventName:    string(input.HookEventName),
		Timestamp:        time.Now(),
	}

	if err := h.deps.Client.CreateTokenUsage(ctx, usage); err != nil {
		h.deps.Log.Error("failed to send token usage", "error", err)
	} else {
		h.deps.Log.Info("file edit recorded", "file", editPayload.FilePath, "edits", len(editPayload.Edits))
	}

	return NewEmptyOutput(), nil
}

// StopHandler handles stop events.
type StopHandler struct {
	deps *HandlerDependencies
}

// Handle processes the stop event.
//
//nolint:varnamelen // stopPayload is clear in context
func (h *StopHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	stopPayload, ok := payload.(*StopPayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	h.deps.Log.Info("agent loop stopped",
		"status", stopPayload.Status,
		"loop_count", stopPayload.LoopCount,
		"conversation_id", input.ConversationID,
	)

	// Create token usage record as session marker
	usage := &domain.TokenUsage{
		ID:               uuid.New(),
		ConversationID:   parseUUID(input.ConversationID),
		GenerationID:     parseUUID(input.GenerationID),
		Model:            h.deps.Config.DefaultModel,
		Stage:            domain.TokenUsageStageStop,
		PromptTokens:     0,
		CompletionTokens: 0,
		TotalTokens:      0,
		PromptChars:      0,
		CompletionChars:  0,
		WorkspaceRoots:   input.WorkspaceRoots,
		UserEmail:        "",
		HookEventName:    string(input.HookEventName),
		Timestamp:        time.Now(),
	}

	if err := h.deps.Client.CreateTokenUsage(ctx, usage); err != nil {
		h.deps.Log.Error("failed to send stop marker", "error", err)
	}

	return NewEmptyOutput(), nil
}

// BeforeShellExecutionHandler handles beforeShellExecution events.
type BeforeShellExecutionHandler struct {
	deps *HandlerDependencies
}

// Handle processes the beforeShellExecution event.
//
//nolint:varnamelen // shellPayload is clear in context
func (h *BeforeShellExecutionHandler) Handle(_ context.Context, input *HookInput, payload any) (*HookOutput, error) {
	shellPayload, ok := payload.(*BeforeShellExecutionPayload)
	if !ok {
		return NewAllowOutput(), nil
	}

	h.deps.Log.Debug("shell execution",
		"command", shellPayload.Command,
		"cwd", shellPayload.Cwd,
		"conversation_id", input.ConversationID,
	)

	// Always allow shell execution (observability only)
	return NewAllowOutput(), nil
}

// AfterShellExecutionHandler handles afterShellExecution events.
type AfterShellExecutionHandler struct {
	deps *HandlerDependencies
}

// Handle processes the afterShellExecution event.
//
//nolint:varnamelen // shellPayload is clear in context
func (h *AfterShellExecutionHandler) Handle(_ context.Context, input *HookInput, payload any) (*HookOutput, error) {
	shellPayload, ok := payload.(*AfterShellExecutionPayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	h.deps.Log.Debug("shell execution completed",
		"command", shellPayload.Command,
		"output_length", len(shellPayload.Output),
		"conversation_id", input.ConversationID,
	)

	return NewEmptyOutput(), nil
}

// BeforeMCPExecutionHandler handles beforeMCPExecution events.
type BeforeMCPExecutionHandler struct {
	deps *HandlerDependencies
}

// Handle processes the beforeMCPExecution event.
//
//nolint:varnamelen // mcpPayload is clear in context
func (h *BeforeMCPExecutionHandler) Handle(_ context.Context, input *HookInput, payload any) (*HookOutput, error) {
	mcpPayload, ok := payload.(*BeforeMCPExecutionPayload)
	if !ok {
		return NewAllowOutput(), nil
	}

	h.deps.Log.Debug("MCP execution",
		"tool", mcpPayload.ToolName,
		"conversation_id", input.ConversationID,
	)

	// Always allow MCP execution (observability only)
	return NewAllowOutput(), nil
}

// BeforeReadFileHandler handles beforeReadFile events.
type BeforeReadFileHandler struct {
	deps *HandlerDependencies
}

// Handle processes the beforeReadFile event.
//
//nolint:varnamelen // readPayload is clear in context
func (h *BeforeReadFileHandler) Handle(_ context.Context, input *HookInput, payload any) (*HookOutput, error) {
	readPayload, ok := payload.(*BeforeReadFilePayload)
	if !ok {
		return NewAllowOutput(), nil
	}

	h.deps.Log.Debug("file read",
		"file", readPayload.FilePath,
		"content_length", len(readPayload.Content),
		"conversation_id", input.ConversationID,
	)

	// Always allow file read (observability only)
	return NewAllowOutput(), nil
}

// parseUUID parses a UUID string, returning a zero UUID on error.
func parseUUID(s string) uuid.UUID {
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil
	}
	return id
}


```

```go
package hook

import (
	"context"
	"fmt"
	"time"

	"github.com/ainslies/ctd/internal/domain"
	"github.com/ainslies/ctd/pkg/logger"
	"github.com/google/uuid"
)

// Handler processes a hook event and returns the appropriate output.
type Handler interface {
	// Handle processes the hook input and returns the output.
	Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error)
}

// HandlerDependencies contains the dependencies for handlers.
type HandlerDependencies struct {
	Client    *CTDClient
	Tokenizer Tokenizer
	Log       *logger.Logger
	Config    *Config
}

// Router routes hook events to the appropriate handler.
type Router struct {
	deps     *HandlerDependencies
	handlers map[HookEventName]Handler
}

// NewRouter creates a new router with all handlers registered.
func NewRouter(deps *HandlerDependencies) *Router {
	router := &Router{
		deps:     deps,
		handlers: make(map[HookEventName]Handler),
	}

	// Register handlers
	router.handlers[EventBeforeSubmitPrompt] = &BeforeSubmitPromptHandler{deps: deps}
	router.handlers[EventAfterAgentResponse] = &AfterAgentResponseHandler{deps: deps}
	router.handlers[EventAfterAgentThought] = &AfterAgentThoughtHandler{deps: deps}
	router.handlers[EventAfterMCPExecution] = &AfterMCPExecutionHandler{deps: deps}
	router.handlers[EventAfterFileEdit] = &AfterFileEditHandler{deps: deps}
	router.handlers[EventStop] = &StopHandler{deps: deps}
	router.handlers[EventBeforeShellExecution] = &BeforeShellExecutionHandler{deps: deps}
	router.handlers[EventAfterShellExecution] = &AfterShellExecutionHandler{deps: deps}
	router.handlers[EventBeforeMCPExecution] = &BeforeMCPExecutionHandler{deps: deps}
	router.handlers[EventBeforeReadFile] = &BeforeReadFileHandler{deps: deps}

	return router
}

// Route routes the input to the appropriate handler.
func (r *Router) Route(ctx context.Context, input *HookInput) (*HookOutput, error) {
	handler, ok := r.handlers[input.HookEventName]
	if !ok {
		// Unknown event, return empty output
		r.deps.Log.Warn("unknown hook event", "event", input.HookEventName)
		return NewEmptyOutput(), nil
	}

	payload, err := input.ParsePayload()
	if err != nil {
		r.deps.Log.Error("failed to parse payload", "error", err, "event", input.HookEventName)
		// Return appropriate default output even on parse error
		return getDefaultOutput(input.HookEventName), nil
	}

	output, err := handler.Handle(ctx, input, payload)
	if err != nil {
		return nil, fmt.Errorf("handler error for %s: %w", input.HookEventName, err)
	}
	return output, nil
}

// getDefaultOutput returns the default output for an event type.
//
//nolint:exhaustive // Only specific events need special handling, others use empty output
func getDefaultOutput(event HookEventName) *HookOutput {
	switch event {
	case EventBeforeSubmitPrompt:
		return NewContinueOutput()
	case EventBeforeShellExecution, EventBeforeMCPExecution, EventBeforeReadFile:
		return NewAllowOutput()
	default:
		return NewEmptyOutput()
	}
}

// BeforeSubmitPromptHandler handles beforeSubmitPrompt events.
type BeforeSubmitPromptHandler struct {
	deps *HandlerDependencies
}

// Handle processes the beforeSubmitPrompt event.
func (h *BeforeSubmitPromptHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	promptPayload, ok := payload.(*BeforeSubmitPromptPayload)
	if !ok {
		return NewContinueOutput(), nil
	}

	// Count tokens
	tokenCount, err := h.deps.Tokenizer.CountTokens(h.deps.Config.DefaultModel, promptPayload.Prompt)
	if err != nil {
		h.deps.Log.Error("failed to count tokens", "error", err)
		tokenCount = 0
	}

	charCount := CountChars(promptPayload.Prompt)

	h.deps.Log.Debug("prompt metrics",
		"tokens", tokenCount,
		"chars", charCount,
		"conversation_id", input.ConversationID,
	)

	// Create token usage record
	usage := &domain.TokenUsage{
		ID:               uuid.New(),
		ConversationID:   parseUUID(input.ConversationID),
		GenerationID:     parseUUID(input.GenerationID),
		Model:            h.deps.Config.DefaultModel,
		Stage:            domain.TokenUsageStagePrompt,
		PromptTokens:     tokenCount,
		CompletionTokens: 0,
		TotalTokens:      tokenCount,
		PromptChars:      charCount,
		CompletionChars:  0,
		WorkspaceRoots:   input.WorkspaceRoots,
		UserEmail:        "",
		HookEventName:    string(input.HookEventName),
		Timestamp:        time.Now(),
	}

	// Send to ctd (best-effort)
	if err := h.deps.Client.CreateTokenUsage(ctx, usage); err != nil {
		h.deps.Log.Error("failed to send token usage", "error", err)
	} else {
		h.deps.Log.Info("token usage recorded", "stage", "prompt", "tokens", tokenCount)
	}

	return NewContinueOutput(), nil
}

// AfterAgentResponseHandler handles afterAgentResponse events.
type AfterAgentResponseHandler struct {
	deps *HandlerDependencies
}

// Handle processes the afterAgentResponse event.
func (h *AfterAgentResponseHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	responsePayload, ok := payload.(*AfterAgentResponsePayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	// Count tokens
	tokenCount, err := h.deps.Tokenizer.CountTokens(h.deps.Config.DefaultModel, responsePayload.Text)
	if err != nil {
		h.deps.Log.Error("failed to count tokens", "error", err)
		tokenCount = 0
	}

	charCount := CountChars(responsePayload.Text)

	h.deps.Log.Debug("completion metrics",
		"tokens", tokenCount,
		"chars", charCount,
		"conversation_id", input.ConversationID,
	)

	// Create token usage record
	usage := &domain.TokenUsage{
		ID:               uuid.New(),
		ConversationID:   parseUUID(input.ConversationID),
		GenerationID:     parseUUID(input.GenerationID),
		Model:            h.deps.Config.DefaultModel,
		Stage:            domain.TokenUsageStageCompletion,
		PromptTokens:     0,
		CompletionTokens: tokenCount,
		TotalTokens:      tokenCount,
		PromptChars:      0,
		CompletionChars:  charCount,
		WorkspaceRoots:   input.WorkspaceRoots,
		UserEmail:        "",
		HookEventName:    string(input.HookEventName),
		Timestamp:        time.Now(),
	}

	// Send token usage to ctd (best-effort)
	if err := h.deps.Client.CreateTokenUsage(ctx, usage); err != nil {
		h.deps.Log.Error("failed to send token usage", "error", err)
	} else {
		h.deps.Log.Info("token usage recorded", "stage", "completion", "tokens", tokenCount)
	}

	// Also send as thought for visibility
	thought := &domain.Thought{
		ID:             uuid.New(),
		ConversationID: parseUUID(input.ConversationID),
		GenerationID:   parseUUID(input.GenerationID),
		Text:           responsePayload.Text,
		DurationMs:     0,
		Model:          h.deps.Config.DefaultModel,
		HookEventName:  string(input.HookEventName),
		CursorVersion:  "",
		WorkspaceRoots: input.WorkspaceRoots,
		UserEmail:      "",
		Timestamp:      time.Now(),
	}

	if err := h.deps.Client.CreateThought(ctx, thought); err != nil {
		h.deps.Log.Error("failed to send thought", "error", err)
	}

	return NewEmptyOutput(), nil
}

// AfterAgentThoughtHandler handles afterAgentThought events.
type AfterAgentThoughtHandler struct {
	deps *HandlerDependencies
}

// Handle processes the afterAgentThought event.
//
//nolint:varnamelen // thoughtPayload is clear in context
func (h *AfterAgentThoughtHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	thoughtPayload, ok := payload.(*AfterAgentThoughtPayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	// Send as thought for visibility
	thought := &domain.Thought{
		ID:             uuid.New(),
		ConversationID: parseUUID(input.ConversationID),
		GenerationID:   parseUUID(input.GenerationID),
		Text:           thoughtPayload.Text,
		DurationMs:     0,
		Model:          h.deps.Config.DefaultModel,
		HookEventName:  string(input.HookEventName),
		CursorVersion:  "",
		WorkspaceRoots: input.WorkspaceRoots,
		UserEmail:      "",
		Timestamp:      time.Now(),
	}

	if err := h.deps.Client.CreateThought(ctx, thought); err != nil {
		h.deps.Log.Error("failed to send thought", "error", err)
	} else {
		h.deps.Log.Info("thought recorded", "event", input.HookEventName)
	}

	return NewEmptyOutput(), nil
}

// AfterMCPExecutionHandler handles afterMCPExecution events.
type AfterMCPExecutionHandler struct {
	deps *HandlerDependencies
}

// Handle processes the afterMCPExecution event.
//
//nolint:varnamelen // mcpPayload is clear in context
func (h *AfterMCPExecutionHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	mcpPayload, ok := payload.(*AfterMCPExecutionPayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	// Count tokens in tool input and result (best effort, ignore errors)
	inputTokens, inputErr := h.deps.Tokenizer.CountTokens(h.deps.Config.DefaultModel, string(mcpPayload.ToolInput))
	if inputErr != nil {
		h.deps.Log.Debug("failed to count input tokens", "error", inputErr)
		inputTokens = 0
	}
	resultTokens, resultErr := h.deps.Tokenizer.CountTokens(h.deps.Config.DefaultModel, string(mcpPayload.ResultJSON))
	if resultErr != nil {
		h.deps.Log.Debug("failed to count result tokens", "error", resultErr)
		resultTokens = 0
	}
	totalTokens := inputTokens + resultTokens

	inputChars := CountChars(string(mcpPayload.ToolInput))
	resultChars := CountChars(string(mcpPayload.ResultJSON))

	h.deps.Log.Debug("tool metrics",
		"tool", mcpPayload.ToolName,
		"input_tokens", inputTokens,
		"result_tokens", resultTokens,
		"conversation_id", input.ConversationID,
	)

	// Create token usage record
	usage := &domain.TokenUsage{
		ID:               uuid.New(),
		ConversationID:   parseUUID(input.ConversationID),
		GenerationID:     parseUUID(input.GenerationID),
		Model:            h.deps.Config.DefaultModel,
		Stage:            domain.TokenUsageStageTool,
		PromptTokens:     inputTokens,
		CompletionTokens: resultTokens,
		TotalTokens:      totalTokens,
		PromptChars:      inputChars,
		CompletionChars:  resultChars,
		WorkspaceRoots:   input.WorkspaceRoots,
		UserEmail:        "",
		HookEventName:    string(input.HookEventName),
		Timestamp:        time.Now(),
	}

	if err := h.deps.Client.CreateTokenUsage(ctx, usage); err != nil {
		h.deps.Log.Error("failed to send token usage", "error", err)
	} else {
		h.deps.Log.Info("token usage recorded", "stage", "tool", "tool", mcpPayload.ToolName, "tokens", totalTokens)
	}

	return NewEmptyOutput(), nil
}

// AfterFileEditHandler handles afterFileEdit events.
type AfterFileEditHandler struct {
	deps *HandlerDependencies
}

// Handle processes the afterFileEdit event.
//
//nolint:varnamelen // editPayload is clear in context
func (h *AfterFileEditHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	editPayload, ok := payload.(*AfterFileEditPayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	// Count characters in edits
	var oldChars, newChars int64
	for _, edit := range editPayload.Edits {
		oldChars += CountChars(edit.OldString)
		newChars += CountChars(edit.NewString)
	}

	h.deps.Log.Debug("file edit metrics",
		"file", editPayload.FilePath,
		"old_chars", oldChars,
		"new_chars", newChars,
		"edit_count", len(editPayload.Edits),
	)

	// Create token usage record for file edits
	usage := &domain.TokenUsage{
		ID:               uuid.New(),
		ConversationID:   parseUUID(input.ConversationID),
		GenerationID:     parseUUID(input.GenerationID),
		Model:            h.deps.Config.DefaultModel,
		Stage:            domain.TokenUsageStageFileEdit,
		PromptTokens:     0,
		CompletionTokens: 0,
		TotalTokens:      0,
		PromptChars:      oldChars,
		CompletionChars:  newChars,
		WorkspaceRoots:   input.WorkspaceRoots,
		UserEmail:        "",
		HookEventName:    string(input.HookEventName),
		Timestamp:        time.Now(),
	}

	if err := h.deps.Client.CreateTokenUsage(ctx, usage); err != nil {
		h.deps.Log.Error("failed to send token usage", "error", err)
	} else {
		h.deps.Log.Info("file edit recorded", "file", editPayload.FilePath, "edits", len(editPayload.Edits))
	}

	return NewEmptyOutput(), nil
}

// StopHandler handles stop events.
type StopHandler struct {
	deps *HandlerDependencies
}

// Handle processes the stop event.
//
//nolint:varnamelen // stopPayload is clear in context
func (h *StopHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	stopPayload, ok := payload.(*StopPayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	h.deps.Log.Info("agent loop stopped",
		"status", stopPayload.Status,
		"loop_count", stopPayload.LoopCount,
		"conversation_id", input.ConversationID,
	)

	// Create token usage record as session marker
	usage := &domain.TokenUsage{
		ID:               uuid.New(),
		ConversationID:   parseUUID(input.ConversationID),
		GenerationID:     parseUUID(input.GenerationID),
		Model:            h.deps.Config.DefaultModel,
		Stage:            domain.TokenUsageStageStop,
		PromptTokens:     0,
		CompletionTokens: 0,
		TotalTokens:      0,
		PromptChars:      0,
		CompletionChars:  0,
		WorkspaceRoots:   input.WorkspaceRoots,
		UserEmail:        "",
		HookEventName:    string(input.HookEventName),
		Timestamp:        time.Now(),
	}

	if err := h.deps.Client.CreateTokenUsage(ctx, usage); err != nil {
		h.deps.Log.Error("failed to send stop marker", "error", err)
	}

	return NewEmptyOutput(), nil
}

// BeforeShellExecutionHandler handles beforeShellExecution events.
type BeforeShellExecutionHandler struct {
	deps *HandlerDependencies
}

// Handle processes the beforeShellExecution event.
//
//nolint:varnamelen // shellPayload is clear in context
func (h *BeforeShellExecutionHandler) Handle(_ context.Context, input *HookInput, payload any) (*HookOutput, error) {
	shellPayload, ok := payload.(*BeforeShellExecutionPayload)
	if !ok {
		return NewAllowOutput(), nil
	}

	h.deps.Log.Debug("shell execution",
		"command", shellPayload.Command,
		"cwd", shellPayload.Cwd,
		"conversation_id", input.ConversationID,
	)

	// Always allow shell execution (observability only)
	return NewAllowOutput(), nil
}

// AfterShellExecutionHandler handles afterShellExecution events.
type AfterShellExecutionHandler struct {
	deps *HandlerDependencies
}

// Handle processes the afterShellExecution event.
//
//nolint:varnamelen // shellPayload is clear in context
func (h *AfterShellExecutionHandler) Handle(_ context.Context, input *HookInput, payload any) (*HookOutput, error) {
	shellPayload, ok := payload.(*AfterShellExecutionPayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	h.deps.Log.Debug("shell execution completed",
		"command", shellPayload.Command,
		"output_length", len(shellPayload.Output),
		"conversation_id", input.ConversationID,
	)

	return NewEmptyOutput(), nil
}

// BeforeMCPExecutionHandler handles beforeMCPExecution events.
type BeforeMCPExecutionHandler struct {
	deps *HandlerDependencies
}

// Handle processes the beforeMCPExecution event.
//
//nolint:varnamelen // mcpPayload is clear in context
func (h *BeforeMCPExecutionHandler) Handle(_ context.Context, input *HookInput, payload any) (*HookOutput, error) {
	mcpPayload, ok := payload.(*BeforeMCPExecutionPayload)
	if !ok {
		return NewAllowOutput(), nil
	}

	h.deps.Log.Debug("MCP execution",
		"tool", mcpPayload.ToolName,
		"conversation_id", input.ConversationID,
	)

	// Always allow MCP execution (observability only)
	return NewAllowOutput(), nil
}

// BeforeReadFileHandler handles beforeReadFile events.
type BeforeReadFileHandler struct {
	deps *HandlerDependencies
}

// Handle processes the beforeReadFile event.
//
//nolint:varnamelen // readPayload is clear in context
func (h *BeforeReadFileHandler) Handle(_ context.Context, input *HookInput, payload any) (*HookOutput, error) {
	readPayload, ok := payload.(*BeforeReadFilePayload)
	if !ok {
		return NewAllowOutput(), nil
	}

	h.deps.Log.Debug("file read",
		"file", readPayload.FilePath,
		"content_length", len(readPayload.Content),
		"conversation_id", input.ConversationID,
	)

	// Always allow file read (observability only)
	return NewAllowOutput(), nil
}

// parseUUID parses a UUID string, returning a zero UUID on error.
func parseUUID(s string) uuid.UUID {
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil
	}
	return id
}


```

```go
package hook

import (
	"context"
	"fmt"
	"time"

	"github.com/ainslies/ctd/internal/domain"
	"github.com/ainslies/ctd/pkg/logger"
	"github.com/google/uuid"
)

// Handler processes a hook event and returns the appropriate output.
type Handler interface {
	// Handle processes the hook input and returns the output.
	Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error)
}

// HandlerDependencies contains the dependencies for handlers.
type HandlerDependencies struct {
	Client    *CTDClient
	Tokenizer Tokenizer
	Log       *logger.Logger
	Config    *Config
}

// Router routes hook events to the appropriate handler.
type Router struct {
	deps     *HandlerDependencies
	handlers map[HookEventName]Handler
}

// NewRouter creates a new router with all handlers registered.
func NewRouter(deps *HandlerDependencies) *Router {
	router := &Router{
		deps:     deps,
		handlers: make(map[HookEventName]Handler),
	}

	// Register handlers
	router.handlers[EventBeforeSubmitPrompt] = &BeforeSubmitPromptHandler{deps: deps}
	router.handlers[EventAfterAgentResponse] = &AfterAgentResponseHandler{deps: deps}
	router.handlers[EventAfterAgentThought] = &AfterAgentThoughtHandler{deps: deps}
	router.handlers[EventAfterMCPExecution] = &AfterMCPExecutionHandler{deps: deps}
	router.handlers[EventAfterFileEdit] = &AfterFileEditHandler{deps: deps}
	router.handlers[EventStop] = &StopHandler{deps: deps}
	router.handlers[EventBeforeShellExecution] = &BeforeShellExecutionHandler{deps: deps}
	router.handlers[EventAfterShellExecution] = &AfterShellExecutionHandler{deps: deps}
	router.handlers[EventBeforeMCPExecution] = &BeforeMCPExecutionHandler{deps: deps}
	router.handlers[EventBeforeReadFile] = &BeforeReadFileHandler{deps: deps}

	return router
}

// Route routes the input to the appropriate handler.
func (r *Router) Route(ctx context.Context, input *HookInput) (*HookOutput, error) {
	handler, ok := r.handlers[input.HookEventName]
	if !ok {
		// Unknown event, return empty output
		r.deps.Log.Warn("unknown hook event", "event", input.HookEventName)
		return NewEmptyOutput(), nil
	}

	payload, err := input.ParsePayload()
	if err != nil {
		r.deps.Log.Error("failed to parse payload", "error", err, "event", input.HookEventName)
		// Return appropriate default output even on parse error
		return getDefaultOutput(input.HookEventName), nil
	}

	output, err := handler.Handle(ctx, input, payload)
	if err != nil {
		return nil, fmt.Errorf("handler error for %s: %w", input.HookEventName, err)
	}
	return output, nil
}

// getDefaultOutput returns the default output for an event type.
//
//nolint:exhaustive // Only specific events need special handling, others use empty output
func getDefaultOutput(event HookEventName) *HookOutput {
	switch event {
	case EventBeforeSubmitPrompt:
		return NewContinueOutput()
	case EventBeforeShellExecution, EventBeforeMCPExecution, EventBeforeReadFile:
		return NewAllowOutput()
	default:
		return NewEmptyOutput()
	}
}

// BeforeSubmitPromptHandler handles beforeSubmitPrompt events.
type BeforeSubmitPromptHandler struct {
	deps *HandlerDependencies
}

// Handle processes the beforeSubmitPrompt event.
func (h *BeforeSubmitPromptHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	promptPayload, ok := payload.(*BeforeSubmitPromptPayload)
	if !ok {
		return NewContinueOutput(), nil
	}

	// Count tokens
	tokenCount, err := h.deps.Tokenizer.CountTokens(h.deps.Config.DefaultModel, promptPayload.Prompt)
	if err != nil {
		h.deps.Log.Error("failed to count tokens", "error", err)
		tokenCount = 0
	}

	charCount := CountChars(promptPayload.Prompt)

	h.deps.Log.Debug("prompt metrics",
		"tokens", tokenCount,
		"chars", charCount,
		"conversation_id", input.ConversationID,
	)

	// Create token usage record
	usage := &domain.TokenUsage{
		ID:               uuid.New(),
		ConversationID:   parseUUID(input.ConversationID),
		GenerationID:     parseUUID(input.GenerationID),
		Model:            h.deps.Config.DefaultModel,
		Stage:            domain.TokenUsageStagePrompt,
		PromptTokens:     tokenCount,
		CompletionTokens: 0,
		TotalTokens:      tokenCount,
		PromptChars:      charCount,
		CompletionChars:  0,
		WorkspaceRoots:   input.WorkspaceRoots,
		UserEmail:        "",
		HookEventName:    string(input.HookEventName),
		Timestamp:        time.Now(),
	}

	// Send to ctd (best-effort)
	if err := h.deps.Client.CreateTokenUsage(ctx, usage); err != nil {
		h.deps.Log.Error("failed to send token usage", "error", err)
	} else {
		h.deps.Log.Info("token usage recorded", "stage", "prompt", "tokens", tokenCount)
	}

	return NewContinueOutput(), nil
}

// AfterAgentResponseHandler handles afterAgentResponse events.
type AfterAgentResponseHandler struct {
	deps *HandlerDependencies
}

// Handle processes the afterAgentResponse event.
func (h *AfterAgentResponseHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	responsePayload, ok := payload.(*AfterAgentResponsePayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	// Count tokens
	tokenCount, err := h.deps.Tokenizer.CountTokens(h.deps.Config.DefaultModel, responsePayload.Text)
	if err != nil {
		h.deps.Log.Error("failed to count tokens", "error", err)
		tokenCount = 0
	}

	charCount := CountChars(responsePayload.Text)

	h.deps.Log.Debug("completion metrics",
		"tokens", tokenCount,
		"chars", charCount,
		"conversation_id", input.ConversationID,
	)

	// Create token usage record
	usage := &domain.TokenUsage{
		ID:               uuid.New(),
		ConversationID:   parseUUID(input.ConversationID),
		GenerationID:     parseUUID(input.GenerationID),
		Model:            h.deps.Config.DefaultModel,
		Stage:            domain.TokenUsageStageCompletion,
		PromptTokens:     0,
		CompletionTokens: tokenCount,
		TotalTokens:      tokenCount,
		PromptChars:      0,
		CompletionChars:  charCount,
		WorkspaceRoots:   input.WorkspaceRoots,
		UserEmail:        "",
		HookEventName:    string(input.HookEventName),
		Timestamp:        time.Now(),
	}

	// Send token usage to ctd (best-effort)
	if err := h.deps.Client.CreateTokenUsage(ctx, usage); err != nil {
		h.deps.Log.Error("failed to send token usage", "error", err)
	} else {
		h.deps.Log.Info("token usage recorded", "stage", "completion", "tokens", tokenCount)
	}

	// Also send as thought for visibility
	thought := &domain.Thought{
		ID:             uuid.New(),
		ConversationID: parseUUID(input.ConversationID),
		GenerationID:   parseUUID(input.GenerationID),
		Text:           responsePayload.Text,
		DurationMs:     0,
		Model:          h.deps.Config.DefaultModel,
		HookEventName:  string(input.HookEventName),
		CursorVersion:  "",
		WorkspaceRoots: input.WorkspaceRoots,
		UserEmail:      "",
		Timestamp:      time.Now(),
	}

	if err := h.deps.Client.CreateThought(ctx, thought); err != nil {
		h.deps.Log.Error("failed to send thought", "error", err)
	}

	return NewEmptyOutput(), nil
}

// AfterAgentThoughtHandler handles afterAgentThought events.
type AfterAgentThoughtHandler struct {
	deps *HandlerDependencies
}

// Handle processes the afterAgentThought event.
func (h *AfterAgentThoughtHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	thoughtPayload, ok := payload.(*AfterAgentThoughtPayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	// Send as thought for visibility
	thought := &domain.Thought{
		ID:             uuid.New(),
		ConversationID: parseUUID(input.ConversationID),
		GenerationID:   parseUUID(input.GenerationID),
		Text:           thoughtPayload.Text,
		DurationMs:     0,
		Model:          h.deps.Config.DefaultModel,
		HookEventName:  string(input.HookEventName),
		CursorVersion:  "",
		WorkspaceRoots: input.WorkspaceRoots,
		UserEmail:      "",
		Timestamp:      time.Now(),
	}

	if err := h.deps.Client.CreateThought(ctx, thought); err != nil {
		h.deps.Log.Error("failed to send thought", "error", err)
	} else {
		h.deps.Log.Info("thought recorded", "event", input.HookEventName)
	}

	return NewEmptyOutput(), nil
}

// AfterMCPExecutionHandler handles afterMCPExecution events.
type AfterMCPExecutionHandler struct {
	deps *HandlerDependencies
}

// Handle processes the afterMCPExecution event.
//
//nolint:varnamelen // mcpPayload is clear in context
func (h *AfterMCPExecutionHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	mcpPayload, ok := payload.(*AfterMCPExecutionPayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	// Count tokens in tool input and result (best effort, ignore errors)
	inputTokens, inputErr := h.deps.Tokenizer.CountTokens(h.deps.Config.DefaultModel, string(mcpPayload.ToolInput))
	if inputErr != nil {
		h.deps.Log.Debug("failed to count input tokens", "error", inputErr)
		inputTokens = 0
	}
	resultTokens, resultErr := h.deps.Tokenizer.CountTokens(h.deps.Config.DefaultModel, string(mcpPayload.ResultJSON))
	if resultErr != nil {
		h.deps.Log.Debug("failed to count result tokens", "error", resultErr)
		resultTokens = 0
	}
	totalTokens := inputTokens + resultTokens

	inputChars := CountChars(string(mcpPayload.ToolInput))
	resultChars := CountChars(string(mcpPayload.ResultJSON))

	h.deps.Log.Debug("tool metrics",
		"tool", mcpPayload.ToolName,
		"input_tokens", inputTokens,
		"result_tokens", resultTokens,
		"conversation_id", input.ConversationID,
	)

	// Create token usage record
	usage := &domain.TokenUsage{
		ID:               uuid.New(),
		ConversationID:   parseUUID(input.ConversationID),
		GenerationID:     parseUUID(input.GenerationID),
		Model:            h.deps.Config.DefaultModel,
		Stage:            domain.TokenUsageStageTool,
		PromptTokens:     inputTokens,
		CompletionTokens: resultTokens,
		TotalTokens:      totalTokens,
		PromptChars:      inputChars,
		CompletionChars:  resultChars,
		WorkspaceRoots:   input.WorkspaceRoots,
		UserEmail:        "",
		HookEventName:    string(input.HookEventName),
		Timestamp:        time.Now(),
	}

	if err := h.deps.Client.CreateTokenUsage(ctx, usage); err != nil {
		h.deps.Log.Error("failed to send token usage", "error", err)
	} else {
		h.deps.Log.Info("token usage recorded", "stage", "tool", "tool", mcpPayload.ToolName, "tokens", totalTokens)
	}

	return NewEmptyOutput(), nil
}

// AfterFileEditHandler handles afterFileEdit events.
type AfterFileEditHandler struct {
	deps *HandlerDependencies
}

// Handle processes the afterFileEdit event.
//
//nolint:varnamelen // editPayload is clear in context
func (h *AfterFileEditHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	editPayload, ok := payload.(*AfterFileEditPayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	// Count characters in edits
	var oldChars, newChars int64
	for _, edit := range editPayload.Edits {
		oldChars += CountChars(edit.OldString)
		newChars += CountChars(edit.NewString)
	}

	h.deps.Log.Debug("file edit metrics",
		"file", editPayload.FilePath,
		"old_chars", oldChars,
		"new_chars", newChars,
		"edit_count", len(editPayload.Edits),
	)

	// Create token usage record for file edits
	usage := &domain.TokenUsage{
		ID:               uuid.New(),
		ConversationID:   parseUUID(input.ConversationID),
		GenerationID:     parseUUID(input.GenerationID),
		Model:            h.deps.Config.DefaultModel,
		Stage:            domain.TokenUsageStageFileEdit,
		PromptTokens:     0,
		CompletionTokens: 0,
		TotalTokens:      0,
		PromptChars:      oldChars,
		CompletionChars:  newChars,
		WorkspaceRoots:   input.WorkspaceRoots,
		UserEmail:        "",
		HookEventName:    string(input.HookEventName),
		Timestamp:        time.Now(),
	}

	if err := h.deps.Client.CreateTokenUsage(ctx, usage); err != nil {
		h.deps.Log.Error("failed to send token usage", "error", err)
	} else {
		h.deps.Log.Info("file edit recorded", "file", editPayload.FilePath, "edits", len(editPayload.Edits))
	}

	return NewEmptyOutput(), nil
}

// StopHandler handles stop events.
type StopHandler struct {
	deps *HandlerDependencies
}

// Handle processes the stop event.
//
//nolint:varnamelen // stopPayload is clear in context
func (h *StopHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	stopPayload, ok := payload.(*StopPayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	h.deps.Log.Info("agent loop stopped",
		"status", stopPayload.Status,
		"loop_count", stopPayload.LoopCount,
		"conversation_id", input.ConversationID,
	)

	// Create token usage record as session marker
	usage := &domain.TokenUsage{
		ID:               uuid.New(),
		ConversationID:   parseUUID(input.ConversationID),
		GenerationID:     parseUUID(input.GenerationID),
		Model:            h.deps.Config.DefaultModel,
		Stage:            domain.TokenUsageStageStop,
		PromptTokens:     0,
		CompletionTokens: 0,
		TotalTokens:      0,
		PromptChars:      0,
		CompletionChars:  0,
		WorkspaceRoots:   input.WorkspaceRoots,
		UserEmail:        "",
		HookEventName:    string(input.HookEventName),
		Timestamp:        time.Now(),
	}

	if err := h.deps.Client.CreateTokenUsage(ctx, usage); err != nil {
		h.deps.Log.Error("failed to send stop marker", "error", err)
	}

	return NewEmptyOutput(), nil
}

// BeforeShellExecutionHandler handles beforeShellExecution events.
type BeforeShellExecutionHandler struct {
	deps *HandlerDependencies
}

// Handle processes the beforeShellExecution event.
//
//nolint:varnamelen // shellPayload is clear in context
func (h *BeforeShellExecutionHandler) Handle(_ context.Context, input *HookInput, payload any) (*HookOutput, error) {
	shellPayload, ok := payload.(*BeforeShellExecutionPayload)
	if !ok {
		return NewAllowOutput(), nil
	}

	h.deps.Log.Debug("shell execution",
		"command", shellPayload.Command,
		"cwd", shellPayload.Cwd,
		"conversation_id", input.ConversationID,
	)

	// Always allow shell execution (observability only)
	return NewAllowOutput(), nil
}

// AfterShellExecutionHandler handles afterShellExecution events.
type AfterShellExecutionHandler struct {
	deps *HandlerDependencies
}

// Handle processes the afterShellExecution event.
//
//nolint:varnamelen // shellPayload is clear in context
func (h *AfterShellExecutionHandler) Handle(_ context.Context, input *HookInput, payload any) (*HookOutput, error) {
	shellPayload, ok := payload.(*AfterShellExecutionPayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	h.deps.Log.Debug("shell execution completed",
		"command", shellPayload.Command,
		"output_length", len(shellPayload.Output),
		"conversation_id", input.ConversationID,
	)

	return NewEmptyOutput(), nil
}

// BeforeMCPExecutionHandler handles beforeMCPExecution events.
type BeforeMCPExecutionHandler struct {
	deps *HandlerDependencies
}

// Handle processes the beforeMCPExecution event.
//
//nolint:varnamelen // mcpPayload is clear in context
func (h *BeforeMCPExecutionHandler) Handle(_ context.Context, input *HookInput, payload any) (*HookOutput, error) {
	mcpPayload, ok := payload.(*BeforeMCPExecutionPayload)
	if !ok {
		return NewAllowOutput(), nil
	}

	h.deps.Log.Debug("MCP execution",
		"tool", mcpPayload.ToolName,
		"conversation_id", input.ConversationID,
	)

	// Always allow MCP execution (observability only)
	return NewAllowOutput(), nil
}

// BeforeReadFileHandler handles beforeReadFile events.
type BeforeReadFileHandler struct {
	deps *HandlerDependencies
}

// Handle processes the beforeReadFile event.
//
//nolint:varnamelen // readPayload is clear in context
func (h *BeforeReadFileHandler) Handle(_ context.Context, input *HookInput, payload any) (*HookOutput, error) {
	readPayload, ok := payload.(*BeforeReadFilePayload)
	if !ok {
		return NewAllowOutput(), nil
	}

	h.deps.Log.Debug("file read",
		"file", readPayload.FilePath,
		"content_length", len(readPayload.Content),
		"conversation_id", input.ConversationID,
	)

	// Always allow file read (observability only)
	return NewAllowOutput(), nil
}

// parseUUID parses a UUID string, returning a zero UUID on error.
func parseUUID(s string) uuid.UUID {
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil
	}
	return id
}


```

```go
package hook

import (
	"context"
	"fmt"
	"time"

	"github.com/ainslies/ctd/internal/domain"
	"github.com/ainslies/ctd/pkg/logger"
	"github.com/google/uuid"
)

// Handler processes a hook event and returns the appropriate output.
type Handler interface {
	// Handle processes the hook input and returns the output.
	Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error)
}

// HandlerDependencies contains the dependencies for handlers.
type HandlerDependencies struct {
	Client    *CTDClient
	Tokenizer Tokenizer
	Log       *logger.Logger
	Config    *Config
}

// Router routes hook events to the appropriate handler.
type Router struct {
	deps     *HandlerDependencies
	handlers map[HookEventName]Handler
}

// NewRouter creates a new router with all handlers registered.
func NewRouter(deps *HandlerDependencies) *Router {
	router := &Router{
		deps:     deps,
		handlers: make(map[HookEventName]Handler),
	}

	// Register handlers
	router.handlers[EventBeforeSubmitPrompt] = &BeforeSubmitPromptHandler{deps: deps}
	router.handlers[EventAfterAgentResponse] = &AfterAgentResponseHandler{deps: deps}
	router.handlers[EventAfterAgentThought] = &AfterAgentThoughtHandler{deps: deps}
	router.handlers[EventAfterMCPExecution] = &AfterMCPExecutionHandler{deps: deps}
	router.handlers[EventAfterFileEdit] = &AfterFileEditHandler{deps: deps}
	router.handlers[EventStop] = &StopHandler{deps: deps}
	router.handlers[EventBeforeShellExecution] = &BeforeShellExecutionHandler{deps: deps}
	router.handlers[EventAfterShellExecution] = &AfterShellExecutionHandler{deps: deps}
	router.handlers[EventBeforeMCPExecution] = &BeforeMCPExecutionHandler{deps: deps}
	router.handlers[EventBeforeReadFile] = &BeforeReadFileHandler{deps: deps}

	return router
}

// Route routes the input to the appropriate handler.
func (r *Router) Route(ctx context.Context, input *HookInput) (*HookOutput, error) {
	handler, ok := r.handlers[input.HookEventName]
	if !ok {
		// Unknown event, return empty output
		r.deps.Log.Warn("unknown hook event", "event", input.HookEventName)
		return NewEmptyOutput(), nil
	}

	payload, err := input.ParsePayload()
	if err != nil {
		r.deps.Log.Error("failed to parse payload", "error", err, "event", input.HookEventName)
		// Return appropriate default output even on parse error
		return getDefaultOutput(input.HookEventName), nil
	}

	output, err := handler.Handle(ctx, input, payload)
	if err != nil {
		return nil, fmt.Errorf("handler error for %s: %w", input.HookEventName, err)
	}
	return output, nil
}

// getDefaultOutput returns the default output for an event type.
//
//nolint:exhaustive // Only specific events need special handling, others use empty output
func getDefaultOutput(event HookEventName) *HookOutput {
	switch event {
	case EventBeforeSubmitPrompt:
		return NewContinueOutput()
	case EventBeforeShellExecution, EventBeforeMCPExecution, EventBeforeReadFile:
		return NewAllowOutput()
	default:
		return NewEmptyOutput()
	}
}

// BeforeSubmitPromptHandler handles beforeSubmitPrompt events.
type BeforeSubmitPromptHandler struct {
	deps *HandlerDependencies
}

// Handle processes the beforeSubmitPrompt event.
func (h *BeforeSubmitPromptHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	promptPayload, ok := payload.(*BeforeSubmitPromptPayload)
	if !ok {
		return NewContinueOutput(), nil
	}

	// Count tokens
	tokenCount, err := h.deps.Tokenizer.CountTokens(h.deps.Config.DefaultModel, promptPayload.Prompt)
	if err != nil {
		h.deps.Log.Error("failed to count tokens", "error", err)
		tokenCount = 0
	}

	charCount := CountChars(promptPayload.Prompt)

	h.deps.Log.Debug("prompt metrics",
		"tokens", tokenCount,
		"chars", charCount,
		"conversation_id", input.ConversationID,
	)

	// Create token usage record
	usage := &domain.TokenUsage{
		ID:               uuid.New(),
		ConversationID:   parseUUID(input.ConversationID),
		GenerationID:     parseUUID(input.GenerationID),
		Model:            h.deps.Config.DefaultModel,
		Stage:            domain.TokenUsageStagePrompt,
		PromptTokens:     tokenCount,
		CompletionTokens: 0,
		TotalTokens:      tokenCount,
		PromptChars:      charCount,
		CompletionChars:  0,
		WorkspaceRoots:   input.WorkspaceRoots,
		UserEmail:        "",
		HookEventName:    string(input.HookEventName),
		Timestamp:        time.Now(),
	}

	// Send to ctd (best-effort)
	if err := h.deps.Client.CreateTokenUsage(ctx, usage); err != nil {
		h.deps.Log.Error("failed to send token usage", "error", err)
	} else {
		h.deps.Log.Info("token usage recorded", "stage", "prompt", "tokens", tokenCount)
	}

	return NewContinueOutput(), nil
}

// AfterAgentResponseHandler handles afterAgentResponse events.
type AfterAgentResponseHandler struct {
	deps *HandlerDependencies
}

// Handle processes the afterAgentResponse event.
func (h *AfterAgentResponseHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	responsePayload, ok := payload.(*AfterAgentResponsePayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	// Count tokens
	tokenCount, err := h.deps.Tokenizer.CountTokens(h.deps.Config.DefaultModel, responsePayload.Text)
	if err != nil {
		h.deps.Log.Error("failed to count tokens", "error", err)
		tokenCount = 0
	}

	charCount := CountChars(responsePayload.Text)

	h.deps.Log.Debug("completion metrics",
		"tokens", tokenCount,
		"chars", charCount,
		"conversation_id", input.ConversationID,
	)

	// Create token usage record
	usage := &domain.TokenUsage{
		ID:               uuid.New(),
		ConversationID:   parseUUID(input.ConversationID),
		GenerationID:     parseUUID(input.GenerationID),
		Model:            h.deps.Config.DefaultModel,
		Stage:            domain.TokenUsageStageCompletion,
		PromptTokens:     0,
		CompletionTokens: tokenCount,
		TotalTokens:      tokenCount,
		PromptChars:      0,
		CompletionChars:  charCount,
		WorkspaceRoots:   input.WorkspaceRoots,
		UserEmail:        "",
		HookEventName:    string(input.HookEventName),
		Timestamp:        time.Now(),
	}

	// Send token usage to ctd (best-effort)
	if err := h.deps.Client.CreateTokenUsage(ctx, usage); err != nil {
		h.deps.Log.Error("failed to send token usage", "error", err)
	} else {
		h.deps.Log.Info("token usage recorded", "stage", "completion", "tokens", tokenCount)
	}

	// Also send as thought for visibility
	thought := &domain.Thought{
		ID:             uuid.New(),
		ConversationID: parseUUID(input.ConversationID),
		GenerationID:   parseUUID(input.GenerationID),
		Text:           responsePayload.Text,
		DurationMs:     0,
		Model:          h.deps.Config.DefaultModel,
		HookEventName:  string(input.HookEventName),
		CursorVersion:  "",
		WorkspaceRoots: input.WorkspaceRoots,
		UserEmail:      "",
		Timestamp:      time.Now(),
	}

	if err := h.deps.Client.CreateThought(ctx, thought); err != nil {
		h.deps.Log.Error("failed to send thought", "error", err)
	}

	return NewEmptyOutput(), nil
}

// AfterAgentThoughtHandler handles afterAgentThought events.
type AfterAgentThoughtHandler struct {
	deps *HandlerDependencies
}

// Handle processes the afterAgentThought event.
func (h *AfterAgentThoughtHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	thoughtPayload, ok := payload.(*AfterAgentThoughtPayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	// Send as thought for visibility
	thought := &domain.Thought{
		ID:             uuid.New(),
		ConversationID: parseUUID(input.ConversationID),
		GenerationID:   parseUUID(input.GenerationID),
		Text:           thoughtPayload.Text,
		DurationMs:     0,
		Model:          h.deps.Config.DefaultModel,
		HookEventName:  string(input.HookEventName),
		CursorVersion:  "",
		WorkspaceRoots: input.WorkspaceRoots,
		UserEmail:      "",
		Timestamp:      time.Now(),
	}

	if err := h.deps.Client.CreateThought(ctx, thought); err != nil {
		h.deps.Log.Error("failed to send thought", "error", err)
	} else {
		h.deps.Log.Info("thought recorded", "event", input.HookEventName)
	}

	return NewEmptyOutput(), nil
}

// AfterMCPExecutionHandler handles afterMCPExecution events.
type AfterMCPExecutionHandler struct {
	deps *HandlerDependencies
}

// Handle processes the afterMCPExecution event.
func (h *AfterMCPExecutionHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	mcpPayload, ok := payload.(*AfterMCPExecutionPayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	// Count tokens in tool input and result (best effort, ignore errors)
	inputTokens, inputErr := h.deps.Tokenizer.CountTokens(h.deps.Config.DefaultModel, string(mcpPayload.ToolInput))
	if inputErr != nil {
		h.deps.Log.Debug("failed to count input tokens", "error", inputErr)
		inputTokens = 0
	}
	resultTokens, resultErr := h.deps.Tokenizer.CountTokens(h.deps.Config.DefaultModel, string(mcpPayload.ResultJSON))
	if resultErr != nil {
		h.deps.Log.Debug("failed to count result tokens", "error", resultErr)
		resultTokens = 0
	}
	totalTokens := inputTokens + resultTokens

	inputChars := CountChars(string(mcpPayload.ToolInput))
	resultChars := CountChars(string(mcpPayload.ResultJSON))

	h.deps.Log.Debug("tool metrics",
		"tool", mcpPayload.ToolName,
		"input_tokens", inputTokens,
		"result_tokens", resultTokens,
		"conversation_id", input.ConversationID,
	)

	// Create token usage record
	usage := &domain.TokenUsage{
		ID:               uuid.New(),
		ConversationID:   parseUUID(input.ConversationID),
		GenerationID:     parseUUID(input.GenerationID),
		Model:            h.deps.Config.DefaultModel,
		Stage:            domain.TokenUsageStageTool,
		PromptTokens:     inputTokens,
		CompletionTokens: resultTokens,
		TotalTokens:      totalTokens,
		PromptChars:      inputChars,
		CompletionChars:  resultChars,
		WorkspaceRoots:   input.WorkspaceRoots,
		UserEmail:        "",
		HookEventName:    string(input.HookEventName),
		Timestamp:        time.Now(),
	}

	if err := h.deps.Client.CreateTokenUsage(ctx, usage); err != nil {
		h.deps.Log.Error("failed to send token usage", "error", err)
	} else {
		h.deps.Log.Info("token usage recorded", "stage", "tool", "tool", mcpPayload.ToolName, "tokens", totalTokens)
	}

	return NewEmptyOutput(), nil
}

// AfterFileEditHandler handles afterFileEdit events.
type AfterFileEditHandler struct {
	deps *HandlerDependencies
}

// Handle processes the afterFileEdit event.
//
//nolint:varnamelen // editPayload is clear in context
func (h *AfterFileEditHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	editPayload, ok := payload.(*AfterFileEditPayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	// Count characters in edits
	var oldChars, newChars int64
	for _, edit := range editPayload.Edits {
		oldChars += CountChars(edit.OldString)
		newChars += CountChars(edit.NewString)
	}

	h.deps.Log.Debug("file edit metrics",
		"file", editPayload.FilePath,
		"old_chars", oldChars,
		"new_chars", newChars,
		"edit_count", len(editPayload.Edits),
	)

	// Create token usage record for file edits
	usage := &domain.TokenUsage{
		ID:               uuid.New(),
		ConversationID:   parseUUID(input.ConversationID),
		GenerationID:     parseUUID(input.GenerationID),
		Model:            h.deps.Config.DefaultModel,
		Stage:            domain.TokenUsageStageFileEdit,
		PromptTokens:     0,
		CompletionTokens: 0,
		TotalTokens:      0,
		PromptChars:      oldChars,
		CompletionChars:  newChars,
		WorkspaceRoots:   input.WorkspaceRoots,
		UserEmail:        "",
		HookEventName:    string(input.HookEventName),
		Timestamp:        time.Now(),
	}

	if err := h.deps.Client.CreateTokenUsage(ctx, usage); err != nil {
		h.deps.Log.Error("failed to send token usage", "error", err)
	} else {
		h.deps.Log.Info("file edit recorded", "file", editPayload.FilePath, "edits", len(editPayload.Edits))
	}

	return NewEmptyOutput(), nil
}

// StopHandler handles stop events.
type StopHandler struct {
	deps *HandlerDependencies
}

// Handle processes the stop event.
//
//nolint:varnamelen // stopPayload is clear in context
func (h *StopHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	stopPayload, ok := payload.(*StopPayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	h.deps.Log.Info("agent loop stopped",
		"status", stopPayload.Status,
		"loop_count", stopPayload.LoopCount,
		"conversation_id", input.ConversationID,
	)

	// Create token usage record as session marker
	usage := &domain.TokenUsage{
		ID:               uuid.New(),
		ConversationID:   parseUUID(input.ConversationID),
		GenerationID:     parseUUID(input.GenerationID),
		Model:            h.deps.Config.DefaultModel,
		Stage:            domain.TokenUsageStageStop,
		PromptTokens:     0,
		CompletionTokens: 0,
		TotalTokens:      0,
		PromptChars:      0,
		CompletionChars:  0,
		WorkspaceRoots:   input.WorkspaceRoots,
		UserEmail:        "",
		HookEventName:    string(input.HookEventName),
		Timestamp:        time.Now(),
	}

	if err := h.deps.Client.CreateTokenUsage(ctx, usage); err != nil {
		h.deps.Log.Error("failed to send stop marker", "error", err)
	}

	return NewEmptyOutput(), nil
}

// BeforeShellExecutionHandler handles beforeShellExecution events.
type BeforeShellExecutionHandler struct {
	deps *HandlerDependencies
}

// Handle processes the beforeShellExecution event.
//
//nolint:varnamelen // shellPayload is clear in context
func (h *BeforeShellExecutionHandler) Handle(_ context.Context, input *HookInput, payload any) (*HookOutput, error) {
	shellPayload, ok := payload.(*BeforeShellExecutionPayload)
	if !ok {
		return NewAllowOutput(), nil
	}

	h.deps.Log.Debug("shell execution",
		"command", shellPayload.Command,
		"cwd", shellPayload.Cwd,
		"conversation_id", input.ConversationID,
	)

	// Always allow shell execution (observability only)
	return NewAllowOutput(), nil
}

// AfterShellExecutionHandler handles afterShellExecution events.
type AfterShellExecutionHandler struct {
	deps *HandlerDependencies
}

// Handle processes the afterShellExecution event.
//
//nolint:varnamelen // shellPayload is clear in context
func (h *AfterShellExecutionHandler) Handle(_ context.Context, input *HookInput, payload any) (*HookOutput, error) {
	shellPayload, ok := payload.(*AfterShellExecutionPayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	h.deps.Log.Debug("shell execution completed",
		"command", shellPayload.Command,
		"output_length", len(shellPayload.Output),
		"conversation_id", input.ConversationID,
	)

	return NewEmptyOutput(), nil
}

// BeforeMCPExecutionHandler handles beforeMCPExecution events.
type BeforeMCPExecutionHandler struct {
	deps *HandlerDependencies
}

// Handle processes the beforeMCPExecution event.
//
//nolint:varnamelen // mcpPayload is clear in context
func (h *BeforeMCPExecutionHandler) Handle(_ context.Context, input *HookInput, payload any) (*HookOutput, error) {
	mcpPayload, ok := payload.(*BeforeMCPExecutionPayload)
	if !ok {
		return NewAllowOutput(), nil
	}

	h.deps.Log.Debug("MCP execution",
		"tool", mcpPayload.ToolName,
		"conversation_id", input.ConversationID,
	)

	// Always allow MCP execution (observability only)
	return NewAllowOutput(), nil
}

// BeforeReadFileHandler handles beforeReadFile events.
type BeforeReadFileHandler struct {
	deps *HandlerDependencies
}

// Handle processes the beforeReadFile event.
//
//nolint:varnamelen // readPayload is clear in context
func (h *BeforeReadFileHandler) Handle(_ context.Context, input *HookInput, payload any) (*HookOutput, error) {
	readPayload, ok := payload.(*BeforeReadFilePayload)
	if !ok {
		return NewAllowOutput(), nil
	}

	h.deps.Log.Debug("file read",
		"file", readPayload.FilePath,
		"content_length", len(readPayload.Content),
		"conversation_id", input.ConversationID,
	)

	// Always allow file read (observability only)
	return NewAllowOutput(), nil
}

// parseUUID parses a UUID string, returning a zero UUID on error.
func parseUUID(s string) uuid.UUID {
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil
	}
	return id
}


```

```go
package hook

import (
	"context"
	"fmt"
	"time"

	"github.com/ainslies/ctd/internal/domain"
	"github.com/ainslies/ctd/pkg/logger"
	"github.com/google/uuid"
)

// Handler processes a hook event and returns the appropriate output.
type Handler interface {
	// Handle processes the hook input and returns the output.
	Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error)
}

// HandlerDependencies contains the dependencies for handlers.
type HandlerDependencies struct {
	Client    *CTDClient
	Tokenizer Tokenizer
	Log       *logger.Logger
	Config    *Config
}

// Router routes hook events to the appropriate handler.
type Router struct {
	deps     *HandlerDependencies
	handlers map[HookEventName]Handler
}

// NewRouter creates a new router with all handlers registered.
func NewRouter(deps *HandlerDependencies) *Router {
	router := &Router{
		deps:     deps,
		handlers: make(map[HookEventName]Handler),
	}

	// Register handlers
	router.handlers[EventBeforeSubmitPrompt] = &BeforeSubmitPromptHandler{deps: deps}
	router.handlers[EventAfterAgentResponse] = &AfterAgentResponseHandler{deps: deps}
	router.handlers[EventAfterAgentThought] = &AfterAgentThoughtHandler{deps: deps}
	router.handlers[EventAfterMCPExecution] = &AfterMCPExecutionHandler{deps: deps}
	router.handlers[EventAfterFileEdit] = &AfterFileEditHandler{deps: deps}
	router.handlers[EventStop] = &StopHandler{deps: deps}
	router.handlers[EventBeforeShellExecution] = &BeforeShellExecutionHandler{deps: deps}
	router.handlers[EventAfterShellExecution] = &AfterShellExecutionHandler{deps: deps}
	router.handlers[EventBeforeMCPExecution] = &BeforeMCPExecutionHandler{deps: deps}
	router.handlers[EventBeforeReadFile] = &BeforeReadFileHandler{deps: deps}

	return router
}

// Route routes the input to the appropriate handler.
func (r *Router) Route(ctx context.Context, input *HookInput) (*HookOutput, error) {
	handler, ok := r.handlers[input.HookEventName]
	if !ok {
		// Unknown event, return empty output
		r.deps.Log.Warn("unknown hook event", "event", input.HookEventName)
		return NewEmptyOutput(), nil
	}

	payload, err := input.ParsePayload()
	if err != nil {
		r.deps.Log.Error("failed to parse payload", "error", err, "event", input.HookEventName)
		// Return appropriate default output even on parse error
		return getDefaultOutput(input.HookEventName), nil
	}

	output, err := handler.Handle(ctx, input, payload)
	if err != nil {
		return nil, fmt.Errorf("handler error for %s: %w", input.HookEventName, err)
	}
	return output, nil
}

// getDefaultOutput returns the default output for an event type.
//
//nolint:exhaustive // Only specific events need special handling, others use empty output
func getDefaultOutput(event HookEventName) *HookOutput {
	switch event {
	case EventBeforeSubmitPrompt:
		return NewContinueOutput()
	case EventBeforeShellExecution, EventBeforeMCPExecution, EventBeforeReadFile:
		return NewAllowOutput()
	default:
		return NewEmptyOutput()
	}
}

// BeforeSubmitPromptHandler handles beforeSubmitPrompt events.
type BeforeSubmitPromptHandler struct {
	deps *HandlerDependencies
}

// Handle processes the beforeSubmitPrompt event.
func (h *BeforeSubmitPromptHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	promptPayload, ok := payload.(*BeforeSubmitPromptPayload)
	if !ok {
		return NewContinueOutput(), nil
	}

	// Count tokens
	tokenCount, err := h.deps.Tokenizer.CountTokens(h.deps.Config.DefaultModel, promptPayload.Prompt)
	if err != nil {
		h.deps.Log.Error("failed to count tokens", "error", err)
		tokenCount = 0
	}

	charCount := CountChars(promptPayload.Prompt)

	h.deps.Log.Debug("prompt metrics",
		"tokens", tokenCount,
		"chars", charCount,
		"conversation_id", input.ConversationID,
	)

	// Create token usage record
	usage := &domain.TokenUsage{
		ID:               uuid.New(),
		ConversationID:   parseUUID(input.ConversationID),
		GenerationID:     parseUUID(input.GenerationID),
		Model:            h.deps.Config.DefaultModel,
		Stage:            domain.TokenUsageStagePrompt,
		PromptTokens:     tokenCount,
		CompletionTokens: 0,
		TotalTokens:      tokenCount,
		PromptChars:      charCount,
		CompletionChars:  0,
		WorkspaceRoots:   input.WorkspaceRoots,
		UserEmail:        "",
		HookEventName:    string(input.HookEventName),
		Timestamp:        time.Now(),
	}

	// Send to ctd (best-effort)
	if err := h.deps.Client.CreateTokenUsage(ctx, usage); err != nil {
		h.deps.Log.Error("failed to send token usage", "error", err)
	} else {
		h.deps.Log.Info("token usage recorded", "stage", "prompt", "tokens", tokenCount)
	}

	return NewContinueOutput(), nil
}

// AfterAgentResponseHandler handles afterAgentResponse events.
type AfterAgentResponseHandler struct {
	deps *HandlerDependencies
}

// Handle processes the afterAgentResponse event.
func (h *AfterAgentResponseHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	responsePayload, ok := payload.(*AfterAgentResponsePayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	// Count tokens
	tokenCount, err := h.deps.Tokenizer.CountTokens(h.deps.Config.DefaultModel, responsePayload.Text)
	if err != nil {
		h.deps.Log.Error("failed to count tokens", "error", err)
		tokenCount = 0
	}

	charCount := CountChars(responsePayload.Text)

	h.deps.Log.Debug("completion metrics",
		"tokens", tokenCount,
		"chars", charCount,
		"conversation_id", input.ConversationID,
	)

	// Create token usage record
	usage := &domain.TokenUsage{
		ID:               uuid.New(),
		ConversationID:   parseUUID(input.ConversationID),
		GenerationID:     parseUUID(input.GenerationID),
		Model:            h.deps.Config.DefaultModel,
		Stage:            domain.TokenUsageStageCompletion,
		PromptTokens:     0,
		CompletionTokens: tokenCount,
		TotalTokens:      tokenCount,
		PromptChars:      0,
		CompletionChars:  charCount,
		WorkspaceRoots:   input.WorkspaceRoots,
		UserEmail:        "",
		HookEventName:    string(input.HookEventName),
		Timestamp:        time.Now(),
	}

	// Send token usage to ctd (best-effort)
	if err := h.deps.Client.CreateTokenUsage(ctx, usage); err != nil {
		h.deps.Log.Error("failed to send token usage", "error", err)
	} else {
		h.deps.Log.Info("token usage recorded", "stage", "completion", "tokens", tokenCount)
	}

	// Also send as thought for visibility
	thought := &domain.Thought{
		ID:             uuid.New(),
		ConversationID: parseUUID(input.ConversationID),
		GenerationID:   parseUUID(input.GenerationID),
		Text:           responsePayload.Text,
		DurationMs:     0,
		Model:          h.deps.Config.DefaultModel,
		HookEventName:  string(input.HookEventName),
		CursorVersion:  "",
		WorkspaceRoots: input.WorkspaceRoots,
		UserEmail:      "",
		Timestamp:      time.Now(),
	}

	if err := h.deps.Client.CreateThought(ctx, thought); err != nil {
		h.deps.Log.Error("failed to send thought", "error", err)
	}

	return NewEmptyOutput(), nil
}

// AfterAgentThoughtHandler handles afterAgentThought events.
type AfterAgentThoughtHandler struct {
	deps *HandlerDependencies
}

// Handle processes the afterAgentThought event.
func (h *AfterAgentThoughtHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	thoughtPayload, ok := payload.(*AfterAgentThoughtPayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	// Send as thought for visibility
	thought := &domain.Thought{
		ID:             uuid.New(),
		ConversationID: parseUUID(input.ConversationID),
		GenerationID:   parseUUID(input.GenerationID),
		Text:           thoughtPayload.Text,
		DurationMs:     0,
		Model:          h.deps.Config.DefaultModel,
		HookEventName:  string(input.HookEventName),
		CursorVersion:  "",
		WorkspaceRoots: input.WorkspaceRoots,
		UserEmail:      "",
		Timestamp:      time.Now(),
	}

	if err := h.deps.Client.CreateThought(ctx, thought); err != nil {
		h.deps.Log.Error("failed to send thought", "error", err)
	} else {
		h.deps.Log.Info("thought recorded", "event", input.HookEventName)
	}

	return NewEmptyOutput(), nil
}

// AfterMCPExecutionHandler handles afterMCPExecution events.
type AfterMCPExecutionHandler struct {
	deps *HandlerDependencies
}

// Handle processes the afterMCPExecution event.
func (h *AfterMCPExecutionHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	mcpPayload, ok := payload.(*AfterMCPExecutionPayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	// Count tokens in tool input and result (best effort, ignore errors)
	inputTokens, inputErr := h.deps.Tokenizer.CountTokens(h.deps.Config.DefaultModel, string(mcpPayload.ToolInput))
	if inputErr != nil {
		h.deps.Log.Debug("failed to count input tokens", "error", inputErr)
		inputTokens = 0
	}
	resultTokens, resultErr := h.deps.Tokenizer.CountTokens(h.deps.Config.DefaultModel, string(mcpPayload.ResultJSON))
	if resultErr != nil {
		h.deps.Log.Debug("failed to count result tokens", "error", resultErr)
		resultTokens = 0
	}
	totalTokens := inputTokens + resultTokens

	inputChars := CountChars(string(mcpPayload.ToolInput))
	resultChars := CountChars(string(mcpPayload.ResultJSON))

	h.deps.Log.Debug("tool metrics",
		"tool", mcpPayload.ToolName,
		"input_tokens", inputTokens,
		"result_tokens", resultTokens,
		"conversation_id", input.ConversationID,
	)

	// Create token usage record
	usage := &domain.TokenUsage{
		ID:               uuid.New(),
		ConversationID:   parseUUID(input.ConversationID),
		GenerationID:     parseUUID(input.GenerationID),
		Model:            h.deps.Config.DefaultModel,
		Stage:            domain.TokenUsageStageTool,
		PromptTokens:     inputTokens,
		CompletionTokens: resultTokens,
		TotalTokens:      totalTokens,
		PromptChars:      inputChars,
		CompletionChars:  resultChars,
		WorkspaceRoots:   input.WorkspaceRoots,
		UserEmail:        "",
		HookEventName:    string(input.HookEventName),
		Timestamp:        time.Now(),
	}

	if err := h.deps.Client.CreateTokenUsage(ctx, usage); err != nil {
		h.deps.Log.Error("failed to send token usage", "error", err)
	} else {
		h.deps.Log.Info("token usage recorded", "stage", "tool", "tool", mcpPayload.ToolName, "tokens", totalTokens)
	}

	return NewEmptyOutput(), nil
}

// AfterFileEditHandler handles afterFileEdit events.
type AfterFileEditHandler struct {
	deps *HandlerDependencies
}

// Handle processes the afterFileEdit event.
func (h *AfterFileEditHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	editPayload, ok := payload.(*AfterFileEditPayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	// Count characters in edits
	var oldChars, newChars int64
	for _, edit := range editPayload.Edits {
		oldChars += CountChars(edit.OldString)
		newChars += CountChars(edit.NewString)
	}

	h.deps.Log.Debug("file edit metrics",
		"file", editPayload.FilePath,
		"old_chars", oldChars,
		"new_chars", newChars,
		"edit_count", len(editPayload.Edits),
	)

	// Create token usage record for file edits
	usage := &domain.TokenUsage{
		ID:               uuid.New(),
		ConversationID:   parseUUID(input.ConversationID),
		GenerationID:     parseUUID(input.GenerationID),
		Model:            h.deps.Config.DefaultModel,
		Stage:            domain.TokenUsageStageFileEdit,
		PromptTokens:     0,
		CompletionTokens: 0,
		TotalTokens:      0,
		PromptChars:      oldChars,
		CompletionChars:  newChars,
		WorkspaceRoots:   input.WorkspaceRoots,
		UserEmail:        "",
		HookEventName:    string(input.HookEventName),
		Timestamp:        time.Now(),
	}

	if err := h.deps.Client.CreateTokenUsage(ctx, usage); err != nil {
		h.deps.Log.Error("failed to send token usage", "error", err)
	} else {
		h.deps.Log.Info("file edit recorded", "file", editPayload.FilePath, "edits", len(editPayload.Edits))
	}

	return NewEmptyOutput(), nil
}

// StopHandler handles stop events.
type StopHandler struct {
	deps *HandlerDependencies
}

// Handle processes the stop event.
//
//nolint:varnamelen // stopPayload is clear in context
func (h *StopHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	stopPayload, ok := payload.(*StopPayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	h.deps.Log.Info("agent loop stopped",
		"status", stopPayload.Status,
		"loop_count", stopPayload.LoopCount,
		"conversation_id", input.ConversationID,
	)

	// Create token usage record as session marker
	usage := &domain.TokenUsage{
		ID:               uuid.New(),
		ConversationID:   parseUUID(input.ConversationID),
		GenerationID:     parseUUID(input.GenerationID),
		Model:            h.deps.Config.DefaultModel,
		Stage:            domain.TokenUsageStageStop,
		PromptTokens:     0,
		CompletionTokens: 0,
		TotalTokens:      0,
		PromptChars:      0,
		CompletionChars:  0,
		WorkspaceRoots:   input.WorkspaceRoots,
		UserEmail:        "",
		HookEventName:    string(input.HookEventName),
		Timestamp:        time.Now(),
	}

	if err := h.deps.Client.CreateTokenUsage(ctx, usage); err != nil {
		h.deps.Log.Error("failed to send stop marker", "error", err)
	}

	return NewEmptyOutput(), nil
}

// BeforeShellExecutionHandler handles beforeShellExecution events.
type BeforeShellExecutionHandler struct {
	deps *HandlerDependencies
}

// Handle processes the beforeShellExecution event.
//
//nolint:varnamelen // shellPayload is clear in context
func (h *BeforeShellExecutionHandler) Handle(_ context.Context, input *HookInput, payload any) (*HookOutput, error) {
	shellPayload, ok := payload.(*BeforeShellExecutionPayload)
	if !ok {
		return NewAllowOutput(), nil
	}

	h.deps.Log.Debug("shell execution",
		"command", shellPayload.Command,
		"cwd", shellPayload.Cwd,
		"conversation_id", input.ConversationID,
	)

	// Always allow shell execution (observability only)
	return NewAllowOutput(), nil
}

// AfterShellExecutionHandler handles afterShellExecution events.
type AfterShellExecutionHandler struct {
	deps *HandlerDependencies
}

// Handle processes the afterShellExecution event.
//
//nolint:varnamelen // shellPayload is clear in context
func (h *AfterShellExecutionHandler) Handle(_ context.Context, input *HookInput, payload any) (*HookOutput, error) {
	shellPayload, ok := payload.(*AfterShellExecutionPayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	h.deps.Log.Debug("shell execution completed",
		"command", shellPayload.Command,
		"output_length", len(shellPayload.Output),
		"conversation_id", input.ConversationID,
	)

	return NewEmptyOutput(), nil
}

// BeforeMCPExecutionHandler handles beforeMCPExecution events.
type BeforeMCPExecutionHandler struct {
	deps *HandlerDependencies
}

// Handle processes the beforeMCPExecution event.
//
//nolint:varnamelen // mcpPayload is clear in context
func (h *BeforeMCPExecutionHandler) Handle(_ context.Context, input *HookInput, payload any) (*HookOutput, error) {
	mcpPayload, ok := payload.(*BeforeMCPExecutionPayload)
	if !ok {
		return NewAllowOutput(), nil
	}

	h.deps.Log.Debug("MCP execution",
		"tool", mcpPayload.ToolName,
		"conversation_id", input.ConversationID,
	)

	// Always allow MCP execution (observability only)
	return NewAllowOutput(), nil
}

// BeforeReadFileHandler handles beforeReadFile events.
type BeforeReadFileHandler struct {
	deps *HandlerDependencies
}

// Handle processes the beforeReadFile event.
//
//nolint:varnamelen // readPayload is clear in context
func (h *BeforeReadFileHandler) Handle(_ context.Context, input *HookInput, payload any) (*HookOutput, error) {
	readPayload, ok := payload.(*BeforeReadFilePayload)
	if !ok {
		return NewAllowOutput(), nil
	}

	h.deps.Log.Debug("file read",
		"file", readPayload.FilePath,
		"content_length", len(readPayload.Content),
		"conversation_id", input.ConversationID,
	)

	// Always allow file read (observability only)
	return NewAllowOutput(), nil
}

// parseUUID parses a UUID string, returning a zero UUID on error.
func parseUUID(s string) uuid.UUID {
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil
	}
	return id
}


```

```go
package hook

import (
	"context"
	"fmt"
	"time"

	"github.com/ainslies/ctd/internal/domain"
	"github.com/ainslies/ctd/pkg/logger"
	"github.com/google/uuid"
)

// Handler processes a hook event and returns the appropriate output.
type Handler interface {
	// Handle processes the hook input and returns the output.
	Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error)
}

// HandlerDependencies contains the dependencies for handlers.
type HandlerDependencies struct {
	Client    *CTDClient
	Tokenizer Tokenizer
	Log       *logger.Logger
	Config    *Config
}

// Router routes hook events to the appropriate handler.
type Router struct {
	deps     *HandlerDependencies
	handlers map[HookEventName]Handler
}

// NewRouter creates a new router with all handlers registered.
func NewRouter(deps *HandlerDependencies) *Router {
	router := &Router{
		deps:     deps,
		handlers: make(map[HookEventName]Handler),
	}

	// Register handlers
	router.handlers[EventBeforeSubmitPrompt] = &BeforeSubmitPromptHandler{deps: deps}
	router.handlers[EventAfterAgentResponse] = &AfterAgentResponseHandler{deps: deps}
	router.handlers[EventAfterAgentThought] = &AfterAgentThoughtHandler{deps: deps}
	router.handlers[EventAfterMCPExecution] = &AfterMCPExecutionHandler{deps: deps}
	router.handlers[EventAfterFileEdit] = &AfterFileEditHandler{deps: deps}
	router.handlers[EventStop] = &StopHandler{deps: deps}
	router.handlers[EventBeforeShellExecution] = &BeforeShellExecutionHandler{deps: deps}
	router.handlers[EventAfterShellExecution] = &AfterShellExecutionHandler{deps: deps}
	router.handlers[EventBeforeMCPExecution] = &BeforeMCPExecutionHandler{deps: deps}
	router.handlers[EventBeforeReadFile] = &BeforeReadFileHandler{deps: deps}

	return router
}

// Route routes the input to the appropriate handler.
func (r *Router) Route(ctx context.Context, input *HookInput) (*HookOutput, error) {
	handler, ok := r.handlers[input.HookEventName]
	if !ok {
		// Unknown event, return empty output
		r.deps.Log.Warn("unknown hook event", "event", input.HookEventName)
		return NewEmptyOutput(), nil
	}

	payload, err := input.ParsePayload()
	if err != nil {
		r.deps.Log.Error("failed to parse payload", "error", err, "event", input.HookEventName)
		// Return appropriate default output even on parse error
		return getDefaultOutput(input.HookEventName), nil
	}

	output, err := handler.Handle(ctx, input, payload)
	if err != nil {
		return nil, fmt.Errorf("handler error for %s: %w", input.HookEventName, err)
	}
	return output, nil
}

// getDefaultOutput returns the default output for an event type.
//
//nolint:exhaustive // Only specific events need special handling, others use empty output
func getDefaultOutput(event HookEventName) *HookOutput {
	switch event {
	case EventBeforeSubmitPrompt:
		return NewContinueOutput()
	case EventBeforeShellExecution, EventBeforeMCPExecution, EventBeforeReadFile:
		return NewAllowOutput()
	default:
		return NewEmptyOutput()
	}
}

// BeforeSubmitPromptHandler handles beforeSubmitPrompt events.
type BeforeSubmitPromptHandler struct {
	deps *HandlerDependencies
}

// Handle processes the beforeSubmitPrompt event.
func (h *BeforeSubmitPromptHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	promptPayload, ok := payload.(*BeforeSubmitPromptPayload)
	if !ok {
		return NewContinueOutput(), nil
	}

	// Count tokens
	tokenCount, err := h.deps.Tokenizer.CountTokens(h.deps.Config.DefaultModel, promptPayload.Prompt)
	if err != nil {
		h.deps.Log.Error("failed to count tokens", "error", err)
		tokenCount = 0
	}

	charCount := CountChars(promptPayload.Prompt)

	h.deps.Log.Debug("prompt metrics",
		"tokens", tokenCount,
		"chars", charCount,
		"conversation_id", input.ConversationID,
	)

	// Create token usage record
	usage := &domain.TokenUsage{
		ID:               uuid.New(),
		ConversationID:   parseUUID(input.ConversationID),
		GenerationID:     parseUUID(input.GenerationID),
		Model:            h.deps.Config.DefaultModel,
		Stage:            domain.TokenUsageStagePrompt,
		PromptTokens:     tokenCount,
		CompletionTokens: 0,
		TotalTokens:      tokenCount,
		PromptChars:      charCount,
		CompletionChars:  0,
		WorkspaceRoots:   input.WorkspaceRoots,
		UserEmail:        "",
		HookEventName:    string(input.HookEventName),
		Timestamp:        time.Now(),
	}

	// Send to ctd (best-effort)
	if err := h.deps.Client.CreateTokenUsage(ctx, usage); err != nil {
		h.deps.Log.Error("failed to send token usage", "error", err)
	} else {
		h.deps.Log.Info("token usage recorded", "stage", "prompt", "tokens", tokenCount)
	}

	return NewContinueOutput(), nil
}

// AfterAgentResponseHandler handles afterAgentResponse events.
type AfterAgentResponseHandler struct {
	deps *HandlerDependencies
}

// Handle processes the afterAgentResponse event.
func (h *AfterAgentResponseHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	responsePayload, ok := payload.(*AfterAgentResponsePayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	// Count tokens
	tokenCount, err := h.deps.Tokenizer.CountTokens(h.deps.Config.DefaultModel, responsePayload.Text)
	if err != nil {
		h.deps.Log.Error("failed to count tokens", "error", err)
		tokenCount = 0
	}

	charCount := CountChars(responsePayload.Text)

	h.deps.Log.Debug("completion metrics",
		"tokens", tokenCount,
		"chars", charCount,
		"conversation_id", input.ConversationID,
	)

	// Create token usage record
	usage := &domain.TokenUsage{
		ID:               uuid.New(),
		ConversationID:   parseUUID(input.ConversationID),
		GenerationID:     parseUUID(input.GenerationID),
		Model:            h.deps.Config.DefaultModel,
		Stage:            domain.TokenUsageStageCompletion,
		PromptTokens:     0,
		CompletionTokens: tokenCount,
		TotalTokens:      tokenCount,
		PromptChars:      0,
		CompletionChars:  charCount,
		WorkspaceRoots:   input.WorkspaceRoots,
		UserEmail:        "",
		HookEventName:    string(input.HookEventName),
		Timestamp:        time.Now(),
	}

	// Send token usage to ctd (best-effort)
	if err := h.deps.Client.CreateTokenUsage(ctx, usage); err != nil {
		h.deps.Log.Error("failed to send token usage", "error", err)
	} else {
		h.deps.Log.Info("token usage recorded", "stage", "completion", "tokens", tokenCount)
	}

	// Also send as thought for visibility
	thought := &domain.Thought{
		ID:             uuid.New(),
		ConversationID: parseUUID(input.ConversationID),
		GenerationID:   parseUUID(input.GenerationID),
		Text:           responsePayload.Text,
		DurationMs:     0,
		Model:          h.deps.Config.DefaultModel,
		HookEventName:  string(input.HookEventName),
		CursorVersion:  "",
		WorkspaceRoots: input.WorkspaceRoots,
		UserEmail:      "",
		Timestamp:      time.Now(),
	}

	if err := h.deps.Client.CreateThought(ctx, thought); err != nil {
		h.deps.Log.Error("failed to send thought", "error", err)
	}

	return NewEmptyOutput(), nil
}

// AfterAgentThoughtHandler handles afterAgentThought events.
type AfterAgentThoughtHandler struct {
	deps *HandlerDependencies
}

// Handle processes the afterAgentThought event.
func (h *AfterAgentThoughtHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	thoughtPayload, ok := payload.(*AfterAgentThoughtPayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	// Send as thought for visibility
	thought := &domain.Thought{
		ID:             uuid.New(),
		ConversationID: parseUUID(input.ConversationID),
		GenerationID:   parseUUID(input.GenerationID),
		Text:           thoughtPayload.Text,
		DurationMs:     0,
		Model:          h.deps.Config.DefaultModel,
		HookEventName:  string(input.HookEventName),
		CursorVersion:  "",
		WorkspaceRoots: input.WorkspaceRoots,
		UserEmail:      "",
		Timestamp:      time.Now(),
	}

	if err := h.deps.Client.CreateThought(ctx, thought); err != nil {
		h.deps.Log.Error("failed to send thought", "error", err)
	} else {
		h.deps.Log.Info("thought recorded", "event", input.HookEventName)
	}

	return NewEmptyOutput(), nil
}

// AfterMCPExecutionHandler handles afterMCPExecution events.
type AfterMCPExecutionHandler struct {
	deps *HandlerDependencies
}

// Handle processes the afterMCPExecution event.
func (h *AfterMCPExecutionHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	mcpPayload, ok := payload.(*AfterMCPExecutionPayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	// Count tokens in tool input and result (best effort, ignore errors)
	inputTokens, inputErr := h.deps.Tokenizer.CountTokens(h.deps.Config.DefaultModel, string(mcpPayload.ToolInput))
	if inputErr != nil {
		h.deps.Log.Debug("failed to count input tokens", "error", inputErr)
		inputTokens = 0
	}
	resultTokens, resultErr := h.deps.Tokenizer.CountTokens(h.deps.Config.DefaultModel, string(mcpPayload.ResultJSON))
	if resultErr != nil {
		h.deps.Log.Debug("failed to count result tokens", "error", resultErr)
		resultTokens = 0
	}
	totalTokens := inputTokens + resultTokens

	inputChars := CountChars(string(mcpPayload.ToolInput))
	resultChars := CountChars(string(mcpPayload.ResultJSON))

	h.deps.Log.Debug("tool metrics",
		"tool", mcpPayload.ToolName,
		"input_tokens", inputTokens,
		"result_tokens", resultTokens,
		"conversation_id", input.ConversationID,
	)

	// Create token usage record
	usage := &domain.TokenUsage{
		ID:               uuid.New(),
		ConversationID:   parseUUID(input.ConversationID),
		GenerationID:     parseUUID(input.GenerationID),
		Model:            h.deps.Config.DefaultModel,
		Stage:            domain.TokenUsageStageTool,
		PromptTokens:     inputTokens,
		CompletionTokens: resultTokens,
		TotalTokens:      totalTokens,
		PromptChars:      inputChars,
		CompletionChars:  resultChars,
		WorkspaceRoots:   input.WorkspaceRoots,
		UserEmail:        "",
		HookEventName:    string(input.HookEventName),
		Timestamp:        time.Now(),
	}

	if err := h.deps.Client.CreateTokenUsage(ctx, usage); err != nil {
		h.deps.Log.Error("failed to send token usage", "error", err)
	} else {
		h.deps.Log.Info("token usage recorded", "stage", "tool", "tool", mcpPayload.ToolName, "tokens", totalTokens)
	}

	return NewEmptyOutput(), nil
}

// AfterFileEditHandler handles afterFileEdit events.
type AfterFileEditHandler struct {
	deps *HandlerDependencies
}

// Handle processes the afterFileEdit event.
func (h *AfterFileEditHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	editPayload, ok := payload.(*AfterFileEditPayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	// Count characters in edits
	var oldChars, newChars int64
	for _, edit := range editPayload.Edits {
		oldChars += CountChars(edit.OldString)
		newChars += CountChars(edit.NewString)
	}

	h.deps.Log.Debug("file edit metrics",
		"file", editPayload.FilePath,
		"old_chars", oldChars,
		"new_chars", newChars,
		"edit_count", len(editPayload.Edits),
	)

	// Create token usage record for file edits
	usage := &domain.TokenUsage{
		ID:               uuid.New(),
		ConversationID:   parseUUID(input.ConversationID),
		GenerationID:     parseUUID(input.GenerationID),
		Model:            h.deps.Config.DefaultModel,
		Stage:            domain.TokenUsageStageFileEdit,
		PromptTokens:     0,
		CompletionTokens: 0,
		TotalTokens:      0,
		PromptChars:      oldChars,
		CompletionChars:  newChars,
		WorkspaceRoots:   input.WorkspaceRoots,
		UserEmail:        "",
		HookEventName:    string(input.HookEventName),
		Timestamp:        time.Now(),
	}

	if err := h.deps.Client.CreateTokenUsage(ctx, usage); err != nil {
		h.deps.Log.Error("failed to send token usage", "error", err)
	} else {
		h.deps.Log.Info("file edit recorded", "file", editPayload.FilePath, "edits", len(editPayload.Edits))
	}

	return NewEmptyOutput(), nil
}

// StopHandler handles stop events.
type StopHandler struct {
	deps *HandlerDependencies
}

// Handle processes the stop event.
func (h *StopHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	stopPayload, ok := payload.(*StopPayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	h.deps.Log.Info("agent loop stopped",
		"status", stopPayload.Status,
		"loop_count", stopPayload.LoopCount,
		"conversation_id", input.ConversationID,
	)

	// Create token usage record as session marker
	usage := &domain.TokenUsage{
		ID:               uuid.New(),
		ConversationID:   parseUUID(input.ConversationID),
		GenerationID:     parseUUID(input.GenerationID),
		Model:            h.deps.Config.DefaultModel,
		Stage:            domain.TokenUsageStageStop,
		PromptTokens:     0,
		CompletionTokens: 0,
		TotalTokens:      0,
		PromptChars:      0,
		CompletionChars:  0,
		WorkspaceRoots:   input.WorkspaceRoots,
		UserEmail:        "",
		HookEventName:    string(input.HookEventName),
		Timestamp:        time.Now(),
	}

	if err := h.deps.Client.CreateTokenUsage(ctx, usage); err != nil {
		h.deps.Log.Error("failed to send stop marker", "error", err)
	}

	return NewEmptyOutput(), nil
}

// BeforeShellExecutionHandler handles beforeShellExecution events.
type BeforeShellExecutionHandler struct {
	deps *HandlerDependencies
}

// Handle processes the beforeShellExecution event.
//
//nolint:varnamelen // shellPayload is clear in context
func (h *BeforeShellExecutionHandler) Handle(_ context.Context, input *HookInput, payload any) (*HookOutput, error) {
	shellPayload, ok := payload.(*BeforeShellExecutionPayload)
	if !ok {
		return NewAllowOutput(), nil
	}

	h.deps.Log.Debug("shell execution",
		"command", shellPayload.Command,
		"cwd", shellPayload.Cwd,
		"conversation_id", input.ConversationID,
	)

	// Always allow shell execution (observability only)
	return NewAllowOutput(), nil
}

// AfterShellExecutionHandler handles afterShellExecution events.
type AfterShellExecutionHandler struct {
	deps *HandlerDependencies
}

// Handle processes the afterShellExecution event.
//
//nolint:varnamelen // shellPayload is clear in context
func (h *AfterShellExecutionHandler) Handle(_ context.Context, input *HookInput, payload any) (*HookOutput, error) {
	shellPayload, ok := payload.(*AfterShellExecutionPayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	h.deps.Log.Debug("shell execution completed",
		"command", shellPayload.Command,
		"output_length", len(shellPayload.Output),
		"conversation_id", input.ConversationID,
	)

	return NewEmptyOutput(), nil
}

// BeforeMCPExecutionHandler handles beforeMCPExecution events.
type BeforeMCPExecutionHandler struct {
	deps *HandlerDependencies
}

// Handle processes the beforeMCPExecution event.
//
//nolint:varnamelen // mcpPayload is clear in context
func (h *BeforeMCPExecutionHandler) Handle(_ context.Context, input *HookInput, payload any) (*HookOutput, error) {
	mcpPayload, ok := payload.(*BeforeMCPExecutionPayload)
	if !ok {
		return NewAllowOutput(), nil
	}

	h.deps.Log.Debug("MCP execution",
		"tool", mcpPayload.ToolName,
		"conversation_id", input.ConversationID,
	)

	// Always allow MCP execution (observability only)
	return NewAllowOutput(), nil
}

// BeforeReadFileHandler handles beforeReadFile events.
type BeforeReadFileHandler struct {
	deps *HandlerDependencies
}

// Handle processes the beforeReadFile event.
//
//nolint:varnamelen // readPayload is clear in context
func (h *BeforeReadFileHandler) Handle(_ context.Context, input *HookInput, payload any) (*HookOutput, error) {
	readPayload, ok := payload.(*BeforeReadFilePayload)
	if !ok {
		return NewAllowOutput(), nil
	}

	h.deps.Log.Debug("file read",
		"file", readPayload.FilePath,
		"content_length", len(readPayload.Content),
		"conversation_id", input.ConversationID,
	)

	// Always allow file read (observability only)
	return NewAllowOutput(), nil
}

// parseUUID parses a UUID string, returning a zero UUID on error.
func parseUUID(s string) uuid.UUID {
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil
	}
	return id
}


```

```go
package hook

import (
	"context"
	"fmt"
	"time"

	"github.com/ainslies/ctd/internal/domain"
	"github.com/ainslies/ctd/pkg/logger"
	"github.com/google/uuid"
)

// Handler processes a hook event and returns the appropriate output.
type Handler interface {
	// Handle processes the hook input and returns the output.
	Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error)
}

// HandlerDependencies contains the dependencies for handlers.
type HandlerDependencies struct {
	Client    *CTDClient
	Tokenizer Tokenizer
	Log       *logger.Logger
	Config    *Config
}

// Router routes hook events to the appropriate handler.
type Router struct {
	deps     *HandlerDependencies
	handlers map[HookEventName]Handler
}

// NewRouter creates a new router with all handlers registered.
func NewRouter(deps *HandlerDependencies) *Router {
	router := &Router{
		deps:     deps,
		handlers: make(map[HookEventName]Handler),
	}

	// Register handlers
	router.handlers[EventBeforeSubmitPrompt] = &BeforeSubmitPromptHandler{deps: deps}
	router.handlers[EventAfterAgentResponse] = &AfterAgentResponseHandler{deps: deps}
	router.handlers[EventAfterAgentThought] = &AfterAgentThoughtHandler{deps: deps}
	router.handlers[EventAfterMCPExecution] = &AfterMCPExecutionHandler{deps: deps}
	router.handlers[EventAfterFileEdit] = &AfterFileEditHandler{deps: deps}
	router.handlers[EventStop] = &StopHandler{deps: deps}
	router.handlers[EventBeforeShellExecution] = &BeforeShellExecutionHandler{deps: deps}
	router.handlers[EventAfterShellExecution] = &AfterShellExecutionHandler{deps: deps}
	router.handlers[EventBeforeMCPExecution] = &BeforeMCPExecutionHandler{deps: deps}
	router.handlers[EventBeforeReadFile] = &BeforeReadFileHandler{deps: deps}

	return router
}

// Route routes the input to the appropriate handler.
func (r *Router) Route(ctx context.Context, input *HookInput) (*HookOutput, error) {
	handler, ok := r.handlers[input.HookEventName]
	if !ok {
		// Unknown event, return empty output
		r.deps.Log.Warn("unknown hook event", "event", input.HookEventName)
		return NewEmptyOutput(), nil
	}

	payload, err := input.ParsePayload()
	if err != nil {
		r.deps.Log.Error("failed to parse payload", "error", err, "event", input.HookEventName)
		// Return appropriate default output even on parse error
		return getDefaultOutput(input.HookEventName), nil
	}

	output, err := handler.Handle(ctx, input, payload)
	if err != nil {
		return nil, fmt.Errorf("handler error for %s: %w", input.HookEventName, err)
	}
	return output, nil
}

// getDefaultOutput returns the default output for an event type.
//
//nolint:exhaustive // Only specific events need special handling, others use empty output
func getDefaultOutput(event HookEventName) *HookOutput {
	switch event {
	case EventBeforeSubmitPrompt:
		return NewContinueOutput()
	case EventBeforeShellExecution, EventBeforeMCPExecution, EventBeforeReadFile:
		return NewAllowOutput()
	default:
		return NewEmptyOutput()
	}
}

// BeforeSubmitPromptHandler handles beforeSubmitPrompt events.
type BeforeSubmitPromptHandler struct {
	deps *HandlerDependencies
}

// Handle processes the beforeSubmitPrompt event.
func (h *BeforeSubmitPromptHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	promptPayload, ok := payload.(*BeforeSubmitPromptPayload)
	if !ok {
		return NewContinueOutput(), nil
	}

	// Count tokens
	tokenCount, err := h.deps.Tokenizer.CountTokens(h.deps.Config.DefaultModel, promptPayload.Prompt)
	if err != nil {
		h.deps.Log.Error("failed to count tokens", "error", err)
		tokenCount = 0
	}

	charCount := CountChars(promptPayload.Prompt)

	h.deps.Log.Debug("prompt metrics",
		"tokens", tokenCount,
		"chars", charCount,
		"conversation_id", input.ConversationID,
	)

	// Create token usage record
	usage := &domain.TokenUsage{
		ID:               uuid.New(),
		ConversationID:   parseUUID(input.ConversationID),
		GenerationID:     parseUUID(input.GenerationID),
		Model:            h.deps.Config.DefaultModel,
		Stage:            domain.TokenUsageStagePrompt,
		PromptTokens:     tokenCount,
		CompletionTokens: 0,
		TotalTokens:      tokenCount,
		PromptChars:      charCount,
		CompletionChars:  0,
		WorkspaceRoots:   input.WorkspaceRoots,
		UserEmail:        "",
		HookEventName:    string(input.HookEventName),
		Timestamp:        time.Now(),
	}

	// Send to ctd (best-effort)
	if err := h.deps.Client.CreateTokenUsage(ctx, usage); err != nil {
		h.deps.Log.Error("failed to send token usage", "error", err)
	} else {
		h.deps.Log.Info("token usage recorded", "stage", "prompt", "tokens", tokenCount)
	}

	return NewContinueOutput(), nil
}

// AfterAgentResponseHandler handles afterAgentResponse events.
type AfterAgentResponseHandler struct {
	deps *HandlerDependencies
}

// Handle processes the afterAgentResponse event.
func (h *AfterAgentResponseHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	responsePayload, ok := payload.(*AfterAgentResponsePayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	// Count tokens
	tokenCount, err := h.deps.Tokenizer.CountTokens(h.deps.Config.DefaultModel, responsePayload.Text)
	if err != nil {
		h.deps.Log.Error("failed to count tokens", "error", err)
		tokenCount = 0
	}

	charCount := CountChars(responsePayload.Text)

	h.deps.Log.Debug("completion metrics",
		"tokens", tokenCount,
		"chars", charCount,
		"conversation_id", input.ConversationID,
	)

	// Create token usage record
	usage := &domain.TokenUsage{
		ID:               uuid.New(),
		ConversationID:   parseUUID(input.ConversationID),
		GenerationID:     parseUUID(input.GenerationID),
		Model:            h.deps.Config.DefaultModel,
		Stage:            domain.TokenUsageStageCompletion,
		PromptTokens:     0,
		CompletionTokens: tokenCount,
		TotalTokens:      tokenCount,
		PromptChars:      0,
		CompletionChars:  charCount,
		WorkspaceRoots:   input.WorkspaceRoots,
		UserEmail:        "",
		HookEventName:    string(input.HookEventName),
		Timestamp:        time.Now(),
	}

	// Send token usage to ctd (best-effort)
	if err := h.deps.Client.CreateTokenUsage(ctx, usage); err != nil {
		h.deps.Log.Error("failed to send token usage", "error", err)
	} else {
		h.deps.Log.Info("token usage recorded", "stage", "completion", "tokens", tokenCount)
	}

	// Also send as thought for visibility
	thought := &domain.Thought{
		ID:             uuid.New(),
		ConversationID: parseUUID(input.ConversationID),
		GenerationID:   parseUUID(input.GenerationID),
		Text:           responsePayload.Text,
		DurationMs:     0,
		Model:          h.deps.Config.DefaultModel,
		HookEventName:  string(input.HookEventName),
		CursorVersion:  "",
		WorkspaceRoots: input.WorkspaceRoots,
		UserEmail:      "",
		Timestamp:      time.Now(),
	}

	if err := h.deps.Client.CreateThought(ctx, thought); err != nil {
		h.deps.Log.Error("failed to send thought", "error", err)
	}

	return NewEmptyOutput(), nil
}

// AfterAgentThoughtHandler handles afterAgentThought events.
type AfterAgentThoughtHandler struct {
	deps *HandlerDependencies
}

// Handle processes the afterAgentThought event.
func (h *AfterAgentThoughtHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	thoughtPayload, ok := payload.(*AfterAgentThoughtPayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	// Send as thought for visibility
	thought := &domain.Thought{
		ID:             uuid.New(),
		ConversationID: parseUUID(input.ConversationID),
		GenerationID:   parseUUID(input.GenerationID),
		Text:           thoughtPayload.Text,
		DurationMs:     0,
		Model:          h.deps.Config.DefaultModel,
		HookEventName:  string(input.HookEventName),
		CursorVersion:  "",
		WorkspaceRoots: input.WorkspaceRoots,
		UserEmail:      "",
		Timestamp:      time.Now(),
	}

	if err := h.deps.Client.CreateThought(ctx, thought); err != nil {
		h.deps.Log.Error("failed to send thought", "error", err)
	} else {
		h.deps.Log.Info("thought recorded", "event", input.HookEventName)
	}

	return NewEmptyOutput(), nil
}

// AfterMCPExecutionHandler handles afterMCPExecution events.
type AfterMCPExecutionHandler struct {
	deps *HandlerDependencies
}

// Handle processes the afterMCPExecution event.
func (h *AfterMCPExecutionHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	mcpPayload, ok := payload.(*AfterMCPExecutionPayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	// Count tokens in tool input and result (best effort, ignore errors)
	inputTokens, inputErr := h.deps.Tokenizer.CountTokens(h.deps.Config.DefaultModel, string(mcpPayload.ToolInput))
	if inputErr != nil {
		h.deps.Log.Debug("failed to count input tokens", "error", inputErr)
		inputTokens = 0
	}
	resultTokens, resultErr := h.deps.Tokenizer.CountTokens(h.deps.Config.DefaultModel, string(mcpPayload.ResultJSON))
	if resultErr != nil {
		h.deps.Log.Debug("failed to count result tokens", "error", resultErr)
		resultTokens = 0
	}
	totalTokens := inputTokens + resultTokens

	inputChars := CountChars(string(mcpPayload.ToolInput))
	resultChars := CountChars(string(mcpPayload.ResultJSON))

	h.deps.Log.Debug("tool metrics",
		"tool", mcpPayload.ToolName,
		"input_tokens", inputTokens,
		"result_tokens", resultTokens,
		"conversation_id", input.ConversationID,
	)

	// Create token usage record
	usage := &domain.TokenUsage{
		ID:               uuid.New(),
		ConversationID:   parseUUID(input.ConversationID),
		GenerationID:     parseUUID(input.GenerationID),
		Model:            h.deps.Config.DefaultModel,
		Stage:            domain.TokenUsageStageTool,
		PromptTokens:     inputTokens,
		CompletionTokens: resultTokens,
		TotalTokens:      totalTokens,
		PromptChars:      inputChars,
		CompletionChars:  resultChars,
		WorkspaceRoots:   input.WorkspaceRoots,
		UserEmail:        "",
		HookEventName:    string(input.HookEventName),
		Timestamp:        time.Now(),
	}

	if err := h.deps.Client.CreateTokenUsage(ctx, usage); err != nil {
		h.deps.Log.Error("failed to send token usage", "error", err)
	} else {
		h.deps.Log.Info("token usage recorded", "stage", "tool", "tool", mcpPayload.ToolName, "tokens", totalTokens)
	}

	return NewEmptyOutput(), nil
}

// AfterFileEditHandler handles afterFileEdit events.
type AfterFileEditHandler struct {
	deps *HandlerDependencies
}

// Handle processes the afterFileEdit event.
func (h *AfterFileEditHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	editPayload, ok := payload.(*AfterFileEditPayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	// Count characters in edits
	var oldChars, newChars int64
	for _, edit := range editPayload.Edits {
		oldChars += CountChars(edit.OldString)
		newChars += CountChars(edit.NewString)
	}

	h.deps.Log.Debug("file edit metrics",
		"file", editPayload.FilePath,
		"old_chars", oldChars,
		"new_chars", newChars,
		"edit_count", len(editPayload.Edits),
	)

	// Create token usage record for file edits
	usage := &domain.TokenUsage{
		ID:               uuid.New(),
		ConversationID:   parseUUID(input.ConversationID),
		GenerationID:     parseUUID(input.GenerationID),
		Model:            h.deps.Config.DefaultModel,
		Stage:            domain.TokenUsageStageFileEdit,
		PromptTokens:     0,
		CompletionTokens: 0,
		TotalTokens:      0,
		PromptChars:      oldChars,
		CompletionChars:  newChars,
		WorkspaceRoots:   input.WorkspaceRoots,
		UserEmail:        "",
		HookEventName:    string(input.HookEventName),
		Timestamp:        time.Now(),
	}

	if err := h.deps.Client.CreateTokenUsage(ctx, usage); err != nil {
		h.deps.Log.Error("failed to send token usage", "error", err)
	} else {
		h.deps.Log.Info("file edit recorded", "file", editPayload.FilePath, "edits", len(editPayload.Edits))
	}

	return NewEmptyOutput(), nil
}

// StopHandler handles stop events.
type StopHandler struct {
	deps *HandlerDependencies
}

// Handle processes the stop event.
func (h *StopHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	stopPayload, ok := payload.(*StopPayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	h.deps.Log.Info("agent loop stopped",
		"status", stopPayload.Status,
		"loop_count", stopPayload.LoopCount,
		"conversation_id", input.ConversationID,
	)

	// Create token usage record as session marker
	usage := &domain.TokenUsage{
		ID:               uuid.New(),
		ConversationID:   parseUUID(input.ConversationID),
		GenerationID:     parseUUID(input.GenerationID),
		Model:            h.deps.Config.DefaultModel,
		Stage:            domain.TokenUsageStageStop,
		PromptTokens:     0,
		CompletionTokens: 0,
		TotalTokens:      0,
		PromptChars:      0,
		CompletionChars:  0,
		WorkspaceRoots:   input.WorkspaceRoots,
		UserEmail:        "",
		HookEventName:    string(input.HookEventName),
		Timestamp:        time.Now(),
	}

	if err := h.deps.Client.CreateTokenUsage(ctx, usage); err != nil {
		h.deps.Log.Error("failed to send stop marker", "error", err)
	}

	return NewEmptyOutput(), nil
}

// BeforeShellExecutionHandler handles beforeShellExecution events.
type BeforeShellExecutionHandler struct {
	deps *HandlerDependencies
}

// Handle processes the beforeShellExecution event.
func (h *BeforeShellExecutionHandler) Handle(_ context.Context, input *HookInput, payload any) (*HookOutput, error) {
	shellPayload, ok := payload.(*BeforeShellExecutionPayload)
	if !ok {
		return NewAllowOutput(), nil
	}

	h.deps.Log.Debug("shell execution",
		"command", shellPayload.Command,
		"cwd", shellPayload.Cwd,
		"conversation_id", input.ConversationID,
	)

	// Always allow shell execution (observability only)
	return NewAllowOutput(), nil
}

// AfterShellExecutionHandler handles afterShellExecution events.
type AfterShellExecutionHandler struct {
	deps *HandlerDependencies
}

// Handle processes the afterShellExecution event.
//
//nolint:varnamelen // shellPayload is clear in context
func (h *AfterShellExecutionHandler) Handle(_ context.Context, input *HookInput, payload any) (*HookOutput, error) {
	shellPayload, ok := payload.(*AfterShellExecutionPayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	h.deps.Log.Debug("shell execution completed",
		"command", shellPayload.Command,
		"output_length", len(shellPayload.Output),
		"conversation_id", input.ConversationID,
	)

	return NewEmptyOutput(), nil
}

// BeforeMCPExecutionHandler handles beforeMCPExecution events.
type BeforeMCPExecutionHandler struct {
	deps *HandlerDependencies
}

// Handle processes the beforeMCPExecution event.
//
//nolint:varnamelen // mcpPayload is clear in context
func (h *BeforeMCPExecutionHandler) Handle(_ context.Context, input *HookInput, payload any) (*HookOutput, error) {
	mcpPayload, ok := payload.(*BeforeMCPExecutionPayload)
	if !ok {
		return NewAllowOutput(), nil
	}

	h.deps.Log.Debug("MCP execution",
		"tool", mcpPayload.ToolName,
		"conversation_id", input.ConversationID,
	)

	// Always allow MCP execution (observability only)
	return NewAllowOutput(), nil
}

// BeforeReadFileHandler handles beforeReadFile events.
type BeforeReadFileHandler struct {
	deps *HandlerDependencies
}

// Handle processes the beforeReadFile event.
//
//nolint:varnamelen // readPayload is clear in context
func (h *BeforeReadFileHandler) Handle(_ context.Context, input *HookInput, payload any) (*HookOutput, error) {
	readPayload, ok := payload.(*BeforeReadFilePayload)
	if !ok {
		return NewAllowOutput(), nil
	}

	h.deps.Log.Debug("file read",
		"file", readPayload.FilePath,
		"content_length", len(readPayload.Content),
		"conversation_id", input.ConversationID,
	)

	// Always allow file read (observability only)
	return NewAllowOutput(), nil
}

// parseUUID parses a UUID string, returning a zero UUID on error.
func parseUUID(s string) uuid.UUID {
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil
	}
	return id
}


```

```go
package hook

import (
	"context"
	"fmt"
	"time"

	"github.com/ainslies/ctd/internal/domain"
	"github.com/ainslies/ctd/pkg/logger"
	"github.com/google/uuid"
)

// Handler processes a hook event and returns the appropriate output.
type Handler interface {
	// Handle processes the hook input and returns the output.
	Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error)
}

// HandlerDependencies contains the dependencies for handlers.
type HandlerDependencies struct {
	Client    *CTDClient
	Tokenizer Tokenizer
	Log       *logger.Logger
	Config    *Config
}

// Router routes hook events to the appropriate handler.
type Router struct {
	deps     *HandlerDependencies
	handlers map[HookEventName]Handler
}

// NewRouter creates a new router with all handlers registered.
func NewRouter(deps *HandlerDependencies) *Router {
	router := &Router{
		deps:     deps,
		handlers: make(map[HookEventName]Handler),
	}

	// Register handlers
	router.handlers[EventBeforeSubmitPrompt] = &BeforeSubmitPromptHandler{deps: deps}
	router.handlers[EventAfterAgentResponse] = &AfterAgentResponseHandler{deps: deps}
	router.handlers[EventAfterAgentThought] = &AfterAgentThoughtHandler{deps: deps}
	router.handlers[EventAfterMCPExecution] = &AfterMCPExecutionHandler{deps: deps}
	router.handlers[EventAfterFileEdit] = &AfterFileEditHandler{deps: deps}
	router.handlers[EventStop] = &StopHandler{deps: deps}
	router.handlers[EventBeforeShellExecution] = &BeforeShellExecutionHandler{deps: deps}
	router.handlers[EventAfterShellExecution] = &AfterShellExecutionHandler{deps: deps}
	router.handlers[EventBeforeMCPExecution] = &BeforeMCPExecutionHandler{deps: deps}
	router.handlers[EventBeforeReadFile] = &BeforeReadFileHandler{deps: deps}

	return router
}

// Route routes the input to the appropriate handler.
func (r *Router) Route(ctx context.Context, input *HookInput) (*HookOutput, error) {
	handler, ok := r.handlers[input.HookEventName]
	if !ok {
		// Unknown event, return empty output
		r.deps.Log.Warn("unknown hook event", "event", input.HookEventName)
		return NewEmptyOutput(), nil
	}

	payload, err := input.ParsePayload()
	if err != nil {
		r.deps.Log.Error("failed to parse payload", "error", err, "event", input.HookEventName)
		// Return appropriate default output even on parse error
		return getDefaultOutput(input.HookEventName), nil
	}

	output, err := handler.Handle(ctx, input, payload)
	if err != nil {
		return nil, fmt.Errorf("handler error for %s: %w", input.HookEventName, err)
	}
	return output, nil
}

// getDefaultOutput returns the default output for an event type.
//
//nolint:exhaustive // Only specific events need special handling, others use empty output
func getDefaultOutput(event HookEventName) *HookOutput {
	switch event {
	case EventBeforeSubmitPrompt:
		return NewContinueOutput()
	case EventBeforeShellExecution, EventBeforeMCPExecution, EventBeforeReadFile:
		return NewAllowOutput()
	default:
		return NewEmptyOutput()
	}
}

// BeforeSubmitPromptHandler handles beforeSubmitPrompt events.
type BeforeSubmitPromptHandler struct {
	deps *HandlerDependencies
}

// Handle processes the beforeSubmitPrompt event.
func (h *BeforeSubmitPromptHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	promptPayload, ok := payload.(*BeforeSubmitPromptPayload)
	if !ok {
		return NewContinueOutput(), nil
	}

	// Count tokens
	tokenCount, err := h.deps.Tokenizer.CountTokens(h.deps.Config.DefaultModel, promptPayload.Prompt)
	if err != nil {
		h.deps.Log.Error("failed to count tokens", "error", err)
		tokenCount = 0
	}

	charCount := CountChars(promptPayload.Prompt)

	h.deps.Log.Debug("prompt metrics",
		"tokens", tokenCount,
		"chars", charCount,
		"conversation_id", input.ConversationID,
	)

	// Create token usage record
	usage := &domain.TokenUsage{
		ID:               uuid.New(),
		ConversationID:   parseUUID(input.ConversationID),
		GenerationID:     parseUUID(input.GenerationID),
		Model:            h.deps.Config.DefaultModel,
		Stage:            domain.TokenUsageStagePrompt,
		PromptTokens:     tokenCount,
		CompletionTokens: 0,
		TotalTokens:      tokenCount,
		PromptChars:      charCount,
		CompletionChars:  0,
		WorkspaceRoots:   input.WorkspaceRoots,
		UserEmail:        "",
		HookEventName:    string(input.HookEventName),
		Timestamp:        time.Now(),
	}

	// Send to ctd (best-effort)
	if err := h.deps.Client.CreateTokenUsage(ctx, usage); err != nil {
		h.deps.Log.Error("failed to send token usage", "error", err)
	} else {
		h.deps.Log.Info("token usage recorded", "stage", "prompt", "tokens", tokenCount)
	}

	return NewContinueOutput(), nil
}

// AfterAgentResponseHandler handles afterAgentResponse events.
type AfterAgentResponseHandler struct {
	deps *HandlerDependencies
}

// Handle processes the afterAgentResponse event.
func (h *AfterAgentResponseHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	responsePayload, ok := payload.(*AfterAgentResponsePayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	// Count tokens
	tokenCount, err := h.deps.Tokenizer.CountTokens(h.deps.Config.DefaultModel, responsePayload.Text)
	if err != nil {
		h.deps.Log.Error("failed to count tokens", "error", err)
		tokenCount = 0
	}

	charCount := CountChars(responsePayload.Text)

	h.deps.Log.Debug("completion metrics",
		"tokens", tokenCount,
		"chars", charCount,
		"conversation_id", input.ConversationID,
	)

	// Create token usage record
	usage := &domain.TokenUsage{
		ID:               uuid.New(),
		ConversationID:   parseUUID(input.ConversationID),
		GenerationID:     parseUUID(input.GenerationID),
		Model:            h.deps.Config.DefaultModel,
		Stage:            domain.TokenUsageStageCompletion,
		PromptTokens:     0,
		CompletionTokens: tokenCount,
		TotalTokens:      tokenCount,
		PromptChars:      0,
		CompletionChars:  charCount,
		WorkspaceRoots:   input.WorkspaceRoots,
		UserEmail:        "",
		HookEventName:    string(input.HookEventName),
		Timestamp:        time.Now(),
	}

	// Send token usage to ctd (best-effort)
	if err := h.deps.Client.CreateTokenUsage(ctx, usage); err != nil {
		h.deps.Log.Error("failed to send token usage", "error", err)
	} else {
		h.deps.Log.Info("token usage recorded", "stage", "completion", "tokens", tokenCount)
	}

	// Also send as thought for visibility
	thought := &domain.Thought{
		ID:             uuid.New(),
		ConversationID: parseUUID(input.ConversationID),
		GenerationID:   parseUUID(input.GenerationID),
		Text:           responsePayload.Text,
		DurationMs:     0,
		Model:          h.deps.Config.DefaultModel,
		HookEventName:  string(input.HookEventName),
		CursorVersion:  "",
		WorkspaceRoots: input.WorkspaceRoots,
		UserEmail:      "",
		Timestamp:      time.Now(),
	}

	if err := h.deps.Client.CreateThought(ctx, thought); err != nil {
		h.deps.Log.Error("failed to send thought", "error", err)
	}

	return NewEmptyOutput(), nil
}

// AfterAgentThoughtHandler handles afterAgentThought events.
type AfterAgentThoughtHandler struct {
	deps *HandlerDependencies
}

// Handle processes the afterAgentThought event.
func (h *AfterAgentThoughtHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	thoughtPayload, ok := payload.(*AfterAgentThoughtPayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	// Send as thought for visibility
	thought := &domain.Thought{
		ID:             uuid.New(),
		ConversationID: parseUUID(input.ConversationID),
		GenerationID:   parseUUID(input.GenerationID),
		Text:           thoughtPayload.Text,
		DurationMs:     0,
		Model:          h.deps.Config.DefaultModel,
		HookEventName:  string(input.HookEventName),
		CursorVersion:  "",
		WorkspaceRoots: input.WorkspaceRoots,
		UserEmail:      "",
		Timestamp:      time.Now(),
	}

	if err := h.deps.Client.CreateThought(ctx, thought); err != nil {
		h.deps.Log.Error("failed to send thought", "error", err)
	} else {
		h.deps.Log.Info("thought recorded", "event", input.HookEventName)
	}

	return NewEmptyOutput(), nil
}

// AfterMCPExecutionHandler handles afterMCPExecution events.
type AfterMCPExecutionHandler struct {
	deps *HandlerDependencies
}

// Handle processes the afterMCPExecution event.
func (h *AfterMCPExecutionHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	mcpPayload, ok := payload.(*AfterMCPExecutionPayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	// Count tokens in tool input and result (best effort, ignore errors)
	inputTokens, inputErr := h.deps.Tokenizer.CountTokens(h.deps.Config.DefaultModel, string(mcpPayload.ToolInput))
	if inputErr != nil {
		h.deps.Log.Debug("failed to count input tokens", "error", inputErr)
		inputTokens = 0
	}
	resultTokens, resultErr := h.deps.Tokenizer.CountTokens(h.deps.Config.DefaultModel, string(mcpPayload.ResultJSON))
	if resultErr != nil {
		h.deps.Log.Debug("failed to count result tokens", "error", resultErr)
		resultTokens = 0
	}
	totalTokens := inputTokens + resultTokens

	inputChars := CountChars(string(mcpPayload.ToolInput))
	resultChars := CountChars(string(mcpPayload.ResultJSON))

	h.deps.Log.Debug("tool metrics",
		"tool", mcpPayload.ToolName,
		"input_tokens", inputTokens,
		"result_tokens", resultTokens,
		"conversation_id", input.ConversationID,
	)

	// Create token usage record
	usage := &domain.TokenUsage{
		ID:               uuid.New(),
		ConversationID:   parseUUID(input.ConversationID),
		GenerationID:     parseUUID(input.GenerationID),
		Model:            h.deps.Config.DefaultModel,
		Stage:            domain.TokenUsageStageTool,
		PromptTokens:     inputTokens,
		CompletionTokens: resultTokens,
		TotalTokens:      totalTokens,
		PromptChars:      inputChars,
		CompletionChars:  resultChars,
		WorkspaceRoots:   input.WorkspaceRoots,
		UserEmail:        "",
		HookEventName:    string(input.HookEventName),
		Timestamp:        time.Now(),
	}

	if err := h.deps.Client.CreateTokenUsage(ctx, usage); err != nil {
		h.deps.Log.Error("failed to send token usage", "error", err)
	} else {
		h.deps.Log.Info("token usage recorded", "stage", "tool", "tool", mcpPayload.ToolName, "tokens", totalTokens)
	}

	return NewEmptyOutput(), nil
}

// AfterFileEditHandler handles afterFileEdit events.
type AfterFileEditHandler struct {
	deps *HandlerDependencies
}

// Handle processes the afterFileEdit event.
func (h *AfterFileEditHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	editPayload, ok := payload.(*AfterFileEditPayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	// Count characters in edits
	var oldChars, newChars int64
	for _, edit := range editPayload.Edits {
		oldChars += CountChars(edit.OldString)
		newChars += CountChars(edit.NewString)
	}

	h.deps.Log.Debug("file edit metrics",
		"file", editPayload.FilePath,
		"old_chars", oldChars,
		"new_chars", newChars,
		"edit_count", len(editPayload.Edits),
	)

	// Create token usage record for file edits
	usage := &domain.TokenUsage{
		ID:               uuid.New(),
		ConversationID:   parseUUID(input.ConversationID),
		GenerationID:     parseUUID(input.GenerationID),
		Model:            h.deps.Config.DefaultModel,
		Stage:            domain.TokenUsageStageFileEdit,
		PromptTokens:     0,
		CompletionTokens: 0,
		TotalTokens:      0,
		PromptChars:      oldChars,
		CompletionChars:  newChars,
		WorkspaceRoots:   input.WorkspaceRoots,
		UserEmail:        "",
		HookEventName:    string(input.HookEventName),
		Timestamp:        time.Now(),
	}

	if err := h.deps.Client.CreateTokenUsage(ctx, usage); err != nil {
		h.deps.Log.Error("failed to send token usage", "error", err)
	} else {
		h.deps.Log.Info("file edit recorded", "file", editPayload.FilePath, "edits", len(editPayload.Edits))
	}

	return NewEmptyOutput(), nil
}

// StopHandler handles stop events.
type StopHandler struct {
	deps *HandlerDependencies
}

// Handle processes the stop event.
func (h *StopHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	stopPayload, ok := payload.(*StopPayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	h.deps.Log.Info("agent loop stopped",
		"status", stopPayload.Status,
		"loop_count", stopPayload.LoopCount,
		"conversation_id", input.ConversationID,
	)

	// Create token usage record as session marker
	usage := &domain.TokenUsage{
		ID:               uuid.New(),
		ConversationID:   parseUUID(input.ConversationID),
		GenerationID:     parseUUID(input.GenerationID),
		Model:            h.deps.Config.DefaultModel,
		Stage:            domain.TokenUsageStageStop,
		PromptTokens:     0,
		CompletionTokens: 0,
		TotalTokens:      0,
		PromptChars:      0,
		CompletionChars:  0,
		WorkspaceRoots:   input.WorkspaceRoots,
		UserEmail:        "",
		HookEventName:    string(input.HookEventName),
		Timestamp:        time.Now(),
	}

	if err := h.deps.Client.CreateTokenUsage(ctx, usage); err != nil {
		h.deps.Log.Error("failed to send stop marker", "error", err)
	}

	return NewEmptyOutput(), nil
}

// BeforeShellExecutionHandler handles beforeShellExecution events.
type BeforeShellExecutionHandler struct {
	deps *HandlerDependencies
}

// Handle processes the beforeShellExecution event.
func (h *BeforeShellExecutionHandler) Handle(_ context.Context, input *HookInput, payload any) (*HookOutput, error) {
	shellPayload, ok := payload.(*BeforeShellExecutionPayload)
	if !ok {
		return NewAllowOutput(), nil
	}

	h.deps.Log.Debug("shell execution",
		"command", shellPayload.Command,
		"cwd", shellPayload.Cwd,
		"conversation_id", input.ConversationID,
	)

	// Always allow shell execution (observability only)
	return NewAllowOutput(), nil
}

// AfterShellExecutionHandler handles afterShellExecution events.
type AfterShellExecutionHandler struct {
	deps *HandlerDependencies
}

// Handle processes the afterShellExecution event.
func (h *AfterShellExecutionHandler) Handle(_ context.Context, input *HookInput, payload any) (*HookOutput, error) {
	shellPayload, ok := payload.(*AfterShellExecutionPayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	h.deps.Log.Debug("shell execution completed",
		"command", shellPayload.Command,
		"output_length", len(shellPayload.Output),
		"conversation_id", input.ConversationID,
	)

	return NewEmptyOutput(), nil
}

// BeforeMCPExecutionHandler handles beforeMCPExecution events.
type BeforeMCPExecutionHandler struct {
	deps *HandlerDependencies
}

// Handle processes the beforeMCPExecution event.
//
//nolint:varnamelen // mcpPayload is clear in context
func (h *BeforeMCPExecutionHandler) Handle(_ context.Context, input *HookInput, payload any) (*HookOutput, error) {
	mcpPayload, ok := payload.(*BeforeMCPExecutionPayload)
	if !ok {
		return NewAllowOutput(), nil
	}

	h.deps.Log.Debug("MCP execution",
		"tool", mcpPayload.ToolName,
		"conversation_id", input.ConversationID,
	)

	// Always allow MCP execution (observability only)
	return NewAllowOutput(), nil
}

// BeforeReadFileHandler handles beforeReadFile events.
type BeforeReadFileHandler struct {
	deps *HandlerDependencies
}

// Handle processes the beforeReadFile event.
//
//nolint:varnamelen // readPayload is clear in context
func (h *BeforeReadFileHandler) Handle(_ context.Context, input *HookInput, payload any) (*HookOutput, error) {
	readPayload, ok := payload.(*BeforeReadFilePayload)
	if !ok {
		return NewAllowOutput(), nil
	}

	h.deps.Log.Debug("file read",
		"file", readPayload.FilePath,
		"content_length", len(readPayload.Content),
		"conversation_id", input.ConversationID,
	)

	// Always allow file read (observability only)
	return NewAllowOutput(), nil
}

// parseUUID parses a UUID string, returning a zero UUID on error.
func parseUUID(s string) uuid.UUID {
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil
	}
	return id
}


```

```go
package hook

import (
	"context"
	"fmt"
	"time"

	"github.com/ainslies/ctd/internal/domain"
	"github.com/ainslies/ctd/pkg/logger"
	"github.com/google/uuid"
)

// Handler processes a hook event and returns the appropriate output.
type Handler interface {
	// Handle processes the hook input and returns the output.
	Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error)
}

// HandlerDependencies contains the dependencies for handlers.
type HandlerDependencies struct {
	Client    *CTDClient
	Tokenizer Tokenizer
	Log       *logger.Logger
	Config    *Config
}

// Router routes hook events to the appropriate handler.
type Router struct {
	deps     *HandlerDependencies
	handlers map[HookEventName]Handler
}

// NewRouter creates a new router with all handlers registered.
func NewRouter(deps *HandlerDependencies) *Router {
	router := &Router{
		deps:     deps,
		handlers: make(map[HookEventName]Handler),
	}

	// Register handlers
	router.handlers[EventBeforeSubmitPrompt] = &BeforeSubmitPromptHandler{deps: deps}
	router.handlers[EventAfterAgentResponse] = &AfterAgentResponseHandler{deps: deps}
	router.handlers[EventAfterAgentThought] = &AfterAgentThoughtHandler{deps: deps}
	router.handlers[EventAfterMCPExecution] = &AfterMCPExecutionHandler{deps: deps}
	router.handlers[EventAfterFileEdit] = &AfterFileEditHandler{deps: deps}
	router.handlers[EventStop] = &StopHandler{deps: deps}
	router.handlers[EventBeforeShellExecution] = &BeforeShellExecutionHandler{deps: deps}
	router.handlers[EventAfterShellExecution] = &AfterShellExecutionHandler{deps: deps}
	router.handlers[EventBeforeMCPExecution] = &BeforeMCPExecutionHandler{deps: deps}
	router.handlers[EventBeforeReadFile] = &BeforeReadFileHandler{deps: deps}

	return router
}

// Route routes the input to the appropriate handler.
func (r *Router) Route(ctx context.Context, input *HookInput) (*HookOutput, error) {
	handler, ok := r.handlers[input.HookEventName]
	if !ok {
		// Unknown event, return empty output
		r.deps.Log.Warn("unknown hook event", "event", input.HookEventName)
		return NewEmptyOutput(), nil
	}

	payload, err := input.ParsePayload()
	if err != nil {
		r.deps.Log.Error("failed to parse payload", "error", err, "event", input.HookEventName)
		// Return appropriate default output even on parse error
		return getDefaultOutput(input.HookEventName), nil
	}

	output, err := handler.Handle(ctx, input, payload)
	if err != nil {
		return nil, fmt.Errorf("handler error for %s: %w", input.HookEventName, err)
	}
	return output, nil
}

// getDefaultOutput returns the default output for an event type.
//
//nolint:exhaustive // Only specific events need special handling, others use empty output
func getDefaultOutput(event HookEventName) *HookOutput {
	switch event {
	case EventBeforeSubmitPrompt:
		return NewContinueOutput()
	case EventBeforeShellExecution, EventBeforeMCPExecution, EventBeforeReadFile:
		return NewAllowOutput()
	default:
		return NewEmptyOutput()
	}
}

// BeforeSubmitPromptHandler handles beforeSubmitPrompt events.
type BeforeSubmitPromptHandler struct {
	deps *HandlerDependencies
}

// Handle processes the beforeSubmitPrompt event.
func (h *BeforeSubmitPromptHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	promptPayload, ok := payload.(*BeforeSubmitPromptPayload)
	if !ok {
		return NewContinueOutput(), nil
	}

	// Count tokens
	tokenCount, err := h.deps.Tokenizer.CountTokens(h.deps.Config.DefaultModel, promptPayload.Prompt)
	if err != nil {
		h.deps.Log.Error("failed to count tokens", "error", err)
		tokenCount = 0
	}

	charCount := CountChars(promptPayload.Prompt)

	h.deps.Log.Debug("prompt metrics",
		"tokens", tokenCount,
		"chars", charCount,
		"conversation_id", input.ConversationID,
	)

	// Create token usage record
	usage := &domain.TokenUsage{
		ID:               uuid.New(),
		ConversationID:   parseUUID(input.ConversationID),
		GenerationID:     parseUUID(input.GenerationID),
		Model:            h.deps.Config.DefaultModel,
		Stage:            domain.TokenUsageStagePrompt,
		PromptTokens:     tokenCount,
		CompletionTokens: 0,
		TotalTokens:      tokenCount,
		PromptChars:      charCount,
		CompletionChars:  0,
		WorkspaceRoots:   input.WorkspaceRoots,
		UserEmail:        "",
		HookEventName:    string(input.HookEventName),
		Timestamp:        time.Now(),
	}

	// Send to ctd (best-effort)
	if err := h.deps.Client.CreateTokenUsage(ctx, usage); err != nil {
		h.deps.Log.Error("failed to send token usage", "error", err)
	} else {
		h.deps.Log.Info("token usage recorded", "stage", "prompt", "tokens", tokenCount)
	}

	return NewContinueOutput(), nil
}

// AfterAgentResponseHandler handles afterAgentResponse events.
type AfterAgentResponseHandler struct {
	deps *HandlerDependencies
}

// Handle processes the afterAgentResponse event.
func (h *AfterAgentResponseHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	responsePayload, ok := payload.(*AfterAgentResponsePayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	// Count tokens
	tokenCount, err := h.deps.Tokenizer.CountTokens(h.deps.Config.DefaultModel, responsePayload.Text)
	if err != nil {
		h.deps.Log.Error("failed to count tokens", "error", err)
		tokenCount = 0
	}

	charCount := CountChars(responsePayload.Text)

	h.deps.Log.Debug("completion metrics",
		"tokens", tokenCount,
		"chars", charCount,
		"conversation_id", input.ConversationID,
	)

	// Create token usage record
	usage := &domain.TokenUsage{
		ID:               uuid.New(),
		ConversationID:   parseUUID(input.ConversationID),
		GenerationID:     parseUUID(input.GenerationID),
		Model:            h.deps.Config.DefaultModel,
		Stage:            domain.TokenUsageStageCompletion,
		PromptTokens:     0,
		CompletionTokens: tokenCount,
		TotalTokens:      tokenCount,
		PromptChars:      0,
		CompletionChars:  charCount,
		WorkspaceRoots:   input.WorkspaceRoots,
		UserEmail:        "",
		HookEventName:    string(input.HookEventName),
		Timestamp:        time.Now(),
	}

	// Send token usage to ctd (best-effort)
	if err := h.deps.Client.CreateTokenUsage(ctx, usage); err != nil {
		h.deps.Log.Error("failed to send token usage", "error", err)
	} else {
		h.deps.Log.Info("token usage recorded", "stage", "completion", "tokens", tokenCount)
	}

	// Also send as thought for visibility
	thought := &domain.Thought{
		ID:             uuid.New(),
		ConversationID: parseUUID(input.ConversationID),
		GenerationID:   parseUUID(input.GenerationID),
		Text:           responsePayload.Text,
		DurationMs:     0,
		Model:          h.deps.Config.DefaultModel,
		HookEventName:  string(input.HookEventName),
		CursorVersion:  "",
		WorkspaceRoots: input.WorkspaceRoots,
		UserEmail:      "",
		Timestamp:      time.Now(),
	}

	if err := h.deps.Client.CreateThought(ctx, thought); err != nil {
		h.deps.Log.Error("failed to send thought", "error", err)
	}

	return NewEmptyOutput(), nil
}

// AfterAgentThoughtHandler handles afterAgentThought events.
type AfterAgentThoughtHandler struct {
	deps *HandlerDependencies
}

// Handle processes the afterAgentThought event.
func (h *AfterAgentThoughtHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	thoughtPayload, ok := payload.(*AfterAgentThoughtPayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	// Send as thought for visibility
	thought := &domain.Thought{
		ID:             uuid.New(),
		ConversationID: parseUUID(input.ConversationID),
		GenerationID:   parseUUID(input.GenerationID),
		Text:           thoughtPayload.Text,
		DurationMs:     0,
		Model:          h.deps.Config.DefaultModel,
		HookEventName:  string(input.HookEventName),
		CursorVersion:  "",
		WorkspaceRoots: input.WorkspaceRoots,
		UserEmail:      "",
		Timestamp:      time.Now(),
	}

	if err := h.deps.Client.CreateThought(ctx, thought); err != nil {
		h.deps.Log.Error("failed to send thought", "error", err)
	} else {
		h.deps.Log.Info("thought recorded", "event", input.HookEventName)
	}

	return NewEmptyOutput(), nil
}

// AfterMCPExecutionHandler handles afterMCPExecution events.
type AfterMCPExecutionHandler struct {
	deps *HandlerDependencies
}

// Handle processes the afterMCPExecution event.
func (h *AfterMCPExecutionHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	mcpPayload, ok := payload.(*AfterMCPExecutionPayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	// Count tokens in tool input and result (best effort, ignore errors)
	inputTokens, inputErr := h.deps.Tokenizer.CountTokens(h.deps.Config.DefaultModel, string(mcpPayload.ToolInput))
	if inputErr != nil {
		h.deps.Log.Debug("failed to count input tokens", "error", inputErr)
		inputTokens = 0
	}
	resultTokens, resultErr := h.deps.Tokenizer.CountTokens(h.deps.Config.DefaultModel, string(mcpPayload.ResultJSON))
	if resultErr != nil {
		h.deps.Log.Debug("failed to count result tokens", "error", resultErr)
		resultTokens = 0
	}
	totalTokens := inputTokens + resultTokens

	inputChars := CountChars(string(mcpPayload.ToolInput))
	resultChars := CountChars(string(mcpPayload.ResultJSON))

	h.deps.Log.Debug("tool metrics",
		"tool", mcpPayload.ToolName,
		"input_tokens", inputTokens,
		"result_tokens", resultTokens,
		"conversation_id", input.ConversationID,
	)

	// Create token usage record
	usage := &domain.TokenUsage{
		ID:               uuid.New(),
		ConversationID:   parseUUID(input.ConversationID),
		GenerationID:     parseUUID(input.GenerationID),
		Model:            h.deps.Config.DefaultModel,
		Stage:            domain.TokenUsageStageTool,
		PromptTokens:     inputTokens,
		CompletionTokens: resultTokens,
		TotalTokens:      totalTokens,
		PromptChars:      inputChars,
		CompletionChars:  resultChars,
		WorkspaceRoots:   input.WorkspaceRoots,
		UserEmail:        "",
		HookEventName:    string(input.HookEventName),
		Timestamp:        time.Now(),
	}

	if err := h.deps.Client.CreateTokenUsage(ctx, usage); err != nil {
		h.deps.Log.Error("failed to send token usage", "error", err)
	} else {
		h.deps.Log.Info("token usage recorded", "stage", "tool", "tool", mcpPayload.ToolName, "tokens", totalTokens)
	}

	return NewEmptyOutput(), nil
}

// AfterFileEditHandler handles afterFileEdit events.
type AfterFileEditHandler struct {
	deps *HandlerDependencies
}

// Handle processes the afterFileEdit event.
func (h *AfterFileEditHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	editPayload, ok := payload.(*AfterFileEditPayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	// Count characters in edits
	var oldChars, newChars int64
	for _, edit := range editPayload.Edits {
		oldChars += CountChars(edit.OldString)
		newChars += CountChars(edit.NewString)
	}

	h.deps.Log.Debug("file edit metrics",
		"file", editPayload.FilePath,
		"old_chars", oldChars,
		"new_chars", newChars,
		"edit_count", len(editPayload.Edits),
	)

	// Create token usage record for file edits
	usage := &domain.TokenUsage{
		ID:               uuid.New(),
		ConversationID:   parseUUID(input.ConversationID),
		GenerationID:     parseUUID(input.GenerationID),
		Model:            h.deps.Config.DefaultModel,
		Stage:            domain.TokenUsageStageFileEdit,
		PromptTokens:     0,
		CompletionTokens: 0,
		TotalTokens:      0,
		PromptChars:      oldChars,
		CompletionChars:  newChars,
		WorkspaceRoots:   input.WorkspaceRoots,
		UserEmail:        "",
		HookEventName:    string(input.HookEventName),
		Timestamp:        time.Now(),
	}

	if err := h.deps.Client.CreateTokenUsage(ctx, usage); err != nil {
		h.deps.Log.Error("failed to send token usage", "error", err)
	} else {
		h.deps.Log.Info("file edit recorded", "file", editPayload.FilePath, "edits", len(editPayload.Edits))
	}

	return NewEmptyOutput(), nil
}

// StopHandler handles stop events.
type StopHandler struct {
	deps *HandlerDependencies
}

// Handle processes the stop event.
func (h *StopHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	stopPayload, ok := payload.(*StopPayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	h.deps.Log.Info("agent loop stopped",
		"status", stopPayload.Status,
		"loop_count", stopPayload.LoopCount,
		"conversation_id", input.ConversationID,
	)

	// Create token usage record as session marker
	usage := &domain.TokenUsage{
		ID:               uuid.New(),
		ConversationID:   parseUUID(input.ConversationID),
		GenerationID:     parseUUID(input.GenerationID),
		Model:            h.deps.Config.DefaultModel,
		Stage:            domain.TokenUsageStageStop,
		PromptTokens:     0,
		CompletionTokens: 0,
		TotalTokens:      0,
		PromptChars:      0,
		CompletionChars:  0,
		WorkspaceRoots:   input.WorkspaceRoots,
		UserEmail:        "",
		HookEventName:    string(input.HookEventName),
		Timestamp:        time.Now(),
	}

	if err := h.deps.Client.CreateTokenUsage(ctx, usage); err != nil {
		h.deps.Log.Error("failed to send stop marker", "error", err)
	}

	return NewEmptyOutput(), nil
}

// BeforeShellExecutionHandler handles beforeShellExecution events.
type BeforeShellExecutionHandler struct {
	deps *HandlerDependencies
}

// Handle processes the beforeShellExecution event.
func (h *BeforeShellExecutionHandler) Handle(_ context.Context, input *HookInput, payload any) (*HookOutput, error) {
	shellPayload, ok := payload.(*BeforeShellExecutionPayload)
	if !ok {
		return NewAllowOutput(), nil
	}

	h.deps.Log.Debug("shell execution",
		"command", shellPayload.Command,
		"cwd", shellPayload.Cwd,
		"conversation_id", input.ConversationID,
	)

	// Always allow shell execution (observability only)
	return NewAllowOutput(), nil
}

// AfterShellExecutionHandler handles afterShellExecution events.
type AfterShellExecutionHandler struct {
	deps *HandlerDependencies
}

// Handle processes the afterShellExecution event.
func (h *AfterShellExecutionHandler) Handle(_ context.Context, input *HookInput, payload any) (*HookOutput, error) {
	shellPayload, ok := payload.(*AfterShellExecutionPayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	h.deps.Log.Debug("shell execution completed",
		"command", shellPayload.Command,
		"output_length", len(shellPayload.Output),
		"conversation_id", input.ConversationID,
	)

	return NewEmptyOutput(), nil
}

// BeforeMCPExecutionHandler handles beforeMCPExecution events.
type BeforeMCPExecutionHandler struct {
	deps *HandlerDependencies
}

// Handle processes the beforeMCPExecution event.
func (h *BeforeMCPExecutionHandler) Handle(_ context.Context, input *HookInput, payload any) (*HookOutput, error) {
	mcpPayload, ok := payload.(*BeforeMCPExecutionPayload)
	if !ok {
		return NewAllowOutput(), nil
	}

	h.deps.Log.Debug("MCP execution",
		"tool", mcpPayload.ToolName,
		"conversation_id", input.ConversationID,
	)

	// Always allow MCP execution (observability only)
	return NewAllowOutput(), nil
}

// BeforeReadFileHandler handles beforeReadFile events.
type BeforeReadFileHandler struct {
	deps *HandlerDependencies
}

// Handle processes the beforeReadFile event.
//
//nolint:varnamelen // readPayload is clear in context
func (h *BeforeReadFileHandler) Handle(_ context.Context, input *HookInput, payload any) (*HookOutput, error) {
	readPayload, ok := payload.(*BeforeReadFilePayload)
	if !ok {
		return NewAllowOutput(), nil
	}

	h.deps.Log.Debug("file read",
		"file", readPayload.FilePath,
		"content_length", len(readPayload.Content),
		"conversation_id", input.ConversationID,
	)

	// Always allow file read (observability only)
	return NewAllowOutput(), nil
}

// parseUUID parses a UUID string, returning a zero UUID on error.
func parseUUID(s string) uuid.UUID {
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil
	}
	return id
}


```

```go
package hook

import (
	"context"
	"fmt"
	"time"

	"github.com/ainslies/ctd/internal/domain"
	"github.com/ainslies/ctd/pkg/logger"
	"github.com/google/uuid"
)

// Handler processes a hook event and returns the appropriate output.
type Handler interface {
	// Handle processes the hook input and returns the output.
	Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error)
}

// HandlerDependencies contains the dependencies for handlers.
type HandlerDependencies struct {
	Client    *CTDClient
	Tokenizer Tokenizer
	Log       *logger.Logger
	Config    *Config
}

// Router routes hook events to the appropriate handler.
type Router struct {
	deps     *HandlerDependencies
	handlers map[HookEventName]Handler
}

// NewRouter creates a new router with all handlers registered.
func NewRouter(deps *HandlerDependencies) *Router {
	router := &Router{
		deps:     deps,
		handlers: make(map[HookEventName]Handler),
	}

	// Register handlers
	router.handlers[EventBeforeSubmitPrompt] = &BeforeSubmitPromptHandler{deps: deps}
	router.handlers[EventAfterAgentResponse] = &AfterAgentResponseHandler{deps: deps}
	router.handlers[EventAfterAgentThought] = &AfterAgentThoughtHandler{deps: deps}
	router.handlers[EventAfterMCPExecution] = &AfterMCPExecutionHandler{deps: deps}
	router.handlers[EventAfterFileEdit] = &AfterFileEditHandler{deps: deps}
	router.handlers[EventStop] = &StopHandler{deps: deps}
	router.handlers[EventBeforeShellExecution] = &BeforeShellExecutionHandler{deps: deps}
	router.handlers[EventAfterShellExecution] = &AfterShellExecutionHandler{deps: deps}
	router.handlers[EventBeforeMCPExecution] = &BeforeMCPExecutionHandler{deps: deps}
	router.handlers[EventBeforeReadFile] = &BeforeReadFileHandler{deps: deps}

	return router
}

// Route routes the input to the appropriate handler.
func (r *Router) Route(ctx context.Context, input *HookInput) (*HookOutput, error) {
	handler, ok := r.handlers[input.HookEventName]
	if !ok {
		// Unknown event, return empty output
		r.deps.Log.Warn("unknown hook event", "event", input.HookEventName)
		return NewEmptyOutput(), nil
	}

	payload, err := input.ParsePayload()
	if err != nil {
		r.deps.Log.Error("failed to parse payload", "error", err, "event", input.HookEventName)
		// Return appropriate default output even on parse error
		return getDefaultOutput(input.HookEventName), nil
	}

	output, err := handler.Handle(ctx, input, payload)
	if err != nil {
		return nil, fmt.Errorf("handler error for %s: %w", input.HookEventName, err)
	}
	return output, nil
}

// getDefaultOutput returns the default output for an event type.
//
//nolint:exhaustive // Only specific events need special handling, others use empty output
func getDefaultOutput(event HookEventName) *HookOutput {
	switch event {
	case EventBeforeSubmitPrompt:
		return NewContinueOutput()
	case EventBeforeShellExecution, EventBeforeMCPExecution, EventBeforeReadFile:
		return NewAllowOutput()
	default:
		return NewEmptyOutput()
	}
}

// BeforeSubmitPromptHandler handles beforeSubmitPrompt events.
type BeforeSubmitPromptHandler struct {
	deps *HandlerDependencies
}

// Handle processes the beforeSubmitPrompt event.
func (h *BeforeSubmitPromptHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	promptPayload, ok := payload.(*BeforeSubmitPromptPayload)
	if !ok {
		return NewContinueOutput(), nil
	}

	// Count tokens
	tokenCount, err := h.deps.Tokenizer.CountTokens(h.deps.Config.DefaultModel, promptPayload.Prompt)
	if err != nil {
		h.deps.Log.Error("failed to count tokens", "error", err)
		tokenCount = 0
	}

	charCount := CountChars(promptPayload.Prompt)

	h.deps.Log.Debug("prompt metrics",
		"tokens", tokenCount,
		"chars", charCount,
		"conversation_id", input.ConversationID,
	)

	// Create token usage record
	usage := &domain.TokenUsage{
		ID:               uuid.New(),
		ConversationID:   parseUUID(input.ConversationID),
		GenerationID:     parseUUID(input.GenerationID),
		Model:            h.deps.Config.DefaultModel,
		Stage:            domain.TokenUsageStagePrompt,
		PromptTokens:     tokenCount,
		CompletionTokens: 0,
		TotalTokens:      tokenCount,
		PromptChars:      charCount,
		CompletionChars:  0,
		WorkspaceRoots:   input.WorkspaceRoots,
		UserEmail:        "",
		HookEventName:    string(input.HookEventName),
		Timestamp:        time.Now(),
	}

	// Send to ctd (best-effort)
	if err := h.deps.Client.CreateTokenUsage(ctx, usage); err != nil {
		h.deps.Log.Error("failed to send token usage", "error", err)
	} else {
		h.deps.Log.Info("token usage recorded", "stage", "prompt", "tokens", tokenCount)
	}

	return NewContinueOutput(), nil
}

// AfterAgentResponseHandler handles afterAgentResponse events.
type AfterAgentResponseHandler struct {
	deps *HandlerDependencies
}

// Handle processes the afterAgentResponse event.
func (h *AfterAgentResponseHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	responsePayload, ok := payload.(*AfterAgentResponsePayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	// Count tokens
	tokenCount, err := h.deps.Tokenizer.CountTokens(h.deps.Config.DefaultModel, responsePayload.Text)
	if err != nil {
		h.deps.Log.Error("failed to count tokens", "error", err)
		tokenCount = 0
	}

	charCount := CountChars(responsePayload.Text)

	h.deps.Log.Debug("completion metrics",
		"tokens", tokenCount,
		"chars", charCount,
		"conversation_id", input.ConversationID,
	)

	// Create token usage record
	usage := &domain.TokenUsage{
		ID:               uuid.New(),
		ConversationID:   parseUUID(input.ConversationID),
		GenerationID:     parseUUID(input.GenerationID),
		Model:            h.deps.Config.DefaultModel,
		Stage:            domain.TokenUsageStageCompletion,
		PromptTokens:     0,
		CompletionTokens: tokenCount,
		TotalTokens:      tokenCount,
		PromptChars:      0,
		CompletionChars:  charCount,
		WorkspaceRoots:   input.WorkspaceRoots,
		UserEmail:        "",
		HookEventName:    string(input.HookEventName),
		Timestamp:        time.Now(),
	}

	// Send token usage to ctd (best-effort)
	if err := h.deps.Client.CreateTokenUsage(ctx, usage); err != nil {
		h.deps.Log.Error("failed to send token usage", "error", err)
	} else {
		h.deps.Log.Info("token usage recorded", "stage", "completion", "tokens", tokenCount)
	}

	// Also send as thought for visibility
	thought := &domain.Thought{
		ID:             uuid.New(),
		ConversationID: parseUUID(input.ConversationID),
		GenerationID:   parseUUID(input.GenerationID),
		Text:           responsePayload.Text,
		DurationMs:     0,
		Model:          h.deps.Config.DefaultModel,
		HookEventName:  string(input.HookEventName),
		CursorVersion:  "",
		WorkspaceRoots: input.WorkspaceRoots,
		UserEmail:      "",
		Timestamp:      time.Now(),
	}

	if err := h.deps.Client.CreateThought(ctx, thought); err != nil {
		h.deps.Log.Error("failed to send thought", "error", err)
	}

	return NewEmptyOutput(), nil
}

// AfterAgentThoughtHandler handles afterAgentThought events.
type AfterAgentThoughtHandler struct {
	deps *HandlerDependencies
}

// Handle processes the afterAgentThought event.
func (h *AfterAgentThoughtHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	thoughtPayload, ok := payload.(*AfterAgentThoughtPayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	// Send as thought for visibility
	thought := &domain.Thought{
		ID:             uuid.New(),
		ConversationID: parseUUID(input.ConversationID),
		GenerationID:   parseUUID(input.GenerationID),
		Text:           thoughtPayload.Text,
		DurationMs:     0,
		Model:          h.deps.Config.DefaultModel,
		HookEventName:  string(input.HookEventName),
		CursorVersion:  "",
		WorkspaceRoots: input.WorkspaceRoots,
		UserEmail:      "",
		Timestamp:      time.Now(),
	}

	if err := h.deps.Client.CreateThought(ctx, thought); err != nil {
		h.deps.Log.Error("failed to send thought", "error", err)
	} else {
		h.deps.Log.Info("thought recorded", "event", input.HookEventName)
	}

	return NewEmptyOutput(), nil
}

// AfterMCPExecutionHandler handles afterMCPExecution events.
type AfterMCPExecutionHandler struct {
	deps *HandlerDependencies
}

// Handle processes the afterMCPExecution event.
func (h *AfterMCPExecutionHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	mcpPayload, ok := payload.(*AfterMCPExecutionPayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	// Count tokens in tool input and result (best effort, ignore errors)
	inputTokens, inputErr := h.deps.Tokenizer.CountTokens(h.deps.Config.DefaultModel, string(mcpPayload.ToolInput))
	if inputErr != nil {
		h.deps.Log.Debug("failed to count input tokens", "error", inputErr)
		inputTokens = 0
	}
	resultTokens, resultErr := h.deps.Tokenizer.CountTokens(h.deps.Config.DefaultModel, string(mcpPayload.ResultJSON))
	if resultErr != nil {
		h.deps.Log.Debug("failed to count result tokens", "error", resultErr)
		resultTokens = 0
	}
	totalTokens := inputTokens + resultTokens

	inputChars := CountChars(string(mcpPayload.ToolInput))
	resultChars := CountChars(string(mcpPayload.ResultJSON))

	h.deps.Log.Debug("tool metrics",
		"tool", mcpPayload.ToolName,
		"input_tokens", inputTokens,
		"result_tokens", resultTokens,
		"conversation_id", input.ConversationID,
	)

	// Create token usage record
	usage := &domain.TokenUsage{
		ID:               uuid.New(),
		ConversationID:   parseUUID(input.ConversationID),
		GenerationID:     parseUUID(input.GenerationID),
		Model:            h.deps.Config.DefaultModel,
		Stage:            domain.TokenUsageStageTool,
		PromptTokens:     inputTokens,
		CompletionTokens: resultTokens,
		TotalTokens:      totalTokens,
		PromptChars:      inputChars,
		CompletionChars:  resultChars,
		WorkspaceRoots:   input.WorkspaceRoots,
		UserEmail:        "",
		HookEventName:    string(input.HookEventName),
		Timestamp:        time.Now(),
	}

	if err := h.deps.Client.CreateTokenUsage(ctx, usage); err != nil {
		h.deps.Log.Error("failed to send token usage", "error", err)
	} else {
		h.deps.Log.Info("token usage recorded", "stage", "tool", "tool", mcpPayload.ToolName, "tokens", totalTokens)
	}

	return NewEmptyOutput(), nil
}

// AfterFileEditHandler handles afterFileEdit events.
type AfterFileEditHandler struct {
	deps *HandlerDependencies
}

// Handle processes the afterFileEdit event.
func (h *AfterFileEditHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	editPayload, ok := payload.(*AfterFileEditPayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	// Count characters in edits
	var oldChars, newChars int64
	for _, edit := range editPayload.Edits {
		oldChars += CountChars(edit.OldString)
		newChars += CountChars(edit.NewString)
	}

	h.deps.Log.Debug("file edit metrics",
		"file", editPayload.FilePath,
		"old_chars", oldChars,
		"new_chars", newChars,
		"edit_count", len(editPayload.Edits),
	)

	// Create token usage record for file edits
	usage := &domain.TokenUsage{
		ID:               uuid.New(),
		ConversationID:   parseUUID(input.ConversationID),
		GenerationID:     parseUUID(input.GenerationID),
		Model:            h.deps.Config.DefaultModel,
		Stage:            domain.TokenUsageStageFileEdit,
		PromptTokens:     0,
		CompletionTokens: 0,
		TotalTokens:      0,
		PromptChars:      oldChars,
		CompletionChars:  newChars,
		WorkspaceRoots:   input.WorkspaceRoots,
		UserEmail:        "",
		HookEventName:    string(input.HookEventName),
		Timestamp:        time.Now(),
	}

	if err := h.deps.Client.CreateTokenUsage(ctx, usage); err != nil {
		h.deps.Log.Error("failed to send token usage", "error", err)
	} else {
		h.deps.Log.Info("file edit recorded", "file", editPayload.FilePath, "edits", len(editPayload.Edits))
	}

	return NewEmptyOutput(), nil
}

// StopHandler handles stop events.
type StopHandler struct {
	deps *HandlerDependencies
}

// Handle processes the stop event.
func (h *StopHandler) Handle(ctx context.Context, input *HookInput, payload any) (*HookOutput, error) {
	stopPayload, ok := payload.(*StopPayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	h.deps.Log.Info("agent loop stopped",
		"status", stopPayload.Status,
		"loop_count", stopPayload.LoopCount,
		"conversation_id", input.ConversationID,
	)

	// Create token usage record as session marker
	usage := &domain.TokenUsage{
		ID:               uuid.New(),
		ConversationID:   parseUUID(input.ConversationID),
		GenerationID:     parseUUID(input.GenerationID),
		Model:            h.deps.Config.DefaultModel,
		Stage:            domain.TokenUsageStageStop,
		PromptTokens:     0,
		CompletionTokens: 0,
		TotalTokens:      0,
		PromptChars:      0,
		CompletionChars:  0,
		WorkspaceRoots:   input.WorkspaceRoots,
		UserEmail:        "",
		HookEventName:    string(input.HookEventName),
		Timestamp:        time.Now(),
	}

	if err := h.deps.Client.CreateTokenUsage(ctx, usage); err != nil {
		h.deps.Log.Error("failed to send stop marker", "error", err)
	}

	return NewEmptyOutput(), nil
}

// BeforeShellExecutionHandler handles beforeShellExecution events.
type BeforeShellExecutionHandler struct {
	deps *HandlerDependencies
}

// Handle processes the beforeShellExecution event.
func (h *BeforeShellExecutionHandler) Handle(_ context.Context, input *HookInput, payload any) (*HookOutput, error) {
	shellPayload, ok := payload.(*BeforeShellExecutionPayload)
	if !ok {
		return NewAllowOutput(), nil
	}

	h.deps.Log.Debug("shell execution",
		"command", shellPayload.Command,
		"cwd", shellPayload.Cwd,
		"conversation_id", input.ConversationID,
	)

	// Always allow shell execution (observability only)
	return NewAllowOutput(), nil
}

// AfterShellExecutionHandler handles afterShellExecution events.
type AfterShellExecutionHandler struct {
	deps *HandlerDependencies
}

// Handle processes the afterShellExecution event.
func (h *AfterShellExecutionHandler) Handle(_ context.Context, input *HookInput, payload any) (*HookOutput, error) {
	shellPayload, ok := payload.(*AfterShellExecutionPayload)
	if !ok {
		return NewEmptyOutput(), nil
	}

	h.deps.Log.Debug("shell execution completed",
		"command", shellPayload.Command,
		"output_length", len(shellPayload.Output),
		"conversation_id", input.ConversationID,
	)

	return NewEmptyOutput(), nil
}

// BeforeMCPExecutionHandler handles beforeMCPExecution events.
type BeforeMCPExecutionHandler struct {
	deps *HandlerDependencies
}

// Handle processes the beforeMCPExecution event.
func (h *BeforeMCPExecutionHandler) Handle(_ context.Context, input *HookInput, payload any) (*HookOutput, error) {
	mcpPayload, ok := payload.(*BeforeMCPExecutionPayload)
	if !ok {
		return NewAllowOutput(), nil
	}

	h.deps.Log.Debug("MCP execution",
		"tool", mcpPayload.ToolName,
		"conversation_id", input.ConversationID,
	)

	// Always allow MCP execution (observability only)
	return NewAllowOutput(), nil
}

// BeforeReadFileHandler handles beforeReadFile events.
type BeforeReadFileHandler struct {
	deps *HandlerDependencies
}

// Handle processes the beforeReadFile event.
func (h *BeforeReadFileHandler) Handle(_ context.Context, input *HookInput, payload any) (*HookOutput, error) {
	readPayload, ok := payload.(*BeforeReadFilePayload)
	if !ok {
		return NewAllowOutput(), nil
	}

	h.deps.Log.Debug("file read",
		"file", readPayload.FilePath,
		"content_length", len(readPayload.Content),
		"conversation_id", input.ConversationID,
	)

	// Always allow file read (observability only)
	return NewAllowOutput(), nil
}

// parseUUID parses a UUID string, returning a zero UUID on error.
func parseUUID(s string) uuid.UUID {
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil
	}
	return id
}


```

All tests pass. Let me update the todos and run a final check:

All linting and tests pass. Let me mark the documentation task as complete and provide a summary:

All tasks are complete. Here's a summary of what was implemented:

## Implementation Summary

The token usage tracking feature has been fully implemented with the following components:

### Server-Side (ctd daemon)
1. **Proto definitions** (`api/proto/v1/token_usage.proto`): `TokenUsage` message and `TokenUsageService` with `CreateTokenUsage` RPC
2. **Database migration** (`migrations/002_token_usage.sql`): `token_usage` table with appropriate indexes
3. **Domain model** (`internal/domain/token_usage.go`): `TokenUsage` struct and `TokenUsageRepository` interface
4. **Repository** (`internal/repository/postgres_token_usage.go`): PostgreSQL implementation with batch insert support
5. **Server** (`internal/server/token_usage.go`): gRPC handler for `CreateTokenUsage`
6. **App wiring** (`internal/app/app.go`): Shared database pool between `ThoughtRepository` and `TokenUsageRepository`

### Hook Binary (ctd-hook)
1. **Config** (`internal/hook/config.go`): Configuration loading via viper with environment variable support
2. **Input parsing** (`internal/hook/input.go`): JSON input parsing for all Cursor hook event types
3. **Tokenizer** (`internal/hook/tokenizer.go`): Composite tokenizer using tiktoken-go with heuristic fallback
4. **Client** (`internal/hook/client.go`): gRPC client for communicating with ctd daemon
5. **Handlers** (`internal/hook/handlers.go`): Event handlers for all 10 hook events
6. **Output** (`internal/hook/output.go`): JSON output formatting for Cursor hook responses
7. **Entry point** (`cmd/ctd-hook/main.go`): Main binary that routes events to handlers

### Build & Deploy
- Updated `Makefile` with `build-hook` and `install-hook` targets
- Created `deploy/hook.yaml.example` and `deploy/hooks.json.example` configuration templates
- Created `docs/hooks.md` documentation

### Testing
- Unit tests for all new components achieving good coverage
- All linting passes with zero issues

---

**User**

Is this part of the documenation in @docs/hooks.md accurate?



Configure Cursor Hooks



Create or update ~/.cursor/hooks.json:

json

Copy

{

  "version": 1,

  "hooks": {

    "beforeSubmitPrompt": [{ "command": "~/.cursor/hooks/ctd-hook" }],

    "afterAgentResponse": [{ "command": "~/.cursor/hooks/ctd-hook" }],

    "afterAgentThought": [{ "command": "~/.cursor/hooks/ctd-hook" }],

    "afterMCPExecution": [{ "command": "~/.cursor/hooks/ctd-hook" }],

    "afterFileEdit": [{ "command": "~/.cursor/hooks/ctd-hook" }],

    "stop": [{ "command": "~/.cursor/hooks/ctd-hook" }]

  }

}



Is `ctd-hook` not supposed to take an argument?

---

**Cursor**

Good question. Let me check the hook input parsing and the Cursor hooks documentation to verify how the hook receives event information.

The documentation is **accurate**. The `ctd-hook` binary does **not** need command-line arguments.

Here's why:

1. **Cursor passes event data via stdin (JSON)**, not command-line arguments
2. The hook reads from `os.Stdin` (line 52 in `main.go`): `input, err := hook.ParseInput(os.Stdin)`
3. The JSON input includes a `hook_event_name` field that identifies which event triggered the hook
4. The same binary handles all event types by routing based on the `HookEventName` field

The input JSON from Cursor looks something like:
```json
{
  "hook_event_name": "beforeSubmitPrompt",
  "conversation_id": "...",
  "generation_id": "...",
  "workspace_roots": [...],
  "prompt": "...",
  ...
}
```

So the configuration in `hooks.json` is correct:
```json
"beforeSubmitPrompt": [{ "command": "~/.cursor/hooks/ctd-hook" }]
```

Cursor invokes the command and pipes the event data to its stdin. The hook parses that JSON, determines the event type from `hook_event_name`, routes to the appropriate handler, and writes a JSON response to stdout.

```json
{
  "hook_event_name": "beforeSubmitPrompt",
  "conversation_id": "...",
  "generation_id": "...",
  "workspace_roots": [...],
  "prompt": "...",
  ...
}
```

```json
"beforeSubmitPrompt": [{ "command": "~/.cursor/hooks/ctd-hook" }]
```

---

**User**

update the make install-hook target to backup the existing configs using date-time stamped backups and install the appropriate configs in additionto installing the binary 

---

**Cursor**

Now let me check what example config files exist in the deploy directory:

I need to create the hook config example file and update the Makefile. Let me first create the example config:

```plaintext
# ctd-hook configuration
# Place this file at ~/.config/ctd/hook.yaml

# URL of the ctd daemon
server_url: http://127.0.0.1:36912

# Log level (debug, info, warn, error)
log_level: info

# Default model for token counting when model is not specified
default_model: gpt-4o

# Timeout for ctd server calls in milliseconds
timeout_ms: 500

```

Now let me update the Makefile with the improved `install-hook` target:

```makefile
.PHONY: help build build-hook build-all run test lint lint-daemon lint-extension clean docker-up docker-down install-deps install install-hook uninstall uninstall-hook

help: ## Display this help message
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

gen: ## Generate Go code from Proto
	@echo "Generating Go code..."
	@buf generate
	@echo "Generated Go code in gen/"

build: ## Build the daemon binary
	@echo "Building ctd..."
	@go build -o bin/ctd ./cmd/thought-daemon

build-hook: ## Build the hook binary
	@echo "Building ctd-hook..."
	@go build -o bin/ctd-hook ./cmd/ctd-hook

build-all: build build-hook ## Build both daemon and hook binaries

install: build ## Install ctd to ~/.local/bin (no sudo required)
	@echo "Installing ctd..."
	@mkdir -p ~/.local/bin
	@cp bin/ctd ~/.local/bin/
	@chmod +x ~/.local/bin/ctd
	@echo " Installed to ~/.local/bin/ctd"
	@echo ""
	@if echo $$PATH | grep -q "$$HOME/.local/bin"; then \
		echo " ~/.local/bin is in your PATH"; \
		echo "  You can now run: ctd"; \
	else \
		echo " ~/.local/bin is not in your PATH"; \
		echo "  Add it by running:"; \
		echo "    echo 'export PATH=\"\$$HOME/.local/bin:\$$PATH\"' >> ~/.bashrc"; \
		echo "    source ~/.bashrc"; \
		echo "  (or ~/.zshrc, ~/.profile, etc. depending on your shell)"; \
	fi
	@echo ""
	@mkdir -p ~/.config/ctd
	@if [ ! -f ~/.config/ctd/config.yaml ]; then \
		cp config.yaml.example ~/.config/ctd/config.yaml; \
		echo " Created default config at ~/.config/ctd/config.yaml"; \
	else \
		echo "ℹ Config already exists at ~/.config/ctd/config.yaml"; \
	fi

extension-install: ## Build and install the VS Code extension to Cursor
	@echo "Installing VS Code extension dependencies..."
	@cd vscode-extension && npm install
	@echo "Building VS Code extension..."
	@cd vscode-extension && npm run build
	@echo "Packaging VS Code extension..."
	@cd vscode-extension && ./node_modules/.bin/vsce package
	@echo "Installing extension to Cursor..."
	@version=$$(cd vscode-extension && node -p "require('./package.json').version"); \
	cursor --uninstall-extension cursor-user.cursor-thought-stream 2>/dev/null || true; \
	cursor --install-extension vscode-extension/cursor-thought-stream-$$version.vsix --force; \
	echo "✅ Extension v$$version installed successfully!"; \
	echo "🔄 Please reload Cursor window (Cmd+Shift+P -> 'Developer: Reload Window')"

install-hook: build-hook ## Install ctd-hook to ~/.cursor/hooks with config files
	@echo "Installing ctd-hook..."
	@TIMESTAMP=$$(date +%Y%m%d-%H%M%S); \
	mkdir -p ~/.cursor/hooks; \
	mkdir -p ~/.config/ctd; \
	if [ -f ~/.cursor/hooks/ctd-hook ]; then \
		echo "[BACKUP] Backing up existing binary to ~/.cursor/hooks/ctd-hook.$$TIMESTAMP"; \
		cp ~/.cursor/hooks/ctd-hook ~/.cursor/hooks/ctd-hook.$$TIMESTAMP; \
	fi; \
	cp bin/ctd-hook ~/.cursor/hooks/; \
	chmod +x ~/.cursor/hooks/ctd-hook; \
	echo "[OK] Installed binary to ~/.cursor/hooks/ctd-hook"; \
	echo ""; \
	if [ -f ~/.cursor/hooks.json ]; then \
		echo "[BACKUP] Backing up existing hooks.json to ~/.cursor/hooks.json.$$TIMESTAMP"; \
		cp ~/.cursor/hooks.json ~/.cursor/hooks.json.$$TIMESTAMP; \
		echo "[INFO] Existing hooks.json preserved - merge manually if needed"; \
		echo "       Example config: deploy/hooks.json.example"; \
	else \
		cp deploy/hooks.json.example ~/.cursor/hooks.json; \
		echo "[OK] Installed hooks.json to ~/.cursor/hooks.json"; \
	fi; \
	echo ""; \
	if [ -f ~/.config/ctd/hook.yaml ]; then \
		echo "[BACKUP] Backing up existing hook.yaml to ~/.config/ctd/hook.yaml.$$TIMESTAMP"; \
		cp ~/.config/ctd/hook.yaml ~/.config/ctd/hook.yaml.$$TIMESTAMP; \
		echo "[INFO] Existing hook.yaml preserved - merge manually if needed"; \
		echo "       Example config: deploy/hook.yaml.example"; \
	else \
		cp deploy/hook.yaml.example ~/.config/ctd/hook.yaml; \
		echo "[OK] Installed hook config to ~/.config/ctd/hook.yaml"; \
	fi; \
	echo ""; \
	echo "[OK] Hook installation complete"; \
	echo "     Restart Cursor to activate hooks"

uninstall: ## Remove ctd from ~/.local/bin
	@echo "Uninstalling ctd..."
	@rm -f ~/.local/bin/ctd
	@echo "[OK] Removed ~/.local/bin/ctd"
	@echo "[INFO] Config remains at ~/.config/ctd/ (remove manually if desired)"

uninstall-hook: ## Remove ctd-hook from ~/.cursor/hooks
	@echo "Uninstalling ctd-hook..."
	@rm -f ~/.cursor/hooks/ctd-hook
	@echo "[OK] Removed ~/.cursor/hooks/ctd-hook"
	@echo "[INFO] Remember to update ~/.cursor/hooks.json to remove hook references"

run: ## Run the daemon locally
	@go run ./cmd/thought-daemon

test: test-daemon test-extension ## Run all tests (daemon + extension)

test-all: test-unit test-integration test-e2e ## Run all tests including E2E

test-race: ## Run all tests with race detector
	@echo "Running all tests with race detector..."
	@go test -v -race ./...

test-unit: ## Run only unit tests (fast)
	@echo "Running unit tests..."
	@go test -v -race -short ./internal/... ./pkg/... ./cmd/...

test-unit-verbose: ## Run unit tests with verbose output
	@echo "Running unit tests (verbose)..."
	@go test -v -race -short ./internal/... ./pkg/... ./cmd/...

test-integration: ## Run integration tests (requires Docker)
	@echo "Running integration tests..."
	@TESTCONTAINERS_RYUK_DISABLED=true go test -v -race -run Integration ./test/integration/...

test-e2e: ## Run end-to-end tests
	@echo "Running E2E tests..."
	@TESTCONTAINERS_RYUK_DISABLED=true go test -v -race ./test/e2e/...

test-performance: ## Run performance tests
	@echo "Running performance tests..."
	@TESTCONTAINERS_RYUK_DISABLED=true go test -v -run=Test ./test/performance/...

test-bench: ## Run benchmark tests
	@echo "Running benchmarks..."
	@go test -bench=. -benchmem -run=^Benchmark ./test/performance/...

test-coverage: ## Run tests with coverage report
	@echo "Running tests with coverage..."
	@go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
	@go tool cover -func=coverage.out | grep total
	@echo ""
	@echo "Coverage report saved to coverage.out"
	@echo "Run 'make test-coverage-html' to view in browser"

test-coverage-html: test-coverage ## Generate and open HTML coverage report
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report saved to coverage.html"
	@open coverage.html || xdg-open coverage.html || echo "Open coverage.html in your browser"

test-verbose: ## Run tests with verbose output
	@go test -v -race ./...

test-one: ## Run a specific test (usage: make test-one TEST=TestName)
	@if [ -z "$(TEST)" ]; then \
		echo "Error: TEST variable not set"; \
		echo "Usage: make test-one TEST=TestName"; \
		exit 1; \
	fi
	@go test -v -race -run $(TEST) ./...

test-watch: ## Watch for changes and run tests (requires entr)
	@echo "Watching for changes (requires 'entr' - install with brew install entr)..."
	@find . -name '*.go' | entr -c make test-unit

test-daemon: ## Run daemon tests (unit + integration)
	@echo "Running daemon tests..."
	@make test-unit
	@make test-integration

test-extension: ## Run extension tests
	@echo "Running extension tests..."
	@cd vscode-extension && npm test

lint: lint-daemon lint-extension ## Run all linters (daemon + extension)

lint-daemon: ## Run Go linter
	@echo "Linting Go code..."
	@golangci-lint run --config .golangci.yml --timeout=5m ./...
	@echo "✓ Go linting passed"

lint-extension: ## Run TypeScript/React linters
	@echo "Linting TypeScript/React code..."
	@cd vscode-extension && npm run lint
	@echo "✓ Extension linting passed"

clean: ## Clean build artifacts
	@rm -rf bin/
	@echo "Cleaned build artifacts"

docker-up: ## Start PostgreSQL container
	@docker-compose up -d
	@echo "Waiting for PostgreSQL to be ready..."
	@sleep 3

docker-down: ## Stop PostgreSQL container
	@docker-compose down

install-deps: ## Install Go dependencies
	@go mod download
	@go mod tidy

.DEFAULT_GOAL := help

```

Let me verify the Makefile syntax by running a dry-run:

The Makefile syntax is valid. Let me run the actual install to test it:

The install works correctly. Let me verify the backups were created:

The backup was created with the timestamp. The `install-hook` target now:

1. **Backs up existing files** with a date-time stamp (`YYYYMMDD-HHMMSS` format):
   - `~/.cursor/hooks/ctd-hook` → `~/.cursor/hooks/ctd-hook.20251125-064314`
   - `~/.cursor/hooks.json` → `~/.cursor/hooks.json.20251125-064314`
   - `~/.config/ctd/hook.yaml` → `~/.config/ctd/hook.yaml.20251125-064314`

2. **Installs the binary** to `~/.cursor/hooks/ctd-hook`

3. **Installs config files** (only if they don't exist):
   - `~/.cursor/hooks.json` from `deploy/hooks.json.example`
   - `~/.config/ctd/hook.yaml` from `deploy/hook.yaml.example`

4. **Preserves existing configs** - if configs already exist, it backs them up but doesn't overwrite, informing the user to merge manually if needed

---

**User**

does the docker make target install all migrations by default ? also I want to persist my db content, I dont want to lose it everytime I rebuild my docker container - what are my options here ?

---

**Cursor**

Let me check the Docker setup and how migrations are handled:

Here's the current setup:

## How Migrations Work

**Yes, migrations are applied automatically.** The `docker-compose.yml` has two mechanisms:

1. **Initial setup** (line 16): `./migrations:/docker-entrypoint-initdb.d` - PostgreSQL runs all `.sql` files in this directory **only on first container creation** (when the data volume is empty)

2. **Migrator service** (lines 24-41): A separate container that runs after postgres is healthy and applies all migrations using `psql`. This ensures migrations run **even on existing volumes**.

Both migrations use `CREATE TABLE IF NOT EXISTS` and `CREATE INDEX IF NOT EXISTS`, making them **idempotent** (safe to run multiple times).

## Data Persistence

**Your data is already persisted.** The named volume `postgres_data` (line 44-45) stores the database files:

```yaml
volumes:
  - postgres_data:/var/lib/postgresql/data
```

### What preserves data:
- `docker-compose down` - Stops containers, **keeps volume**
- `docker-compose up -d` - Restarts, data intact
- `docker-compose restart` - Data intact

### What destroys data:
- `docker-compose down -v` - The `-v` flag **removes volumes**
- `docker volume rm cursor-thought-db_postgres_data` - Explicitly removes volume
- `docker system prune --volumes` - Removes all unused volumes

### To verify your volume exists:
```bash
docker volume ls | grep postgres_data
```

### To check data persists:
```bash
# Stop and remove containers (NOT volumes)
docker-compose down

# Bring back up
docker-compose up -d

# Your data should still be there
```

### Options for Additional Durability

If you want even more protection:

**Option 1: Bind mount to a specific directory** (instead of named volume)
```yaml
volumes:
  - ./data/postgres:/var/lib/postgresql/data  # Local directory
```

**Option 2: Regular backups**
```bash
# Add to Makefile
docker-backup:
    docker exec cursor-thought-db pg_dump -U ctd cursor_thoughts > backup-$(date +%Y%m%d).sql

docker-restore:
    cat $(BACKUP) | docker exec -i cursor-thought-db psql -U ctd cursor_thoughts
```

Would you like me to add backup/restore targets to the Makefile?

```yaml
volumes:
  - postgres_data:/var/lib/postgresql/data
```

```shellscript
docker volume ls | grep postgres_data
```

```shellscript
# Stop and remove containers (NOT volumes)
docker-compose down

# Bring back up
docker-compose up -d

# Your data should still be there
```

```yaml
volumes:
  - ./data/postgres:/var/lib/postgresql/data  # Local directory
```

```shellscript
# Add to Makefile
docker-backup:
    docker exec cursor-thought-db pg_dump -U ctd cursor_thoughts > backup-$(date +%Y%m%d).sql

docker-restore:
    cat $(BACKUP) | docker exec -i cursor-thought-db psql -U ctd cursor_thoughts
```

---

**User**

Implement option 2 and bring down my database and bring it up and make sure all the migrations are run

---

**Cursor**

```makefile
.PHONY: help build build-hook build-all run test lint lint-daemon lint-extension clean docker-up docker-down install-deps install install-hook uninstall uninstall-hook

help: ## Display this help message
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

gen: ## Generate Go code from Proto
	@echo "Generating Go code..."
	@buf generate
	@echo "Generated Go code in gen/"

build: ## Build the daemon binary
	@echo "Building ctd..."
	@go build -o bin/ctd ./cmd/thought-daemon

build-hook: ## Build the hook binary
	@echo "Building ctd-hook..."
	@go build -o bin/ctd-hook ./cmd/ctd-hook

build-all: build build-hook ## Build both daemon and hook binaries

install: build ## Install ctd to ~/.local/bin (no sudo required)
	@echo "Installing ctd..."
	@mkdir -p ~/.local/bin
	@cp bin/ctd ~/.local/bin/
	@chmod +x ~/.local/bin/ctd
	@echo " Installed to ~/.local/bin/ctd"
	@echo ""
	@if echo $$PATH | grep -q "$$HOME/.local/bin"; then \
		echo " ~/.local/bin is in your PATH"; \
		echo "  You can now run: ctd"; \
	else \
		echo " ~/.local/bin is not in your PATH"; \
		echo "  Add it by running:"; \
		echo "    echo 'export PATH=\"\$$HOME/.local/bin:\$$PATH\"' >> ~/.bashrc"; \
		echo "    source ~/.bashrc"; \
		echo "  (or ~/.zshrc, ~/.profile, etc. depending on your shell)"; \
	fi
	@echo ""
	@mkdir -p ~/.config/ctd
	@if [ ! -f ~/.config/ctd/config.yaml ]; then \
		cp config.yaml.example ~/.config/ctd/config.yaml; \
		echo " Created default config at ~/.config/ctd/config.yaml"; \
	else \
		echo "ℹ Config already exists at ~/.config/ctd/config.yaml"; \
	fi

extension-install: ## Build and install the VS Code extension to Cursor
	@echo "Installing VS Code extension dependencies..."
	@cd vscode-extension && npm install
	@echo "Building VS Code extension..."
	@cd vscode-extension && npm run build
	@echo "Packaging VS Code extension..."
	@cd vscode-extension && ./node_modules/.bin/vsce package
	@echo "Installing extension to Cursor..."
	@version=$$(cd vscode-extension && node -p "require('./package.json').version"); \
	cursor --uninstall-extension cursor-user.cursor-thought-stream 2>/dev/null || true; \
	cursor --install-extension vscode-extension/cursor-thought-stream-$$version.vsix --force; \
	echo "✅ Extension v$$version installed successfully!"; \
	echo "🔄 Please reload Cursor window (Cmd+Shift+P -> 'Developer: Reload Window')"

install-hook: build-hook ## Install ctd-hook to ~/.cursor/hooks with config files
	@echo "Installing ctd-hook..."
	@TIMESTAMP=$$(date +%Y%m%d-%H%M%S); \
	mkdir -p ~/.cursor/hooks; \
	mkdir -p ~/.config/ctd; \
	if [ -f ~/.cursor/hooks/ctd-hook ]; then \
		echo "[BACKUP] Backing up existing binary to ~/.cursor/hooks/ctd-hook.$$TIMESTAMP"; \
		cp ~/.cursor/hooks/ctd-hook ~/.cursor/hooks/ctd-hook.$$TIMESTAMP; \
	fi; \
	cp bin/ctd-hook ~/.cursor/hooks/; \
	chmod +x ~/.cursor/hooks/ctd-hook; \
	echo "[OK] Installed binary to ~/.cursor/hooks/ctd-hook"; \
	echo ""; \
	if [ -f ~/.cursor/hooks.json ]; then \
		echo "[BACKUP] Backing up existing hooks.json to ~/.cursor/hooks.json.$$TIMESTAMP"; \
		cp ~/.cursor/hooks.json ~/.cursor/hooks.json.$$TIMESTAMP; \
		echo "[INFO] Existing hooks.json preserved - merge manually if needed"; \
		echo "       Example config: deploy/hooks.json.example"; \
	else \
		cp deploy/hooks.json.example ~/.cursor/hooks.json; \
		echo "[OK] Installed hooks.json to ~/.cursor/hooks.json"; \
	fi; \
	echo ""; \
	if [ -f ~/.config/ctd/hook.yaml ]; then \
		echo "[BACKUP] Backing up existing hook.yaml to ~/.config/ctd/hook.yaml.$$TIMESTAMP"; \
		cp ~/.config/ctd/hook.yaml ~/.config/ctd/hook.yaml.$$TIMESTAMP; \
		echo "[INFO] Existing hook.yaml preserved - merge manually if needed"; \
		echo "       Example config: deploy/hook.yaml.example"; \
	else \
		cp deploy/hook.yaml.example ~/.config/ctd/hook.yaml; \
		echo "[OK] Installed hook config to ~/.config/ctd/hook.yaml"; \
	fi; \
	echo ""; \
	echo "[OK] Hook installation complete"; \
	echo "     Restart Cursor to activate hooks"

uninstall: ## Remove ctd from ~/.local/bin
	@echo "Uninstalling ctd..."
	@rm -f ~/.local/bin/ctd
	@echo "[OK] Removed ~/.local/bin/ctd"
	@echo "[INFO] Config remains at ~/.config/ctd/ (remove manually if desired)"

uninstall-hook: ## Remove ctd-hook from ~/.cursor/hooks
	@echo "Uninstalling ctd-hook..."
	@rm -f ~/.cursor/hooks/ctd-hook
	@echo "[OK] Removed ~/.cursor/hooks/ctd-hook"
	@echo "[INFO] Remember to update ~/.cursor/hooks.json to remove hook references"

run: ## Run the daemon locally
	@go run ./cmd/thought-daemon

test: test-daemon test-extension ## Run all tests (daemon + extension)

test-all: test-unit test-integration test-e2e ## Run all tests including E2E

test-race: ## Run all tests with race detector
	@echo "Running all tests with race detector..."
	@go test -v -race ./...

test-unit: ## Run only unit tests (fast)
	@echo "Running unit tests..."
	@go test -v -race -short ./internal/... ./pkg/... ./cmd/...

test-unit-verbose: ## Run unit tests with verbose output
	@echo "Running unit tests (verbose)..."
	@go test -v -race -short ./internal/... ./pkg/... ./cmd/...

test-integration: ## Run integration tests (requires Docker)
	@echo "Running integration tests..."
	@TESTCONTAINERS_RYUK_DISABLED=true go test -v -race -run Integration ./test/integration/...

test-e2e: ## Run end-to-end tests
	@echo "Running E2E tests..."
	@TESTCONTAINERS_RYUK_DISABLED=true go test -v -race ./test/e2e/...

test-performance: ## Run performance tests
	@echo "Running performance tests..."
	@TESTCONTAINERS_RYUK_DISABLED=true go test -v -run=Test ./test/performance/...

test-bench: ## Run benchmark tests
	@echo "Running benchmarks..."
	@go test -bench=. -benchmem -run=^Benchmark ./test/performance/...

test-coverage: ## Run tests with coverage report
	@echo "Running tests with coverage..."
	@go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
	@go tool cover -func=coverage.out | grep total
	@echo ""
	@echo "Coverage report saved to coverage.out"
	@echo "Run 'make test-coverage-html' to view in browser"

test-coverage-html: test-coverage ## Generate and open HTML coverage report
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report saved to coverage.html"
	@open coverage.html || xdg-open coverage.html || echo "Open coverage.html in your browser"

test-verbose: ## Run tests with verbose output
	@go test -v -race ./...

test-one: ## Run a specific test (usage: make test-one TEST=TestName)
	@if [ -z "$(TEST)" ]; then \
		echo "Error: TEST variable not set"; \
		echo "Usage: make test-one TEST=TestName"; \
		exit 1; \
	fi
	@go test -v -race -run $(TEST) ./...

test-watch: ## Watch for changes and run tests (requires entr)
	@echo "Watching for changes (requires 'entr' - install with brew install entr)..."
	@find . -name '*.go' | entr -c make test-unit

test-daemon: ## Run daemon tests (unit + integration)
	@echo "Running daemon tests..."
	@make test-unit
	@make test-integration

test-extension: ## Run extension tests
	@echo "Running extension tests..."
	@cd vscode-extension && npm test

lint: lint-daemon lint-extension ## Run all linters (daemon + extension)

lint-daemon: ## Run Go linter
	@echo "Linting Go code..."
	@golangci-lint run --config .golangci.yml --timeout=5m ./...
	@echo "✓ Go linting passed"

lint-extension: ## Run TypeScript/React linters
	@echo "Linting TypeScript/React code..."
	@cd vscode-extension && npm run lint
	@echo "✓ Extension linting passed"

clean: ## Clean build artifacts
	@rm -rf bin/
	@echo "Cleaned build artifacts"

docker-up: ## Start PostgreSQL container and run migrations
	@docker-compose up -d
	@echo "Waiting for PostgreSQL and migrations to complete..."
	@sleep 5
	@docker-compose logs migrator 2>/dev/null | tail -10 || true
	@echo "[OK] Database ready"

docker-down: ## Stop PostgreSQL container (preserves data)
	@docker-compose down
	@echo "[OK] Containers stopped (data preserved in volume)"

docker-reset: ## Stop PostgreSQL and DELETE all data
	@echo "WARNING: This will delete all database data!"
	@read -p "Are you sure? [y/N] " confirm && [ "$$confirm" = "y" ] || exit 1
	@docker-compose down -v
	@echo "[OK] Containers and volumes removed"

docker-backup: ## Backup database to timestamped SQL file
	@mkdir -p backups
	@BACKUP_FILE="backups/ctd-backup-$$(date +%Y%m%d-%H%M%S).sql"; \
	echo "Backing up database to $$BACKUP_FILE..."; \
	docker exec cursor-thought-db pg_dump -U ctd cursor_thoughts > "$$BACKUP_FILE"; \
	echo "[OK] Backup saved to $$BACKUP_FILE"

docker-restore: ## Restore database from backup (usage: make docker-restore BACKUP=backups/file.sql)
	@if [ -z "$(BACKUP)" ]; then \
		echo "Error: BACKUP variable not set"; \
		echo "Usage: make docker-restore BACKUP=backups/ctd-backup-20251125-123456.sql"; \
		echo ""; \
		echo "Available backups:"; \
		ls -1 backups/*.sql 2>/dev/null || echo "  (none found)"; \
		exit 1; \
	fi
	@if [ ! -f "$(BACKUP)" ]; then \
		echo "Error: Backup file not found: $(BACKUP)"; \
		exit 1; \
	fi
	@echo "Restoring from $(BACKUP)..."
	@cat "$(BACKUP)" | docker exec -i cursor-thought-db psql -U ctd cursor_thoughts
	@echo "[OK] Database restored from $(BACKUP)"

docker-logs: ## Show PostgreSQL container logs
	@docker-compose logs -f postgres

install-deps: ## Install Go dependencies
	@go mod download
	@go mod tidy

.DEFAULT_GOAL := help

```

Now let me update the `.PHONY` line to include the new targets:

```makefile
.PHONY: help build build-hook build-all run test lint lint-daemon lint-extension clean docker-up docker-down docker-reset docker-backup docker-restore docker-logs install-deps install install-hook uninstall uninstall-hook

help: ## Display this help message
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

gen: ## Generate Go code from Proto
	@echo "Generating Go code..."
	@buf generate
	@echo "Generated Go code in gen/"

build: ## Build the daemon binary
	@echo "Building ctd..."
	@go build -o bin/ctd ./cmd/thought-daemon

build-hook: ## Build the hook binary
	@echo "Building ctd-hook..."
	@go build -o bin/ctd-hook ./cmd/ctd-hook

build-all: build build-hook ## Build both daemon and hook binaries

install: build ## Install ctd to ~/.local/bin (no sudo required)
	@echo "Installing ctd..."
	@mkdir -p ~/.local/bin
	@cp bin/ctd ~/.local/bin/
	@chmod +x ~/.local/bin/ctd
	@echo " Installed to ~/.local/bin/ctd"
	@echo ""
	@if echo $$PATH | grep -q "$$HOME/.local/bin"; then \
		echo " ~/.local/bin is in your PATH"; \
		echo "  You can now run: ctd"; \
	else \
		echo " ~/.local/bin is not in your PATH"; \
		echo "  Add it by running:"; \
		echo "    echo 'export PATH=\"\$$HOME/.local/bin:\$$PATH\"' >> ~/.bashrc"; \
		echo "    source ~/.bashrc"; \
		echo "  (or ~/.zshrc, ~/.profile, etc. depending on your shell)"; \
	fi
	@echo ""
	@mkdir -p ~/.config/ctd
	@if [ ! -f ~/.config/ctd/config.yaml ]; then \
		cp config.yaml.example ~/.config/ctd/config.yaml; \
		echo " Created default config at ~/.config/ctd/config.yaml"; \
	else \
		echo "ℹ Config already exists at ~/.config/ctd/config.yaml"; \
	fi

extension-install: ## Build and install the VS Code extension to Cursor
	@echo "Installing VS Code extension dependencies..."
	@cd vscode-extension && npm install
	@echo "Building VS Code extension..."
	@cd vscode-extension && npm run build
	@echo "Packaging VS Code extension..."
	@cd vscode-extension && ./node_modules/.bin/vsce package
	@echo "Installing extension to Cursor..."
	@version=$$(cd vscode-extension && node -p "require('./package.json').version"); \
	cursor --uninstall-extension cursor-user.cursor-thought-stream 2>/dev/null || true; \
	cursor --install-extension vscode-extension/cursor-thought-stream-$$version.vsix --force; \
	echo "✅ Extension v$$version installed successfully!"; \
	echo "🔄 Please reload Cursor window (Cmd+Shift+P -> 'Developer: Reload Window')"

install-hook: build-hook ## Install ctd-hook to ~/.cursor/hooks with config files
	@echo "Installing ctd-hook..."
	@TIMESTAMP=$$(date +%Y%m%d-%H%M%S); \
	mkdir -p ~/.cursor/hooks; \
	mkdir -p ~/.config/ctd; \
	if [ -f ~/.cursor/hooks/ctd-hook ]; then \
		echo "[BACKUP] Backing up existing binary to ~/.cursor/hooks/ctd-hook.$$TIMESTAMP"; \
		cp ~/.cursor/hooks/ctd-hook ~/.cursor/hooks/ctd-hook.$$TIMESTAMP; \
	fi; \
	cp bin/ctd-hook ~/.cursor/hooks/; \
	chmod +x ~/.cursor/hooks/ctd-hook; \
	echo "[OK] Installed binary to ~/.cursor/hooks/ctd-hook"; \
	echo ""; \
	if [ -f ~/.cursor/hooks.json ]; then \
		echo "[BACKUP] Backing up existing hooks.json to ~/.cursor/hooks.json.$$TIMESTAMP"; \
		cp ~/.cursor/hooks.json ~/.cursor/hooks.json.$$TIMESTAMP; \
		echo "[INFO] Existing hooks.json preserved - merge manually if needed"; \
		echo "       Example config: deploy/hooks.json.example"; \
	else \
		cp deploy/hooks.json.example ~/.cursor/hooks.json; \
		echo "[OK] Installed hooks.json to ~/.cursor/hooks.json"; \
	fi; \
	echo ""; \
	if [ -f ~/.config/ctd/hook.yaml ]; then \
		echo "[BACKUP] Backing up existing hook.yaml to ~/.config/ctd/hook.yaml.$$TIMESTAMP"; \
		cp ~/.config/ctd/hook.yaml ~/.config/ctd/hook.yaml.$$TIMESTAMP; \
		echo "[INFO] Existing hook.yaml preserved - merge manually if needed"; \
		echo "       Example config: deploy/hook.yaml.example"; \
	else \
		cp deploy/hook.yaml.example ~/.config/ctd/hook.yaml; \
		echo "[OK] Installed hook config to ~/.config/ctd/hook.yaml"; \
	fi; \
	echo ""; \
	echo "[OK] Hook installation complete"; \
	echo "     Restart Cursor to activate hooks"

uninstall: ## Remove ctd from ~/.local/bin
	@echo "Uninstalling ctd..."
	@rm -f ~/.local/bin/ctd
	@echo "[OK] Removed ~/.local/bin/ctd"
	@echo "[INFO] Config remains at ~/.config/ctd/ (remove manually if desired)"

uninstall-hook: ## Remove ctd-hook from ~/.cursor/hooks
	@echo "Uninstalling ctd-hook..."
	@rm -f ~/.cursor/hooks/ctd-hook
	@echo "[OK] Removed ~/.cursor/hooks/ctd-hook"
	@echo "[INFO] Remember to update ~/.cursor/hooks.json to remove hook references"

run: ## Run the daemon locally
	@go run ./cmd/thought-daemon

test: test-daemon test-extension ## Run all tests (daemon + extension)

test-all: test-unit test-integration test-e2e ## Run all tests including E2E

test-race: ## Run all tests with race detector
	@echo "Running all tests with race detector..."
	@go test -v -race ./...

test-unit: ## Run only unit tests (fast)
	@echo "Running unit tests..."
	@go test -v -race -short ./internal/... ./pkg/... ./cmd/...

test-unit-verbose: ## Run unit tests with verbose output
	@echo "Running unit tests (verbose)..."
	@go test -v -race -short ./internal/... ./pkg/... ./cmd/...

test-integration: ## Run integration tests (requires Docker)
	@echo "Running integration tests..."
	@TESTCONTAINERS_RYUK_DISABLED=true go test -v -race -run Integration ./test/integration/...

test-e2e: ## Run end-to-end tests
	@echo "Running E2E tests..."
	@TESTCONTAINERS_RYUK_DISABLED=true go test -v -race ./test/e2e/...

test-performance: ## Run performance tests
	@echo "Running performance tests..."
	@TESTCONTAINERS_RYUK_DISABLED=true go test -v -run=Test ./test/performance/...

test-bench: ## Run benchmark tests
	@echo "Running benchmarks..."
	@go test -bench=. -benchmem -run=^Benchmark ./test/performance/...

test-coverage: ## Run tests with coverage report
	@echo "Running tests with coverage..."
	@go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
	@go tool cover -func=coverage.out | grep total
	@echo ""
	@echo "Coverage report saved to coverage.out"
	@echo "Run 'make test-coverage-html' to view in browser"

test-coverage-html: test-coverage ## Generate and open HTML coverage report
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report saved to coverage.html"
	@open coverage.html || xdg-open coverage.html || echo "Open coverage.html in your browser"

test-verbose: ## Run tests with verbose output
	@go test -v -race ./...

test-one: ## Run a specific test (usage: make test-one TEST=TestName)
	@if [ -z "$(TEST)" ]; then \
		echo "Error: TEST variable not set"; \
		echo "Usage: make test-one TEST=TestName"; \
		exit 1; \
	fi
	@go test -v -race -run $(TEST) ./...

test-watch: ## Watch for changes and run tests (requires entr)
	@echo "Watching for changes (requires 'entr' - install with brew install entr)..."
	@find . -name '*.go' | entr -c make test-unit

test-daemon: ## Run daemon tests (unit + integration)
	@echo "Running daemon tests..."
	@make test-unit
	@make test-integration

test-extension: ## Run extension tests
	@echo "Running extension tests..."
	@cd vscode-extension && npm test

lint: lint-daemon lint-extension ## Run all linters (daemon + extension)

lint-daemon: ## Run Go linter
	@echo "Linting Go code..."
	@golangci-lint run --config .golangci.yml --timeout=5m ./...
	@echo "✓ Go linting passed"

lint-extension: ## Run TypeScript/React linters
	@echo "Linting TypeScript/React code..."
	@cd vscode-extension && npm run lint
	@echo "✓ Extension linting passed"

clean: ## Clean build artifacts
	@rm -rf bin/
	@echo "Cleaned build artifacts"

docker-up: ## Start PostgreSQL container and run migrations
	@docker-compose up -d
	@echo "Waiting for PostgreSQL and migrations to complete..."
	@sleep 5
	@docker-compose logs migrator 2>/dev/null | tail -10 || true
	@echo "[OK] Database ready"

docker-down: ## Stop PostgreSQL container (preserves data)
	@docker-compose down
	@echo "[OK] Containers stopped (data preserved in volume)"

docker-reset: ## Stop PostgreSQL and DELETE all data
	@echo "WARNING: This will delete all database data!"
	@read -p "Are you sure? [y/N] " confirm && [ "$$confirm" = "y" ] || exit 1
	@docker-compose down -v
	@echo "[OK] Containers and volumes removed"

docker-backup: ## Backup database to timestamped SQL file
	@mkdir -p backups
	@BACKUP_FILE="backups/ctd-backup-$$(date +%Y%m%d-%H%M%S).sql"; \
	echo "Backing up database to $$BACKUP_FILE..."; \
	docker exec cursor-thought-db pg_dump -U ctd cursor_thoughts > "$$BACKUP_FILE"; \
	echo "[OK] Backup saved to $$BACKUP_FILE"

docker-restore: ## Restore database from backup (usage: make docker-restore BACKUP=backups/file.sql)
	@if [ -z "$(BACKUP)" ]; then \
		echo "Error: BACKUP variable not set"; \
		echo "Usage: make docker-restore BACKUP=backups/ctd-backup-20251125-123456.sql"; \
		echo ""; \
		echo "Available backups:"; \
		ls -1 backups/*.sql 2>/dev/null || echo "  (none found)"; \
		exit 1; \
	fi
	@if [ ! -f "$(BACKUP)" ]; then \
		echo "Error: Backup file not found: $(BACKUP)"; \
		exit 1; \
	fi
	@echo "Restoring from $(BACKUP)..."
	@cat "$(BACKUP)" | docker exec -i cursor-thought-db psql -U ctd cursor_thoughts
	@echo "[OK] Database restored from $(BACKUP)"

docker-logs: ## Show PostgreSQL container logs
	@docker-compose logs -f postgres

install-deps: ## Install Go dependencies
	@go mod download
	@go mod tidy

.DEFAULT_GOAL := help

```

Now let's bring down the database and bring it back up:

There's an issue with the migration script - the glob pattern isn't expanding correctly. Let me fix the docker-compose.yml:

```dockercompose
version: '3.8'

services:
  postgres:
    image: postgres:16-alpine
    container_name: cursor-thought-db
    environment:
      POSTGRES_DB: cursor_thoughts
      POSTGRES_USER: ctd
      POSTGRES_PASSWORD: ctd_dev_pass
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
      # Keep this for fresh installs
      - ./migrations:/docker-entrypoint-initdb.d
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ctd -d cursor_thoughts"]
      interval: 5s
      timeout: 5s
      retries: 5

  # Explicit migration service to ensure schema is applied even on existing volumes
  migrator:
    image: postgres:16-alpine
    restart: "no"
    depends_on:
      postgres:
        condition: service_healthy
    volumes:
      - ./migrations:/migrations:ro
    environment:
      PGPASSWORD: ctd_dev_pass
    entrypoint: ["/bin/sh", "-c"]
    command:
      - |
        echo "Running migrations..."
        for f in /migrations/*.sql; do
          if [ -f "$${f}" ]; then
            echo "Applying $${f}..."
            psql -h postgres -U ctd -d cursor_thoughts -f "$${f}"
          fi
        done
        echo "Migrations complete."

volumes:
  postgres_data:

```

---

**User**

whats taking so long ? 

---

**Cursor**

Let me check the container status and logs:

The migrator ran but failed because the glob didn't expand. The fix I made hasn't been applied yet (the container used the old config). Let me restart:

Still not finding the files. Let me check what's in the migrations directory inside the container:

The volume mount is empty. This is a Docker Desktop for Mac issue with `/Volumes` paths. Let me check the actual path:

The files exist locally but Docker can't see them. This is likely a Docker Desktop file sharing issue with `/Volumes`. Let me check if `/Volumes/Development` is shared in Docker:

Docker Desktop on macOS isn't sharing `/Volumes/Development`. You need to add it to Docker's file sharing settings. Let me check if there's a symlink or alternative path:

---

**User**

NB context: I use colima and the docker cli - not Docker Desktopso 

---

**Cursor**

Ah, Colima! That's the issue. Colima needs the path mounted into the VM. Let me check your Colima config:

Only `/Users/jamesainslie` is mounted, not `/Volumes/Development`. You have two options:

**Option 1: Add `/Volumes/Development` to Colima mounts** (requires restart):
```bash
colima stop
colima start --mount /Volumes/Development:w
```

**Option 2 (Quick fix): Restart Colima with the additional mount**:

```shellscript
colima stop
colima start --mount /Volumes/Development:w
```

You need to add `/Volumes/Development` as a mount. Run:

```bash
colima stop
colima start --mount /Users/jamesainslie:w --mount /Volumes/Development:w
```

Or to make it permanent, edit `~/.colima/default/colima.yaml` and add to the mounts section:

```yaml
mounts:
  - location: /Users/jamesainslie
    writable: true
  - location: /Volumes/Development
    writable: true
```

Would you like me to proceed with restarting Colima with the additional mount? This will briefly stop all your running containers.

```shellscript
colima stop
colima start --mount /Users/jamesainslie:w --mount /Volumes/Development:w
```

```shellscript
colima stop
colima start --mount /Users/jamesainslie:w --mount /Volumes/Development:w
```

```yaml
mounts:
  - location: /Users/jamesainslie
    writable: true
  - location: /Volumes/Development
    writable: true
```

---

**User**

Make it permanent and restart colima please. Do so safely for my db 

please continue

