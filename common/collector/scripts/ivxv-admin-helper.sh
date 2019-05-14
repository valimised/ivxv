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
  echo "    ivxv-admin-helper copy-logs-to-logmon <hostname> <logmonitor-address>"
  echo "        Copy IVXV service log files to Log Monitor"
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
# copy_logs_to_logmon {{{
# Copy IVXV service log files to Log Monitor
copy_logs_to_logmon() {
  HOST_NAME="$1"
  LOGMON_ADDR="$2"
  LOGFILE_PATTERN='ivxv-????-??-??.log'

  cd /var/log

  # check if log file exists
  if [ "$(echo ${LOGFILE_PATTERN})" = "${LOGFILE_PATTERN}" ]; then
    echo "Log files not found in pattern ${LOGFILE_PATTERN}"
    exit 1
  fi

  # copy log files to Log Monitor
  for SRC_FILE in ${LOGFILE_PATTERN}
  do
    TGT_FILE="$(
      echo ${SRC_FILE} |
        sed --regexp-extended --expression="s/^ivxv-(.+).log/${HOST_NAME}-\1-ivxv.log/"
    )"

    rsync ${SRC_FILE} ${LOGMON_ADDR}:/var/log/ivxv/${TGT_FILE}
  done
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
  copy-logs-to-logmon)
    copy_logs_to_logmon "$2" "$3"
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
