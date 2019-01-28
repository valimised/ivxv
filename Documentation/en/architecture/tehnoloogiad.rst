..  IVXV arhitecture

.. _technology:

Technology Used
========================

Collector Service Programming Language
---------------------------------------

The core functionality of the collector service is programmed using the Go
programming language, which meets the following procurement requirements:

* Static typing
* Automatic memory management
* Open source code compiler
* Concurrency

The collector service administration service is programmed in Python.


Programming Language of Applications
-------------------------------------

The applications are programmed in Java, which meets the procurement
requirements regarding the widespread nature and sustainability of the
programming language.

Project Dependencies
---------------------

The third-party components used in the project together with their motivated
usage need are listed in the tables below. There are separate tables for
packaging and operating the framework as well as developing and testing the
framework.

All components used in the collector service have an open source code.

.. tabularcolumns:: |p{0.4\linewidth}|p{0.1\linewidth}|p{0.25\linewidth}|p{0.25\linewidth}|
.. list-table::
   Third-party components used for the work of the IVXV framework
   :header-rows: 1

   *  - Name
      - Version
      - License
      - Usage need

   *  - `Bootstrap <http://getbootstrap.com>`_
      - 3.3.7
      - MIT
      - Design of the user interface of the collector service administration service

   *  - Bouncy Castle
      - 1.58
      - MIT
      - ASN1 handling, support functions of the data type BigInteger

   *  - `Bottle <https://bottlepy.org/>`_
      - 0.12.13
      - MIT
      - Framework for executing the collector service administration service web interface

   *  - CAL10N
      - 0.7.7
      - MIT
      - Multilanguage support, translation file validation

   *  - Digidoc 4j
      - 2.1.0.
      - LGPL
      - BDoc container handling

   *  - Digidoc 4j DSS
      - 5.2.d4j.3
      - LGPL
      - Digidoc 4j dependency

   *  - Apache Commons (cli 1.4, codec 1.10, collections 4.1, io 2.5, lang 3.6, logging 1.2, compress 1.3)
      - -
      - Apache License v2.0
      - Digidoc 4j and PDFBox dependencies

   *  - Apache HttpComponents
      - 4.5.3
      - Apache License v2.0
      - Digidoc 4j dependency

   *  - Apache Santuario
      - 2.0.9
      - Apache License v2.0
      - Digidoc 4j dependency

   *  - Google Guava
      - 20.0
      - Apache License v2.0
      - Digidoc 4j dependency

   *  - JDigiDoc
      - 3.12.1
      - LGPL
      - Digidoc 4j dependency

   *  - StaX
      - 1.0-2
      - Apache License v2.0
      - Digidoc 4j dependency

   *  - log4j
      - 1.2.6
      - Apache License 2.0
      - Digidoc 4j dependency

   *  - Woodstox
      - 4.4.1
      - Apache License 2.0
      - Digidoc 4j dependency

   *  - Xalan-Java
      - 2.7.2
      - Apache License 2.0
      - Digidoc 4j dependency

   *  - Xml Apis
      - 1.3.04
      - Apache License 2.0
      - Digidoc 4j dependency

   *  - `Docopt <http://docopt.org/>`_
      - 0.6.2
      - MIT
      - Execution of the command line interface of the collector service administration utilities

   *  - `etcd <https://coreos.com/etcd>`_
      - 3.2.17
      - Apache License v2.0
      - Distributed key value database used as a storage service

   *  - `github.com/ghodss/yaml <https://github.com/ghodss/yaml>`_
      - 73d445a
      - MIT
      - etcd client library dependency

   *  - `gopkg.in/yaml.v2 <https://gopkg.in/yaml.v2>`_
      - 4c78c97
      - Apache License v2.0
      - github.com/ghodss/yaml dependency

   *  - `github.com/golang/protobuf <https://github.com/golang/protobuf>`_
      - 224aaba
      - BSD 2.0
      - etcd client library dependency

   *  - `github.com/grpc-ecosystem/go-grpc-prometheus <https://github.com/grpc-ecosystem/go-grpc-prometheus>`_
      - 6b7015e
      - Apache License v2.0
      - etcd client library dependency

   *  - `github.com/grpc-ecosystem/grpc-gateway <https://github.com/grpc-ecosystem/grpc-gateway>`_
      - 6863684
      - BSD 2.0
      - etcd client library dependency

   *  - `google.golang.org/grpc <https://google.golang.org/grpc>`_
      - 1.0.4
      - Apache License v2.0
      - etcd client library dependency

   *  - `golang.org/x/net <https://golang.org/x/net>`_
      - f249948
      - BSD 2.0
      - etcd client library dependency

   *  - `Prometheuse klientteek <https://prometheus.io>`_
      - 0.8
      - Apache License v2.0
      - etcd client library dependency

   *  - `github.com/beorn7/perks/quantile <https://github.com/beorn7/perks>`_
      - 4c0e845
      - MIT
      - Prometheus client library dependency

   *  - `github.com/matttproud/golang_protobuf_extensions <https://github.com/matttproud/golang_protobuf_extensions>`_
      - 1.0.0
      - Apache License v2.0
      - Prometheus client library dependency

   *  - Gradle
      - 3.0
      - Apache License v2.0
      - Java applications build tool

   *  - `HAProxy <http://www.haproxy.org/>`_
      - 1.8.8
      - GPL v2
      - TCP proxy used as a proxy service

   *  - IvyPot
      - 0.4
      - Apache License v2.0
      - A Gradle build tool extension for managing dependencies and building applications offline

   *  - Jackson
      - 2.8.9
      - Apache License v2.0
      - Reading and writing JSON files

   *  - `jQuery <https://jquery.org/>`_
      - 3.1.0
      - MIT
      - User interface of the collector service administration service

   *  - Logback
      - 1.2.3
      - Eclipse Public License v1.0 or LGPL v2.1
      - Logging API SLF4J implementation

   *  - Logback JSON
      - 0.1.5
      - Eclipse Public License v1.0 or LGPL v2.1
      - Logback logger extension for compiling log entries in the JSON schema format using the Jackson library

   *  - `metisMenu <https://github.com/onokumus/metisMenu>`_
      - 1.1.3
      - MIT
      - User interface of the collector service administration service

   *  - PDFBox
      - 2.0.8
      - Apache License v2.0
      - PDF report generation support for Java applications

   *  - `PyYAML <http://pyyaml.org/>`_
      - 3.12
      - MIT
      - Collector service configuration files’ processing support for the administration service

   *  - `Schematics <https://github.com/schematics/schematics>`_
      - 2.0.1
      - BSD
      - Collector service configuration files’ validation support for the administration service

   *  - SLF4J
      - 1.7.25
      - MIT
      - Standard logging API

   *  - SnakeYAML
      - 1.18
      - Apache License v2.0
      - Reading data in the YAML format

   *  - `SB Admin 2 <https://github.com/BlackrockDigital/startbootstrap-sb-admin-2>`_
      - 3.3.7+1
      - MIT
      - Design of the user interface of the collector service administration service

.. list-table::
   Third-party components used by the IVXV framework tests
   :header-rows: 1

   *  - Name
      - Version
      - License
      - Usage need

   *  - Hamcrest
      - 1.3
      - BSD
      - A more readable use of assert-methods in Java unit tests

   *  - JUnit
      - 4.12
      - Eclipse Public License v1.0
      - Java testing framework

   *  - JUnitParams
      - 1.1.0
      - Apache License v2.0
      - Test parameterization support

   *  - Mockito
      - 2.10.0
      - MIT
      - Support for mocking the dependencies of the code being tested

   *  - Byte Buddy
      - 1.6.14
      - Apache License v2.0
      - Mockito dependency

   *  - Objenesis
      - 2.5
      - Apache License v2.0
      - Mockito dependency

   *  - libdigidoc2
      - 3.10.4.1218
      - LGPL
      - Generating test data

   *  - libdigidocpp-tools
      - 3.13.6.1371
      - LGPL
      - Generating test data

.. list-table::
   Third-party tools used to develop and/or test the IVXV framework
   :header-rows: 1

   *  - Name
      - Version
      - License
      - Usage need

   *  - `Behave <https://github.com/behave/behave>`_
      - 1.2.6
      - BSD
      - Regression test driver (*Behavior-driven development*)

   *  - `Docker <http://www.docker.com/>`_
      - 18.06 (or newer)
      - Apache License 2.0
      - Environment for conducting regression tests – software containers

   *  - `Docker Compose <http://www.docker.com/>`_
      - 1.22.0
      - Apache License 2.0
      - Environment for conducting regression tests – software container management

   *  - `Sphinx <http://www.sphinx-doc.org/>`_
      - 1.8.2
      - BSD
      - Environment for document generation
