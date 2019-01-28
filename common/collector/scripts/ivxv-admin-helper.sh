#!/bin/sh

# IVXV Internet voting framework

# Helper script for ivxv-admin account

# Usage: ivxv-admin-helper <action> [<arg> ...]

set -e


usage() {
  echo "Usage:"
  echo "    ivxv-admin-helper check-service-config <service-type> <service-id>"
  echo "        Check service configuration"
  echo
  echo "    ivxv-admin-helper restart-service <service-type> <service-id>"
  echo "                                      <systemctl-service-id>"
  echo "        Restart service"
}

# ACTIONS {{{

# check_config {{{
# Check service configuration
check_config() {
  SERVICE_TYPE="$1"
  SERVICE_ID="$2"

  # check service config
  test -e /etc/default/ivxv && . /etc/default/ivxv
  "/usr/bin/ivxv-${SERVICE_TYPE}" ${EXTRAOPTS} -instance "${SERVICE_ID}" -check input
}
# }}}
# restart_service {{{
# Restart service
restart_service() {
  SERVICE_TYPE="$1"
  SERVICE_ID="$2"
  SYSTEMCTL_SERVICE_ID="$3"

  ivxv-admin-helper check-service-config "${SERVICE_TYPE}" "${SERVICE_ID}"
  systemctl restart --user "${SYSTEMCTL_SERVICE_ID}"
}
# }}}

# }}}

# parse CLI arguments {{{
ACTION="$1"
case "${ACTION}" in
  check-service-config)
    check_config "$2" "$3"
  ;;
  restart-service)
    restart_service "$2" "$3" "$4"
  ;;
  --help)
    usage
    exit 0
  ;;
  *)
    echo "Unknown action: ${ACTION}"
    echo
    usage
    exit 1
  ;;
esac
# }}}

exit 0

# vim:foldmethod=marker:
