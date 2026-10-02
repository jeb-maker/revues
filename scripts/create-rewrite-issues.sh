#!/usr/bin/env bash
# Crée les issues GitHub du rewrite depuis docs/rewrite/WORK_PACKAGES.md.
# Prérequis : gh authentifié avec droit d'écriture sur le repo.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
WP_FILE="$ROOT/docs/rewrite/WORK_PACKAGES.md"
MAP_FILE="$ROOT/docs/rewrite/ISSUE_MAP.md"

if ! command -v gh >/dev/null 2>&1; then
  echo "gh CLI requis" >&2
  exit 1
fi

if [[ ! -f "$WP_FILE" ]]; then
  echo "missing $WP_FILE" >&2
  exit 1
fi

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

# Split WORK_PACKAGES.md into one file per WP-xxx section.
awk '
  /^## WP-/ {
    if (out) close(out)
    id = $2
    gsub(/[^A-Za-z0-9-]/, "", id)
    out = "'"$tmp"'/" id ".md"
    print > out
    next
  }
  out { print >> out }
' "$WP_FILE"

echo "# Mapping WP → issues GitHub" >"$MAP_FILE"
echo "" >>"$MAP_FILE"
echo "Généré par \`scripts/create-rewrite-issues.sh\`." >>"$MAP_FILE"
echo "" >>"$MAP_FILE"
echo "| WP | Issue | Titre |" >>"$MAP_FILE"
echo "|----|-------|-------|" >>"$MAP_FILE"

for f in "$tmp"/WP-*.md; do
  [[ -f "$f" ]] || continue
  id="$(basename "$f" .md)"
  title="$(head -n1 "$f" | sed 's/^## //')"
  labels_line="$(grep -m1 '^Labels:' "$f" || true)"
  labels_csv="${labels_line#Labels: }"
  labels_csv="$(echo "$labels_csv" | tr -d ' ')"

  body_file="$tmp/${id}.body.md"
  {
    echo "## Contexte"
    echo ""
    echo "Epic rewrite API-first + SvelteKit — voir [docs/rewrite/README.md](../docs/rewrite/README.md), [ADR-001](../docs/ADR-001-api-first-svelte.md)."
    echo ""
    # Drop the H2 title line from body; keep the rest.
    tail -n +2 "$f"
  } >"$body_file"

  gh_args=(issue create --title "$title" --body-file "$body_file")
  if [[ -n "$labels_csv" ]]; then
    IFS=',' read -r -a label_arr <<<"$labels_csv"
    for lab in "${label_arr[@]}"; do
      [[ -n "$lab" ]] && gh_args+=(--label "$lab")
    done
  fi

  echo "Creating: $title"
  set +e
  url="$(gh "${gh_args[@]}" 2>"$tmp/${id}.err")"
  status=$?
  set -e
  if [[ $status -ne 0 ]]; then
    # Retry without labels if some labels are missing on the repo.
    echo "  label create failed, retry without labels…"
    url="$(gh issue create --title "$title" --body-file "$body_file")"
  fi
  num="$(echo "$url" | grep -oE '[0-9]+$')"
  echo "| $id | #$num | $title |" >>"$MAP_FILE"
  echo "  -> #$num $url"
done

echo ""
echo "Mapping écrit dans $MAP_FILE"
echo "Mettre à jour les dépendances « Bloqué par: WP-xxx » dans les issues avec les #N du mapping."
