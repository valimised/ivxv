..  IVXV kogumisteenuse haldusteenuse kirjeldus

Auditirakendus
==============

Auditirakendus koosneb tööriistadest *mixer* ja *decrypt*.  Kõigi tööriistade
kasutamine eeldab allkirjastatud usaldusjuure ja konkreetse tööriista
seadistuste olemasolu. Alljärgnevalt kirjeldame konkreetsete tööriistade
seadistusi.

.. _auditor-mix:

E-häälte miksimistõendi kontroll
--------------------------------

Tööriist *mixer* kontrollib Verificatumi lugemistõendi korrektsust.

:mixer.protinfo: Verificatumi segamistõendi protokollifaili asukoht.

:mixer.proofdir: Verificatumi segamistõendi asukoht.

:file:`auditor.mixer.yaml`:

.. literalinclude:: config-examples/auditor.mixer.yaml
   :language: yaml
   :linenos:

.. _auditor-decrypt:

E-häälte lugemistõendi kontroll
-------------------------------

Tööriist *decrypt* kontrollib dekrüpteerimistõendi korrektsust.

:decrypt.input: Dekrüpteerimistõendi asukoht

:decrypt.pub: Dekrüpteerimiseks kasutatud salajasele võtmele vastava avaliku
              võtme asukoht.

:decrypt.out: Dekrüpteerimistõendi kontrolli tulemuste asukoht. Tegemist peab
              olema kataloogiga kuhu salvestatakse sedelid, mille
              dekrüpteerimistõend oli kehtetu.

:file:`auditor.decrypt.yaml`:

.. literalinclude:: config-examples/auditor.decrypt.yaml
   :language: yaml
   :linenos:

.. vim: sts=3 sw=3 et:
