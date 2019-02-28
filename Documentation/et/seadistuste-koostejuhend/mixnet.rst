..  IVXV juhend Verificatumi miksneti ettevalmistamiseks ning kasutamiseks

E-häälte miksimine
====================================================

.. _mix-install:

Miksneti Verificatum paigaldamine
---------------------------------

Eeldused
^^^^^^^^

Juhend on kasutamiseks distributsiooniga Ubuntu 18.04 LTS (Bionic Beaver) ja
see eeldab, et käske käivitatakse lihtkasutaja õigustest, kellel on õigus
privileegide eskaleerimiseks `sudo` käsu abil. Lisaks on eeldatud järgmiste
failide olemasolu kasutaja kodukaustas:

Github repositooriumist (https://github.com/vvk-ehk/intcheck):

* :file:`intcheck.py` - tööriist kataloogide täielikkuse kontrolliks

IVXV tarnefailist:

* :file:`gmpmee.dirsha256sum` - gmpmee kataloogi räsi;

* :file:`vmgj.dirsha256sum` - vmgj kataloogi räsi;

* :file:`vcr.dirsha256sum` - vcr kataloogi räsi;

* :file:`vmn.dirsha256sum` - vmn kataloogi räsi;

* :file:`ivxv-verificatum-1.4.4-runner.zip` - IVXV adapter Verificatumi
  kasutamiseks.

Valimise korraldaja käest:

* :file:`data/bb-4.json` - anonümiseeritud e-urn;

* :file:`data/pub.pem` - häälte krüpteerimiseks kasutatud võti.

Kataloogis :file:`data/` ei tohi olla ühtegi teist faili.

Pärast protsessi lõppu on kataloogis :file:`data/` vajalikud järgnevad failid:

* :file:`shuffled.json` - miksitud e-urn;

* :file:`proof.zip` - korrektse miksimise tõend.


Verificatumi ehitamine
^^^^^^^^^^^^^^^^^^^^^^

Ehitamiseks vajalike pakkide paigaldamine::

    sudo apt-get install --no-install-recommends -y autoconf autoconf automake \
    build-essential libgmp-dev libtool git openjdk-8-jdk-headless \
    python unzip wget

Verificatumi lähtekoodi allalaadimine::

    git clone https://github.com/verificatum/verificatum-gmpmee gmpmee
    git clone https://github.com/verificatum/vmgj
    git clone https://github.com/verificatum/vcr
    git clone https://github.com/verificatum/vmn

Lähtekoodist puhaste arhiivide loomine täielikkuse kontrolliks::

    cd gmpmee
    git checkout d781e4a
    rm -rf .git/
    cd ../vmgj
    git checkout 82b57dd
    rm -rf .git/
    cd ../vcr
    git checkout 6dba049
    rm -rf .git/
    cd ../vmn
    git checkout 2a0719e
    rm -rf .git/
    cd ..

Verificatumi lähtekoodi täielikkuse kontrollimine::

    chmod +x ./intcheck.py
    ./intcheck.py verify gmpmee gmpmee.dirsha256sum
    ./intcheck.py verify vmgj vmgj.dirsha256sum
    ./intcheck.py verify vcr vcr.dirsha256sum
    ./intcheck.py verify vmn vmn.dirsha256sum

`gmpmee` ehitamine::

    cd gmpmee/
    make -f Makefile.build
    ./configure
    make
    sudo make install

`vmgj` ehitamine::

    cd ../vmgj/
    make -f Makefile.build
    ./configure
    make
    sudo make install

`vcr` ehitamine::

    cd ../vcr/
    make -f Makefile.build
    ./configure --enable-vmgj
    make
    sudo make install

`vmn` ehitamine::

    cd ../vmn/
    make -f Makefile.build
    ./configure
    make
    sudo make install


IVXV Verificatumi adapteri ja käivitusskripti lahtipakkimine::

    cd ..
    unzip ivxv-verificatum-1.4.4-runner.zip

Verificatumi teekide kopeerimine adapteri väliste teekide kataloogi::

    cp /usr/local/share/java/verificatum-vmgj-1.2.1.jar mixer/lib/verificatum-vmgj.jar
    cp /usr/local/share/java/verificatum-vcr-vmgj-3.0.3.jar mixer/lib/verificatum-vcr-vmgj.jar
    cp /usr/local/share/java/verificatum-vmn-3.0.3.jar mixer/lib/verificatum-vmn.jar
    cp /usr/local/lib/libgmpmee.so.0.0.0 mixer/lib/libgmpmee.so.0
    cp /usr/local/lib/libvmgj-1.2.1.so mixer/lib/libvmgj-1.2.1.so



.. _mix-mix:


E-häälte miksimine
----------------------------------------

Verificatumi miksneti käivitamine::

    cd data
    ../mixer/bin/mix.py --pubkey pub.pem --ballotbox bb-4.json \
    --shuffled shuffled.json --proof-zipfile proof.zip shuffle


.. _mix-verify:

Miksimistõendi verifitseerimine
-------------------------------

Verificatumi adapteri abil saab miskimistõendit ka verifitseerida::

    cd ..
    mkdir verify
    cp data/proof.zip verify
    cd verify
    ../mixer/bin/mix.py verify --proof-zipfile proof.zip
