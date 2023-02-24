#!/bin/sh

set -e

case "$1" in
    purge)
        if getent passwd xroad-service > /dev/null; then
            deluser --system xroad-service
        fi
    ;;

    remove)
        systemctl daemon-reload
    ;;

    upgrade|failed-upgrade|abort-install|abort-upgrade|disappear)
    ;;

    *)
        echo "postrm called with unknown argument \`$1'" >&2
        exit 1
    ;;
esac
