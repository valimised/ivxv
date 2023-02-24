#!/bin/sh

set -e

case "$1" in
    configure)
        if ! getent passwd xroad-service > /dev/null; then
            adduser --system --no-create-home xroad-service
        fi
        systemctl daemon-reload
    ;;

    *)
        echo "postinst called with unknown argument \`$1'" >&2
        exit 1
    ;;
esac
