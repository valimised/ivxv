..  IVXV API

========
Java API
========

Java lähtekoodi API nägemiseks kasutatakse standardvahendite hulgas olevat
`javadoc <http://www.oracle.com/technetwork/java/javase/documentation/index-jsp-135444.html>`_ tööriista.

Javadoc tööriistale tuleb anda Java lähtekoodi sisaldavad kaustad ning
väljundkaust, kuhu genereeritud HTML panna. Seega käsu

::

   javadoc -sourcepath \
      common/java/src/main/java/:auditor/src/main/java:key/src/main/java:processor/src/main/java \
      -subpackages ee.ivxv \
      -d javadoc

tulemusena luuakse kaust ``javadoc``, mis sisaldab endas IVXV Java lähtekoodi
dokumentatsiooni.
