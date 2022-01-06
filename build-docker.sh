#!/bin/sh

# This is more a reminder to myself on building the image than anything else!

DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"
docker build -t spiral90210/fileserve:latest .
