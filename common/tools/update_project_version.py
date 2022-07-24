# IVXV Internet voting framework
"""
Update project version in subcomponents.

Release string is defined in debian/changelog.
Version string is release string without suffix (e.g. ~dev).
"""

import os
import re

from debian import debfile


def update_version(filename, pattern, repl):
    """Update version data in file."""

    # ignore files missing in some repositories
    if not os.path.exists(filename):
        return

    file_content = []
    updated = None
    pattern = re.compile(pattern)
    with open(filename) as fp:
        for line in fp:
            if pattern.match(line):
                oldline = line
                line = pattern.sub(repl, line)
                updated = line != oldline
            file_content.append(line)
    assert updated is not None

    if updated:
        with open(filename, 'w') as fp:
            fp.write(''.join(file_content))


def main():
    """Main routine."""
    # detect debian version
    with open('debian/changelog') as fp:
        debfile_data = debfile.Changelog(fp.read())
        release = str(debfile_data.version)

    version = release.split('~')[0]
    py_version = release
    # fix version for Python package (PEP 440)
    if '~' in py_version:
        py_version = py_version.replace('~', '.')
        try:
            int(py_version[-1])
        except ValueError:
            py_version += '0'
    debfile_year = int(debfile_data.date.split(' ')[3])
    copyright = '2016-{}, Cybernetica AS'.format(debfile_year)

    update_version("setup.py",
                   r"( +version=').+(',.*)", "\\g<1>%s\\2" % py_version)
    update_version("collector-admin/ivxv_admin/__init__.py",
                   r"(__version__ = ').+(')", "\\g<1>%s\\2" % py_version)
    update_version("collector-admin/ivxv_admin/__init__.py",
                   r"(DEB_PKG_VERSION = ').+(')", "\\g<1>%s\\2" % release)
    update_version("common/java/common-build.gradle",
                   "^(version ').+('.*)", "\\g<1>%s\\2" % release)
    update_version("tests/features/steps/__init__.py",
                   r"(DEB_VERSION = ').+(')", "\\g<1>%s\\2" % release)
    update_version("collector-admin/site/ivxv/about.html",
                   r'(.+id="version">).+(<)', '\\g<1>%s\\2' % release)
    update_version("Documentation/documents.py",
                   r"^(release *= *').+('.*)", "\\g<1>%s\\2" % release)
    update_version("Documentation/documents.py",
                   r"^(copyright *= *').+('.*)", "\\g<1>%s\\2" % copyright)
    update_version("Documentation/et/seadistuste_koostejuhend/mixnet.rst",
                   r"(.*ivxv-verificatum-).+(-runner.zip)", "\\g<1>%s\\2" % release)
    update_version('tests/templates/test-report/conf.py',
                   "^(version *= *').+('.*)", "\\g<1>%s\\2" % version)
    update_version('tests/templates/test-report/conf.py',
                   "^(release *= *').+('.*)", "\\g<1>%s\\2" % release)
    update_version("Documentation/et/audiitor/audit.rst",
                   r"^(.*(auditor|key|processor)-).+(.zip)", "\\g<1>%s\\3" % release)


if __name__ == '__main__':
    exit(main())
