# Ralph Setup

Bootstrap [Ralph](https://www.aihero.dev/getting-started-with-ralph) in any project - the AI coding loop that lets you ship while you sleep.

## What is Ralph?

Ralph is a technique created by [Geoffrey Huntley](https://ghuntley.com/ralph/) for running AI coding agents in a loop:

1. You describe what you want to build
2. Claude interviews you and creates a PRD (task list)
3. Claude implements tasks one at a time, committing after each
4. You come back to working code

## Quick Start

### 1. Install (one time)

```bash
curl -fsSL https://raw.githubusercontent.com/xaelophone/ralph-setup/main/install.sh | bash
exec $SHELL  # Reload PATH
```

### 2a. New Project (start from scratch)

```bash
mkdir my-new-app && cd my-new-app
git init
setup-ralph                # Creates CLAUDE.md + progress.txt
claude                     # Claude will interview you and create PRD.md
# Once PRD.md exists with tasks:
ralph-loop                 # Run autonomously overnight
```

### 2b. Existing Project (add Ralph to existing code)

```bash
cd my-existing-project
setup-ralph                # Creates CLAUDE.md + progress.txt

# Option A: Let Claude analyze and create a PRD
claude                     # Tell Claude what you want to build/fix
                           # It will create PRD.md with tasks

# Option B: Create PRD.md yourself with tasks, then run:
ralph-loop                 # Claude works through your task list
```

### 2c. From a GitHub Issue (bridges strategic + tactical)

If you plan features as GitHub Issues (recommended for teams):

```bash
cd my-project
setup-ralph                # Initialize Ralph files

ralph-gh link 42           # Link to GitHub Issue #42
ralph-gh sync              # Create PRD.md from issue content
# Edit PRD.md to break into atomic 🤖/🧑 tasks

claude                     # Implement
ralph-gh post              # Post progress back to the issue
```

### 3. Check Progress

Come back in the morning to find:
- ✅ Completed tasks in `PRD.md`
- 📝 Work log in `progress.txt`
- 🧑 Human tasks listed in `HANDOFF.md` (if any)

## Tools Included

| Tool | Description |
|------|-------------|
| `ralph-gh` | Bridge between GitHub Issues and local Ralph files |
| `setup-ralph` | Initializes a project with CLAUDE.md and progress.txt |
| `ralph-loop` | Autonomous Claude runner with completion detection |

### setup-ralph

Creates the Ralph workflow files in your project:

```bash
setup-ralph
# Creates:
#   CLAUDE.md    - Instructions for Claude (the Ralph workflow)
#   progress.txt - Log of completed work
```

### ralph-loop

Runs Claude autonomously with real-time completion detection:

```bash
ralph-loop                    # Run with defaults
ralph-loop --max-iterations 5 # Limit iterations
ralph-loop --resume           # Resume crashed session
ralph-loop --status           # Show session status
```

**Key features:**
- 🔄 Automatic restart when Claude's context fills up
- ✅ Detects `<promise>COMPLETE</promise>` token to continue to next task
- 💾 Session persistence with crash recovery
- 📝 Cross-iteration context injection
- 🤖 Skips human tasks (🧑), works only on AI tasks (🤖)

**Output example:**
```
  ██████╗  █████╗ ██╗     ██████╗ ██╗  ██╗
  ██╔══██╗██╔══██╗██║     ██╔══██╗██║  ██║
  ██████╔╝███████║██║     ██████╔╝███████║
  ██╔══██╗██╔══██║██║     ██╔═══╝ ██╔══██║
  ██║  ██║██║  ██║███████╗██║     ██║  ██║
  ╚═╝  ╚═╝╚═╝  ╚═╝╚══════╝╚═╝     ╚═╝  ╚═╝

  Autonomous Claude Loop v2.0.0

═══════════════════════════════════════════════════════════════
Iteration 3 of 100
───────────────────────────────────────────────────────────────
Progress:    5/12 tasks (41%)
Remaining:   7 🤖 tasks, 3 🧑 tasks
Current:     Implement user authentication
═══════════════════════════════════════════════════════════════
```

### ralph-gh

Bridges GitHub Issues (strategic planning) with local Ralph files (tactical execution):

```bash
ralph-gh link 42           # Link project to issue #42
ralph-gh sync              # Create PRD.md from issue content
ralph-gh post              # Post progress summary as issue comment
ralph-gh status            # Show current link and progress
```

**The two-layer approach:**

| Layer | Where | Purpose |
|-------|-------|---------|
| **Strategic** | GitHub Issues | Team visibility, stakeholder access, cross-project planning |
| **Tactical** | PRD.md + progress.txt | Within-session task tracking, context recovery |

This lets you keep GitHub Issues for high-level planning while using Ralph's local files for the granular task execution that Claude needs.

## For a Full TUI: ralph-tui

I was building my own terminal UI (`rwatch` in Go/Bubbletea) to support the core ralph loop when I came across [ralph-tui](https://github.com/subsy/ralph-tui) - a beautifully polished implementation that does everything I wanted and more. Rather than reinvent the wheel, I'm recommending it here:

```bash
# Install Bun first (if needed)
curl -fsSL https://bun.sh/install | bash
exec $SHELL

# Install ralph-tui
bun install -g ralph-tui

# Run
ralph-tui
```

**ralph-tui provides:**
- 🎨 Beautiful React-based TUI
- 🔌 Plugin system for different trackers (JSON, Beads)
- 🔔 Desktop notifications
- 📊 Subagent tracing (see Claude's tool calls)
- 🔄 Session resume with full state

**When to use which:**
| Scenario | Use |
|----------|-----|
| Headless/CI, minimal dependencies | `ralph-loop` |
| Interactive development, fancy UI | `ralph-tui` |

> **Note:** The experimental `rwatch` Go code is still in this repo under `cmd/rwatch/` and `internal/` if you want to build your own native TUI. Run `make build` to compile it.

## Task Markers

Mark tasks in your PRD.md to control what runs autonomously:

```markdown
- [ ] 🤖 Create API endpoint     # Claude does this
- [ ] 🧑 Set up AWS account      # Human does this (skipped)
- [x] 🤖 Write database schema   # Already done
```

## The Completion Protocol

`ralph-loop` and `ralph-tui` detect when Claude finishes a task using a special token:

```
<promise>COMPLETE</promise>
```

Claude outputs this after:
1. Implementing the task
2. Passing tests
3. Committing code
4. Updating progress.txt

This lets the loop know to continue to the next task automatically.

## Example PRD

Claude will create something like this:

```markdown
# My App - PRD

> **Legend:** 🤖 = AI task | 🧑 = Human task

## Tasks

### Phase 1: Foundation
- [x] 🤖 Set up project structure
- [x] 🤖 Create data models
- [ ] 🤖 Implement API endpoints
- [ ] 🧑 Set up production database

### Phase 2: Features
- [ ] 🤖 Add authentication
- [ ] 🤖 Create user dashboard
- [ ] 🧑 Configure OAuth provider
```

## How It Works

```
┌─────────────────────────────────────────────────────────────┐
│  1. Read PRD.md (task list) and progress.txt (history)      │
│  2. Pick the highest-priority incomplete 🤖 task            │
│  3. Implement ONLY that task                                │
│  4. Run tests and type checks (must pass!)                  │
│  5. Mark task complete in PRD.md                            │
│  6. Update progress.txt with what was done                  │
│  7. Commit changes                                          │
│  8. Output <promise>COMPLETE</promise>                      │
│  9. Loop detects token, continues to next task              │
└─────────────────────────────────────────────────────────────┘
```

## Tips

### Task Sizing (Critical!)

**Good tasks** (~15-30 min):
- "Create login form component"
- "Add API endpoint for /users"
- "Write tests for auth service"

**Bad tasks** (too big - break them down):
- "Implement user authentication"
- "Build the entire dashboard"

### Overnight Runs

For best results running overnight:

1. Mark tasks clearly with 🤖 or 🧑
2. Ensure tests exist and pass
3. Run `ralph-loop` or `ralph-tui`
4. Check HANDOFF.md in the morning for blocked/human tasks

## Troubleshooting

| Problem | Solution |
|---------|----------|
| "CLAUDE.md not found" | Run `setup-ralph` first |
| "Another ralph-loop running" | Delete `.ralph.lock` or use `--resume` |
| Claude doesn't complete tasks | Add completion protocol to CLAUDE.md |
| Tasks too big | Ask Claude to "break this down into smaller tasks" |

## Credits

- **Ralph creator**: [Geoffrey Huntley](https://ghuntley.com/ralph/)
- **ralph-tui**: [subsy/ralph-tui](https://github.com/subsy/ralph-tui) - Full-featured TUI
- **This guide based on**: [Matt Pocock's guide](https://www.aihero.dev/getting-started-with-ralph)

## License

MIT - Do whatever you want with it!
