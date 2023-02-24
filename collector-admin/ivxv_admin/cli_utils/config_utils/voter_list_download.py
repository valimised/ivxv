# IVXV Internet voting framework
"""CLI utility to download voter list from VIS."""

import datetime
import json
import os
import shutil
import sys
import zipfile

import requests

from ... import command_file
from ...config import CONFIG, cfg_path
from ...db import IVXVManagerDb
from ...event_log import register_service_event
from ...lib import get_current_voter_list_changeset_no
from .. import init_cli_util, log
from .command_load import register_cfg, validate_lists_consistency


def main():
    """Download voter list from VIS to IVXV Collector Management Service."""
    args = init_cli_util(
        """
        Download next available voter list changeset from VIS to IVXV Collector
        Management Service.

        Usage:
            ivxv-voter-list-download [--output-report=<filepath>] [--log-level=<level>]

        Options:
            --output-report=<filepath>  Write JSON report about HTTP request to VIS.
            --log-level=<level>         Logging level [Default: INFO].
    """
    )

    # validate CLI arguments
    output_report_filepath = args["--output-report"]

    # suppress error if executed from crontab before loading election config
    if not os.path.exists("/etc/ivxv/election.bdoc") and not sys.stdout.isatty():
        return 0

    # detect state of the last changeset
    with IVXVManagerDb() as db:
        changeset_no = get_current_voter_list_changeset_no(db)
        changeset_state = db.get_value(f"list/voters{changeset_no:04d}-state")

    # stop on INVALID changeset
    if changeset_state == "INVALID":
        log.info(
            "State of the last registered voter list changeset #%d is %s",
            changeset_no,
            changeset_state,
        )
        log.info("Skipping download of the next voter list changeset")
        if output_report_filepath:
            with open(output_report_filepath, "w") as fd:
                json.dump(
                    {"voter-list-changeset": {"status": f"Skipped"}}, fd, indent=True
                )
        return 0
    changeset_no += 1

    # get params from election config
    command_file.log.setLevel(args["--log-level"])
    cfg = command_file.load_collector_cmd_file("election", "/etc/ivxv/election.bdoc")
    if not cfg:
        return 1
    election_id = cfg["identifier"]
    vis_base_url_template = f"{cfg['vis']['url']}{{endpoint}}"
    ca_certs_filepath = None
    if cfg["vis"].get("ca"):
        ca_certs_filepath = cfg_path("vis_path", "ca.pem")
        with open(ca_certs_filepath, "w") as fd:
            fd.write("\n".join(cfg["vis"]["ca"]))

    # download voter list
    output_filepath = f"{CONFIG.get('vis_path')}/voters{changeset_no:04d}.zip"
    url = vis_base_url_template.format(endpoint="ehs-election-voters-changeset")
    log.info("Query voter list update #%d from %r", changeset_no, url)
    request_params = {"electionCode": election_id, "changeset": changeset_no}
    resp = requests.get(
        url,
        verify=ca_certs_filepath,
        cert=(
            "/etc/ssl/certs/ivxv-admin-client.crt",
            "/etc/ssl/private/ivxv-admin-client.key",
        ),
        params=request_params,
    )
    log.info(
        "VIS responded to voter list update #%d: %r %s",
        changeset_no,
        resp.status_code,
        resp.reason,
    )
    timestamp = datetime.datetime.utcnow().strftime("%Y-%m-%dT%H:%M:%SZ")
    version = f"{url} {timestamp}"

    # write query report
    if output_report_filepath:
        report = {
            "voter-list-changeset": {
                "operation": f"Download voter list update #{changeset_no} from VIS",
                "request": {"url": resp.url, "method": "GET", "params": request_params},
                "response": {
                    "status": "OK" if resp.ok else "ERROR",
                    "status_code": resp.status_code,
                    "reason": resp.reason,
                    "time_elapsed": str(resp.elapsed),
                    "headers": dict([key, val] for key, val in resp.headers.items()),
                    "content": resp.text,
                },
            },
        }
        with open(output_report_filepath, "w") as fd:
            json.dump(report, fd, indent=True)

    # register download event / handle download errors
    if resp.status_code == 404:
        log.info("Voter list changeset #%d it not available in VIS", changeset_no)
        return 0
    if resp.status_code != 200:
        log.error("Server responded with status %d - %r", resp.status_code, resp.reason)
        register_service_event(
            "VOTER_LIST_DOWNLOAD_FAILED",
            level="INFO",
            service="admin",
            params={"changeset_no": changeset_no},
        )
        return 1
    register_service_event(
        "VOTER_LIST_DOWNLOADED",
        level="INFO",
        service="admin",
        params={"changeset_no": changeset_no},
    )

    # write downloaded content to file
    write_voter_list_zip(resp.content, version, output_filepath)

    # register voter list
    if changeset_no:
        register_voter_list(output_filepath, timestamp, version)

    return 0


def write_voter_list_zip(content, version, output_filepath):
    """Write voter list ZIP file."""
    # write list to temporary ZIP file
    tmp_filepath = f"{output_filepath}.tmp"
    with open(tmp_filepath, "bw") as fd:
        tmp_filepath = fd.name
        log.info("Writing VIS response to temporary file %r", tmp_filepath)
        fd.write(content)

    # add version to ZIP
    comment = f"Version: {version}"
    log.info("Adding comment %r to %r", comment, tmp_filepath)
    with zipfile.ZipFile(tmp_filepath, "a") as zip_file:
        zip_file.comment = bytes(comment, "ASCII")

    # move ZIP file to final place
    log.info("Moving temporary file %r to %r", tmp_filepath, output_filepath)
    shutil.move(tmp_filepath, output_filepath)


def register_voter_list(filepath, timestamp, version):
    """Register voter list in management service."""
    cfg_data = command_file.load_collector_cmd_file("voters", filepath)
    is_cfg_valid = cfg_data is not None
    if is_cfg_valid:
        is_cfg_valid = validate_lists_consistency("voters", filepath)

    register_service_event(
        "CMD_LOAD", params={"cmd_type": "voters", "version": version}
    )
    register_cfg(
        cmd_type="voters",
        cfg_data=cfg_data,
        cfg_filename=filepath,
        cfg_timestamp=timestamp,
        cfg_version=version,
        autoapply=is_cfg_valid,
        state="PENDING" if is_cfg_valid else "INVALID",
    )
    register_service_event(
        "CMD_LOADED", params={"cmd_type": "voters", "version": version}
    )
