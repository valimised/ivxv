..  IVXV protocols

================================================================================
Electronic Vote Verification
================================================================================

An electronic vote is verified in the processing application, in the collector
service, in the voter application and in the verification application. An
electronic vote is checked the most thoroughly inside the ballot box in the
processing application, where it is decided whether this specific vote is to be
sent to counting or not. For each single vote, checks are made in the voter
application analogously to the processing application to make sure that the
collector service has qualified the vote in a way that allows the checks made in
the processing application to be performed successfully. The checks analogous to
the voter application are performed by the verification application.

Checks in the Collector Service
----------------------------------

The voter application sends to the collector service in the set of the signed
vote the following:

#. An encrypted ballot
#. The voter’s signature on the encrypted ballot
#. The voter’s signature certificate

The collector service performs the following checks as a minimum:

#. The signer of the vote is included in the voter list
#. The signed vote is presented in the correct container format
#. The digital signature on the encrypted ballot is correct
#. The vote signer’s certificate was valid at the moment the vote was accepted

To verify the validity of the vote signer’s certificate, the collector service
requests information from the certificate status service. The collector service
verifies the response of the certificate status service on the status of the
certificate and adds the response to the elements that qualify a vote.

The collector service registers the fact of storing the vote in the external
registration service by signing the registration request and storing the
registration certificate signed by the registration service as one of the
vote-qualifying elements.

The collector service sends all of the elements qualifying the vote collected by
it to the voter application with the unique vote identifier.

Checks in the Voter Application
-----------------------------------

On the basis of the voter’s plaintext will, the voter application prepares an
encrypted ballot and signs it with the voter’s chosen digital signing means.

The role of the voter application after the vote is signed is to make sure that
the collector service has acted according to protocol when getting
vote-qualifying elements and that the vote has been stored so that it will be
taken into account by the processing application.

The voter application performs at a minimum the following checks:

#. The collector service got the certificate status from the certificate status
   service authorized for the voter’s certificate. The voter application checks
   the signature on the certificate status response

#. The collector service registered the vote signed by the voter in the
   authorized registration service. The voter application checks that the
   request made by the collector service was signed by the collector service and
   referred to a correctly signed vote. The voter application checks that the
   registration service response has been signed by the correct registration
   service provider and includes the request signed by the collector service

If the elements needed to qualify a vote cannot be checked, the voter
application will notify the user.


Checks in the Verification Application
----------------------------------------

The verification application receives the following information from the voter
application:

#. The randomness used to generate an encrypted ballot
#. The unique identifier of a signed vote in the collector service

The verification application uses the unique identifier of the vote from the
collector service to retrieve the following information:

#. The encrypted ballot
#. The voter’s signature on the encrypted ballot
#. The voter’s signing certificate
#. Vote-qualifying elements, incl. the certificate status and the registration
   certificate

The verification application makes the following checks:

#. The signed vote is presented in the correct container format
#. The digital signature on the encrypted ballot is correct
#. The certificate of the vote signer was valid at the moment the vote was
   accepted, which is confirmed by a correct certificate status
#. The vote is correctly registered in the correct registration service

After these checks have been performed, the verification application displays
the data of the person who signed the vote.

In addition, the verification application uses the randomness used to generate
the encrypted ballot also to decrypt the encrypted ballot.

**NB!** The randomness
used to encrypt one vote can be used to decrypt only that vote. To decrypt
several different votes, the private component of the vote secrecy key pair is
needed.

The verification application makes sure that the plaintext retrieved as a result
of decryption matches the required syntax of the plaintext intent.

The verification application displays the form-compliant intent to allow the
verifier to make sure that the intent is correct.


Checks in the Processing Application
--------------------------------------

The processing application checks each vote separately, making sure among other
things that the views on the content of the ballot box given by each
collector service and registration service are consistent. The processing
application then decides, which vote is the last in terms of time and is to be
directed to the next stage of processing, as a result of which, the vote may
make it to counting.

The input of the processing application:

#. A list of registration requests received by the registration service
#. A list of voter lists implemented in the collector service
#. The ballot box delivered by the collector service that includes per each vote
   an encrypted ballot, the voter’s signature on an encrypted ballot, the
   voter’s signing certificate, the status of the certificate, and
   the registration certificate

The processing application checks the consistency of the registration service
and the collector service and issues the differences:

#. Votes for which there is a registration request in the collector service, but
   the response for which has not made it to the collector service
#. Votes for which there is a registration request in the collector service, but
   which have not been delivered by the collector service

The processing application checks every single vote for the following:

#. The signer of the vote is in the voter list
#. The signed vote is presented in the correct container format
#. The digital signature on the encrypted ballot is correct
#. The certificate of the signer of the vote was valid at the moment the vote
   was accepted, which is verified by the correct certificate status
#. The vote is registered correctly in the correct registration service

The processing application decides which of the voter’s votes was last and thus
moves on to the next stage of processing. I.e. one of the vote-qualifying
elements plays the role of a fixer when the vote is stored, and on the basis of
that element, the sequence of single votes in time is generated. Depending on
the IVXV profile, this element can be in the composition of the certificate
status (BDOC-TM), as a separate timestamp (BDOC-TS), or in the composition of
the registration certificate (BDOC-TS).
