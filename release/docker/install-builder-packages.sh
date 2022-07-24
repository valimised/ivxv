# IVXV Internet voting framework

# Script to install packages to release builder docker image

set -e

apt-get update

# Install apt-utils to avoid warning messages during install
apt-get install --yes apt-utils

# Upgrade existing packages
apt-get upgrade --yes

# Install management tools
apt-get install --yes --quiet curl git language-pack-et locales-all

# Install build tools
apt-get install --yes --quiet build-essential debhelper dh-exec dh-python dh-systemd dpkg-dev fakeroot libfile-fcntllock-perl make python3-all

# Install java
apt-get install --yes openjdk-11-jdk-headless

# Install golang
apt-get install --yes --quiet golang-1.14-go

# Install documentation tools
apt-get install --yes --quiet dvipng graphviz latexmk plantuml python3-debian python3-pip python3-setuptools texlive-fonts-recommended texlive-lang-european texlive-latex-extra texlive-plain-generic latexdiff

# Install validation tools
apt-get install --yes --quiet python3-jsonschema

# Install code checking tools
apt-get install --yes --quiet flake8

# Clean APT files
apt-get clean
rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

exit 0
