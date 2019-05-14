..  IVXV protocols

=========================
Communications Protocols
=========================

Interface
-----------

The microservices of the collector service directed at the voter communicate
with the voter application and the verification application via the JSON-RPC
protocol.

:id: JSON-RPC information request identifier
:method: RPC method
:params: Parameters of a specific RPC method

.. literalinclude:: ../../common/examples/json.rpc.method.query.json
   :language: json
   :linenos:

:error: Possible error code or ``null`` when there is no error
:id: JSON-RPC information request identifier, has to match the id used in the request
:result: Method-based response data structure

.. literalinclude:: ../../common/examples/json.rpc.method.response.json
   :language: json
   :linenos:

During the first information request exchange with an IVXV microservice, a
HEX-coded unique session identifier  (``result.SessionID``), is issued to the
communicating application; this identifier will henceforth be used by the
application in all requests sent to the collector service
(``params.SessionID``). The session identifier is used to integrate various RPC
requests related to voting into one session. This integration is informative and
its purpose is to simplify log analysis; decisions related to the electoral
district to which the vote belongs and/or other essential
aspects are made on the basis of digitally signed data.

TLS is used as the transport protocol. The encrypted channel is terminated in a
specific microservice.  To enable sharing the load and implementing
microservices flexibly, a TLS SNI extension is used that
allows the proxy services to direct TLS into the correct microservice instance
without terminating the flow.  The proxy service is typically available in port
443 of the external interface of the collector service.

Retrieving the List of Choices
---------------------------------

Retrieving the list of choices means that the voter application has to
communicate with the list service (SNI ``choices.ivxv.invalid``). Retrieving the
list of choices means that the voter has to be authenticated and their electoral
district has to be identified.

The voter application makes the request ``RPC.VoterChoices`` to retrieve the
lists.

:params.AuthMethod: The supported choices are the methods  ``tls`` and ``ticket``.
:params.OS: The operation system in which the voter application is used.

Request ``RPC.VoterChoices`` when authenticating with an ID card – authentication
happens at the TLS protocol level during the processing of the request using the
authentication certificate of the ID card.

.. literalinclude:: ../../common/examples/id.rpc.voterchoices.query.json
   :language: json
   :linenos:

Request ``RPC.VoterChoices`` when authenticating with mobile ID – before making
the request, the DigiDocService proxy service (SNI ``dds.ivxv.invalid``) has to
be used to retrieve a signed authentication certificate.

:params.AuthToken: Certificate signed using the authentication service,
                   which includes the voter’s unique identifier.
:params.SessionID: Since in the case of mobile ID, an interaction to retrieve
                   the authentication certificate has preceded list retrieval,
                   there exists a session identifier that needs to be used.


.. literalinclude:: ../../common/examples/mid.rpc.voterchoices.query.json
   :language: json
   :linenos:

Response of the list service to the request ``RPC.VoterChoices``.

:result.Choices: The voter’s district identifier  ``VoterDistrict``
:result.List: BASE64-encoded district choice list ``DistrictChoices``
:result.Voted: If the voter has already cast their vote, it is  ``true``; if
               not, it will not be displayed among the field responses.


.. literalinclude:: ../../common/examples/id.rpc.voterchoices.response.json
   :language: json
   :linenos:

Possible error codes to the request ``RPC.VoterChoices``.

:BAD_CERTIFICATE: An error in the voter’s personal identification certificate.
:BAD_REQUEST: There is an error in the request.
:INELIGIBLE_VOTER: The voter does not have the right to vote.
:INTERNAL_SERVER_ERROR: An error in the functioning of the internal server.
:UNAUTHENTICATED: An unauthenticated request.
:VOTER_TOO_YOUNG: The voter is too young to vote.
:VOTING_END: The voting period has ended.


Sending a Signed Vote to Storage
-------------------------------------------

Sending a signed vote to storage means that the voter application has to
communicate with the voting application (SNI ``voting.ivxv.invalid``).

The voter application makes the request ``RPC.Vote`` to send the signed vote to
be stored.

:params.AuthMethod: The supported choices are the methods ``tls`` and ``ticket``.
:params.Choices: The voter’s district identifier  ``VoterDistrict``, which was
                 valid when the list of choices was retrieved. Correct use of
                 the parameter allows the collector service to warn the voter if
                 their district has changed compared to the start of voting.
:params.OS: The operation system in which the voter application is used.
:params.Type: The format of a signed vote. At the moment, the only supported
              value is ``bdoc``.
:params.Vote: BASE64-encoded vote ``SignedVote`` in the format specified above.

Request ``RPC.Vote`` when authenticating with an ID card.

.. literalinclude:: ../../common/examples/id.rpc.vote.query.json
   :language: json
   :linenos:

Request ``RPC.Vote`` when authenticating with mobile ID.

.. literalinclude:: ../../common/examples/mid.rpc.vote.query.json
   :language: json
   :linenos:


The voting service’s response to the request  ``RPC.Vote``.

:result.Qualification.ocsp:
:result.Qualification.tspreg:
    Additional proof retrieved by the collector service to qualify and correctly
    register the vote ``SignedVote`` created by the voter application. The
    composition of the response depends on the specific settings of the
    collector service; in this case the standard OCSP protocol is used to check
    that the voter’s signature certificate is valid, and the PKIX timestamp
    protocol based registration service is used to both fix the time of casting
    the vote and registering the electronic vote in an external independent
    service. Both the OCSP response and the timestamp in the PKIX format with
    any additions necessary for the registration service are sent to the voter
    application to be checked.
:result.TestVote:
    If the vote was cast before voting started and was counted as a test vote,
    then ``true``; if not, this field will not be included in the response. In
    case of a test vote, the voter application will display a relevant warning
    to the voter.
:result.VoteID:
    The vote’s identifier in the storage service; based on that, the
    verification application can demand access to the vote for later analysis.

.. literalinclude:: ../../common/examples/id.rpc.vote.response.json
   :language: json
   :linenos:

Possible error codes in case of the request ``RPC.Vote``.

:BAD_CERTIFICATE: An error in the voter’s personal identification or signing
                  certificate.
:BAD_REQUEST: Error in the request.
:IDENTITY_MISMATCH: The personal identification codes in the personal
                    identification certificate and the signing certificate do
                    not match.
:INELIGIBLE_VOTER: The voter does not have the right to vote.
:INTERNAL_SERVER_ERROR: An error in the functioning of the internal server.
:OUTDATED_CHOICES: The district to which the voter belongs has changed since the
                   moment the list was retrieved.
:UNAUTHENTICATED: An unauthenticated request.
:VOTER_TOO_YOUNG: The voter is too young to vote.
:VOTING_END: The voting period has ended.


Voting with Mobile ID
------------------------

Using mobile ID as a means of signing and authentication means that a support
service (SNI ``dds.ivxv.invalid``) that integrates with the service DigiDocService
has to be used to retrieve an authentication certificate before the list of
choices is retrieved and to sign the vote before storing it.


Retrieving an Authentication Certificate
*****************************************

The voter application makes the request ``RPC.Authenticate`` to start mobile ID
authentication.

:params.OS: The operation system in which the voter application is used.
:params.PhoneNo: The mobile phone number of the person using the mobile ID.

.. literalinclude:: ../../common/examples/mid.rpc.authenticate.query.json
   :language: json
   :linenos:

:result.ChallengeID: The mobile ID verification code to be displayed in the
                     voter application.
:result.SessionCode: The mobile ID session identifier for further poll requests.


.. literalinclude:: ../../common/examples/mid.rpc.authenticate.response.json
   :language: json
   :linenos:

Possible error codes in case of the request ``RPC.Authenticate``.

:BAD_REQUEST: An error in the request.
:INTERNAL_SERVER_ERROR: An error in the functioning of the internal server.
:MID_BAD_CERTIFICATE: An error in the voter’s mobile ID personal identification
                      certificate.
:MID_NOT_USER: The phone number does not belong to the mobile ID client.
:VOTING_END: The voting period has ended.

The voter application makes the request ``RPC.AuthenticateStatus`` to assess the
status of the authentication process.

:params.OS: The operation system in which the voter application is used.
:params.SessionCode: Authentication session identifier.

.. literalinclude:: ../../common/examples/mid.rpc.authenticatestatus.query.json
   :language: json
   :linenos:


:result.AuthToken: The authentication certification to be presented to other
                   IVXV services, or ``null``, if the request is still being
                   processed.
:result.GivenName: The voter’s given name in case of successful authentication.
:result.PersonalCode: The voter’s personal identification code in case of
                      successful authentication.
:result.Status: Request status – ``POLL`` means the request has to be repeated, ``OK``
                means the authentication was successful. Other fields of the
                response only contain information if the value is
                ``OK``.
:result.Surname: The voter’s surname in case of successful authentication.


.. literalinclude:: ../../common/examples/mid.rpc.authenticatestatus.response.json
   :language: json
   :linenos:

.. literalinclude:: ../../common/examples/mid.rpc.authenticatestatus2.response.json
   :language: json
   :linenos:

Possible error codes in case of the request ``RPC.AuthenticateStatus``.

:BAD_REQUEST: An error in the request.
:INTERNAL_SERVER_ERROR: An error in the functioning of the internal server.
:MID_ABSENT: The voter’s mobile phone is not available.
:MID_CANCELED: The voter cancelled the mobile ID session.
:MID_EXPIRED: The mobile ID session has expired.
:MID_GENERAL: An error in the functioning of the mobile ID service.
:VOTING_END: The voting period has ended.


Signing a Vote
*********************

The voter application makes the request ``RPC.GetCertificate`` to get the
signing certificate.


:params.AuthMethod: Only the authentication method ``ticket`` is supported.
:params.AuthToken: Mobile ID authentication certificate.
:params.OS: The operation system in which the voter application is used.
:params.PhoneNo: The phone number of the person signing the vote.

.. literalinclude:: ../../common/examples/mid.rpc.getcertificate.query.json
   :language: json
   :linenos:


:result.Certificate: Signing certificate in the X509 format.

.. literalinclude:: ../../common/examples/mid.rpc.getcertificate.response.json
   :language: json
   :linenos:

Possible error codes in case of the request ``RPC.GetCertificate``.

:BAD_REQUEST: An error in the request.
:INTERNAL_SERVER_ERROR: An error in the functioning of the internal server.
:MID_BAD_CERTIFICATE: An error in the voter’s mobile ID signing certificate.
:MID_GENERAL: An error in the functioning of the mobile ID service.
:MID_NOT_USER: The phone number does not belong to the mobile ID client.
:VOTING_END: The voting period has ended.


The voter application makes the request ``RPC.Sign`` to initiate vote signing.

:params.AuthMethod: Only the authentication method ``ticket`` is supported.
:params.AuthToken: Mobile ID authentication certificate.
:params.Hash: BASE64-encoded electronic vote SHA-256 hash.
:params.OS: The operation system in which the voter application is used.
:params.PhoneNo: The phone number of the person signing the vote.

.. literalinclude:: ../../common/examples/mid.rpc.sign.query.json
   :language: json
   :linenos:

:result.ChallengeID: The mobile ID verification code to be displayed in the
                     voter application.
:result.SessionCode: The mobile ID session identifier for further poll requests.


.. literalinclude:: ../../common/examples/mid.rpc.sign.response.json
   :language: json
   :linenos:

Possible error codes in case of the request ``RPC.Sign``.

:BAD_REQUEST: An error in the request.
:INTERNAL_SERVER_ERROR: An error in the functioning of the internal server.
:MID_BAD_CERTIFICATE: An error in the voter’s mobile ID signing certificate.
:MID_NOT_USER: The phone number does not belong to the mobile ID client.
:VOTING_END: The voting period has ended.


The voter application makes the request ``RPC.SignStatus`` to assess the status
of the signing process.

:params.OS: The operation system in which the voter application is used.
:params.SessionCode: Mobile ID session identifier.

.. literalinclude:: ../../common/examples/mid.rpc.signstatus.query.json
   :language: json
   :linenos:

:result.Signature: If the response ``Status`` field is ``OK``, BASE64-encoded PKCS1
                   signature, otherwise ``null``.
:result.Status: Request status – ``POLL`` means the request has to be repeated, ``OK``
                means the signing was successful. Other fields of the response
                only contain information if the value is ``OK``.


.. literalinclude:: ../../common/examples/mid.rpc.signstatus.response.json
   :language: json
   :linenos:

.. literalinclude:: ../../common/examples/mid.rpc.signstatus2.response.json
   :language: json
   :linenos:

Possible error codes in case of the request ``RPC.SignStatus``.

:BAD_REQUEST: An error in the request.
:INTERNAL_SERVER_ERROR: An error in the functioning of the internal server.
:MID_ABSENT: The voter’s mobile phone is not available.
:MID_BAD_CERTIFICATE: An error in the voter’s mobile ID signing certificate.
:MID_CANCELED: The voter cancelled the mobile ID session.
:MID_EXPIRED: The mobile ID session has expired.
:MID_GENERAL: An error in the functioning of the mobile ID service.
:VOTING_END: The voting period has ended.


Vote Verification
-------------------

The verification application makes the request ``RPC.Verify`` to download the
signed vote and the certificates qualifying the vote from the collector service.

:params.OS: The operation system in which the verification application is used.
:params.VoteID: The identifier of the vote in the storage service retrieved from
                the voter application via a QR code.

.. literalinclude:: ../../common/examples/ver.rpc.verify.query.json
   :language: json
   :linenos:

:result.Qualification.ocsp:
:result.Qualification.tspreg:
    See the chapter on vote verification.


:result.Type: The signed vote format. At the moment, the only supported value is
              ``bdoc``.
:result.Vote: BASE64-encoded vote ``SignedVote`` in the format specified above.

.. literalinclude:: ../../common/examples/ver.rpc.verify.response.json
   :language: json
   :linenos:

Possible error codes in case of the request ``RPC.Verify``.

:BAD_REQUEST: An error in the request.
:INTERNAL_SERVER_ERROR: An error in the functioning of the internal server.
:VOTING_END: The voting period has ended.
