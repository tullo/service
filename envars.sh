#!/usr/bin/env bash
# Using /usr/bin/env is a workaround to run the first bash found on the PATH.

# Usage:
#   sudo ./go-install.sh 1.24.2
# Releases:
#   https://go.dev/dl/
echo "Generating '.envrc'"

# Define the raw components of the connection string
BASE_URL="postgresql://root@localhost:26257/garagesales"
SSL_PARAMS="sslmode=verify-full&sslcert=${PWD}/certs/client.root.crt&sslkey=${PWD}/certs/client.root.key&sslrootcert=${PWD}/certs/ca.crt"

# URL-encode only the parameter values
ENCODED_SSL_PARAMS=$(python3 -c "import urllib.parse; print('&'.join([f'{key}={urllib.parse.quote(value)}' for key, value in [param.split('=') for param in '$SSL_PARAMS'.split('&')]]))")

# Construct the final URL
FINAL_URL="${BASE_URL}?${ENCODED_SSL_PARAMS}"

# Write to the .envrc file
printf 'export COCKROACH_URL="postgresql://root@localhost:26257"\n' > .envrc
printf 'export DATABASE_URL="%s"\n' "$FINAL_URL" >> .envrc
