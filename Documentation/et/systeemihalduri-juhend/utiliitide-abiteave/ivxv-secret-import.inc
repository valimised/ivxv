.. Elektroonilise hääletamise infosüsteem IVXV

.. Käsureautiliidi ivxv-secret-import kasutusjuhend

.. SEE FAIL ON GENEREERITUD AUTOMAATTESTI POOLT:
.. * features/steps/autodoc.py
.. * features/autodoc-utilities.feature

.. _ivxv-secret-import:

ivxv-secret-import
``````````````````

:command:`ivxv-secret-import --help`:

.. code-block:: text

   Import secret data files to services.

   This utility loads file that contains secret data to services.

   Supported secret types are:

       tls-cert - TLS certificate for service.

               Certificate (and key) is used for securing
               communication between services and service instances.

       tls-key - TLS key for service.

               Key is used together with service certificate.

       tsp-regkey - PKIX TSP registration key for voting services.

               Key is used for signing Time Stamp Protocol requests.

               Key file must be in PEM format and
               must be not password protected.

       dds-token-key - Mobile ID identity token for
                       choices, dds and voting services.

               Key file must be 32 bytes long.

   Usage: ivxv-secret-import [--service=<service-id>] <secret-type> <keyfile>
