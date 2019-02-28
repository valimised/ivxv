..  IVXV kogumisteenuse haldusteenuse kirjeldus

.. _app-auditor:

Auditirakendus
==============

IVXV võtmerakendus võimaldab kasutada tõestatavat dekrüpteerimist -
koos tulemusega väljastatakse lugemistõend e-häälte korrektse avamise
kohta. Vältimaks häälte salajasuse rikkumist lugemistõendi kontrollil
võimaldab IVXV kasutada häälte miksimist, mis säilitab häälte sisu
kuid eemaldab krüptograafiliselt seose konkreetse hääle ja selle hääle
andnud isiku vahel.

IVXV kasutab e-häälte miksimiseks tarkvara Verificatum, mis võtab
sisendiks krüpteeritud hääled ning annab väljundiks miksitud
krüpteeritud hääled ja miksimistõendi.

Miksimistõendi ja lugemistõendi kontroll toimub auditirakendusega, mis
koosneb tööriistadest *convert*, *mixer* ja *decrypt*.

#. Tööriist *convert* kontrollib, teisenduste korrektsust IVXV
   andmevormingute ja Verificatumi andmevormingute vahel.
#. Tööriist *mixer* kontrollib miksimistõendi korrektsust.
#. Tööriist *decrypt* kontrollib lugemistõendi korrektsust.

E-häälte korrektse kokkulugemise kontrolliks on vajalik ja piisav
kasutada kõiki kolme auditirakenduse tööriista.

Suurte e-urni ja tõendite korral võib olla tarvilik suurendada rakenduse
mälupiirangut. Selleks tuleb seada keskonnamuutuja ``AUDITOR_OPTS`` väärtusega
``-Xmx{N}G``, kus ``{N}`` on mälupiirangu suurus gigabaitides. Näiteks 10
gigabaidi mälupiirangu seadmiseks tuleb seada keskkonnamuutuja
``AUDITOR_OPTS=-Xmx10G``.

Kõigi tööriistade kasutamine eeldab allkirjastatud usaldusjuure ja
konkreetse tööriista seadistuste olemasolu. Alljärgnevalt kirjeldame
konkreetsete tööriistade seadistusi.

.. _auditor-convert:

E-häälte korrektse teisendamise kontroll
----------------------------------------

Verificatumi poolt koostatud miksimistõendi formaat on erinev IVXV raamistikus
kasutatavast formaadist, samuti erinevad IVXV ning Verificatumi
krüpteeritud häälte formaadid. IVXV raamistikku on pakendatud
adapterid formaaditeisendusteks, auditirakendus pakub võimalust nende
teisenduste korrektsuse kontrolliks.

Tööriist *convert* kontrollib, et Verificatumi poolt väljastatud
miksimistõend vastab failidele IVXV raamistikus.

:convert.input_bb: IVXV miksimiseelse e-urni asukoht.

:convert.output_bb: IVXV miksimisjärgse e-urni asukoht.

:convert.pub: IVXV avaliku võtme asukoht.

:convert.protinfo: Verificatumi miksimise protokollifaili asukoht.

:convert.proofdir: Verificatumi miksimistõendi asukoht.

:file: `auditor.convert.yaml`:

.. literalinclude:: config-examples/auditor.convert.yaml
   :language: yaml
   :linenos:

.. _auditor-mix:

E-häälte miksimistõendi kontroll
--------------------------------

Tööriist *mixer* kontrollib Verificatumi miksimistõendi korrektsust.

:mixer.protinfo: Verificatumi miksimistõendi protokollifaili asukoht.

:mixer.proofdir: Verificatumi miksimistõendi asukoht.

:mixer.threaded: Kasuta mitmelõimelist implementatsiooni. Vaikimisi
                 väärtus on väär. Kasutatavate lõimede arv sõltub
                 käsurea-argumentidest. Käsurea-argumentide puudumise
                 korral valitakse optimaalne lõimede arv lähtudes
                 tuvastatud tuumade arvust.

:file:`auditor.mixer.yaml`:

.. literalinclude:: config-examples/auditor.mixer.yaml
   :language: yaml
   :linenos:

.. _auditor-decrypt:

E-häälte lugemistõendi kontroll
-------------------------------

Tööriist *decrypt* kontrollib lugemistõendi korrektsust.

:decrypt.input: Lugemistõendi asukoht

:decrypt.pub: Dekrüpteerimiseks kasutatud salajasele võtmele vastava avaliku
              võtme asukoht.

:decrypt.out: Lugemistõendi kontrolli tulemuste asukoht. Tegemist on
              kataloogiga kuhu salvestatakse sedelid, mille
              lugemistõend oli kehtetu.

:file:`auditor.decrypt.yaml`:

.. literalinclude:: config-examples/auditor.decrypt.yaml
   :language: yaml
   :linenos:
