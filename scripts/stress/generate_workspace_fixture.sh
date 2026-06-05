#!/usr/bin/env bash
set -euo pipefail

TARGET_DIR="${1:-${WORKSPACE_PATH:-./.tmp/stress-workspace}}"
DIR_COUNT="${DIR_COUNT:-12}"
FILES_PER_DIR="${FILES_PER_DIR:-15}"
LINES_PER_FILE="${LINES_PER_FILE:-40}"
LARGE_FILES="${LARGE_FILES:-4}"

mkdir -p "$TARGET_DIR"
TARGET_DIR="$(cd "$TARGET_DIR" && pwd)"

rm -rf "$TARGET_DIR"/* "$TARGET_DIR"/.git "$TARGET_DIR"/node_modules "$TARGET_DIR"/.runtime "$TARGET_DIR"/.tmp
mkdir -p "$TARGET_DIR"/.git/objects "$TARGET_DIR"/node_modules/demo "$TARGET_DIR"/.runtime/cache "$TARGET_DIR"/.tmp/session

for dir_index in $(seq 1 "$DIR_COUNT"); do
  dir_path="$TARGET_DIR/module-$dir_index"
  mkdir -p "$dir_path/nested"

  for file_index in $(seq 1 "$FILES_PER_DIR"); do
    file_path="$dir_path/file-$file_index.txt"
    {
      echo "module-$dir_index file-$file_index"
      echo "search-keyword baseline entry"
      for line_index in $(seq 1 "$LINES_PER_FILE"); do
        printf 'line %03d in module-%d file-%d contains search-keyword and fixture text\n' "$line_index" "$dir_index" "$file_index"
      done
    } > "$file_path"
  done

  cat > "$dir_path/nested/overview.md" <<EOF
# module-$dir_index
This fixture directory is generated for stress testing.
search-keyword appears here for recursive text search.
EOF
done

for large_index in $(seq 1 "$LARGE_FILES"); do
  large_path="$TARGET_DIR/large-$large_index.txt"
  {
    for line_index in $(seq 1 4000); do
      printf 'large file %d line %04d search-keyword repeated content for search pressure\n' "$large_index" "$line_index"
    done
  } > "$large_path"
done

cat > "$TARGET_DIR/.git/ignored.txt" <<EOF
search-keyword should not be counted from ignored directories.
EOF

cat > "$TARGET_DIR/node_modules/demo/ignored.txt" <<EOF
search-keyword should not be counted from node_modules.
EOF

cat > "$TARGET_DIR/README.md" <<EOF
# Stress Fixture Workspace
Generated at: $(date -u +"%Y-%m-%dT%H:%M:%SZ")
EOF

echo "[stress] fixture generated: $TARGET_DIR"
echo "[stress] dirs=$DIR_COUNT files_per_dir=$FILES_PER_DIR lines_per_file=$LINES_PER_FILE large_files=$LARGE_FILES"
