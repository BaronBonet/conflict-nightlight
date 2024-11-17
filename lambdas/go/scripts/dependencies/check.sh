#!/usr/bin/env bash
set -euo pipefail

DIR_PATH=$(cd "$(dirname "${BASH_SOURCE:-$0}")" && pwd)
. "${DIR_PATH}/../../build_dependencies_versions"

LOCAL_BIN="${DIR_PATH}/../../.local/bin"

exit_code=0
if [[ ! -f "$LOCAL_BIN/mockery" ]]; then
  echo "mockery is not installed."
  exit_code=1
fi

if [[ $exit_code != 0 ]]; then
  exit ${exit_code}
fi

ACTUAL_MOCKERY_VERSION="$("${LOCAL_BIN}/mockery" --version)"

if [[ "${ACTUAL_MOCKERY_VERSION}" != "v${MOCKERY_VERSION}" ]]; then
  echo "mockery version ($ACTUAL_MOCKERY_VERSION) mismatch - expected ${MOCKERY_VERSION}"
  exit_code=1
fi

exit ${exit_code}
