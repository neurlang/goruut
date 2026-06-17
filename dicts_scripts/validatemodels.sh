find . -type f -name 'weights6.*.json.zlib' -print0 |
while IFS= read -r -d '' f; do
  if zlib-flate -uncompress < "$f" 2>/dev/null | jq -e . >/dev/null 2>&1; then
    echo "OK (valid JSON): $f"
  else
    echo "BAD JSON: $f"
  fi
done
