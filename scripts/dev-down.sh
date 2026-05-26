#!/usr/bin/env zsh
set -euo pipefail

docker compose -f deployments/local/docker-compose.yml down
