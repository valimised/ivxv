==============================
IVXV Internet voting framework
==============================

<General introduction to IVXV.>

----------
 Building
----------

Building the collector components requires Go 1.9 and the management
applications require Java 8::

        sudo apt-get install --no-install-recommends openjdk-8-jdk-headless golang-1.9-go


Next, external dependencies need to be acquired. See common/external/README.rst
for instructions on this.

Finally, to build, test, and clean the entire codebase, just do

::

        make
        make test
        make clean

Individual components can be built, tested, and cleaned with

::

        make <component>
        make test-<component>
        make clean-<component>

where ``<component>`` is the name name of the components subdirectory.

All components are built in release mode by default. To build in development
mode, which enables dummy modules and other components not safe for release,
the make variable ``DEVELOPMENT`` must be set either on the command line or in
the environment. Every target of the root Makefile can be called with a
``-dev`` suffix, which sets this variable during the make.

Debian packages can be built using ``dpkg-buildpackage`` (or ``debuild`` if
preferred) as usual.

-------------
 Development
-------------

Setting GOPATH
--------------

The IVXV project uses multiple GOPATH directories, one for each Go component.
During ``make`` commands, the GOPATH will be automatically set, but if a
developer (or their IDE) wishes to run some manual Go commands, then this can
be tedious to specify. To help with this

::

        make gopath

will print out the GOPATH used by the build system. So for example, when using
Bash,

::

        export GOPATH="$(make gopath)"

will set the correct GOPATH in the current shell.

