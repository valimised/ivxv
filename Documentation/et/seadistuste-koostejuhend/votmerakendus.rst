..  IVXV kogumisteenuse haldusteenuse kirjeldus

Võtmerakendus
=============

Võtmerakendus koosneb tööriistadest *groupgen*, *init*, *decrypt* ja *util*.
Kõigi tööriistade kasutamine eeldab allkirjastatud usaldusjuure ja konkreetse
tööriista seadistuste olemasolu. Alljärgnevalt kirjeldame konkreetsete
tööriistade seadistusi.

.. _key-groupgen:

Häälte salastamise võtme spetsifikatsiooni genereerimine (VALIKULINE)
---------------------------------------------------------------------

Häälte salastamise võtme spetsifikatsioon genereeritakse kasutades tööriista
*groupgen*. Võtmerakendus kasutab ElGamal'i krüptosüsteemi. Vastavalt aluseks
olevale rühmale tuleb valida turvaparameeter, mis on hiljem aluseks rühma ja
võtme genereerimisele.

Võtmespetsifikatsiooni genereerimine on ajaliselt mahukas tegevus, mis võib
olenevalt riistvarast kesta tunde. Ühekordselt genereeritud rühm on
mitmekordselt kasutatav.

:groupgen.paramtype: ElGamal'i krüptosüsteemi töö aluseks oleva rühma
                     tüüp. Toetatud väärtused:

                     #. mod - jäägiklassiring Zp

:groupgen.length: ElGamal'i krüptosüsteemi töö aluseks olevat rühma iseloomustav
                  turvaparameeter. Jäägiklassiringide korral on sobiv väärtus
                  2048, mis on samaväärne 2048 bitise RSA turvalisusega.


:groupgen.random_source: Juhuarvugeneraatori sisendiks kasutatavate allikate
                         loetelu.

:groupgen.random_source.random_source_type: Juhuarvugeneraatori allika tüüp.

:groupgen.random_source.random_source_path: Juhuarvugeneraatori allika
                                            seadistatav asukoht.  Argument on
                                            valikuline sõltuvalt allika tüübist.


:file:`key.groupgen.yaml`:

.. literalinclude:: config-examples/key.groupgen.yaml
   :language: yaml
   :linenos:

.. _key-init:

Häälte salastamise võtme genereerimine
--------------------------------------

Häälte salastamise võtme genereerimiseks kasutatakse võtmerakenduse tööriista
*init*. Võti genereeritakse seadistustes näidatud läviskeemiga MofN, mis
tähendab, et N võtmehaldurist peavad häälte dekrüpteerimisel osalema vähemalt M
haldurit, vastasel juhul ei ole dekrüpteerimine võimalik.


:init.identifier: Valimise unikaalne identifikaator.

:init.out:

       Võtmerakenduse tööriista *init* väljundkataloog. Sellesse kataloogi
       tekivad

       #. PEM vormingus allkirjavõtme sertifikaat
       #. PEM vormingus krüpteerimisvõtme sertifikaat
       #. PEM vormingus krüpteerimisvõti

:init.skiptest: Võtmeosakute kontrolltestide vahelejätmine.

----

:init.paramtype: ElGamal krüptosüsteemi aluseks oleva rühma parameetrid, mis
                 ühtlasi määravad võtme turvataseme.

:init.paramtype.mod: Jäägiklassiringi määravad parameetrid kümnendesituses.
                     Parameetrid võib luua võtmerakenduse tööriista *groupgen*
                     kasutades.

:init.paramtype.mod.p: Jäägiklassiringi moodul.

:init.paramtype.mod.g: Jäägiklassiringi generaator.

----

:init.signaturekeylen: Võtmerakenduse poolt genereeritava allkirjastamise võtme
                       pikkus.
:init.issuercn:
:init.signcn:
:init.signsn:
:init.enccn:
:init.encsn: Võtmerakenduse poolt loodavate sertifikaatide DN väärtused.

----

:init.required_randomness: Juhuslikkuse allikatest loetava kohustuslik suurus
                           baitides.

:init.random_source: Juhuarvugeneraatori sisendiks kasutatavate allikate
                     loetelu.

:init.random_source.random_source_type: Juhuarvugeneraatori allika tüüp.

:init.random_source.random_source_path: Juhuarvugeneraatori allika seadistatav
                                        asukoht. Argument on valikuline
                                        sõltuvalt allika tüübist.

----

:init.genprotocol: Võtme genereerimiseks kasutatava algoritmi ja läviskeemi
                   spetsifikatsioon.

:init.genprotocol.desmedt:

      Algoritmi Desmedt korral genereeritakse võti usaldatava osakujagaja poolt
      ehk võtmerakenduse mälus. Privaatvõtme osakud talletatakse kiipkaartidel.

      Täiendavalt tuleb määrata läviskeemi osaliste arv ja minimaalne kvoorum.

      Kaartide arv 7 - võimalikud kvoorumid 1,2,3,4 - soovitatav kvoorum 4
      Kaartide arv 8 - võimalikud kvoorumid 1,2,3,4 - soovitatav kvoorum 4
      Kaartide arv 9 - võimalikud kvoorumid 1,2,3,4,5 - soovitatav kvoorum 5

:init.genprotocol.desmedt.threshold: Läviskeemi M väärtus - kvoorum.

:init.genprotocol.desmedt.parties: Läviskeemi N väärtus.

:file:`key.init.yaml`:

.. literalinclude:: config-examples/key.init.yaml
   :language: yaml
   :linenos:

.. _key-decrypt:

Elektrooniliste häälte dekrüpteerimine
--------------------------------------

Elektrooniliste häälte dekrüpteerimiseks kasutatakse võtmerakenduse tööriista
*decrypt*. Dekrüpteerimise õnnestumiseks peab osalema läviskeemi poolt määratud
kvoorumi jagu võtmehaldureid. Kui rakendati skeemi 5of9, siis osaleb
dekrüpteerimisel täpselt 5 võtmehaldurit. Vähema arvu haldurite korral ei ole
dekrüpteerimine võimalik.

:decrypt.identifier:

        Valimise unikaalne identifikaator.

----

:decrypt.protocol:

:decrypt.protocol.recover:

      Algoritmi Desmedt korral genereeritakse võti usaldatava osakujagaja poolt
      ehk võtmerakenduse mälus. Privaatvõtme osakud talletatakse kiipkaartidel.

      Täiendavalt tuleb määrata läviskeemi osaliste arv ja minimaalne kvoorum.

      Kaartide arv 7 - võimalikud kvoorumid 1,2,3,4 - soovitatav kvoorum 4
      Kaartide arv 8 - võimalikud kvoorumid 1,2,3,4 - soovitatav kvoorum 4
      Kaartide arv 9 - võimalikud kvoorumid 1,2,3,4,5 - soovitatav kvoorum 5

:decrypt.protocol.recover.threshold:

      Läviskeemi M väärtus - kvoorum.

:decrypt.protocol.recover.parties:

      Läviskeemi N väärtus.


----

:decrypt.anonballotbox:

      Töötlemisrakenduse või miksimisrakenduse poolt loodud e-urn anonüümistatud
      häältega.

:decrypt.anonballotbox_checksum:

      Anonüümistatud häältega e-urni allkirjastatud SHA256 kontrollsummafail.

:decrypt.questioncount:

      Küsimuste arv anonüümistatud e-urnis. Vaikimisi väärtus on 1.

:decrypt.candidates:

      Valimise valikute nimekiri allkirjastatud kujul.

:decrypt.districts:

      Valimise ringkondade nimekiri allkirjastatud kujul.

:decrypt.provable:

      Valikuline korrektse dekrüpteerimise tõestuse väljastamine. Vaikimisi
      väärtus on tõene.

:decrypt.out:

      Võtmerakenduse tööriista *decrypt* väljundkataloog. Eduka dekrüpteerimise
      korral tekivad siia kausta:

      #. Elektroonilise hääletamise tulemus
      #. Elektroonilise hääletamise tulemuse signatuur
      #. Loend kehtetutest sedelitest
      #. Lugemistõend


:file:`key.decrypt.yaml`:

.. literalinclude:: config-examples/key.decrypt.yaml
   :language: yaml
   :linenos:


Täiendavad tööriistad
---------------------

:util.listreaders: Loetle ühendatud kaardilugejad.

----

:util.testkey: Alamargumentide seadistuse korral viiakse läbi võtmeosakute
               kooskõlalisuse ja korrasoleku testimine.

:util.testkey.out: Krüpteerimise avaliku võtme asukoha kataloog.

:util.testkey.threshold: Testimiseks kasutatav lävi.

:util.testkey.parties: Testimiseks kasutatav osapoolte arv.

:file:`key.util.yaml`:

.. literalinclude:: config-examples/key.util.yaml
   :language: yaml
   :linenos:


.. vim: sts=3 sw=3 et:
