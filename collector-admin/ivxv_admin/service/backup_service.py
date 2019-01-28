# IVXV Internet voting framework
"""Backup service management helper."""

import subprocess

from .logging import log


def install_backup_crontab():
    """Install crontab for backup automation."""
    subprocess.run(['env', 'VISUAL=ivxv-backup-crontab', 'crontab', '-e'])


def remove_backup_crontab():
    """Remove backup automation crontab if exist."""
    log.info('Removing crontab (if exist)')
    proc = subprocess.run(['crontab', '-r'])
    assert proc.returncode in [0, 1], 'Unexpeced exit code for crontab command'
