# IVXV Internet voting framework
"""
Management daemon for collector management service.
"""

import datetime
import json
import os
import re
import subprocess

from bottle import post, request, response, route, run, get

from . import MANAGEMENT_DAEMON_PORT
from .config import CONFIG

FILE_UPLOAD_PATH = CONFIG.get('file_upload_path')


@route('/')
def index():
    """Index page."""
    return '<b>This is a web server for collector management service</b>!'


@post('/apply-command')
def apply_command():
    """Upload command file."""
    filename = request.forms.get('filename')
    original_filename = request.forms.get('original_filename')
    cmd_type = request.forms.get('cmd_type')
    file_path = os.path.join(FILE_UPLOAD_PATH, filename)

    cmd = ['ivxv-cmd-load', cmd_type, file_path]
    process = subprocess.run(cmd, stdout=subprocess.PIPE,
                             stderr=subprocess.PIPE)

    out = (process.stderr.decode('utf-8') + '\n' +
           process.stdout.decode('utf-8'))
    body = dict(message='Fail "%s" on edukalt üles laaditud' %
                original_filename)

    if re.search(r'error', out, re.IGNORECASE):
        body = out
    elif cmd_type in ['technical', 'election', 'choices', 'voters']:
        cmd = ['ivxv-config-apply', '--type=' + cmd_type]
        process = subprocess.run(
            cmd, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
        out = process.stderr.decode(
            'utf-8') + '\n' + process.stdout.decode('utf-8')
        if re.search(r'error', out, re.IGNORECASE):
            body = out

    # start response
    response.content_type = 'application/json'
    return json.dumps(body)


@get('/download-ballot-box')
def download_ballots():
    """Export votes into ballot box folder."""
    timestamp = '{:%Y.%m.%d_%H.%M}'.format(datetime.datetime.now())
    # FIXME: Hardcoded folder location, because
    # CONFIG_DEFAULTS.get('exported_votes_path') == "%(ivxv_admin_data_path)s/ballot-box"
    folder = '/var/lib/ivxv/ballot-box/'
    filename = 'exported-votes-{}.zip'.format(timestamp)
    path = os.path.join(folder, filename)
    cmd = ['ivxv-votes-export', path]
    subprocess.run(cmd)

    os.chmod(path, 0o666)

    # start response
    response.content_type = 'application/json'
    return json.dumps(filename)


def daemon():
    """Daemon main loop."""
    run(host='localhost', port=MANAGEMENT_DAEMON_PORT)
