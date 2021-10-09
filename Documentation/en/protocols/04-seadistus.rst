..  IVXV protocols

================================================================================
Election Definition
================================================================================

The organizer of the election defines the election. At Estonian national
elections, all people who have the right to vote are divided into one or several
electoral districts.  A voter who belongs to a specific electoral district can
only choose between the candidates of that district.

To define an election, at least the following have to be defined:

#. The unique identifiers of the election and the unique identifiers of the
   questions
#. A complete list of electoral districts and polling divisions
#. A list of people with the right to vote and the electoral districts to which
   they belong
#. A list of candidates and the electoral districts to which they belong

Election and Question Identifiers
================================================================================

One data set relating to the elections is integrated using a unique election
identifier.  Typically, people vote on one specific question at one election.
However, it is possible to pose several questions at an election. All questions
are separated using a unique identifier.

The length of identifiers is limited to 28 characters from the ASCII code table.
The identifiers to be used for a specific election are specified each time in
the election settings.  The applications implementing the specification have to
refuse to process data that identifies an election/question that is not included
in the list of elections/questions set out for the application.

.. code-block:: bnf

    election-identifier := 1*28CHAR

    question-identifier := 1*28CHAR


List of Electoral Districts and Polling Stations
================================================================================

Candidates can be set up for elections only in specific electoral districts.
Districts are used to give voters voting choices:

#. Each voter belongs to a pre-determined district

#. In all polling stations of one district, voters can only choose between the
   candidates of this district

At Estonian national elections, the following are distinguished between: local
government elections, Riigikogu (Estonian parliament) elections, European
Parliament elections, and referendums.

Local government elections are organized pursuant to the Local Government
Council Election Act.  The elections are organized at the local government level
and each local government has its own voting result. The electoral districts are
specified at the local government level pursuant to the rules specified in the
election act.

Riigikogu elections are organized pursuant to the Riigikogu Election Act.  The
elections are organized at state level. The state is divided into 12 electoral
districts. The voting result is determined for every district.

European Parliament elections are organized pursuant to the European Parliament
Election Act.  The elections are organized at state level and the voting result
is the same for all local governments. The entire country is one
electoral district.

Referendums are organized pursuant to the Referendum Act. The elections are
organized at state level and the voting result is the same for all local
governments. The entire country is one electoral district.

Various elections are not different on the basis of online voting data forms
and procedures.  Various district distributions are handled by the Elections
Infosystem.

Candidates can be set up for elections only in a specific electoral district.
Voters are divided between districts. The voter can only choose between the
candidates running in their district.

In local government council elections, voting happens at the level of Estonian
local governments (parishes, cities), and thus the classification of `Estonian
administrative units and settlements (EHAK) <http://metaweb.stat.ee/>`_ is used
in the online voting protocol suite to specify electoral districts
and to show to which district voters and choices belong.

For example:

* The EHAK code for the Pirita city district in Tallinn City is 0596
* The EHAK code for Anija Parish is 0141.

As agreed, for state-level elections, the EHAK code for the district is 0.

At Riigikogu and European Parliament elections and referendums, a fictitious
unit is set up in each electoral district for voters who is living abroad. The
polling station number for those voters is 0 and the EHAK code is 0000.

.. code-block:: bnf

   ehak-code = 1*10DIGIT

   ehak-district = ehak-code
   no-district = 1*10DIGIT

   district = ehak-district '.' no-district

   station = ehak-station '.' no-station
   station-legacy = ehak-station TAB no-station TAB district-legacy


The JSON schema of the electoral district list is defined as follows. The
elements of the object :token:`region_dict` are indexed with the element type
:token:`ehak-code`. The elements of the object :token:`district_dict` are
indexed with the element type :token:`district`.

.. literalinclude:: ../../common/schema/ivxv.districts.schema
   :language: json
   :linenos:

Example:

.. literalinclude:: ../../common/schema/ivxv.districts.schema.example
   :language: json

The list of electoral districts is received from the Election Infosystem and the
JSON file is delivered to the online voting system as a digitally signed
ASICE file.


List of Voters
================================================================================

The voter list includes the eligible voters’ names and personal identification codes,
their district number. The voter list is uploaded to the system in the following format:

.. code-block:: bnf

    voter-personalcode = 11DIGIT
    voter-name = 1*100UTF-8-CHAR
    action = "lisamine" | "kustutamine"
    adminunit-code = 1*4UTF-8-CHAR | "FOREIGN"
    electoral-district = adminunit-code TAB no-district
    reason = "" | "tõkend" | "valimisringkonna vahetus" | "muu"
    version-no = "2"
    list-type = "algne" | "muudatused"

    voter = voter-personalcode TAB voter-name TAB action TAB electoral-district TAB reason LF

    voter-list = version-no LF election-identifier LF list-type LF *voter

The legacy systems’ data structures include the field version number, whose
length is limited to 2 characters. The value of the field is 2.

The voter list can either be original or amended. The original list only allows
adding voters; the amended list also allows removing voters from the list. The
voter entry can also include additional information – a reason to be in the specific
amended list.


The data contain the following:

#. The type (``list-type``) "``algne``" means the original large list uploaded to the system
   before e-voting starts, and "``muudatused``" means the cumulative updates made later

#. The action (``action``) "``lisamine``" means adding a new voter
   and "``kustutamine``" means removing. When a voter moves from one unit
   (administrative division or electoral district) to
   another, then one deleting entry is made in the amendments of the voter list
   to delete the voter from their previous unit, and one adding entry
   is made to add the voter to the voter list in their new unit. In
   the original list, all entries are the "``lisamine``" type

#. The electoral district (``electoral-district``) identifies the
   administrative division and electoral district where the voter votes.
   Value is EHAK code for municipalities and for districts of Tallinn; or
   "0000" for voter who lives permanently in the foreign country.

#. The reason (``reason``) is used in deleting entries to note the reason for the deletion.
   The reason field has to be empty for adding entries. If the reason is
   ``tõkend``, this means that from the moment the amendment is implemented, the
   voter with this personal identification code will not be allowed to vote
   anymore. If the reason is a ``valimisringkonna vahetus``, it means that the voter is
   deleted from one electoral district, because they are added to another. In this
   case, a deleting entry has to be accompanied by an adding entry (this is
   checked). If the voter is removed from the list for some other reason (death,
   moving to a district that is not part of the elections), the reason has to be
   ``muu`` or can be left empty. This field is informative.


Signing the Voter List
----------------------------

The voter list is retrieved from the Population Register, which is run by the IT
and Development Centre of the Ministry of the Interior. The legacy format text
file comes with a signature file made by the Population Register by taking the
SHA256 hash from the original voter list and signing that hash with a 2048-bit
RSA key. The public key generated by the Population Register is made available
to the online voting information system, and this key is used to check the
integrity of the voter lists.  The schema has been used since the 2015 Riigikogu
elections.

..
   Key generation:
   openssl genrsa -out voterlist-signer.key 2048
   Public key extraction:
   openssl rsa -in voterlist-signer.key -pubout -out voterlist-signer.pem
   Digital signing:
   openssl dgst -sha256 -sign voterlist-signer.key -out voters.signature voters
   Verification:
   openssl dgst -sha256 -verify voterlist-signer.pem -signature voters.signature voters

Version of the Applied List
-----------------------------

The applied voter list at a certain point in time depends on the original list
and the amendments made – what kind of amendments have been made and in which
order. So as to identify this situation unambiguously, a list version has to be
calculated.

**NB!** This version is not related to the version number in the list file that
determines the list format version.

The version is calculated as follows::

    v_0 = ""
    v_n = base64(sha256(v_{n-1} | base64(sha256(nk_n))))

where ``nk_n`` is the list loaded in ``n`` (counting starts from one, i.e. the
original list is  ``nk_1``), ``v_n`` is the voter list version after it has been
loaded, ``""`` is an empty string and ``|`` is the string connection operation.

Records about the implemented list version are kept by the collector service and
the processing application, which guarantee that the specific vote is counted in
the right district.

List of Choices
================================================================================

The list of choices includes data on the candidates (at elections) or answers
(at referendums).  At elections, the list includes not only the candidate’s data
but also the name of their political party.

There are two systemic differences at elections that are visible to voters
during online voting:

#. At referendums, voters do not choose between the candidates of political
   parties, but answer "yes" or "no" to the referendum question;

#. At Riigikogu, local government and European Parliament elections, voters vote
   for one candidate, who may or may not belong to a political party/list.

The protocol suite encodes the voter’s possible choices in the district as a
numerical value of up to 11 characters, which is encoded in the list of choices
with the EHAK code of the district.  Only the choices of their electoral
district can be available to a voter. The voter application has to ensure this
function, and the application calculating the voting result has to check it.

.. code-block:: bnf

    choice-no = 1*11DIGIT
    district-choice = ehak-district '.' choice-no


The JSON schema of the list of choices has been defined as follows. The elements
of the object :token:`district_dict` are indexed with the :token:`district` type
element.  The elements of the object :token:`list-choices` are indexed with the
:token:`district-choice` type element.

.. literalinclude:: ../../common/schema/ivxv.choices.schema
   :language: json
   :linenos:

Example:

.. literalinclude:: ../../common/schema/ivxv.choices.schema.example
   :language: json
