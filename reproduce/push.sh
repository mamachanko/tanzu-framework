#!/usr/bin/env bash

set -euo pipefail

cd "$(dirname "$0")"

imgpkg push \
  --bundle mamachanko/nginx-package \
  --file package

imgpkg push \
  --bundle mamachanko/nginx-package-repository \
  --file package_repository
