================================
 IVXV Internet voting framework
================================
-----------------------
 Processor application
-----------------------

Processor is a command line application for validating and post-processing ballot box in off-line mode.

Processor functionality is provided by the tools described below. The tools should be executed in the given order, each tool providing input for the next. The tools:

* *check* - validates the ballot box and ballot signatures and compares it to the provided voters lists, voting districts lists and registration data.
* *squash* - removes all but the latest ballot per each voter.
* *revoke* - revokes voter ballots according to the provided revocation lists.
* *anonymize* - removes voters data from the ballot box and provides anonymized ballot box ready for the *key application*.

Building
--------

IVXV java applications have 2 levels of build systems:

* *make* - the build system facade. Must be installed on the user's machine.
* *gradle* - the implementation of the build system. Gradle is located under ``common/external/gradle-3.0``, with the executable ``bin/gradle(.bat)``.

Building:

* ``make`` or
* ``make all`` or
* ``gradle build installDist`` - build and test the application.
* ``make clean`` or
* ``gradle clean`` - clean build resources, i.e. the directory ``build``.

Application executable is ``build/install/processor/bin/processor(.bat)``.

Distributable application packages are provided under ``build/distributions/``.

Executing
---------

Usage:

* ``processor -h`` - show application help
* ``processor <tool> -h`` - show the tool help
* ``processor <tool> --certs certs.bdoc --conf processor.yaml.bdoc`` - execute a tool with the *--certs* and *--conf* arguments
* ``processor check --certs certs.bdoc --conf processor.yaml.bdoc``
* ``processor squash --certs certs.bdoc --conf processor.yaml.bdoc``
* ``PROCESSOR_OPTS=-Xmx8G processor squash --certs certs.bdoc --conf processor.yaml.bdoc`` - the same as above, but increase maximal memory usage to 8GB.

The ``certs.bdoc`` container is expected to contain all CA, OCSP and TSA certificates required for validating any signed input (including the ``certs.bdoc`` itself) and the file ``ivxv.properties`` that refers to the certificates. Example of ``ivxv.properties``:

::

  ca = cert1.pem, cert2.pem
  ocsp = ocsp.pem
  tsa = tsa.pem

Example of ``processor.yaml``:

.. code-block:: yaml

  check:
    ballotbox: votes.zip
    districts: TESTKOV2017.districts.json
    registrationlist: ocsp.zip
    signingkey: voterfile.pub.key
    voterlists:
      -
        path: 00.TESTKOV2017.gen.voters
        signature: 00.TESTKOV2017.gen.voters.signature
      -
        path: 03.TESTKOV2017.gen.voters
        signature: 03.TESTKOV2017.gen.voters.signature
      -
        path: 06.TESTKOV2017.gen.voters
        signature: 06.TESTKOV2017.gen.voters.signature
      -
        path: 09.TESTKOV2017.gen.voters
        signature: 09.TESTKOV2017.gen.voters.signature
    out: out-1

  squash:
    ballotbox: out-1/bb-1.json
    out: out-2

  revoke:
    ballotbox: out-2/bb-2.json
    revocationlists:
      - 12.TESTKOV2017.gen.revoke.json
      - 13.TESTKOV2017.gen.revoke.json
      - 14.TESTKOV2017.gen.revoke.json
      - 15.TESTKOV2017.gen.revoke.json
    out: out-3

  anonymize:
    ballotbox: out-3/bb-3.json
    out: out-4

