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
  The tool also verifies whether the tally has been computed correctly and
  whether the files resulting from the decryption are consistent with one
  another.
  It outputs ciphertexts for which verification of the decryption proof failed.
* *integrity* - tool for verifying the logs of the processing application. The
  tool uses the logs output by the processing application as well as the
  original and anonymised ballot boxes to verify the audit trail of the logs.

Building
--------

IVXV java applications have 2 levels of build systems:

* *make* - the build system facade. Must be installed on the user's machine.
* *gradle* - the implementation of the build system. Gradle is located under
  ``common/external/gradle-8.11``, with the executable ``bin/gradle(.bat)``.

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

As the command-line arguments have to be defined precisely for correct
operation, we strongly recommend using a parameters file. Refer to the
configuration preparation documentation for the example configurations.

* Verify the correctness of file format conversion::

    auditor convert --conf app-conf.bdoc --params auditor-app-conf.bdoc

* Verify the correcntess of the shuffle::

    auditor mixer --conf app-conf.bdoc --params auditor-app-conf.bdoc

* Verify the correctness of decryption::

    auditor decrypt --conf app-conf.bdoc --params auditor-app-conf.bdoc

* Verify the integrity of processing logs::

    auditor integrity --conf app-conf.bdoc --params auditor-app-conf.bdoc

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
    proofs: decout/proof
    pub: initout/pub.pem
    discarded: decout/invalid
    anon_bb: bb-4.json
    plain_bb: decout/TESTQUESTION.plain
    tally: decout/TESTQUESTION.tally
    candidates: choices.bdoc
    districts: districts.bdoc
    out: auditout/
    invalidity_proofs: decout/proof-invalid

  integrity:
    ballotbox: votes.zip
    anon_bb: bb-4.json
    log_accepted: out-1/TESTQUESTION.check.log1
    log_squashed: out-2/TESTQUESTION.squash.log2
    log_revoked: out-3/TESTQUESTION.revoke.log2
    log_anonymised: out-4/TESTQUESTION.anonymize.log3
    bb_errors: out-1/ballotbox_errors.txt
