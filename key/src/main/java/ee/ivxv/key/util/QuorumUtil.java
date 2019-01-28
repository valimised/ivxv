package ee.ivxv.key.util;

import ee.ivxv.common.service.smartcard.IndexedBlob;
import java.util.ArrayList;
import java.util.HashSet;
import java.util.List;
import java.util.Set;

public class QuorumUtil {

    /**
     * <pre>
     * {@literal
     * quorumSize == count():
     *  result list size is 1 and contains the object this was called on
     * quorumSize < count():
     *  creates 3 quorums such that all blobs are at least in one quorum
     *  but none are in every quorum.
     *  1st quorum: first n blobs among all cards
     *  2nd quorum: last n blobs among all cards
     *  3rd quorum: first n-1 blobs and the last blob among all blobs
     *  where n is parameter quorumSize.
     *  }
     * </pre>
     *
     * @param blobs List of blobs
     * @param quorumSize number of blobs in a single quorum
     * @throws IllegalArgumentException if {@code quorumSize > count()}
     */
    static public List<Set<IndexedBlob>> getQuorumList(List<IndexedBlob> blobs, int quorumSize) {
        int size = blobs.size();
        if (quorumSize > size) {
            throw new IllegalArgumentException("Quorum size has to smaller than total blob count");
        }
        List<Set<IndexedBlob>> res = new ArrayList<>();
        if (quorumSize == size) {
            res.add(new HashSet<>(blobs));
        } else {
            Set<IndexedBlob> q1 = new HashSet<>(blobs.subList(0, quorumSize));

            Set<IndexedBlob> q2 = new HashSet<>(blobs.subList(size - quorumSize, size));

            Set<IndexedBlob> q3 = new HashSet<>(blobs.subList(0, quorumSize - 1));
            q3.add(blobs.get(size - 1));

            res.add(q1);
            res.add(q2);
            res.add(q3);
        }
        return res;
    }
}
