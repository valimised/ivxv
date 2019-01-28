..  IVXV API

Java rakenduste tõlkimine
=========================

1. Tekitada tühi töökaust :file:`<kaust>`::

       mkdir <kaust>

2. Kopeerida sellesse common teegi jar fail :file:`common-<versioon>.jar`
   (asub tarnitud rakenduste :file:`lib/` kaustas)::

       cp <path>/<to>/lib/common-<versioon>.jar <kaust>

3. Minna töökausta::

       cd <kaust>

4. Pakkida common teek lahti::

       unzip common-<version>.jar

   .. note::

      Tekib hulk faile ja kaustu, millest tõlkimise kontekstis on oluline
      ainult kaust :file:`i18n/`.

5. Muuta tõlkeid kaustas :file:`i18n` asuvates tõlkefailides. Näiteks
   :file:`i18n/common-msg_et.properties`, võti :file:`r_ivl_description` on PDF
   e-hääletajate nimekirja (``r`` ehk "raport", ``ivl`` ehk "internet voter list")
   päis.

6. Uuendada common teegi jar-is ühte muudetud tõlkefaili::

       zip -r common-<version>.jar i18n/common-msg_et.properties

   Või korraga kogu tõlgete kausta, kui muudatusi tehti mitmes tõlkefailis::

       zip -r common-<version>.jar i18n/

7. Kopeerida common teek algsele kohale, kirjutades esialgse faili üle::

       cp common-<versioon>.jar ../<path>/<to>/lib/

8. Eemaldada töökaust::

       cd ..
       rm -r <kaust>

