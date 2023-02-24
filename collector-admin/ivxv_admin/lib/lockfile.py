# IVXV Internet voting framework
"""Lockfile implementation for collector management service."""

import argparse
import atexit
import fcntl
import logging
import os
import sys
import time

from ..config import cfg_path

# create logger
log = logging.getLogger(
    f"{__package__}.lockfile" if __name__ == "__main__" else __name__
)


class PidLocker:
    """Pidfile creator and locker.

    Create pidfile and keep it locked until program exit.
    Raise IOError on locking failure.

    >>> try:
    ...     PidLocker("/path/to/pidfile")
    ... except IOError:
    ...     print('Locking failed')
    ...     exit(1)
    """

    filepath = None  #: Pidfile name
    fp = None  #: Pidfile descriptor

    def __init__(self, pidfile_name):
        """Create and lock pidfile.

        :raises IOError: On locking failure
        """
        self.filepath = pidfile_name
        self.fp = open(pidfile_name, "ab")
        fcntl.flock(self.fp, fcntl.LOCK_EX | fcntl.LOCK_NB)
        atexit.register(self.rm_pid)
        self.fp.write(bytes(f"{os.getpid()}\n", "ASCII"))
        self.fp.flush()

    def rm_pid(self):
        """Remove pidfile."""
        try:
            os.remove(self.filepath)
        except FileNotFoundError:
            log.warning("PID file %r already removed", self.filepath)

    @classmethod
    def get_pidfile_path(cls, pidfile_name):
        """Get pidfile path."""
        pidfile_path = cfg_path("ivxv_admin_data_path", pidfile_name)
        return pidfile_path

    @classmethod
    def pidfile_exists(cls, pidfile_name):
        """Check if pidfile exist."""
        pidfile_path = cls.get_pidfile_path(pidfile_name)
        return os.path.exists(pidfile_path)

    @classmethod
    def rm_stale_pidfile(cls, pidfile_name):
        """Remove stale or invalid pidfile."""
        if not cls.pidfile_exists(pidfile_name):
            return

        pidfile_path = cls.get_pidfile_path(pidfile_name)
        with open(pidfile_path) as fp:
            pidline = fp.readline()

        reason = "" if pidline.strip() else "empty"
        if not reason:
            try:
                pid = int(pidline)
            except ValueError:
                reason = "invalid"
        if not reason and not os.path.exists(f"/proc/{pid}"):
            reason = "stale"
        if reason:
            log.error("Removing %s pidfile %r", reason, pidfile_path)
            os.remove(pidfile_path)


def main():
    """Lockfile testing utility."""
    # parse CLI args
    parser = argparse.ArgumentParser(
        description="""
        Create lock file and related process to hold it for a specified period.
        """
    )
    parser.add_argument("--file", required=True, type=str, help="Lock file name")
    parser.add_argument(
        "--time", required=True, type=int, help="Period to hold lock file (in seconds)"
    )
    parser.add_argument(
        "--attempts",
        required=False,
        type=int,
        default=1,
        help="Attempts to acquire lock file",
    )
    parser.add_argument(
        "--background", action="store_true", help="Fork process to background"
    )

    args = parser.parse_args()
    assert args.time >= 0, "--time can't have negative value"
    assert args.attempts > 0, "--attempts must have positive value"

    filepath = PidLocker.get_pidfile_path(args.file)
    attempts_left = args.attempts
    proc_type = "background" if args.background else "foreground"

    # fork to background if requested
    if args.background:
        log.info("Forking to background")
        pid = os.fork()
        if pid:
            log.info("Forked child process %d, exiting foreground process", pid)
            return 0

    # acquire lockfile
    PidLocker.rm_stale_pidfile(args.file)
    log.info("Creating lockfile %r for PID %d in %s", filepath, os.getpid(), proc_type)
    while attempts_left:
        attempts_left -= 1
        try:
            PidLocker(filepath)
        except BlockingIOError as err:
            if not attempts_left:
                log.error("Failed to create lockfile (%s), no attempts left", err)
                return 1
            log.warning(
                "Failed to create lockfile %r, waiting for 1 second (%d attempts left)",
                filepath,
                attempts_left,
            )
            time.sleep(1)
            continue
        log.info("Lockfile %r successfully acquired", filepath)
        break

    # sleep
    log.info("Sleeping for %d seconds", args.time)
    time.sleep(args.time)

    # cleanup
    log.info("Removing lockfile %r for PID %d in %s", filepath, os.getpid(), proc_type)

    return 0


if __name__ == "__main__":
    sys.exit(main())
