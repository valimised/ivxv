Smart card initialization tool
==============================

The tool `erase_and_install.sh` formats the smart card PKCS15 file system and
initializes authentication tokens. It takes a single argument, which denotes the
number of readers for which to perform the tasks. The readers are indexed by the
sequence of being plugged into the system.

Internally, the tool uses `pkcs15-init` tool for the tasks which may display
requests for entering PIN code. The codes are automatically provided to the tool
and user does not need to input anything.

The PIN and PUK codes are given in the file `card_options.conf` and also
displayed to the user. Default PIN is 0000.
