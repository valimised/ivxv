================================
 IVXV Internet voting framework
================================
----------------
 Voting service
----------------

The voting service

- verifies eligibility of voters,
- verifies signatures on submitted votes,
- generates vote identifiers,
- qualifies votes as configured, which usually entails
  - acquiring revocation information for signing certificates, and
  - registering votes with a registration service,
- stores votes, identifiers, and qualifying properties in the `storage
  service`_, and
- sends voters the vote identifiers and qualifying properties for
  verification.

The vote container and signature format, qualifying properties, and storage
protocols are all abstracted and can be replaced with new implementations.

.. _`storage service`: https://ivxv.ee/storage
