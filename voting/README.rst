================================
 IVXV Internet voting framework
================================
----------------
 Voting service
----------------

The voting service

- verifies eligibility of voters,
- verifies signatures on submitted votes,
- acquires time marks to prove that the voter's signature certificate was valid
  during submission,
- generates vote identifiers,
- registers the vote with a registration service,
- stores the voter identity, certificate, vote, vote identifier, time mark, and
  registration confirmation in the `storage service`_, and
- sends the voter the vote identifier and registration confirmation for
  verification.

The vote container and signature format, time mark, registration, and storage
protocols are all abstracted and can be replaced with new implementations.

.. _`storage service`: https://ivxv.ee/storage
