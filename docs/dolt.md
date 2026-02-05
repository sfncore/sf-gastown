# Dolt Backend for Gas Town Beads

> Reference guide for using Dolt as the storage backend for Gas Town beads.

## Dolt Operating Modes

Gas Town supports **two different Dolt modes**:

### 1. Embedded Mode (Recommended)
- Database stored in: `.beads/dolt/beads/.dolt/`
- Direct file access, no server needed
- Single-writer (like SQLite but with Dolt features)
- Simpler, no server management required
- **Use this mode for standard Gas Town operations**

### 2. Server Mode
- Database stored in: `.dolt-data/<rig>/`
- MySQL-compatible server on port 3307
- Multi-client access
- Requires `gt dolt start` to be running
- **Use this mode only when you need concurrent SQL access**

**⚠️ Important**: Don't mix modes! Using both simultaneously causes database lock errors.

## What is Dolt?

Dolt is a SQL database with Git-like versioning. It provides:
- **Git semantics**: Branch, merge, diff, and log for your data
- **MySQL compatibility**: Standard SQL queries and tools
- **Multi-client access**: No single-writer limitation like SQLite (server mode only)
- **Versioning**: Track every change to your beads

## Why Use Dolt?

| Feature | SQLite | Dolt |
|---------|--------|------|
| Single-file storage | ✅ | ❌ |
| Multi-client access | ❌ | ✅ |
| `bd diff` between commits | ❌ | ✅ |
| `bd history` for issues | ❌ | ✅ |
| Branch/merge beads | ❌ | ✅ |
| SQL compatibility | Partial | Full MySQL |

## Quick Start

### Check Current Backend

```bash
cat .beads/metadata.json
```

**SQLite:**
```json
{"database": "beads.db", "jsonl_export": "issues.jsonl"}
```

**Dolt:**
```json
{"database": "dolt", "jsonl_export": "issues.jsonl", "backend": "dolt"}
```

### Migrate from SQLite to Dolt (Embedded Mode)

See [Migrating to Dolt](./migrating-to-dolt.md) for complete instructions.

**Quick version (Embedded Mode - Recommended):**
```bash
# Migrate town-level beads
cd ~/gt && bd migrate --to-dolt --yes

# Migrate rig beads
cd ~/gt/sfgastown/mayor/rig && bd migrate --to-dolt --yes

# Verify
bd status
```

**Note**: Do NOT run `gt dolt start` when using embedded mode. The server is only needed for server mode.

## Embedded Mode (Default)

Embedded mode stores Dolt databases directly in the `.beads/dolt/` directory. This is the recommended mode for Gas Town.

### Characteristics
- **Storage**: `.beads/dolt/beads/.dolt/`
- **Access**: Direct file access via `bd` commands
- **Server**: Not required
- **Concurrent access**: Single-writer (same as SQLite)

### When to Use
- Standard Gas Town operations
- Single-agent workflows
- When simplicity is preferred

## Server Mode (Optional)

Server mode runs a MySQL-compatible Dolt server for multi-client access.

### Start/Stop Server

```bash
# Start the server (only if you need multi-client SQL access)
gt dolt start

# Check status
gt dolt status

# View logs
gt dolt logs

# Stop server
gt dolt stop
```

### Server Configuration

- **Port**: 3307 (avoids MySQL conflict on 3306)
- **User**: root (no password for localhost)
- **Data directory**: `~/gt/.dolt-data/`
- **Databases**: One per rig (hq, sfgastown, etc.)

### Initialize Server Database

```bash
# Only needed for server mode
gt dolt init-rig myproject
```

Creates: `~/gt/.dolt-data/myproject/`

### When to Use
- Direct SQL queries from multiple clients
- External tools connecting via MySQL protocol
- Concurrent read access needed

## Beads Commands with Dolt

### Version Control Commands

```bash
# Compare beads between branches
bd diff main feature-branch

# Show changes in last 5 commits
bd diff HEAD~5 HEAD

# View full history for an issue
bd history hq-123
bd history hq-123 --limit 10

# Compare specific commits
bd diff abc123 def456
```

### Standard Commands (Work Normally)

All standard beads commands work the same:

```bash
bd ready              # Find available work
bd list               # List issues
bd show hq-123        # Show issue details
bd create --title="..." # Create issue
bd update hq-123 --status=in_progress
bd close hq-123
bd dep add hq-456 hq-123
```

## Direct SQL Access

### Query Beads with SQL

```bash
# Connect to Dolt SQL shell
gt dolt sql

# Or run specific queries
cd ~/gt/.beads/dolt/beads
dolt sql -q "SELECT COUNT(*) FROM issues WHERE status = 'open'"
```

### Common Queries

```sql
-- Count issues by status
SELECT status, COUNT(*) FROM issues GROUP BY status;

-- Find high priority open issues
SELECT id, title, priority 
FROM issues 
WHERE status = 'open' AND priority <= 1
ORDER BY priority;

-- Issues created this week
SELECT id, title, created_at
FROM issues
WHERE created_at > DATE_SUB(NOW(), INTERVAL 7 DAY);

-- Dependencies for a specific issue
SELECT depends_on_id, type 
FROM dependencies 
WHERE issue_id = 'hq-123';

-- Issues with comments
SELECT i.id, i.title, COUNT(c.id) as comment_count
FROM issues i
LEFT JOIN comments c ON i.id = c.issue_id
GROUP BY i.id
HAVING comment_count > 0;
```

## Database Structure

### Tables

| Table | Purpose |
|-------|---------|
| `issues` | Main issue data |
| `dependencies` | Issue relationships |
| `comments` | Issue comments |
| `events` | Activity log |
| `labels` | Issue labels |
| `config` | Beads configuration |
| `routes` | Cross-rig routing |

### Storage Locations

**Embedded Mode (Default):**
```
~/gt/
├── .beads/
│   ├── dolt/beads/.dolt/     # Dolt database files (embedded mode)
│   ├── issues.jsonl          # JSONL export (synced)
│   └── metadata.json         # Backend config
│
~/gt/sfgastown/mayor/rig/.beads/
├── dolt/beads/.dolt/         # Rig Dolt database (embedded mode)
└── metadata.json
```

**Server Mode (Optional):**
```
~/gt/.dolt-data/              # Centralized Dolt server data
├── hq/                       # Town-level database
│   └── .dolt/
├── sfgastown/                # Rig database
│   └── .dolt/
└── config.yaml               # Server config
```

## Advanced Features

### Branching Beads

```bash
# Create branch
cd ~/gt/.beads/dolt/beads
dolt checkout -b experiment

# Make changes via beads commands
bd create --title="Experimental feature"

# Commit to Dolt
dolt add .
dolt commit -m "Added experimental feature"

# Switch back
dolt checkout main
```

### Merging Changes

```bash
cd ~/gt/.beads/dolt/beads
dolt merge experiment
dolt commit -m "Merged experiment"
```

### Backup and Restore

```bash
# Create backup
cd ~/gt/.beads/dolt/beads
dolt backup create ~/backups/beads-backup

# Restore from backup
dolt backup restore ~/backups/beads-backup
```

## Troubleshooting

### "Dolt server is not running"

**For Embedded Mode**: This is expected! No server needed.

**For Server Mode**:
```bash
gt dolt start
```

### "Database is locked by another dolt process"

This happens when both embedded and server modes try to access the database simultaneously.

**Solution**:
```bash
# 1. Stop the server if running
gt dolt stop

# 2. Kill any stuck bd processes
lsof | grep "sfgastown.*dolt" | awk '{print $2}' | sort -u | xargs kill

# 3. Use embedded mode (recommended)
# bd commands will work directly without server
```

### "No rig databases found"

**For Embedded Mode**: Run migration:
```bash
cd ~/gt/sfgastown/mayor/rig
bd migrate --to-dolt --yes
```

**For Server Mode**:
```bash
gt dolt init-rig <rig-name>
gt dolt start
```

### Migration Issues

```bash
# Rollback to SQLite
bd migrate --to-sqlite --yes

# Check backup
ls .beads/beads.backup-pre-dolt-*.db
```

### Foreign Key Warnings

Cross-rig dependencies (e.g., `external:sfgastown:st-xxx`) show warnings during migration. This is expected - the issues exist in other rigs' databases.

### Connection Issues

```bash
# Check if server is listening
lsof -i :3307

# Restart server
gt dolt stop
gt dolt start
```

## Migration Back to SQLite

If needed, you can migrate back:

```bash
bd migrate --to-sqlite --yes
```

This converts the Dolt database back to SQLite format.

## Choosing a Mode

| Use Case | Recommended Mode | Why |
|----------|-----------------|-----|
| Standard Gas Town workflows | **Embedded** | Simpler, no server management |
| Single-agent operations | **Embedded** | Direct file access |
| Multi-client SQL queries | **Server** | MySQL protocol support |
| External tool integration | **Server** | Standard SQL connectivity |
| Concurrent read access | **Server** | Multiple clients can read |

**Default Recommendation**: Use **Embedded Mode** unless you specifically need multi-client SQL access.

## Performance Tips

### Embedded Mode
- No server overhead
- Same performance characteristics as SQLite
- Best for standard Gas Town operations

### Server Mode
- **Keep server running**: Dolt performs best when the server stays up
- **Use JSONL for sync**: The `issues.jsonl` export is still used for git sync
- **Regular commits**: Dolt tracks all changes - commit periodically for cleaner history
- **Backup before major changes**: Use `dolt backup` or rely on automatic pre-migration backups

## References

- [Dolt Documentation](https://docs.dolthub.com/)
- [Migrating to Dolt](./migrating-to-dolt.md) - Complete migration guide
- `gt dolt --help` - Dolt management commands
- `bd migrate --help` - Migration options
- `bd diff --help` - Compare beads between refs
- `bd history --help` - Issue history

---

**Summary**: Gas Town supports two Dolt modes:
- **Embedded Mode** (recommended): Direct file access, no server needed, simpler
- **Server Mode**: MySQL-compatible server for multi-client access

Use **Embedded Mode** for standard workflows. Use **Server Mode** only when you need concurrent SQL access from multiple clients.

**Note:** Dolt is optional. SQLite works fine for most use cases. Use Dolt when you need versioning, branching, or multi-client access.
