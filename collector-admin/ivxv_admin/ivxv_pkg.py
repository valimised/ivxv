# IVXV Internet voting framework
"""
IVXV debian package information.
"""

# package version number
VERSION = '0.9'

COLLECTOR_PACKAGE_FILENAMES = {
    'ivxv-admin': 'ivxv-admin_%s_all.deb' % VERSION,
    'ivxv-choices': 'ivxv-choices_%s_amd64.deb' % VERSION,
    'ivxv-common': 'ivxv-common_%s_all.deb' % VERSION,
    'ivxv-dds': 'ivxv-dds_%s_amd64.deb' % VERSION,
    'ivxv-log': 'ivxv-log_%s_all.deb' % VERSION,
    'ivxv-proxy': 'ivxv-proxy_%s_amd64.deb' % VERSION,
    'ivxv-storage': 'ivxv-storage_%s_amd64.deb' % VERSION,
    'ivxv-verification': 'ivxv-verification_%s_amd64.deb' % VERSION,
    'ivxv-voting': 'ivxv-voting_%s_amd64.deb' % VERSION,
}
