..  IVXV protocols

================================================================================
Overview
================================================================================

The online voting protocol suite (hereinafter the protocol suite) defines
communication between the components of the online voting system, the data
structures used, the algorithms and interfaces to external systems. The
communication of messages is presented in UML interaction diagrams, which define
the sequence of messages. The specifications of data structures have
representations in the Backus-Naur or JSON schema. The following are used to
separate data structure fields: line break symbol ``LF`` with the ASCII code
``0x0A`` and the tabulator symbol ``TAB`` with the ASCII code ``0x09``. The
algorithms are presented as a pseudocode.

NB! All fields in the data structures of the protocol suite require strict
adherence to the permitted symbols and the minimum and maximum lengths of the
fields. Using additional spaces, tabulators, etc. is prohibited and applications
implementing the specification have to refuse processing data that do not comply
with the format.

The protocol suite defines the online voting protocol and the support structures
necessary to implement this protocol.

Online Voting Protocol
================================================================================

The online voting protocol specifies:

#. The electronic vote format that allows defining the voter’s intent
   unambiguously at a specific election
#. The encryption of the electronic vote to ensure secrecy
#. The digital signing of an electronic vote to ensure integrity and voter
   identification
#. Qualifying the electronic vote by a collector service to stand for vote
   acceptance

The protocol requires the election organizer to have defined the election and
generated for vote encryption a key pair, whose public component has been made
available to the voter application.

The voter’s intent moves via the protocol to the ballot box stored in the
collector service and is taken into consideration in the shaping of the result
as follows:

#. The voter uses the voter application to formulate their intent electronically

   #. the expression of intent is shown as an electronic vote

   #. the vote is encrypted

   #. the encrypted vote is digitally signed

#. The collector service stores the electronic vote

   #. elements confirming the validity of the voter certificate are requested
      for the digitally signed vote

   #. the electronic vote is registered in the external registering service

   #. the voter can check the qualified electronic vote with the verification
      application

#. The voter can use the verification application to make sure that their vote
   is handled correctly by the collector service

#. At the end of the voting period the collector service issues the ballot box
   to the organizer of the election, and the registration service issues a list
   of the votes registered by the collector service

#. The organizer of the election calculates the voting result

   #. it is made sure that all the votes registered in the registration service
      have been delivered in the composition of the ballot box

   #. encrypted votes and digital signatures are separated

   #. encrypted votes are decrypted

   #. the voting result is calculated based on the decrypted votes

The protocol is analogous to the protocol of postal voting, where the voter’s
intent travels to the National Electoral Committee in two envelopes – the outer
envelope has inside it another envelope, which in turn contains the voting
ballot stating the voter’s intent. The outer envelope carries the information
that allows identifying the voter, thus making it possible to check the voter’s
right to vote, among other things. The inner envelope is anonymous to protect
the secrecy of the vote. The inner envelopes and the outer envelopes are
separated before the votes are counted.

In online voting, the inner envelope is an encrypted vote and the outer
envelope is a digitally signed document.
