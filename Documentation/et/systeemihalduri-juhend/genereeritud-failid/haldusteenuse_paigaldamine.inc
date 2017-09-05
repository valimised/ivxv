.. Elektroonilise hääletamise infosüsteem IVXV

.. Haldusteenuse paigaldamine:

.. SEE FAIL ON GENEREERITUD AUTOMAATTESTI POOLT:
.. * steps/test_util/report.py

..  Feature: Kogumisteenuse paigaldamine kõrgkäideldava konfiguratsiooniga
             features/collector-ha-setup.feature:6
   Scenario: Kogumisteenuse haldusteenuse paigaldus
             features/collector-ha-setup.feature:470
       Step: Haldur paigaldab haldusteenuse
             features/collector-ha-setup.feature:475

.. code-block:: text

   root@admin # dpkg -i /etc/ivxv/debs/ivxv-common_0.9_all.deb /etc/ivxv/debs/ivxv-admin_0.9_all.deb
   Selecting previously unselected package ivxv-common.
   (Andmebaasi lugemine ... 12008 files and directories currently installed.)
   Preparing to unpack .../debs/ivxv-common_0.9_all.deb ...
   Unpacking ivxv-common (0.9) ...
   Selecting previously unselected package ivxv-admin.
   Preparing to unpack .../debs/ivxv-admin_0.9_all.deb ...
   Unpacking ivxv-admin (0.9) ...
   Paki ivxv-common (0.9) paikasättimine ...
   # Adding user group 'ivxv'
   Adding group `ivxv' (GID 109) ...
   ...
   Signature ok
   subject=/CN=ivxv-admin-ui/O=IVXV kogumisteenuse haldusteenus/L=Somewhere/C=EE
   Getting Private key
   # Generating strong Diffie-Hellman group file
   Generating DH parameters, 2048 bit long safe prime, generator 2
   This is going to take a long time
   .......................................+........................................
   ................................................................+...............
   ..............................................................................+.
   .........++*++*
   # Starting Apache web server
   # Restarting rsyslog log server