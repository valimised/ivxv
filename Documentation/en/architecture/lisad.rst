..  IVXV arhitecture

Annexes
=======

Building the Collector Service with a Patched Go Standard Library
-----------------------------------------------------------------

Several ID cards and digital IDs exist in Estonia whose certificates include an
incorrectly encoded RSA public key. The Go standard library refuses to accept
such faulty certificates. At the same time, there are too many to do nothing
about it.

One solution would be to compile the IVXV collector service subservices using a
patched Go standard library. The delivery comes with a  ``ivxv-golang`` package,
which contains patches for accepting faulty certificates of this type and the
means for implementing them.

The patched standard library should be built in the same environment where the
IVXV collector service is built, i.e. Ubuntu 18.04.

First, all the dependencies listed in the ``README.rst`` file in the
``ivxv-golang`` folder have to be installed.  Then, with ``make`` command , the
newest Go 1.9 source code is downloaded from Ubuntu depositories, patched, built
and tested for the use of faulty certificates.  If it is successful, the
``source/`` subfolder includes two necessary .deb packages:

- ``golang-1.9-src_1.9.4-rubuntu1_amd64.deb`` and

- ``golang-1.7-go_1.9.4-1ubuntu1_amd64.deb``.

These have to be installed in the computer building the IVXV collector service
before IVXV is compiled: then the Go standard library patched during preparation
is used.
