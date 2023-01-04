#!/usr/bin/env bash


# So that when a command fails, bash exits.
set -o errexit

# This will make the script fail, when accessing an unset variable.
set -o nounset

# the return value of a pipeline is the value of the last (rightmost) command
# to exit with a non-zero status, or zero if all commands in the pipeline exit
# successfully.
set -o pipefail

# This helps in debugging your scripts. TRACE=1 ./script.sh
if [[ "${TRACE-0}" == "1" ]]; then set -o xtrace; fi



if [[ "${1-}" =~ ^-*h(elp)?$ ]]; then
    echo 'Usage: ./check-codeowners.sh
Bash script to validate if each component has an owner.
'
    exit
fi

CODEOWNERS=".github/CODEOWNERS"

# Get component folders from the project and checks that they have
# an owner in $CODEOWNERS
# Code from: https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/.github/workflows/scripts/check-codeowners.sh
check_code_owner_existence() {
  MODULES=$(find . -type f -name "go.mod" -exec dirname {} \; | sort | grep -E '^./' | cut -c 3-)
  MISSING_COMPONENTS=0
  for module in ${MODULES}
  do
    # For a component path exact match, need to add '/ ' to end of module as
    # each line in the CODEOWNERS file is of the format:
    # <component_path_relative_from_project_root>/<min_1_space><owner_1><space><owner_2><space>..<owner_n>
    # This is because the path separator at end is dropped while searching for
    # modules and there is at least 1 space separating the path from the owners.
    if ! grep -q "^$module/ " "$CODEOWNERS"; then
      # If there is not an exact match to component path, there might be a parent folder
      # which has an owner and would therefore implicitly include the component
      # path as a sub folder e.g. 'internal/aws' is listed in $CODEOWNERS
      # which accounts for internal/aws/awsutil, internal/aws/k8s etc.
      PREFIX_MODULE_PATH=$(echo $module | cut -d/ -f 1-2)
      if ! grep -wq "^$PREFIX_MODULE_PATH/ " "$CODEOWNERS"; then
          ((MISSING_COMPONENTS=MISSING_COMPONENTS+1))
          echo "FAIL: \"$module\" not included in CODEOWNERS"
      fi
    fi
  done
  if [ "$MISSING_COMPONENTS" -gt 0 ]; then
    echo "---"
    echo "FAIL: there are $MISSING_COMPONENTS components not included in CODEOWNERS and not known in the ALLOWLIST"
    exit 1
  else
    echo "---"
    echo "SUCCED: all components has a codeowner"
    exit 0
  fi
}

main() {
    check_code_owner_existence
}

main "$@"

