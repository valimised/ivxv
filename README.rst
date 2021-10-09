==============================
IVXV Internet voting framework
==============================

<General introduction to IVXV.>

----------
 Building
----------

Building the collector components requires Go 1.9 and the management
applications require Java 11::

        sudo apt-get install --no-install-recommends openjdk-11-jdk-headless golang-1.9-go

Next, external dependencies need to be acquired. If working off of an offline
copy of the IVXV repository, then they are already included and this step can
be skipped. Otherwise run

::

        make external

to download the external dependecy repository.

Finally, to build, test, and clean the entire codebase, just do

::

        make all
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

Updating external dependencies
------------------------------

During development, if changes are pushed to the external dependencies
repository, then the reference needs to be updated in this repository to link
those changes to the current revision. This is done with

::

        make update-external
        git add common/external
        git commit
        git push

After the reference update has been pulled by other developers, they will get
an entry in their ``git status`` output, indicating that external has changed::

        modified:   common/external (new commits)

To fetch the changes to their local working tree, developers need to rerun

::

        make external

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


----------
 Releasing
----------

Release builds are made using test system.

To make release build:

* run ``dch --release`` to finalize the changelog for a release

* run ``make release``

        *Note!* This creates a new virtual environment for building the
        release, so if you have installed a custom binaries or libraries to the
        local machine (e.g., the patched Go standard library from ivxv-golang),
        then those will not be used. In this case, build the Debian packages
        manually.
