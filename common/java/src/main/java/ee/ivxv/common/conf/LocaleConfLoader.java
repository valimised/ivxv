package ee.ivxv.common.conf;

import ch.qos.cal10n.BaseName;
import ch.qos.cal10n.LocaleData;
import ee.ivxv.common.service.i18n.MessageException;
import ee.ivxv.common.util.Util;
import java.io.IOException;
import java.io.InputStream;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.List;
import java.util.Locale;
import java.util.Properties;
import java.util.stream.Stream;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

public class LocaleConfLoader {

    private static final Logger log = LoggerFactory.getLogger(LocaleConfLoader.class);

    private static final String LANG_PROPERTIES = "lang.properties";
    private static final String LANGS_KEY = "languages";

    /**
     * Loads locale configuration into the specified instance and returns it.
     * 
     * @param localeConf The destination instance to load the locale configuration into.
     * @return Returns <tt>localeConf</tt>.
     */
    public static LocaleConf load(LocaleConf localeConf) {
        log.info("Loading language properties from {}", LANG_PROPERTIES);
        try (InputStream in = Util.getResource(LANG_PROPERTIES)) {
            load(in, localeConf);
            return localeConf;
        } catch (MessageException e) {
            throw e;
        } catch (Exception e) {
            throw new MessageException(e, Msg.e_lang_conf_not_found, LANG_PROPERTIES);
        }
    }

    static LocaleConf load(InputStream langProperties, LocaleConf localeConf) throws IOException {
        Properties langProps = loadLangProperties(langProperties);
        List<Locale> locales = new ArrayList<>();

        if (!langProps.containsKey(LANGS_KEY)) {
            throw new MessageException(Msg.e_langs_property_missing, LANG_PROPERTIES, LANGS_KEY);
        }
        String[] langs = langProps.getProperty(LANGS_KEY).split(",");
        log.info("Loaded languages: {}", Arrays.asList(langs));
        Stream.of(langs).map(String::trim).filter(s -> !s.isEmpty())
                .forEach(s -> locales.add(new Locale(s)));

        if (locales.isEmpty()) {
            throw new MessageException(Msg.e_langs_empty, LANG_PROPERTIES, LANGS_KEY);
        }

        localeConf.setAllLocales(locales);

        return localeConf;
    }

    private static Properties loadLangProperties(InputStream langProperties) throws IOException {
        if (langProperties == null) {
            throw new MessageException(Msg.e_lang_conf_not_found, LANG_PROPERTIES);
        }

        Properties props = new Properties();

        props.load(langProperties);

        return props;
    }

    @BaseName("bootstrap-i18n.common-conf-localeconfloader-msg")
    @LocaleData(defaultCharset = "UTF-8", value = {})
    public enum Msg {
        e_lang_conf_not_found, e_langs_property_missing, e_langs_empty
    }
}
