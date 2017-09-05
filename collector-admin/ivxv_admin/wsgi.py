# IVXV Internet voting framework
"""
WSGI application for collector management service.

Default path for application is "/ivxv/cgi/"
"""

import datetime
import json
import os
import urllib
import urllib.request

import bottle

from bottle import request, response

from . import MANAGEMENT_DAEMON_URL, ROLES
from . import (COLLECTOR_STATE_CONFIGURED, COLLECTOR_STATE_FAILURE,
               COLLECTOR_STATE_NOT_INSTALLED, COLLECTOR_STATE_INSTALLED,
               COLLECTOR_STATE_PARTIAL_FAILURE)
from .config import CONFIG

FILE_UPLOAD_PATH = CONFIG.get('file_upload_path')


APP = bottle.Bottle()
# Default value for "Expires" header (don't allow to cache responses)
EXPIRES_DEFAULT = datetime.datetime(1970, 1, 1)


def abort(code=500, text='Unknown Error.'):
    """Aborts execution and causes a HTTP error."""
    body = json.dumps(dict(message=text))
    raise bottle.HTTPResponse(body, code,
                              headers={'content_type': 'application/json'})


@APP.route('/context.json')
def context():
    """Output context data."""
    # read user information from Apache environment
    user_cn = request.environ.get('SSL_CLIENT_S_DN_CN')
    assert user_cn
    assert ',' in user_cn

    # parse user information
    surname, name, idcode = user_cn.split(',')
    user_data = {
        'cn': user_cn,
        'user_name': '{} {}'.format(name, surname),
        'idcode': idcode,
        'role': [],
        'role-description': [],
        'permissions': [],
    }

    # detect user role
    for role in sorted(ROLES.keys()):
        access_filename = os.path.join(CONFIG.get('permissions_path'),
                                       '-'.join([user_cn, role]))
        if os.path.exists(access_filename):
            user_data['role'].append(role)
            user_data['role-description'].append(ROLES[role]['description'])
            user_data['permissions'] += ROLES[role]['permissions']

    # prepare context data
    if not user_data['role']:
        context_data = {}
        user_data.update({
            'role': ['none'],
            'role-description': [ROLES['none']['description']],
            'permissions': ROLES['none']['permissions'],
        })
    else:
        filepath = os.path.join(CONFIG['admin_ui_data_path'], 'status.json')
        with open(filepath) as ifp:
            status_json = json.load(ifp)
            context_data = {'collector': status_json['collector'],
                            'voting':
                                {'id': status_json['election']['election-id'],
                                 'stage': status_json['election']['phase']}}
    user_data['permissions'] = sorted(set(user_data['permissions']))
    context_data['current-user'] = user_data

    if 'collector' in context_data:
        status_info = {
            COLLECTOR_STATE_NOT_INSTALLED: ['Paigaldamata', 'info'],
            COLLECTOR_STATE_INSTALLED: ['Paigaldatud', 'info'],
            COLLECTOR_STATE_CONFIGURED: ['Seadistatud', 'success'],
            COLLECTOR_STATE_FAILURE: ['Tõrge', 'danger'],
            COLLECTOR_STATE_PARTIAL_FAILURE: ['Osaline tõrge', 'danger'],
        }
        status = context_data['collector']['status']
        context_data['collector'].update({
            'status-description': status_info[status][0],
            'status-indication': status_info[status][1],
        })

    # start response
    response.content_type = 'application/json'
    response.expires = EXPIRES_DEFAULT
    return json.dumps({'data': context_data})


@APP.post('/add_user')
def add_user():
    """Add user account."""
    raise NotImplementedError()


@APP.post('/change_user_role')
def change_user_role():
    """Change role of the existing user."""
    raise NotImplementedError()


@APP.post('/download-ballot-box')
def download_ballot_box():
    """Download the ballot box"""
    daemon_url = MANAGEMENT_DAEMON_URL + 'download-ballot-box'
    with urllib.request.urlopen(daemon_url) as req_fp:
        daemon_response = req_fp.read(1024)

    # start response
    response.content_type = 'application/json'
    response.expires = EXPIRES_DEFAULT
    return daemon_response


@APP.post('/apply_voting_list')
def apply_voting_list():
    """Upload and apply voting list."""
    upload = request.files.get('list')
    upload.save('/tmp/uploaded_voting_list')
    raise NotImplementedError()


@APP.post('/upload-config')
def apply_config():
    """
    Upload and apply command file.

    Command file is a digitally signed file that contains one of the following
    commands:

        1. User management command (add, change role)
        2. Apply collector config (contains config file)
        3. Apply voting list (contains voting list)
        4. Download ballot-box
    """
    upload = request.files.get('upload')
    if upload is None:
        abort(400, 'Üleslaaditav fail on määramata')

    # save file to local directory
    assert os.path.isdir(FILE_UPLOAD_PATH)
    filename = '-'.join([datetime.datetime.now().strftime('%s.%f'),
                         upload.filename])
    file_path = os.path.join(FILE_UPLOAD_PATH, filename)
    upload.save(file_path)

    # forward command to management daemon
    daemon_url = MANAGEMENT_DAEMON_URL + 'apply-command'
    post_data = urllib.parse.urlencode({'filename': filename,
                                        'original_filename': upload.filename,
                                        'cmd_type': request.forms.get('type')})
    post_data = post_data.encode('UTF-8')
    with urllib.request.urlopen(daemon_url, post_data) as req_fp:
        daemon_response = req_fp.read(1024)

    # start response
    response.content_type = 'application/json'
    response.expires = EXPIRES_DEFAULT
    return daemon_response

# vim:set sts=4 sw=4 et ft=python:
