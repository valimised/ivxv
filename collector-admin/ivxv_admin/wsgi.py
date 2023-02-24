# IVXV Internet voting framework
"""
WSGI application for collector management service.

Default path for application is "/ivxv/cgi/"
"""

import datetime
import json
import logging
import os
import re
import urllib
import urllib.request

import bottle
from bottle import request, response

from . import (
    COLLECTOR_STATE_CONFIGURED,
    COLLECTOR_STATE_FAILURE,
    COLLECTOR_STATE_INSTALLED,
    COLLECTOR_STATE_NOT_INSTALLED,
    COLLECTOR_STATE_PARTIAL_FAILURE,
    EVENT_LOG_FILENAME,
    MANAGEMENT_DAEMON_URL,
    USER_ROLES,
    __version__,
)
from .config import CONFIG, cfg_path

FILE_UPLOAD_PATH = CONFIG['file_upload_path']

APP = bottle.Bottle()
#: Default value for "Expires" header (don't allow to cache responses)
EXPIRES_DEFAULT = datetime.datetime(1970, 1, 1)

log = logging.getLogger(__name__)


def abort(code=500, text='Unknown Error.'):
    """
    Abort execution and raise a HTTP error.

    :raises bottle.HTTPResponse:
    """
    body = json.dumps(dict(message=text))
    raise bottle.HTTPResponse(
        body, code, headers={'content_type': 'application/json'})


@APP.hook('before_request')
def log_query():
    """Log query data."""
    log.info('%s %s', request.method, request.urlparts.path)
    if request.forms:
        log.info('POST params: %s', dict(request.forms))
    for key, fileobj in request.files.items():
        log.info("FILE: field %r, filename %r", key, fileobj.raw_filename)


@APP.route('/context.json')
def context():
    """
    Output context data (logged in user, collector state).

    :return: Context data as JSON
    :rtype: str
    """
    # read user information from Apache environment
    user_cn = request.environ.get('SSL_CLIENT_S_DN_CN')
    assert user_cn
    assert ',' in user_cn

    # parse user information
    surname, name, idcode = user_cn.split(',')
    user_data = {
        'cn': user_cn,
        'user_name': f'{name} {surname}',
        'idcode': idcode,
        'role': [],
        'role-description': [],
        'permissions': [],
    }

    # detect user role
    for role in sorted(USER_ROLES):
        access_filename = cfg_path(
            'permissions_path', '-'.join([user_cn, role]))
        if os.path.exists(access_filename):
            user_data['role'].append(role)
            user_data['role-description'].append(
                USER_ROLES[role]['description'])
            user_data['permissions'] += USER_ROLES[role]['permissions']

    # prepare context data
    if not user_data['role']:
        context_data = {}
        user_data.update({
            'role': ['none'],
            'role-description': [USER_ROLES['none']['description']],
            'permissions': USER_ROLES['none']['permissions'],
        })
    else:
        filepath = cfg_path('admin_ui_data_path', 'status.json')
        with open(filepath) as ifp:
            state_json = json.load(ifp)
            context_data = {
                'collector': state_json['collector'],
                'election': {
                    'id': state_json['election']['election-id'],
                    'stage': state_json['election']['phase']
                }
            }
            context_data["collector"]["version"] = __version__
    user_data['permissions'] = sorted(set(user_data['permissions']))
    context_data['current-user'] = user_data

    if 'collector' in context_data:
        state_info = {
            COLLECTOR_STATE_NOT_INSTALLED: ['Paigaldamata', 'info'],
            COLLECTOR_STATE_INSTALLED: ['Paigaldatud', 'info'],
            COLLECTOR_STATE_CONFIGURED: ['Seadistatud', 'success'],
            COLLECTOR_STATE_FAILURE: ['Tõrge', 'danger'],
            COLLECTOR_STATE_PARTIAL_FAILURE: ['Osaline tõrge', 'danger'],
        }
        state = context_data['collector']['state']
        context_data['collector'].update({
            'state-description': state_info[state][0],
            'state-indication': state_info[state][1],
        })

    # start response
    response.content_type = 'application/json'
    response.expires = EXPIRES_DEFAULT
    return json.dumps({'data': context_data})


@APP.post('/download-ballot-box')
@APP.post('/download-consolidated-ballot-box')
@APP.post('/skip-voters-list')
@APP.post('/download-voter-detail-stats')
@APP.post('/download-voting-sessions')
@APP.post('/download-processor-input')
def fwd_request():
    """Forward request to IVXV Management Service Daemon."""
    path = request.fullpath.split('/')[-1]
    daemon_url = MANAGEMENT_DAEMON_URL + path

    data = urllib.parse.urlencode(request.forms).encode("UTF-8")
    fwd_request = urllib.request.Request(daemon_url, data=data, method="GET")
    max_response_size = 1024 * 1024 * 1024  # 1GB
    with urllib.request.urlopen(fwd_request) as req_fp:
        daemon_response = req_fp.read(max_response_size)
        if len(daemon_response) == max_response_size:
            response.content_type = "application/json"
            return json.dumps(
                {
                    "success": False,
                    "message": f"Vastus on suurem kui {max_response_size} baiti",
                }
            )
        response.content_type = req_fp.headers["Content-Type"] or "application/json"
        for header in ["Content-Length", "Content-Disposition"]:
            if req_fp.headers[header]:
                response.set_header(header, req_fp.headers[header])

    # start response
    response.expires = EXPIRES_DEFAULT
    return daemon_response


@APP.post('/upload-config')
def upload_cfg():
    """
    Upload command file.

    Command file is a digitally signed file that contains one of the following
    commands:

    1. User management command (add, change role)
    2. Collector config command (contains config file or voting list)
    """
    upload = request.files.get('upload')
    if upload is None:
        abort(400, 'Üleslaaditav fail on määramata')

    # save file to local directory
    assert os.path.isdir(FILE_UPLOAD_PATH)
    filename = '-'.join(
        [datetime.datetime.now().strftime('%s.%f'), upload.filename])
    file_path = os.path.join(FILE_UPLOAD_PATH, filename)
    upload.save(file_path)

    # forward command to management daemon
    daemon_url = MANAGEMENT_DAEMON_URL + 'upload-command'
    post_data = urllib.parse.urlencode({
        'filename': filename,
        'original_filename': upload.filename,
        'cmd_type': request.forms.get('type')
    })
    post_data = post_data.encode('UTF-8')
    with urllib.request.urlopen(daemon_url, post_data) as req_fp:
        daemon_response = req_fp.read()

    # start response
    response.content_type = 'application/json'
    response.expires = EXPIRES_DEFAULT
    return daemon_response


@APP.route('/eventlog')
def eventlog():
    """Convert Collector event log file to valid JSON.

    :return: Event log as JSON
    :rtype: str
    """
    records = []
    filepath = cfg_path('ivxv_admin_data_path', EVENT_LOG_FILENAME)
    try:
        with open(filepath) as fp:
            while True:
                line = fp.readline()
                if not line:
                    break
                records.append(json.loads(line))
    except OSError as err:
        log.error(err)
        abort(text=str(err))

    # start response
    response.content_type = 'application/json'
    response.expires = EXPIRES_DEFAULT
    return json.dumps({'data': records})


@APP.route('/ballot-box-state')
def get_ballot_box_state():
    """Output state of ballot boxes.

    :return: State as JSON
    :rtype: str
    """
    state = []
    for filename in os.listdir(CONFIG['exported_votes_path']):
        if not filename.endswith('.log'):
            continue
        timestamp = datetime.datetime.strptime(
            re.sub(r'.+-(\d.+).log', '\\1', filename),
            '%Y.%m.%d_%H.%M')
        filepath = os.path.join(CONFIG['exported_votes_path'], filename)
        with open(filepath) as fp:
            log_lines = fp.readlines()
            if len(log_lines) > 12:
                log_lines = log_lines[:6] + ['...\n'] + log_lines[-6:]
        zip_filename = filename.replace('.log', '.zip')
        file_state = (
            'ready'
            if 'Collected votes archive is written to' in log_lines[-1]
            else 'prepare')
        state.append({
            'timestamp': timestamp.strftime('%d.%m.%Y %H:%M'),
            'filename': zip_filename,
            'state': file_state,
            'log': ''.join(log_lines),
        })

    # start response
    response.content_type = 'application/json'
    response.expires = EXPIRES_DEFAULT
    return json.dumps({'data': state})
