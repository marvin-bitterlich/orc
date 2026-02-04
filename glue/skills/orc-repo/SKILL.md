---
name: orc-repo
description: Guide through adding a new repo to global configuration. Use when user says /orc-repo or wants to add a repository for workbench creation.
---

# Repository Configuration Skill

Guide users through adding a repository to ORC's global configuration.

## Usage

```
/orc-repo <name>
/orc-repo              (will prompt for name)
```

## Flow

### Step 1: Gather Repo Name

If name not provided, ask:
- "What should this repo be called?" (slug format: lowercase, hyphens)

### Step 2: Gather Repo URL

Ask for the git remote URL:
- "What is the repository URL?" (e.g., git@github.com:org/repo.git)

This is optional - leave blank for local-only repos.

### Step 3: Gather Local Path

Ask where the repo is (or will be) cloned:
- "Where is the repo cloned locally?" (defaults to ~/src/<name>)

Verify the path exists:
```bash
ls -d <path>
```

If it doesn't exist, warn but continue (user may clone later).

### Step 4: Gather Default Branch

Ask for the default branch:
- "What is the default branch?" (defaults to main)

### Step 5: Create Repository Record

```bash
orc repo create <name> --url <url> --path <path> --default-branch <branch>
```

Capture the created REPO-xxx ID.

### Step 6: Verify Creation

```bash
orc repo show REPO-xxx
```

### Step 7: Confirm Ready

Output:
```
Repository configured:
  REPO-xxx: <name>
  URL: <url>
  Path: <path>
  Branch: <branch>

This repo can now be used when creating workbenches:
  orc workbench create my-workbench --repo-id REPO-xxx
```

## Example Session

```
User: /orc-repo intercom-web

Agent: What is the repository URL?
       (e.g., git@github.com:org/repo.git, or leave blank for local-only)

User: git@github.com:intercom/intercom-web.git

Agent: Where is the repo cloned locally?
       [Default: ~/src/intercom-web]

User: ~/src/intercom-web

Agent: What is the default branch?
       [Default: main]

User: master

Agent: [runs orc repo create intercom-web --url git@github.com:intercom/intercom-web.git --path ~/src/intercom-web --default-branch master]

Agent: Repository configured:
         REPO-xxx: intercom-web
         URL: git@github.com:intercom/intercom-web.git
         Path: ~/src/intercom-web
         Branch: master

       This repo can now be used when creating workbenches:
         orc workbench create my-workbench --repo-id REPO-xxx
```

## Error Handling

- If `orc repo create` fails with "already exists", show existing repo info
- If path doesn't exist, warn but allow creation (user may clone later)
- If URL is invalid format, ask again
