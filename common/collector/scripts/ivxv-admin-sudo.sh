#!/bin/sh

# IVXV Internet voting framework

# Helper script for ivxv-admin account
# for actions that require privilege escalation

# Usage: ivxv-admin-sudo <action> [<arg> ...]

set -e

usage() {
  echo "Usage:"
  echo "    ivxv-admin-sudo backup-ballot-box <voting-host> <service-id> "
  echo "                                      <backup-filename>"
  echo "        Backup ballot box (in backup service)"
  echo
  echo "    ivxv-admin-sudo backup-log <log-host> <backup-timestamp>"
  echo "        Backup log file (in backup service)"
  echo
  echo "    ivxv-admin-sudo create-ssh-access <account-name>"
  echo "        Create management service access to account in service host"
  echo
  echo "    ivxv-admin-sudo init-host"
  echo "        Initialize service host"
  echo
  echo "    ivxv-admin-sudo init-service <service-id>"
  echo "        Initialize service data directory. "
  echo "        Value 'backup' is used for all backup services"
  echo
  echo "    ivxv-admin-sudo install-pkg <package-filename>"
  echo "        Install IVXV package with dependencies"
  echo
  echo "    ivxv-admin-sudo prepare-ballot-box-backup <service-id> <backup-filename>"
  echo "        Prepare votes backup file in voting service"
  echo
  echo "    ivxv-admin-sudo remove-admin-root-access"
  echo "        Remove management service access to service host root account"
  echo
  echo "    ivxv-admin-sudo rsyslog-config-apply"
  echo "        Apply rsyslog config file for IVXV logging"
}

# ACTIONS {{{

# backup_ballot_box {{{
# Backup current ballot box
backup_ballot_box() {
  VOTING_HOST="$1"
  SERVICE_ID="$2"
  BACKUP_FILENAME="$3"
  BACKUP_DIR="/var/backups/ivxv/ballot-box"
  echo "# Preparing ballot box backup file in voting service ${SERVICE_ID}"
  ssh "ivxv-voting@${VOTING_HOST}" ivxv-admin-sudo prepare-ballot-box-backup "${SERVICE_ID}" "${BACKUP_FILENAME}"
  echo "# Copying backup file ${BACKUP_FILENAME} to backup service"
  scp "ivxv-voting@${VOTING_HOST}":"${BACKUP_FILENAME}" "${BACKUP_DIR}/tmp-${BACKUP_FILENAME}"
  mv "${BACKUP_DIR}/tmp-${BACKUP_FILENAME}" "${BACKUP_DIR}/${BACKUP_FILENAME}"
  echo "# Removing backup file ${BACKUP_FILENAME} from voting service"
  ssh "${VOTING_HOST}" rm -f "${BACKUP_FILENAME}"
}
# }}}

# backup_log {{{
# Backup log
backup_log() {
  LOG_HOST="$1"
  BACKUP_TIMESTAMP="$2"
  BACKUP_TARGET_FILEPATH="/var/backups/ivxv/log/${LOG_HOST}-${BACKUP_TIMESTAMP}.tar.gz"
  echo "# Copying log files from log collector to backup service"
  RET=0
  (
    ssh ${LOG_HOST} \
      tar --create --gzip --verbose '/var/log/ivxv*.log' \
      > "${BACKUP_TARGET_FILEPATH}"
  ) || RET=$?
  # ignore exit code 1 (file in grown during archiving process)
  if [ "${RET}" = 1 ]; then
    RET=0
  fi
  return "${RET}"
}
# }}}

# create_ssh_access {{{
create_ssh_access() {
  ACCOUNT_NAME="$1"
  echo "# Validating account '${ACCOUNT_NAME}'"
  echo "${ACCOUNT_NAME}" | grep --quiet --extended-regexp "^(ivxv-|haproxy)"

  echo "# Installing SSH authorized_keys file for '${ACCOUNT_NAME}'"
  ACCOUNT_HOME="$(getent passwd "${ACCOUNT_NAME}" | cut -d: -f 6)"
  test -d "${ACCOUNT_HOME}/.ssh" ||
    mkdir --verbose "${ACCOUNT_HOME}/.ssh"
  chown --changes "${ACCOUNT_NAME}:ivxv" "${ACCOUNT_HOME}/.ssh"
  chmod --changes 700 "${ACCOUNT_HOME}/.ssh"

  touch "${ACCOUNT_HOME}/.ssh/authorized_keys"
  chown --changes "${ACCOUNT_NAME}:ivxv" "${ACCOUNT_HOME}/.ssh/authorized_keys"
  chmod --changes 600 "${ACCOUNT_HOME}/.ssh/authorized_keys"
  grep ivxv-admin-account ~ivxv-admin/.ssh/authorized_keys \
    > "${ACCOUNT_HOME}/.ssh/authorized_keys"
}
# }}}

# init_host {{{
# initialize service host
init_host() {
  echo "# Removing IVXV service packages"
  dpkg --purge \
    ivxv-choices \
    ivxv-log \
    ivxv-mid \
    ivxv-proxy \
    ivxv-storage \
    ivxv-verification \
    ivxv-voting
  echo "# Removing IVXV config files"
  rm --force --verbose \
    /etc/ivxv/choices.bdoc \
    /etc/ivxv/election.bdoc \
    /etc/ivxv/technical.bdoc \
    /etc/ivxv/trust.bdoc \
    /etc/ivxv/voters-??.bdoc

  RSYSLOG_CFG_FILE="/etc/rsyslog.d/ivxv-logging.conf"
  if [ -e "${RSYSLOG_CFG_FILE}" ]; then
    echo "# Removing rsyslog config file"
    rm -fv "${RSYSLOG_CFG_FILE}"
    echo "# Restarting rsyslog service"
    systemctl restart rsyslog
  fi
}
# }}}

# init_service {{{
# initialize service data directory
init_service() {
  SERVICE_ID="$1"
  if [ "${SERVICE_ID}" = backup ]; then
    echo "# Removing backup service files"
    rm --recursive --force --verbose /var/backups/ivxv/management-conf/*
    rm --recursive --force --verbose /var/backups/ivxv/log/*
    rm --recursive --force --verbose /var/backups/ivxv/ballot-box/*
  else
    echo "# Removing service directory"
    rm --recursive --force --verbose "/var/lib/ivxv/service/${SERVICE_ID}"
  fi
}
# }}}

# install_package {{{
# install debian package dependencies
install_package() {
  DEB_FILENAME="$1"

  # operate in dumb terminal
  TERM=dumb
  export TERM
  DEBIAN_FRONTEND=noninteractive
  export DEBIAN_FRONTEND

  echo "# Performing dpkg database audit"
  dpkg --audit

  # install dependencies
  DEB_FILEPATH="/etc/ivxv/debs/${DEB_FILENAME}"
  DEPS_ORIG="$(dpkg -I "${DEB_FILEPATH}" | grep 'Depends:' | cut -d: -f2)"
  echo "# Installing package ${DEB_FILENAME} dependencies"
  DEPS="$(echo "${DEPS_ORIG}" | sed --regexp-extended --expression='s/\([^\)]+\)//g' --expression='s/,//g')"
  apt-get --yes install ${DEPS}

  # install package
  echo "# Installing package ${DEB_FILENAME}"
  dpkg -i "${DEB_FILEPATH}"
}
# }}}

# prepare_ballot_box_backup {{{
# Prepare votes backup file
prepare_ballot_box_backup() {
  SERVICE_ID="$1"
  BACKUP_FILENAME="$2"
  echo "# Creating ballot box backup file ${BACKUP_FILENAME}"
  rm -fv "${BACKUP_FILENAME}"
  RETVAL=""
  ivxv-voteexp -instance "${SERVICE_ID}" "${BACKUP_FILENAME}" || RETVAL="$?"
  if [ "${RETVAL}" = 2 ]; then
    echo "NOTE: The ivxv-voteexp utility exited with non-fatal errors"
  elif [ "${RETVAL}" ]; then
    echo "ERROR: ivxv-voteexp utility exited with error code ${RETVAL}"
    exit "${RETVAL}"
  fi
}
# }}}

# remove_admin_root_access {{{
# remove management service access to service host root account
remove_admin_root_access() {
  echo "# Removing ivxv-admin key from root SSH authorized_keys file"
  sed --in-place \
      --regexp-extended \
      --expression="s/^(ssh-rsa .+ ivxv-admin-account)\$/# \\1/" \
      /root/.ssh/authorized_keys
}
# }}}

# rsyslog_config_apply {{{
# apply IVXV logging config to rsyslog daemon
rsyslog_config_apply() {
  echo "# Installing IVXV logging config for rsyslog"
  cp --verbose /etc/ivxv/ivxv-logging.conf /etc/rsyslog.d/ivxv-logging.conf
  chmod --changes 644 /etc/rsyslog.d/ivxv-logging.conf
  chown --changes root:root /etc/rsyslog.d/ivxv-logging.conf
  echo "# Restarting rsyslog service"
  systemctl restart rsyslog
}
# rsyslog_config_apply }}}

# }}}

# parse CLI arguments {{{
ACTION="$1"
case "${ACTION}" in
  backup-ballot-box)
    backup_ballot_box "$2" "$3" "$4"
  ;;
  backup-log)
    backup_log "$2" "$3"
  ;;
  create-ssh-access)
    create_ssh_access "$2"
  ;;
  init-host)
    init_host
  ;;
  init-service)
    init_service "$2"
  ;;
  install-pkg)
    install_package "$2"
  ;;
  prepare-ballot-box-backup)
    prepare_ballot_box_backup "$2" "$3"
  ;;
  remove-admin-root-access)
    remove_admin_root_access
  ;;
  rsyslog-config-apply)
    rsyslog_config_apply
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
