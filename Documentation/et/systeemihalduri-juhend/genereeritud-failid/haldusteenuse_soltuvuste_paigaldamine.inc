.. Elektroonilise hääletamise infosüsteem IVXV

.. Haldusteenuse sõltuvuste paigaldamine:

.. SEE FAIL ON GENEREERITUD AUTOMAATTESTI POOLT:
.. * steps/test_util/report.py

..  Feature: Kogumisteenuse käivitamine kõrgkäideldava konfiguratsiooniga
             features/collector-ha-setup.feature:6
   Scenario: Kogumisteenuse haldusteenuse paigaldus
             features/collector-ha-setup.feature:336
       Step: Haldur paigaldab haldusteenuse
             features/collector-ha-setup.feature:342

.. code-block:: text

   root@admin # apt-get install --yes --quiet adduser openssh-server openssl rsyslog rsyslog-relp sudo tzdata locales libc6 python3-bottle python3-dateutil python3-debian python3-docopt python3-jinja2 python3-openssl python3-yaml python3:any apache2 cron language-pack-et libapache2-mod-wsgi-py3 python3-gdbm python3-pkg-resources rsync
   Reading package lists...
   Building dependency tree...
   Reading state information...
   adduser is already the newest version (3.116ubuntu1).
   cron is already the newest version (3.0pl1-128.1ubuntu1).
   libc6 is already the newest version (2.27-3ubuntu1).
   locales is already the newest version (2.27-3ubuntu1).
   locales on määratud käsitsi paigaldatuks.
   python3-dateutil is already the newest version (2.6.1-1).
   python3-debian is already the newest version (0.1.32).
   ...
   apache2 is already the newest version (2.4.29-1ubuntu4.5).
   language-pack-et is already the newest version (1:18.04+20180712).
   openssh-server is already the newest version (1:7.6p1-4ubuntu0.1).
   openssl is already the newest version (1.1.0g-2ubuntu4.3).
   python3 is already the newest version (3.6.7-1~18.04).
   python3 on määratud käsitsi paigaldatuks.
   python3-gdbm is already the newest version (3.6.7-1~18.04).
   tzdata is already the newest version (2018i-0ubuntu0.18.04).
   0 uuendatud, 0 värskelt paigaldatud, 0 eemaldada ja 15 uuendamata.