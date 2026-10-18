#!/usr/bin/env bash
set -euo pipefail

# Run from backend. Regenerate in temporary files without modifying tracked mocks.
temporary_directory=$(mktemp -d)
trap 'rm -rf "$temporary_directory"' EXIT

module_path=$(go list -m -f '{{.Path}}')
checked=0
while IFS= read -r source; do
  while IFS= read -r directive; do
    arguments=()
    read -r -a arguments <<< "${directive#//go:generate go tool mockgen }"
    destination=""
    for argument in "${arguments[@]}"; do
      case "$argument" in
        -destination=*) destination=${argument#-destination=} ;;
      esac
    done
    if [[ -z "$destination" ]]; then
      echo "Missing -destination in $source" >&2
      exit 1
    fi
    generated_package="$module_path/$(dirname "$source")"
    destination_directory=$(dirname "$destination")
    if [[ "$destination_directory" != "." ]]; then
      generated_package="$generated_package/$destination_directory"
    fi
    generated_file="$temporary_directory/$checked.go"
    (
      cd "$(dirname "$source")"
      go tool mockgen "${arguments[@]}" -self_package="$generated_package" -destination="$generated_file"
    )
    if ! diff -u "$(dirname "$source")/$destination" "$generated_file"; then
      echo 'Mocks are outdated. Run make mocks-generate.' >&2
      exit 1
    fi
    checked=$((checked + 1))
  done < <(sed -n '/^\/\/go:generate go tool mockgen /p' "$source")
done < <(find internal/app -name generate.go -type f | sort)

if [[ "$checked" -eq 0 ]]; then
  echo 'No mock generation directives found.' >&2
  exit 1
fi
printf 'Checked %s generated mock files.\n' "$checked"
