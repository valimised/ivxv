#!/bin/sh

# IVXV Internet voting framework

# Helper script for ivxv-admin account
# for actions that require privilege escalation

# Usage: ivxv-admin-helper <action> [<arg> ...]

set -e

usage() {
  echo "Usage:"
  echo "    ivxv-admin-helper configure-etcd-apt-source"
  echo "        Configure APT source for etcd"
  echo
  echo "    ivxv-admin-helper create-ssh-access <account-name>"
  echo "        Create management service access to account in service host"
  echo
  echo "    ivxv-admin-helper init-host"
  echo "        Initialize service host"
  echo
  echo "    ivxv-admin-helper init-service <service-id>"
  echo "        Initialize service data directory"
  echo
  echo "    ivxv-admin-helper install-pkg <package-filename>"
  echo "        Install IVXV package with dependencies"
  echo
  echo "    ivxv-admin-helper remove-admin-root-access"
  echo "        Remove management service access to service host root account"
  echo
  echo "    ivxv-admin-helper rsyslog-config-apply"
  echo "        Apply rsyslog config file for IVXV logging"
}

# ACTIONS {{{

# configure_etcd_apt_source {{{
# configure APT source for etcd
configure_etcd_apt_source() {
  # Add APT preferences file to define priority of etcd package.
  # https://wiki.debian.org/AptPreferences
  #
  # - Priority for etcd from zesty/zesty-updates is 700
  # - Default priority for packages is 500
  # - Priority for packages from zesty and zesty-updates is 100
  #
  # $ apt-cache policy etcd
  # etcd:
  #   Installed: (none)
  #   Candidate: 3.1.0-1
  #   Version table:
  #      3.1.0-1 700
  #         100 http://ee.archive.ubuntu.com/ubuntu zesty/universe amd64 Packages
  #      2.2.5+dfsg-1ubuntu1 500
  #         500 http://ee.archive.ubuntu.com/ubuntu xenial-updates/universe amd64 Packages
  #      2.2.5+dfsg-1 500
  #         500 http://ee.archive.ubuntu.com/ubuntu xenial/universe amd64 Packages
  echo "# Installing /etc/apt/preferences.d/ivxv-etcd-zesty.pref"
  cat > /etc/apt/preferences.d/ivxv-etcd-zesty.pref << EOF
Package: *
Pin: release o=Ubuntu,a=zesty
Pin-Priority: 100

Package: *
Pin: release o=Ubuntu,a=zesty-updates
Pin-Priority: 100

Package: *
Pin: release o=Ubuntu,a=zesty-security
Pin-Priority: 100

Package: etcd
Pin: release o=Ubuntu,a=zesty
Pin-Priority: 700

Package: etcd
Pin: release o=Ubuntu,a=zesty-updates
Pin-Priority: 700

Package: etcd
Pin: release o=Ubuntu,a=zesty-security
Pin-Priority: 700
EOF

  echo "# Installing /etc/apt/sources.list.d/zesty.list"
  cat > /etc/apt/sources.list.d/zesty.list << EOF
# IVXV Internet voting framework
# Package sources for "Ubuntu 17.04 (Zesty Zapus)"

# main
deb http://ee.archive.ubuntu.com/ubuntu zesty main universe
#deb-src http://ee.archive.ubuntu.com/ubuntu zesty main universe

# updates
deb http://ee.archive.ubuntu.com/ubuntu zesty-updates main universe
#deb-src http://ee.archive.ubuntu.com/ubuntu zesty-updates main universe

deb http://security.ubuntu.com/ubuntu zesty-security main universe
#deb-src http://security.ubuntu.com/ubuntu zesty-security main universe
EOF

  echo "# Updating package list"
  apt-get update
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
    ivxv-dds \
    ivxv-log \
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
}
# }}}

# init_service {{{
# initialize service data directory
init_service() {
  SERVICE_ID="$1"
  echo "# Removing service directory"
  rm --recursive --force --verbose "/var/lib/ivxv/${SERVICE_ID}"
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
  echo "# Installing package dependencies"
  DEPS="$(echo "${DEPS_ORIG}" | sed --regexp-extended --expression='s/\([^\)]+\)//g' --expression='s/,//g')"
  apt-get --yes install ${DEPS}

  # install package
  echo "# Installing package"
  dpkg -i "${DEB_FILEPATH}"
}
# }}}

# remove_admin_root_access {{{
# remove management service access to service host root account
remove_admin_root_access() {
  echo "# Removing ivxv-admin key from root SSH authorized_keys file"
  sed --in-place \
      --regexp-extended \
      --expression="s/^(ssh-rsa .+ ivxv-admin-account)\$/# \\\\1/" \
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
  configure-etcd-apt-source)
    configure_etcd_apt_source
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

# vim:sts=2 sw=2 et foldmethod=marker:
