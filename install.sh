#!/bin/bash
set -euo pipefail

cd "$(dirname "$0")"

stow -v -R -t ~ "boxed"
