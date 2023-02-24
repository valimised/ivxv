### collector/cmd/verifier

Can verify documents that are signed with:

- id-card
- mobile-id
- smart-id

#### If and only if you have the following trust.yaml configuration:
```
# Usaldusjuure seadistus YAML-struktuurina

container:
bdoc:
bdocsize: 104857600  # 100 MiB
filesize: 104857600  # 100 MiB
roots:
- !container devel_root.crt
- !container sk_test_root.crt
- !container EE-GovCA2018.crt
- !container EE_Certification_Centre_Root_CA.crt
intermediates:
- !container devel_intermediate.crt
- !container sk_test_intermediate.crt
- !container ESTEID-SK_2015.crt
- !container esteid2018.crt
- !container EID-SK_2016.pem.crt
profile: TS
ocsp:
responders:
- !container sk_test_ocsp.crt
- !container SK_OCSP_RESPONDER_2011.crt
- !container EID-SK_2016_OCSP_RESPONDER_2018.pem.cer
tsp:
signers:
- !container sk_test_tsa.crt
- !container SK_TIMESTAMPING_AUTHORITY_2019.crt
- !container SK_TIMESTAMPING_AUTHORITY_2020.crt
- !container SK_TIMESTAMPING_AUTHORITY_2021.crt
- !container SK_TIMESTAMPING_AUTHORITY_2022.crt
delaytime: 10

authorizations:
- PEREKONNANIMI,NIMI,ISIKUKOOD
```

#### P.S Don't forget to include all these listed certificates into container
