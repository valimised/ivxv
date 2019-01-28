====================================================
IVXV Internet voting framework external dependencies
====================================================

This directory contains source code and build artifacts of all
external dependencies of the IVXV Internet voting framework.

Dependencies required in this directory are listed in
:file:`../../Documentation/et/arhitektuur/tehnoloogiad.rst`.

Note that the source code for all Go dependencies must be in a `gopath/src/`
subdirectory.

Java
----

To get Java applications' (auditor, key, processor) dependencies on a Linux
machine, run the script :file:`install_java_dep`.

In order to achieve the same on Windows, do the following:

1. Install Gradle 3.0 globally (or see point 4)
2. In :file:`../java/` directory, run :code:`gradle syncRemoteRepositories`
3. Manually download and add the following jars
  - commons-codec/commons-codec/1.10/commons-codec-1.10.jar
  - xml-apis/xml-apis/1.3.04/xml-apis-1.3.04.jar
4. To build or test Java applications with Make, Gradle 3.0 has to be manually
installed into this directory. See :file:`../java/common.mk` for exact path.
