# Gas Town Crew Workflow

## The Golden Rule: Never Work on Main

**Crew members should always work on feature branches, never directly on `main`.**

### Why?

The `gt stabilize` command syncs worktrees by fast-forwarding to origin/main:
```bash
git pull origin main --ff-only
```

**This requires:**
- Local main has no uncommitted changes
- Local main is a direct ancestor of origin/main (no divergent commits)

Working on main with local commits breaks this assumption and causes stabilize to fail silently.

---

## Proper Workflow

### 1. Start a Feature Branch

```bash
cd ~/gt/<rig>/crew/<your-name>/rig

# Create and checkout feature branch
git checkout -b feature/my-feature-name

# Or based on an integration branch for epic work
git checkout -b feature/gt-auth-epic/my-task integration/st-auth-epic
```

### 2. Do Your Work

```bash
# Make changes
vim internal/cmd/myfeature.go

# Commit to your feature branch (not main!)
git add .
git commit -m "feat: implement awesome feature"
git push origin feature/my-feature-name
```

### 3. Create Merge Request

```bash
# Create bead for tracking
bd create --title="Implement awesome feature" --type=feature

# Submit to merge queue
gt done  # or gt mq submit
```

### 4. Keep Main Clean

```bash
# Switch back to main
git checkout main

# Ensure it's clean (no uncommitted changes)
git status  # Should be "nothing to commit, working tree clean"

# Ensure it's tracking origin
git log --oneline -1  # Should match origin/main
```

### 5. Stabilize Works Perfectly

```bash
# Mayor updates gastown source
cd ~/gt/gastown/mayor/rig
git fetch upstream
git merge upstream/main
make install
gt stabilize

# ✓ Crew worktrees fast-forward cleanly
# ✓ Your feature branches are unaffected
# ✓ Everyone stays in sync
```

---

## What Happens If You Work on Main

### Scenario 1: Uncommitted Changes
```bash
cd ~/gt/gastown/crew/dunks/rig
# Edit files but don't commit
gt stabilize  # ✗ Fails silently

# Output shows:
# ✓ Worktrees synced  # <-- LIE! It actually failed

# Reality: Your changes are stashed or blocking the pull
```

### Scenario 2: Commits on Main
```bash
cd ~/gt/gastown/crew/dunks/rig
git commit -m "WIP: my local changes"  # On main!
gt stabilize

# ✗ Error: "Not possible to fast-forward, aborting"
# ✓ Worktrees synced  # <-- Still reports success!

# Reality: Your main branch diverged from origin
```

---

## Recovery: If You Already Committed to Main

### Option 1: Move commits to feature branch (recommended)

```bash
cd ~/gt/<rig>/crew/<your-name>/rig

# Create feature branch from current main
git checkout -b feature/my-wip

# Go back to main and reset to origin
git checkout main
git fetch origin
git reset --hard origin/main

# Now your WIP is safe on feature branch
git checkout feature/my-wip
```

### Option 2: Force reset (discard work)

```bash
cd ~/gt/<rig>/crew/<your-name>/rig
git checkout main
git fetch origin
git reset --hard origin/main

# ⚠️ WARNING: This discards your local commits!
```

---

## Integration Branches for Epics

For large features spanning multiple beads:

```bash
# Create integration branch for epic
gt mq integration create gt-auth-epic

# Crew members branch from integration branch
git checkout -b feature/gt-auth-epic/login integration/st-auth-epic
git checkout -b feature/gt-auth-epic/logout integration/st-auth-epic

# When done, land epic
gt mq integration land gt-auth-epic
```

See: [Integration Branches](../concepts/integration-branches.md)

---

## Quick Checklist

**Before running `gt stabilize`:**

- [ ] Crew worktrees on main have no uncommitted changes
- [ ] Crew worktrees on main have no local commits
- [ ] All WIP is on feature branches
- [ ] Feature branches are pushed to origin

**If stabilize fails:**

1. Check crew worktrees: `cd ~/gt/<rig>/crew/* && git status`
2. Move WIP to feature branches
3. Reset main to origin: `git reset --hard origin/main`
4. Run stabilize again

---

## Related

- [Integration Branches](../concepts/integration-branches.md) - Epic work grouping
- [Polecat Lifecycle](../concepts/polecat-lifecycle.md) - Ephemeral worker workflow
- [Stabilize Command](../reference.md#gt-stabilize) - Post-update recovery

---

## Key Insight

**Gas Town is designed for branch-based workflows.**

- Main = clean synchronization point
- Feature branches = your work in progress  
- Integration branches = epic coordination
- Stabilize = fast-forward sync to origin

Working on main defeats the entire architecture. Use branches!

---

## Polecats: The Model Worker

**Polecats automatically follow the correct workflow.** When a polecat spawns:

1. Creates worktree from `origin/main` (never local main)
2. Gets unique branch: `polecat/<name>-<timestamp>`
3. Works and commits to their branch
4. Submits MR from their branch
5. Self-destructs when done

**Crew members should work the same way:**
```bash
# Spawn (crew member)
git checkout -b feature/my-work origin/main

# Work
vim file.go
git commit -m "feat: my changes"

# Submit
gt done  # or gt mq submit

# Cleanup (manual for crew)
git checkout main
git reset --hard origin/main
```

**Key difference:** Polecats are ephemeral (auto-cleanup), crew members are persistent (manual cleanup). Both should use feature branches.
