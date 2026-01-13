# Ralph Setup

Bootstrap [Ralph](https://www.aihero.dev/getting-started-with-ralph) in any project - the AI coding loop that lets you ship while you sleep.

## What is Ralph?

Ralph is a technique created by [Geoffrey Huntley](https://ghuntley.com/ralph/) for running AI coding agents in a loop:

1. You describe what you want to build
2. Claude interviews you and creates a PRD (task list)
3. Claude implements tasks one at a time, committing after each
4. You come back to working code

## Installation

### Option 1: Clone this repo

```bash
git clone https://github.com/xaelophone/ralph-setup.git ~/ralph-setup
chmod +x ~/ralph-setup/setup-ralph

# Add to PATH (in ~/.zshrc or ~/.bashrc)
export PATH="$HOME/ralph-setup:$PATH"
```

### Option 2: Copy to existing scripts directory

```bash
# If you already have ~/bin/ralph or similar
curl -fsSL https://raw.githubusercontent.com/xaelophone/ralph-setup/main/setup-ralph -o ~/bin/ralph/setup-ralph
chmod +x ~/bin/ralph/setup-ralph
```

## Usage

```bash
# In any new or existing project
cd my-project
git init                    # If not already a git repo
setup-ralph                 # Initialize Ralph
claude                      # Start Claude Code
```

### For Frontend Projects (Recommended)

Enable browser tools for a visual feedback loop. **[Full setup guide →](https://code.claude.com/docs/en/chrome)**

**Prerequisites:**
- [Google Chrome](https://www.google.com/chrome/)
- [Claude in Chrome extension](https://chromewebstore.google.com/detail/claude/fcoeoabgfenejglbffodgkkbkcdhcgfn) (v1.0.36+)
- Claude Code CLI (v2.0.73+) — run `claude update` to ensure latest
- Paid Claude plan (Pro, Team, or Enterprise)

```bash
claude --chrome             # Start with Chrome integration
# Or enable permanently: run /chrome → "Enable by default"
```

With Chrome enabled, Claude can verify UI changes in a real browser before committing—navigating pages, checking console errors, and testing interactions.

Then tell Claude what you want to build. It will:
1. **Interview you** about the project
2. **Generate PRD.md** with small, prioritized tasks
3. **Implement tasks** one by one, committing after each

## How It Works

`setup-ralph` creates two files:

| File | Purpose |
|------|---------|
| `CLAUDE.md` | Instructions for Claude (the Ralph workflow) |
| `progress.txt` | Log of completed work |

Claude reads `CLAUDE.md` automatically and knows to:
- **Interview you** if no PRD.md exists
- **Implement tasks** if PRD.md has incomplete items
- **Celebrate & offer next steps** if PRD.md is complete

## The Ralph Cycle

```
┌─────────────────────────────────────────────────────────────┐
│  1. Read PRD.md (task list) and progress.txt (history)      │
│  2. Pick the highest-priority incomplete task               │
│  3. Implement ONLY that task                                │
│  4. Run tests and type checks (must pass!)                  │
│  5. For UI work: verify in browser (if Chrome enabled)      │
│  6. Mark task complete in PRD.md                            │
│  7. Update progress.txt with what was done                  │
│  8. Commit changes                                          │
│  9. Repeat until PRD complete                               │
└─────────────────────────────────────────────────────────────┘
```

## Example PRD

Claude will create something like this:

```markdown
# My App - PRD

## Overview
Building a CLI todo app in Python

## Tasks
- [x] Set up project structure with pyproject.toml
- [x] Create Todo dataclass model
- [ ] Implement add command
- [ ] Implement list command
- [ ] Implement complete command
- [ ] Add JSON persistence
- [ ] Write tests

## Technical Notes
- Python 3.11+, Click for CLI, pytest for tests
```

## Commands to Tell Claude

| You Say | Claude Does |
|---------|-------------|
| "Let's build X" | Interviews you, creates PRD |
| "Continue" / "Next task" | Implements next incomplete task |
| "What's left?" | Shows remaining tasks |
| "Add feature Y" | Adds new tasks to PRD |
| "Let's start fresh" | Archives old PRD, interviews for new one |

## Tips

### Task Sizing (Critical!)

**Good tasks** (~15-30 min):
- "Create login form component"
- "Add API endpoint for /users"
- "Write tests for auth service"

**Bad tasks** (too big - break them down):
- "Implement user authentication"
- "Build the entire dashboard"

### Test Requirements

Every commit must:
- ✅ Pass all existing tests
- ✅ Pass type checks (if applicable)
- ✅ Not break the build

### When to Use Ralph

**Good for:**
- Greenfield projects
- Feature implementations
- Refactoring with clear goals

**Not ideal for:**
- Quick bug fixes (just ask directly)
- Exploration/research

## Troubleshooting

| Problem | Solution |
|---------|----------|
| Claude doesn't follow workflow | Run `setup-ralph` again to recreate CLAUDE.md |
| Tasks are too big | Ask Claude to "break this down into smaller tasks" |
| Tests keep failing | Tell Claude "fix the tests before moving on" |
| Want to start over | Delete PRD.md, tell Claude "let's plan something new" |

## Credits

- **Ralph creator**: [Geoffrey Huntley](https://ghuntley.com/ralph/)
- **This implementation based on**: [Matt Pocock's guide](https://www.aihero.dev/getting-started-with-ralph) and [video walkthrough](https://www.youtube.com/watch?v=_IK18goX4X8)
- **Additional tips**: [11 Tips for AI Coding with Ralph](https://www.aihero.dev/tips-for-ai-coding-with-ralph-wiggum)

## License

MIT - Do whatever you want with it!
