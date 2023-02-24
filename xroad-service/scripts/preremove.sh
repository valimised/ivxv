#!/bin/sh

set -e

case "$1" in
    remove)
        systemctl unmask xroad-service.service
        systemctl stop xroad-service.service
        systemctl disable xroad-service.service
    ;;

    upgrade|failed-upgrade|deconfigure)
    ;;

    *)
        echo "prerm called with unknown argument \`$1'" >&2
        exit 1
    ;;
esac
