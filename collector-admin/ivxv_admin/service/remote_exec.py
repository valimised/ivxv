# IVXV Internet voting framework
"""Remote execution helper for microservice management."""

import subprocess

from .logging import log


def exec_remote_cmd(cmd, **kw):
    """Perform remote operation in service host.

    Execute service management command in service host (over OpenSSH SSH
    client) or copy file to service host file system (over secure copy).

    :param cmd: Command for subprocess.run()
    :type cmd: list
    :param kw: Parameters for subprocess.run()
    :type kw: dict
    :return: subprocess.CompletedProcess
    """
    assert cmd[0] in ('ssh', 'scp')

    # patch command
    if cmd[0] == 'ssh':
        cmd.insert(1, '-T')  # disable pseudo-terminal allocation
    cmd.insert(1, '-o')  # set preferred authentication method
    cmd.insert(2, 'PreferredAuthentications=publickey')

    # execute command
    log.debug('Executing command: %s', ' '.join(cmd))
    proc = subprocess.run(cmd, **kw)
    if proc.returncode:
        log.debug('Command finished with error code %d', proc.returncode)
    else:
        log.debug('Command successfully executed')

    return proc
