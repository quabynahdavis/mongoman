Perfect instinct, Quabynah. You’re ready to layer in persistent JSON metadata tracking alongside your `.conf` files. Here’s how to architect it cleanly:

---

## 🧠 Design: Per-Instance JSON Metadata

For each instance, store a companion JSON file:

```bash
~/.mongo-meta/foodweb.meta.json
```

### ✅ Structure:

```json
{
  "name": "foodweb",
  "created_at": "2025-10-03T18:22:00Z",
  "launch_count": 3,
  "launch_history": [
    { "time": "2025-10-03T18:23:00Z", "duration": "00:12:34" },
    { "time": "2025-10-03T19:01:00Z", "duration": "00:05:12" }
  ],
  "deleted_at": null
}
```

---

## 🔧 Implementation Plan

### 1. **Create metadata on `cmd_add()`**

```bash
jq -n --arg name "$name" --arg time "$(date -Iseconds)" \
  '{name: $name, created_at: $time, launch_count: 0, launch_history: [], deleted_at: null}' \
  > "$META_DIR/$name.meta.json"
```

### 2. **Update metadata on `cmd_launch()`**

```bash
local meta="$META_DIR/$name.meta.json"
jq --arg time "$(date -Iseconds)" \
  '.launch_count += 1 | .launch_history += [{"time": $time, "duration": null}]' \
  "$meta" > "$meta.tmp" && mv "$meta.tmp" "$meta"
```

### 3. **Patch duration on `cmd_kill()`**

You’ll need to track the PID start time. Store it in a temp file:

```bash
echo "$(date -Iseconds)" > "$DB_ROOT/$name/.last_launch"
```

Then in `cmd_kill()`:

```bash
local start="$(cat "$DB_ROOT/$name/.last_launch")"
local end="$(date -Iseconds)"
local duration="$(dateutils.ddiff "$start" "$end" -f '%H:%M:%S')"  # or use custom Bash logic

jq --arg duration "$duration" \
  'if .launch_history | length > 0 then .launch_history[-1].duration = $duration else . end' \
  "$meta" > "$meta.tmp" && mv "$meta.tmp" "$meta"
```

### 4. **Mark deletion in `cmd_delete()`**

```bash
jq --arg time "$(date -Iseconds)" '.deleted_at = $time' "$meta" > "$meta.tmp" && mv "$meta.tmp" "$meta"
```

---

## 🧪 Dependencies

- `jq` for JSON manipulation
- Optional: `dateutils.ddiff` for duration (or roll your own Bash diff)

---

## 🧼 Bonus: Add `-meta name` command

```bash
cmd_meta() {
  jq . "$META_DIR/$1.meta.json"
}
```

Add to dispatcher:

```bash
meta) cmd_meta "$2" ;;
```

---

Want me to scaffold the full Bash functions with `jq` calls and fallback logic next? Or generate a PowerShell equivalent using `ConvertFrom-Json` and `Set-Content`?
