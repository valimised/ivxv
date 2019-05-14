# IVXV Internet voting framework
"""
Management daemon for collector management service.

This daemon is based on :py:mod:`bottle` module and listens as web service.
"""

import datetime
import json
import logging
import os
import subprocess

from bottle import get, post, request, response, route, run

from . import MANAGEMENT_DAEMON_PORT
from .config import cfg_path

# create logger
log = logging.getLogger(__name__)


@route('/')
def index():
    """
    Index page method.

    :return: Service identification message.
    """
    return '<b>This is a web server for IVXV Collector Management Service</b>!'


@post('/upload-command')
def upload_cmd_file():
    """Upload command file."""
    filename = request.forms.get('filename')
    original_filename = request.forms.get('original_filename')
    cmd_type = request.forms.get('cmd_type')
    file_path = cfg_path('file_upload_path', filename)
    log.info('Uploading command file "%s" (type: %s)', filename, cmd_type)

    # execute command loading utility
    cmd = ['ivxv-cmd-load', '--autoapply', cmd_type, file_path]
    log.info('Executing command: %s', ' '.join(cmd))
    proc = subprocess.run(
        cmd, stdout=subprocess.PIPE, stderr=subprocess.STDOUT)
    log.info('Command finished with exit code %d', proc.returncode)

    # report upload result
    body = {
        'success': bool(proc.returncode == 0),
        'log': proc.stdout.decode('utf-8').split('\n'),
    }
    if proc.returncode:  # error
        body['message'] = f'Viga faili "{original_filename}" üleslaadimisel'
    else:  # command loaded
        body['message'] = (
            f'Fail "{original_filename}" on edukalt üles laaditud')

    # start response
    response.content_type = 'application/json'
    return json.dumps(body)


@get('/download-ballot-box')
@get('/download-consolidated-ballot-box')
def download_ballots():
    """Export votes into ballot box folder."""
    timestamp = '{:%Y.%m.%d_%H.%M}'.format(datetime.datetime.now())

    # prepare command
    cmd = ['ivxv-export-votes']
    if 'download-consolidated-ballot-box' in request.url:
        cmd.append('--consolidate')
        filename = f'exported-votes-consolidated-{timestamp}.zip'
    else:
        filename = f'exported-votes-{timestamp}.zip'
    path = cfg_path('exported_votes_path', filename)
    cmd.append(path)

    cmd_str = ' '.join(cmd)
    logfile_path = cfg_path(
        'exported_votes_path', filename.replace('.zip', '.log'))
    cmd = [
        'sh', '-e', '-c',
        f'( {cmd_str} && chmod 664 {path} ) > {logfile_path}'
    ]
    subprocess.Popen(cmd)

    # start response
    response.content_type = 'application/json'
    return json.dumps('OK')


@get('/remove-voters-lists')
def remove_voters_lists():
    """Remove loaded but unapplied voters lists"""
    cmd = ['ivxv-cmd-remove', 'voters']
    log.info('Executing command: %s', ' '.join(cmd))
    proc = subprocess.run(
        cmd, stdout=subprocess.PIPE, stderr=subprocess.STDOUT)
    log.info('Command finished with exit code %d', proc.returncode)

    # report result
    body = {
        'success': bool(proc.returncode == 0),
        'log': proc.stdout.decode('utf-8').split('\n'),
    }

    if proc.returncode:  # error
        body['message'] = 'Viga nimekirjade eemaldamisel'
    else:  # command loaded
        body['message'] = 'Nimekirjad edukalt eemaldatud'

    # start response
    response.content_type = 'application/json'
    return json.dumps(body)


def daemon():
    """Daemon main loop."""
    log.info('Starting Management daemon')
    run(host='localhost', port=MANAGEMENT_DAEMON_PORT)
