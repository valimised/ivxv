..  IVXV protocols

====================================================
Qualifying an Electronic Vote for Storage
====================================================

Qualified Vote
=====================

The work of the voter application results in sending a double envelope to the
collector service; that double envelope includes the voter’s intent in an
encrypted form, the voter’s signature on the encrypted intent in the signature
and container format agreed upon, and the voter’s signing certificate in the
X509 format.

So as to store a vote successfully, the IVXV protocol foresees registering the
vote with an external registering service provider, and making the registering
certificate available to the voter application. The election organizer can
prescribe steps in addition to registering in order to quality a vote – such as
getting a certificate status concerning the certificate used to sign the vote.

All qualifying elements requested by the collector service that determine the
status of the vote in the later stages of processing have to be presented to the
voter application, and if demanded, also to the verification application to make
sure that the voter learns about the proper processing of their vote in due
course.

OCSP Certificate Status
-------------------------

OCSP (*Online Certificate Status Protocol*) is a standard protocol to request the
validity information of X509 certificates. The collector service can use this
protocol to learn the validity of the certificate used to sign a vote. The OCSP
response says that the certificate was valid at the moment the request was made,
but does not connect the OCSP response with a specific source.

OCSP-TM Certificate Status
---------------------------

The BDOC 2.1 standard specifies the BDOC-TM profile, where the certificate
status retrieved with the OCSP protocol also functions as a timestamp, which
confirms that the specific signature existed before the OCSP certificate status
was retrieved.

RFC3161 Timestamp
-----------------

The RFC3161 timestamp protocol is used to retrieve confirmation from the trust
service that some set of data already existed before a certain point in time. In
the BDOC-TS context, the signature element ``SignatureValue`` is timestamped in
its canonicalized form. The classic OCSP response with a timestamp in an RFC
3161 format qualifies the BDOC-TS signature.


Vote Registration
====================================================

The IVXV registration protocol has been specified in the document “Registration
service for the electronic voting information system IVXV”. The registration
service functions on the basis of the RFC3161 timestamp protocol. The protocol
has been extended so that the collector service could give its signature to the
timestamp request, which makes it possible to later retrieve a comparative extract
from the registration service. The existence of an independent
registration service decreases the risk of votes being  'lost' by the collector
service.

There is no inherent need to bind the registration protocol to the timestamp
protocol.


Storage
====================================================

Storing an electronic vote in the collector service means:

#. Accepting a vote from the voter application and verifying the voter’s
   signature
#. Potential qualification of the vote – such as verifying the validity of the
   certificate at a time close to the moment the vote was signed
#. Registering the vote in an independent registration service
#. Relaying the elements that qualify the vote to the voter application

Various combinations of the signature format and the vote-qualifying service may
generate various IVXV profiles. In a specific document, the IVXV profile looks
as follows:

#. The signed vote format is BDOC-TS
#. The certificate status protocol is a standard OCSP
#. The RFC3161 timestamp used to qualify BDOC-TS is also used as registration
   verification
