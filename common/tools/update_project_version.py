# IVXV Internet voting framework
"""
Update project version in subcomponents
"""

import os
import re

from debian import debfile


def update_version(filename, pattern, repl):
    """Update version data in file."""
    file_content = []
    updated = None
    pattern = re.compile(pattern)
    with open(filename) as fp:
        for line in fp:
            if pattern.match(line):
                oldline = line
                line = pattern.sub(repl, line)
                updated = line == oldline
            file_content.append(line)
    assert updated is not None

    if not updated:
        with open(filename, 'w') as fp:
            fp.write(''.join(file_content))


def main():
    """Main routine."""
    # detect debian version
    with open('debian/changelog') as fp:
        version = str(debfile.Changelog(fp.read()).version)

    update_version("setup.py",
                   r"( +version=').+(',.*)", "\\g<1>%s\\2" % version)
    update_version("collector-admin/ivxv_admin/ivxv_pkg.py",
                   r"(VERSION = ').+(')", "\\g<1>%s\\2" % version)
    update_version("common/java/common-build.gradle",
                   "^(version ').+('.*)", "\\g<1>%s\\2" % version)
#    update_version("tests/features/steps/__init__.py",
#                   r"(DEB_VERSION = ').+(')", "\\g<1>%s\\2" % version)
    update_version("collector-admin/site/ivxv/about.html",
                   r'(.+id="version">).+(<)', '\\g<1>%s\\2' % version)
    doc_dirname = 'Documentation/et'
    for dirname in os.listdir(doc_dirname):
        conf_filename = os.path.join(doc_dirname, dirname, 'conf.py')
        if os.path.exists(conf_filename):
            update_version(conf_filename,
                           "^(version *= *').+('.*)", "\\g<1>%s\\2" % version)
#    update_version('tests/templates/test-report/conf.py',
#                   "^(version *= *').+('.*)", "\\g<1>%s\\2" % version)


if __name__ == '__main__':
    exit(main())
