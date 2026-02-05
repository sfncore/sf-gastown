# Migrating Beads from SQLite to Dolt

> Complete guide for migrating Gas Town beads databases from SQLite to Dolt storage backend.

## Overview

Gas Town supports two storage backends for beads:
- **SQLite** (default): Single-file database, simple but limited to single-writer
- **Dolt**: Git-like database with versioning, branching, and multi-client support

Dolt enables powerful features like:
- `bd diff` - Compare beads between commits/branches
- `bd history` - View complete issue history
- Multi-client access (no single-writer limitation)
- Git-like versioning and branching

## Prerequisites

- Dolt installed (`which dolt`)
- `gt` CLI installed and working
- `bd` CLI installed and working
- Current beads using SQLite (check: `cat .beads/metadata.json`)

## Migration Steps

### 1. Initialize Dolt Rig Databases

Each rig needs its own Dolt database:

```bash
# Initialize town-level (hq) database
gt dolt init-rig hq

# Initialize rig databases (repeat for each rig)
gt dolt init-rig sfgastown
gt dolt init-rig myproject
```

### 2. Start the Dolt Server

```bash
gt dolt start
```

This starts a MySQL-compatible server on port 3307.

### 3. Inspect Migration (Optional but Recommended)

Check what will be migrated:

```bash
# For town-level beads
cd ~/gt
bd migrate --inspect --to-dolt

# For rig beads
cd ~/gt/sfgastown/mayor/rig
bd migrate --inspect --to-dolt
```

### 4. Run the Migration

```bash
# For town-level beads
cd ~/gt
bd migrate --to-dolt --yes

# For each rig
cd ~/gt/sfgastown/mayor/rig
bd migrate --to-dolt --yes
```

The `--yes` flag auto-confirms cleanup prompts.

### 5. Verify Migration

```bash
# Check status
bd status

# Test Dolt-specific features
bd diff main HEAD
bd history <issue-id> --limit 5

# List issues
bd list --limit 10
```

### 6. Clean Up (After Verification)

Once you've verified everything works:

```bash
# Remove old SQLite database
rm ~/gt/.beads/beads.db
rm ~/gt/sfgastown/mayor/rig/.beads/beads.db
```

## What Gets Migrated

- All issues and their metadata
- Dependencies between issues
- Events and history
- Configuration keys
- Labels and comments

## What Changes

### Before (SQLite)
```json
// .beads/metadata.json
{
  "database": "beads.db",
  "jsonl_export": "issues.jsonl"
}
```

### After (Dolt)
```json
// .beads/metadata.json
{
  "database": "dolt",
  "jsonl_export": "issues.jsonl",
  "backend": "dolt"
}
```

### Directory Structure Changes

```
.beads/
├── beads.db                    # OLD - can delete after verification
├── beads.db-shm
├── beads.db-wal
├── dolt/                       # NEW - Dolt database
│   └── beads/
│       └── .dolt/
├── metadata.json               # UPDATED
└── issues.jsonl                # Still exported
```

## Troubleshooting

### "Dolt server is not running"
```bash
gt dolt start
```

### "No rig databases found"
```bash
# Re-initialize
gt dolt init-rig <rig-name>
```

### Migration fails or data looks wrong
```bash
# Check backup exists
ls .beads/beads.backup-pre-dolt-*.db

# Rollback to SQLite
bd migrate --to-sqlite --yes
```

### Foreign key warnings after migration
These are expected for cross-rig dependencies (e.g., `external:sfgastown:st-xxx`). The issues exist in other rigs' databases, so they're not found during validation but work correctly at runtime.

## Rollback

If you need to switch back to SQLite:

```bash
bd migrate --to-sqlite --yes
```

## Post-Migration

### Dolt Server Management

```bash
# Check status
gt dolt status

# View logs
gt dolt logs

# Stop server
gt dolt stop

# Restart
gt dolt stop && gt dolt start
```

### New Capabilities

Now you can use:

```bash
# Compare beads between branches
bd diff main feature-branch

# Show changes in last 5 commits
bd diff HEAD~5 HEAD

# View full issue history
bd history <issue-id>

# SQL access
gt dolt sql -q "SELECT * FROM issues WHERE status = 'open'"
```

## One-Liner Migration

For automated/scripted migration:

```bash
#!/bin/bash
set -e

# Initialize and start
gt dolt init-rig hq
gt dolt init-rig sfgastown  # add more rigs as needed
gt dolt start

# Migrate town
cd ~/gt && bd migrate --to-dolt --yes

# Migrate rigs
cd ~/gt/sfgastown/mayor/rig && bd migrate --to-dolt --yes
# cd ~/gt/otherproject/mayor/rig && bd migrate --to-dolt --yes

echo "Migration complete! Verify with: bd status"
```

## Verification Checklist

After migration, verify:

- [ ] `bd status` shows correct issue counts
- [ ] `bd list` displays issues
- [ ] `bd diff main HEAD` works
- [ ] `bd history <issue-id>` shows history
- [ ] `gt dolt status` shows server running
- [ ] `.beads/metadata.json` has `"backend": "dolt"`

## References

- `gt dolt --help` - Dolt management commands
- `bd migrate --help` - Migration options
- `bd diff --help` - Compare beads between refs
- `bd history --help` - Issue history
- [Dolt Backend Guide](./dolt.md) - Full Dolt reference

---

**Note:** Backups are automatically created before migration. Keep them until you've verified the migration succeeded.
