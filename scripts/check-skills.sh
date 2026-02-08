#!/bin/bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
SKILLS_DIR="$PROJECT_ROOT/glue/skills"
ARCH_FILE="$PROJECT_ROOT/ARCHITECTURE.md"

errors=()

# Check 1: Every skill has valid frontmatter (name + description)
echo "Checking skill frontmatter..."
for skill_dir in "$SKILLS_DIR"/*/; do
    [[ -d "$skill_dir" ]] || continue
    skill_name=$(basename "$skill_dir")
    skill_file="$skill_dir/SKILL.md"

    if [[ ! -f "$skill_file" ]]; then
        errors+=("MISSING: $skill_name/SKILL.md does not exist")
        continue
    fi

    # Extract frontmatter (between first two ---)
    frontmatter=$(awk '/^---$/{if(++n==2)exit}n==1' "$skill_file")

    if [[ -z "$frontmatter" ]]; then
        errors+=("FRONTMATTER: $skill_name has no YAML frontmatter")
        continue
    fi

    if ! echo "$frontmatter" | grep -q "^name:"; then
        errors+=("FRONTMATTER: $skill_name missing 'name:' field")
    fi

    if ! echo "$frontmatter" | grep -q "^description:"; then
        errors+=("FRONTMATTER: $skill_name missing 'description:' field")
    fi
done

# Get list of skills in glue/skills/
skills_on_disk=()
for skill_dir in "$SKILLS_DIR"/*/; do
    [[ -d "$skill_dir" ]] || continue
    skills_on_disk+=("$(basename "$skill_dir")")
done

# Get list of skills in ARCHITECTURE.md (from markdown tables)
# Format: | skill-name | Description |
skills_in_arch=()
while IFS= read -r line; do
    # Extract skill name from table row (first column after |)
    skill=$(echo "$line" | sed -n 's/^| *\([a-z][a-z0-9-]*\) *|.*/\1/p')
    if [[ -n "$skill" && "$skill" != "Skill" ]]; then
        skills_in_arch+=("$skill")
    fi
done < <(grep -E '^\| *[a-z][a-z0-9-]* *\|' "$ARCH_FILE" 2>/dev/null || true)

# Check 2: Every skill on disk is in ARCHITECTURE.md
echo "Checking skills are documented in ARCHITECTURE.md..."
for skill in "${skills_on_disk[@]}"; do
    found=false
    for arch_skill in "${skills_in_arch[@]}"; do
        if [[ "$skill" == "$arch_skill" ]]; then
            found=true
            break
        fi
    done
    if [[ "$found" == "false" ]]; then
        errors+=("UNDOCUMENTED: $skill exists in glue/skills/ but not in ARCHITECTURE.md")
    fi
done

# Check 3: Every skill in ARCHITECTURE.md exists on disk
echo "Checking documented skills exist..."
for arch_skill in "${skills_in_arch[@]}"; do
    found=false
    for skill in "${skills_on_disk[@]}"; do
        if [[ "$skill" == "$arch_skill" ]]; then
            found=true
            break
        fi
    done
    if [[ "$found" == "false" ]]; then
        errors+=("MISSING: $arch_skill listed in ARCHITECTURE.md but not in glue/skills/")
    fi
done

# Report results
if [[ ${#errors[@]} -eq 0 ]]; then
    echo "✓ All skills have valid frontmatter"
    echo "✓ All skills documented in ARCHITECTURE.md"
    exit 0
else
    echo ""
    for err in "${errors[@]}"; do
        echo "  $err"
    done
    echo ""
    echo "✗ ${#errors[@]} skill issues found"
    exit 1
fi
