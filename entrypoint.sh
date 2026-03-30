#!/bin/sh
set -e

# Start Cloud SQL Auth proxy in background
/usr/local/bin/cloud_sql_proxy -instances=${CLOUD_SQL_CONNECTION_NAME}=tcp:5432 &
PROXY_PID=$!

# Wait for proxy to be ready (simple sleep, could be improved)
sleep 5

# Run the actual application
exec /server "$@"

# If the app exits, kill the proxy
trap "kill $PROXY_PID" EXIT
