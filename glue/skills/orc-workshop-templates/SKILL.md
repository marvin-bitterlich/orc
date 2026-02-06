---
name: orc-workshop-templates
description: Manage workshop templates. Use when user says /orc-workshop-templates or wants to view, add, edit, or remove workshop templates.
---

# Workshop Templates Management

Interactive skill to view and manage workshop templates stored in `~/.orc/workshop-templates.json`.

## Usage

```
/orc-workshop-templates                    # Interactive menu
/orc-workshop-templates list               # Show all templates
/orc-workshop-templates add                # Add new template
/orc-workshop-templates edit <name>        # Edit existing template
/orc-workshop-templates remove <name>      # Remove template
```

## Config File

Templates are stored in `~/.orc/workshop-templates.json`:

```json
{
  "templates": {
    "standard": ["intercom", "infrastructure"],
    "muster": ["intercom", "muster", "muster-deployer"],
    "ami": ["intercom", "ami", "infrastructure"],
    "provisioner": ["intercom", "intercom-provisioner", "infrastructure"],
    "events": ["intercom", "infrastructure", "event-management-system"]
  }
}
```

## Flow

### Step 1: Load Templates

```bash
cat ~/.orc/workshop-templates.json
```

If file doesn't exist, create with defaults:
```json
{
  "templates": {
    "standard": ["intercom", "infrastructure"],
    "muster": ["intercom", "muster", "muster-deployer"],
    "ami": ["intercom", "ami", "infrastructure"],
    "provisioner": ["intercom", "intercom-provisioner", "infrastructure"],
    "events": ["intercom", "infrastructure", "event-management-system"]
  }
}
```

### Step 2: Handle Subcommand

**If `list` or no subcommand (interactive mode):**

Display current templates:

```
Workshop Templates (~/.orc/workshop-templates.json)

| Template    | Repos                                    |
|-------------|------------------------------------------|
| standard    | intercom, infrastructure                 |
| muster      | intercom, muster, muster-deployer        |
| ami         | intercom, ami, infrastructure            |
| provisioner | intercom, intercom-provisioner, infrastructure |
| events      | intercom, infrastructure, event-management-system |
```

If interactive (no subcommand), ask:

> What would you like to do?
> 1. Add a new template
> 2. Edit an existing template
> 3. Remove a template
> 4. Done

**If `add`:**

Go to Add Template flow.

**If `edit <name>`:**

Go to Edit Template flow with given name.

**If `remove <name>`:**

Go to Remove Template flow with given name.

### Add Template Flow

1. Ask for template name:
   > "Template name? (lowercase, no spaces)"

2. Validate name doesn't already exist

3. Ask for repos:
   > "Which repos should this template include?"
   > "Enter repo names separated by commas (e.g., intercom, infrastructure)"

4. Show available repos for reference:
   ```bash
   orc repo list
   ```

5. Confirm:
   > "Create template 'mytemplate' with repos: intercom, myrepo?"
   > [y/n]

6. Save to config file

### Edit Template Flow

1. If name not provided, ask:
   > "Which template to edit?"
   > [show list]

2. Show current repos:
   > "Template 'muster' currently includes: intercom, muster, muster-deployer"

3. Ask for new repos:
   > "Enter new repo list (or press enter to keep current)"

4. Confirm and save

### Remove Template Flow

1. If name not provided, ask:
   > "Which template to remove?"
   > [show list]

2. Confirm:
   > "Remove template 'mytemplate'? This cannot be undone."
   > [y/n]

3. Remove from config and save

### Step 3: Save Config

Write updated config back to `~/.orc/workshop-templates.json`:

```bash
cat > ~/.orc/workshop-templates.json << 'EOF'
{
  "templates": {
    ...updated templates...
  }
}
EOF
```

Confirm: "Templates updated."

## Example Session (list)

```
User: /orc-workshop-templates list

Agent: Workshop Templates (~/.orc/workshop-templates.json)

       | Template    | Repos                                          |
       |-------------|------------------------------------------------|
       | standard    | intercom, infrastructure                       |
       | muster      | intercom, muster, muster-deployer              |
       | ami         | intercom, ami, infrastructure                  |
       | provisioner | intercom, intercom-provisioner, infrastructure |
       | events      | intercom, infrastructure, event-management-system |
```

## Example Session (add)

```
User: /orc-workshop-templates add

Agent: Template name? (lowercase, no spaces)

User: dlq

Agent: Which repos should this template include?
       Enter repo names separated by commas.

       Available repos:
       - intercom
       - infrastructure
       - muster
       - ami
       ...

User: intercom, infrastructure, dlq-admin

Agent: Create template 'dlq' with repos: intercom, infrastructure, dlq-admin?

User: y

Agent: ✓ Template 'dlq' created

       Workshop Templates:
       ...updated list...
```

## Example Session (interactive)

```
User: /orc-workshop-templates

Agent: Workshop Templates (~/.orc/workshop-templates.json)
       [shows table]

       What would you like to do?
       1. Add a new template
       2. Edit an existing template
       3. Remove a template
       4. Done

User: 2

Agent: Which template to edit?
       1. standard
       2. muster
       3. ami
       4. provisioner
       5. events

User: 2

Agent: Template 'muster' currently includes: intercom, muster, muster-deployer

       Enter new repo list (or press enter to keep current):

User: intercom, muster, muster-deployer, muster-config

Agent: Update template 'muster' with repos: intercom, muster, muster-deployer, muster-config?

User: y

Agent: ✓ Template 'muster' updated
```

## Error Handling

| Error | Remediation |
|-------|-------------|
| Config file missing | Create with defaults |
| Template name exists (add) | "Template 'x' already exists. Use edit instead." |
| Template not found (edit/remove) | "Template 'x' not found." Show available templates. |
| Invalid JSON | "Config file corrupted. Reset to defaults? [y/n]" |
| Empty repo list | "Template must include at least one repo." |

## Notes

- Template names should be lowercase with no spaces
- Repo names must match registered repos (use `orc repo list` to verify)
- Changes take effect immediately for `/orc-workshop`
