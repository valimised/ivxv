package ee.ivxv.common.service.bbox;

/**
 * Ref represents logical reference to an entry in a either ballot box or registration data. Used in
 * error reporting.
 */
public interface Ref {

    class BbRef implements Ref {
        public final String voter;
        public final String ballot;

        public BbRef(String voter, String ballot) {
            this.voter = voter;
            this.ballot = ballot;
        }

        @Override
        public int hashCode() {
            return (voter == null ? 0 : voter.hashCode()) * 31 //
                    + (ballot == null ? 0 : ballot.hashCode());
        }

        @Override
        public boolean equals(Object obj) {
            if (this == obj) {
                return true;
            }
            if (obj == null || getClass() != obj.getClass()) {
                return false;
            }
            BbRef o = (BbRef) obj;
            if (ballot == null && o.ballot != null || ballot != null && !ballot.equals(o.ballot)) {
                return false;
            }
            if (voter == null && o.voter != null || voter != null && !voter.equals(o.voter)) {
                return false;
            }
            return true;
        }
    }

    class RegRef implements Ref {
        public final String ref;

        public RegRef(String ref) {
            this.ref = ref;
        }

        @Override
        public int hashCode() {
            return ref == null ? 0 : ref.hashCode();
        }

        @Override
        public boolean equals(Object obj) {
            if (this == obj) {
                return true;
            }
            if (obj == null || getClass() != obj.getClass()) {
                return false;
            }
            RegRef o = (RegRef) obj;
            if (ref == null && o.ref != null || ref != null && !ref.equals(o.ref)) {
                return false;
            }
            return true;
        }
    }

}
