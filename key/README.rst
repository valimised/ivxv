================================
 IVXV Internet voting framework
================================
-----------------
 Key application
-----------------

Key application is a command line application for generating key parameters,
generating private keys and decrypting ballots.

The application functionality is provided by the tools described below:

* *util* - utility functions such as listing smart card readers and inserted
  smart cards or testing correct functionality of key shares.
* *groupgen* - generate group parameters suitable for generating a key pair.
* *init* - generate the decryption and signing key in a distributed manner on
  smart cards, construct the corresponding public keys and output the public
  keys and certificate corresponding to the encryption key.
* *decrypt* - take the anonymized ballot box from the *processor*, construct the
  decryption key from the shares on the smart cards, decrypt the encrypted
  ballots and provide proofs of correct decryption.

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

Application executable is ``build/install/key/bin/key(.bat)``.

Distributable application packages are provided under ``build/distributions/``.

Randomness sources
------------------

During key parameter generation and key initialization, random bytes are
required. The IVXV framework provides several interfaces for obtaining the
random bytes:

* *file* - randomness source is a file with finite length.
* *stream* - randomness source is a stream with infinite length. For example, on
  Linux, `/dev/urandom` is one such stream.
* *DPRNG* - deterministic pseudo-random number generator uses the seed and a
  hash function to stretch the seed into stream of bytes. A digest of input file
  is computed and the digest is used as seed.
* *system* - system's native random source is used. On Windows, this is provided
  by CryptoAPI and on Linux, this is provided by getrandom().
* *user* - external entropy source is used. A program is executed which should
  listen on the port provided as a command line argument. It is sent big-endian
  integer denoting the amount of bytes to be read and it should reply with
  exactly that amount of entropy. In practice, this application should read user
  source (mouse movements, keyboard interaction). The path should point to
  application location which should be run.

Furthermore, the entropy source inputs are conditioned using SHAKE-256
extendable digest function. We refer to documentation of entropy mixing for
further details.

The tools can take several randomness sources and in this case, the order of
inputs matters. There can be several inputs of same type.

For input types *file*, *stream* and *DPRNG*, a required argument *path* must be
defined.

Key parameters
--------------

The group generation tool can generate the group parameters for integer groups
and elliptic groups and the initialization tool can use the corresponding
groups. The following types of groups are implemented:

* group of integers modulo a prime. The group generation tool generates a safe
  prime, i.e. a prime p such that p = 2*q+1, where q is also a prime, and a
  generator g. For such a choice of parameters, the multiplicative subgroup with
  generator g has order q. The group generation tool takes the value `mod` for
  argument `--paramtype`. The argument `--length` defines the bit-length of the
  value p.

  In this case, the subargument `--mod` with the corresponding values for the
  argument `--paramtype` has to be set in the initialization tool.
* elliptic curve group. The group generation tool takes an elliptic curve from a
  list of predefined curves. The group generation tool takes the value `ec` for
  argument `--paramtype`. The argument `--length` defines the bit-length of the
  underlying field.

  In this case, the subargument `--ec` with the corresponding elliptic curve
  group name has to be defined in the initialization tool.

Key initialization and decryption protocols
-------------------------------------------

Key could be initialized using different protocols. Currently, only a protocol
due to Desmedt has been implemented.

Furthermore, decryption could be performed using different protocol. Currently,
only a protocol which reconstructs a key in memory has been implemented.

A description of protocol interfaces, protocol linking and descriptions of
implemented protocols is given in documentation.

Sample executions
-----------------

As the command-line arguments have to be defined precisely for correct
operation, we strongly recommend using a parameters file. To encourage the
approach, the examples use the stored parameters file from the repository.

* Display the card readers and inserted cards::

    key util --conf ../examples/app-conf.bdoc --listreaders

  The application should display::

    ID | NIMI                                       | KAARDIGA
    0  | Gemalto Ezio Shield (S1370135510111) 00 00 | Ei

* Generate group parameters for integers modulo a prime for prime with
  bit-length 2048 bits::

    key groupgen --conf ../examples/app-conf.bdoc --params \
    ../examples/key-app-conf-3-parties.bdoc

  The application should display (the actual values differ)::

    Zp(p=7714308592189717986218840932759665300259388772179838475553570421760647588928648139668265831789568076388443118361277912418352218030834963525256914386555943, q=3857154296094858993109420466379832650129694386089919237776785210880323794464324069834132915894784038194221559180638956209176109015417481762628457193277971, g=1320131981645001575383816722884941834282340771961541298839439470294413867893836077812664006160241399746686856405124801824622449176630605416180422631415563)

* Initialize key with an group of integers modulo a prime, with the key shared
  between three shareholders with a threshold of two shareholders required for
  reconstruction::

    key init --conf ../examples/app-conf.bdoc --params \
    ../examples/key-app-conf-3-parties.bdoc

  The key initialization tool is interactive. Two scenarios are supported:

  1. **lazy**-initialization - the number of smart card readers is less than the
     number of shareholders (smart cards).
  2. **fast**-initialization - the number of smart card readers is at least as
     the number of shareholders (smart cards).

  The application asks the user which smart card will be inserted into which
  smart card reader. Then, the application asks the user to insert the
  corresponding smart card to the corresponding smart card reader. If PIN-code
  for storing and accessing the blobs on smart card is required, then it is
  asked from the user.
  
  In case of lazy-initialization, if card change is needed in the reader, the
  application instructs the user to remove the smart card from the smart card
  reader and insert another card. The PIN-codes are not stored between
  insertions, so the user has to insert PIN-code for every smart card insertion.

  After initializing the key, a test decryption is performed to verify the
  correct operation of smart cards. This test can be omitted by setting
  `skiptest` argument during initialization and running utility tool later.

  Four files are output into the output folder (in the example parameters case,
  `initoutmod/` folder): ElGamal encryption key in DER encoding, ElGamal
  encryption key in PEM encoding, self-signed certificate for RSA verification
  key and certificate for the ElGamal encryption key signed by the corresponding
  RSA signing key.

* Test the keyshares by encrypting a randomly generated ciphertext and decoding
  it using a keyshares::

    key util --conf ../examples/app-conf.bdoc --params \
    ../examples/key-app-conf-3-parties.bdoc

  Similarly to the initialization tool, the utility tool asks interactively,
  which smart card to insert to which smart card reader.

* Decrypt the anonymized ballot box::

    key decrypt --conf ../examples/app-conf.bdoc --params \
    ../examples/key-app-conf-3-parties.bdoc

  Similarly to the initialization tool, the decryption tool asks interactively,
  which smart card is inserted to which smart card reader.

  The tool creates four files in the output directory (`decout/` in our case):
  tally, signature of the tally (verifiable using the verification key output
  in previous step), proof of correct decryption and list of invalid ballots.

Sample configuration
--------------------

.. code-block:: yaml

  util:
    listreaders: true
    testkey:
      out: initout
      threshold: 2
      parties: 3

  groupgen:
    paramtype: mod
    length: 2048
    random_source:
    - random_source_type: file
      random_source_path: key-app-conf-3-parties.bdoc
    - random_source_type: system
    - random_source_type: DPRNG
      random_source_path: key-app-conf-3-parties.bdoc

  init:
    identifier: TEST ELECTION
    paramtype:
      mod:
        p: 31006623739572055594316376496810078655821270818754776758101749724523127678743816496005396628221427409620979185538693981332233517023057755223387686225836450702330899860650955137758634321834579766029698693186809115755521515897667860071986274175450351763184797360034660544443454913687056349390882016990934446816120149080730628207368885887225431671992925980886511322581579324848309495164952603484101370988563607786382332191775515738664908046528895919529565837085370042158593639745744812919291872615090822187022072663753788349131123599938206656494617028004229197148296470034927606948396845266919660981060450815399206881783
        g: 20059753780648374919602788861160364382456464965482375582965295943301854939416182323173140858574092810792118344852095286738382046691448892104316682098008314350665701027152563473740684121450331097788027653033348674923482123773625083470304297639241714585564590789474492899516890105701623740106839386638356003242965179652567872352520346502216110134705958819357550844390509563395644009912704419231922712876118179968831219832352718118433349430475939433575982671204894741140789395862949917210862142735116719181778539869412092203546842680413734885218779101834458151531152658413650017524889336760758622216128148652402856928435
    out: initout
    skiptest: true
    signaturekeylen: 2048
    issuercn: TEST
    signcn: SIGNATURE
    signsn: 1
    enccn: ENCRYPTION
    encsn: 2
    required_randomness: 128
    genprotocol:
      desmedt:
        threshold: 2
        parties: 3
    random_source:
      - random_source_type: file
        random_source_path: key-app-conf-3-parties.bdoc
      - random_source_type: system
      - random_source_type: DPRNG
        random_source_path: key-app-conf-3-parties.bdoc

  decrypt:
    identifier: TEST ELECTION
    protocol:
      recover:
        threshold: 2
        parties: 3
    anonballotbox: bb.json
    anonballotbox_checksum: bb.json.sha256sum.bdoc
    candidates: choices.bdoc
    districts: districts.bdoc
    provable: true
    out: decout

Notes
-----

Java smartcard library
~~~~~~~~~~~~~~~~~~~~~~

Running this application might require setting a JVM parameter
`sun.security.smartcardio.library` for proper smart card terminal functionality.
Notably required when using Oracle JDK. Run the JVM with argument::

    -Dsun.security.smartcardio.library=/path/to/libpcsclite

On Ubuntu 16.04 LTS the path for the libpcslite is
`/lib/x86_64-linux-gnu/libpcsclite.so.1`

Smart card support
~~~~~~~~~~~~~~~~~~

The key application has been tested with Aventra ActiveSecurity MyEID (ATR 3B F5
18 00 00 81 31 FE 45 4D 79 45 49 44 9A). The smart card can be obtained from
`Aventra <http://www.aventra.fi/webshop/?route=product/product&product_id=53>`.

Smart card initialization
~~~~~~~~~~~~~~~~~~~~~~~~~

The key application depends on the existence of fresh PKCS#15 file system on
smart card. An example tool `erase_and_install.sh` is provided in the
`tools/pkcs15_initialization/` directory for erasing the smart card and
installing the file system.

Interfering programs
~~~~~~~~~~~~~~~~~~~~

As the tested smart card supports only a single session, then other programs may
interfere with smart card communication. Check and verify that no other program
is running and using the smart card.

Lazy initialization
~~~~~~~~~~~~~~~~~~~

Even though lazy initialization is supported, different unrecoverable errors can
be encountered. In this case, the initialization tool has to be run with freshly
formatted cards.

Some examples of reasons which can cause unrecoverable errors:

* Windows Smart Card Service card autodetect card probing
* incomplete insertion of smart card
* insertion of wrong smart card into terminal

Platform support
~~~~~~~~~~~~~~~~

The key application has been tested on Linux and Windows.

* On Linux, `opensc` is required.
* On Windows, follow the installation procedure

.. todo:: add the installation procedure

Smart card reader support
~~~~~~~~~~~~~~~~~~~~~~~~~

.. todo:: add list of supported smart card readers
