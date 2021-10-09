package ee.ivxv.common.util;

import ee.ivxv.common.M;
import ee.ivxv.common.model.SkipCommand;
import ee.ivxv.common.service.i18n.MessageException;
import java.io.InputStream;
import java.util.Map;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.yaml.snakeyaml.Yaml;

public class SkipCommandUtil {

    private static final Logger log = LoggerFactory.getLogger(SkipCommandUtil.class);

    public static SkipCommand readSkipCommand(InputStream in) {

        Yaml yaml = new Yaml();
        Map<String, Object> obj = (Map<String, Object>)yaml.load(in);
        return new SkipCommand(
                String.valueOf(obj.get("changeset")),
                String.valueOf(obj.get("election")),
                String.valueOf(obj.get("skip_voter_list")));
    }

}
