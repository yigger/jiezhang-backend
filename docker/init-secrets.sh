#!/bin/sh
set -eu
umask 077

# Persist random local credentials; restarts must keep the MySQL passwords.
for name in mysql-root-password mysql-password session-token-secret; do
    path="/run/dev-secrets/$name"
    if [ ! -s "$path" ]; then
        od -An -N32 -tx1 /dev/urandom | tr -d ' \n' > "$path.tmp"
        printf '\n' >> "$path.tmp"
        mv "$path.tmp" "$path"
    fi
done
