package ee.ivxv.common.service.container.bdoc;

import ee.ivxv.common.service.container.Container;
import ee.ivxv.common.service.container.DataFile;
import ee.ivxv.common.service.container.Signature;
import ee.ivxv.common.service.container.Subject;
import java.io.InputStream;
import java.time.Instant;
import java.util.Date;
import java.util.List;
import java.util.Optional;
import java.util.stream.Collectors;
import java.util.stream.Stream;
import org.digidoc4j.X509Cert;
import org.digidoc4j.X509Cert.SubjectName;
import org.digidoc4j.impl.bdoc.BDocSignature;

class DataConverter {

    static Container convert(org.digidoc4j.Container c) {
        return new Container(convertFiles(c.getDataFiles()), convertSignatures(c.getSignatures()));
    }

    static List<DataFile> convertFiles(List<org.digidoc4j.DataFile> files) {
        return files.stream().map(DataConverter::convertFile).collect(Collectors.toList());
    }

    static DataFile convertFile(org.digidoc4j.DataFile df) {
        return new DataFileImpl(df);
    }

    static List<Signature> convertSignatures(List<org.digidoc4j.Signature> signatures) {
        return signatures.stream().map(DataConverter::convertSignature)
                .collect(Collectors.toList());
    }

    static Signature convertSignature(org.digidoc4j.Signature s) {
        Subject signer = getSubject(s.getSigningCertificate());
        Instant time = Stream.of(s.getTrustedSigningTime(), s.getClaimedSigningTime())
                .filter(d -> d != null).findFirst().map(Date::toInstant).orElse(null);
        byte[] value = null;
        if (s instanceof BDocSignature) {
            // The base64 representation of the value is the contents of <SignatureValue> tag.
            value = ((BDocSignature) s).getOrigin().getSignatureValue();
        }

        return new Signature(signer, time, value);
    }

    private static Subject getSubject(X509Cert cert) {
        if (cert == null) {
            return null;
        }
        String serial = cert.getSubjectName(SubjectName.SERIALNUMBER);
        String first = cert.getSubjectName(SubjectName.GIVENNAME);
        String last = cert.getSubjectName(SubjectName.SURNAME);

        if (serial == null || first == null || last == null) {
            // Common name is "SURNAME,GIVENNAME,SERIALNUMBER", possibly in quotes
            String cn = Optional.ofNullable(cert.getSubjectName(SubjectName.CN))
                    .map(n -> n.replaceAll("^\"|\"$", "")).orElse("");
            String[] splits = cn.split(",", 3);

            if (splits.length > 2) {
                if (serial == null) {
                    serial = splits[2];
                }
                if (first == null) {
                    first = splits[1];
                }
                if (last == null) {
                    last = splits[0];
                }
            }
        }

        String name = Stream.of(first, last) //
                .filter(s -> s != null) //
                .map(String::trim) //
                .filter(s -> !s.isEmpty()) //
                .collect(Collectors.joining(" "));

        return new Subject(serial, name);
    }

    /**
     * DataFileImpl is implementation of DataFile specific to digidoc4j library.
     */
    static class DataFileImpl implements DataFile {

        private final org.digidoc4j.DataFile df;

        DataFileImpl(org.digidoc4j.DataFile df) {
            this.df = df;
        }

        @Override
        public String getName() {
            return df.getName();
        }

        @Override
        public InputStream getStream() {
            return df.getStream();
        }

    }
}
