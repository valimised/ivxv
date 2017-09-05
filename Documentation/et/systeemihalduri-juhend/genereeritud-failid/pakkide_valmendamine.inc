.. Elektroonilise hääletamise infosüsteem IVXV

.. Pakkide ehitamise käigus ühikteste ei käivitata (``DEB_BUILD_OPTIONS=nocheck``), kuna ühiktestid pole käesoleva testikomplekti fookuses. Pakendatavast tarkvarast kompileeritakse arendusversioon (``DEVELOPMENT=1``), mis võimaldab kasutada võimaldab kasutada ainult arenduseks ja testimiseks mõeldud mooduleid.

.. SEE FAIL ON GENEREERITUD AUTOMAATTESTI POOLT:
.. * steps/test_util/report.py

..  Feature: IVXV raamistiku ehitamine
             features/_build-ivxv.feature:5
   Scenario: IVXV kogumisteenuse tarkvara ja dokumentatsiooni ehitamine
             features/_build-ivxv.feature:8
       Step: Ehitaja kompileerib koodihoidlast IVXV deb-pakid
             features/_build-ivxv.feature:15

.. code-block:: shell-session

   builder@builder $ dpkg-buildpackage -b -uc
   dpkg-buildpackage: source package ivxv
   dpkg-buildpackage: source version 0.9
   dpkg-buildpackage: source distribution UNRELEASED
   dpkg-buildpackage: source changed by IVXV Developer <info@ivotingcentre.ee>
   dpkg-buildpackage: host architecture amd64
    dpkg-source --before-build ivxv.git
    fakeroot debian/rules clean
   dh clean --with systemd --with python3
      dh_testdir
      debian/rules override_dh_auto_clean
   ...
   dpkg-deb: building package 'ivxv-verification' in '../ivxv-verification_0.9_amd64.deb'.
   dpkg-deb: building package 'ivxv-voting' in '../ivxv-voting_0.9_amd64.deb'.
   dpkg-deb: building package 'ivxv-storage' in '../ivxv-storage_0.9_amd64.deb'.
   dpkg-deb: building package 'ivxv-dds' in '../ivxv-dds_0.9_amd64.deb'.
   dpkg-deb: building package 'ivxv-log' in '../ivxv-log_0.9_all.deb'.
    dpkg-genchanges -b >../ivxv_0.9_amd64.changes
   dpkg-genchanges: binary-only upload (no source code included)
    dpkg-source --after-build ivxv.git
   dpkg-buildpackage: binary-only upload (no source included)