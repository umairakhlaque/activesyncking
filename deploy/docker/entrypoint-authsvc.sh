#!/bin/sh
# When called with no arguments, start the auth service (normal operation).
# When Fly.io calls release_command="/migrate", exec the migration binary.
set -e
if [ $# -eq 0 ]; then
    exec /authsvc
else
    exec "$@"
fi
