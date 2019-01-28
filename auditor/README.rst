================================
 IVXV Internet voting framework
================================
---------------------
 Auditor application
---------------------

The auditor application is an application for verifying the correctness of
operations of other tools.

The application functionality is provided by the tools described below:

* *convert* - tool for verifying that the proof of shuffle provided by
  Verificatum corresponds to the anonymized ballot boxes. The input ballot box
  is the anonymized ballot box output by the processor tool and the output
  ballot box is the anonymized ballot box output by the Verificatum prover
  application. The provided public key must be output by key generation
  tool.
* *mixer* - tool for verifying the correctness of a proof of shuffle. The tool
  takes as input the protocol information file provided to Verificatum during
  shuffle and the proof directory output by it.
* *decrypt* - tool for verifying the correctness of decryption. The tool uses
  public key and the proof of correct decryption output by the decryption tool.
  It outputs ciphertexts for which verification of the decryption proof failed.

Building
--------

IVXV java applications have 2 levels of build systems:

* *make* - the build system facade. Must be installed on the user's machine.
* *gradle* - the implementation of the build system. Gradle is located under
  ``common/external/gradle-3.0``, with the executable ``bin/gradle(.bat)``.

Building:

* ``make`` or
* ``make all`` or
* ``gradle build installDist`` - build and test the application.
* ``make clean`` or
* ``gradle clean`` - clean build resources, i.e. the directory ``build``.

Application executable is ``build/install/auditor/bin/auditor(.bat)``.

Distributable application packages are provided under ``build/distributions/``.

Sample executions
-----------------

An example configuration is provided in the examples directory. The example
configuration uses unit-test data which is not available in distribution
releases.

* Verify the correctness of file format conversion::

    auditor convert --conf ../examples/app-conf.bdoc --params
    ../examples/auditor-app-conf.bdoc

* Verify the correcntess of the shuffle::

    auditor mixer --conf ../examples/app-conf.bdoc --params
    ../examples/auditor-app-conf.bdoc

* Verify the correctness of decryption::

    auditor mixer --conf ../examples/app-conf.bdoc --params
    ../examples/auditor-app-conf.bdoc

Sample configuration
--------------------

.. code-block:: yaml

  convert:
    input_bb: bb-4.json
    output_bb: bb-5.json
    pub: initout/pub.pem
    protinfo: mixnet/ProtocolInformation.xml
    proofdir: mixnet/

  mixer:
    protinfo: mixnet/ProtocolInformation.xml
    proofdir: mixnet/
    threaded: true

  decrypt:
    input: decout/proof
    pub: initout/pub.pem
    out: auditout/
