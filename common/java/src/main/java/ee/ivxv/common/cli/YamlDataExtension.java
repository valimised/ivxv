package ee.ivxv.common.cli;

import java.util.List;
import java.util.Arrays;
import java.util.stream.Collectors;

class YamlDataExtension {

    private static final String VOTERLISTS_DIR_TAG = "voterlists_dir"; // const
    private static final String DIRECTORY_TRAILING_SLASH = "/"; // const
    private static String voterlists_dir_name = ""; // var


    /**
     * Get private variable "voterlists_dir_name". Default value for "voterlists_dir_name" = ""
     *
     * @return private variable "voterlists_dir_name" or "" as a default value
     */
    static String getVoterlistsDirName() {
        return voterlists_dir_name;
    }

    /**
     * Check whether "tagName" matches "VOTERLISTS_DIR_TAG"
     *
     * @param tagName Yaml tag name
     * @return true if matches, otherwise - false
     */
    static boolean isVoterListsDirTag(String tagName) {
        return tagName.equals((VOTERLISTS_DIR_TAG));
    }

    /**
     * Set private variable "voterlists_dir_name". Default value for "voterlists_dir_name" = "".
     * If Exception is thrown - do nothing and let "processor" application handle that case
     *
     * @param tagValue Yaml tag value
     */
    static void setVoterListsDirName(Object tagValue) {
        try {
            voterlists_dir_name = (String) tagValue + DIRECTORY_TRAILING_SLASH;
        } catch (Exception ex) {
            YamlData.log.error(getExceptionStackAsString(ex));
        }
    }

    /**
     * Get Exception's stack trace as a human-readable String
     *
     * @param ex Exception
     * @return Exception's stack trace as a String
     */
    public static String getExceptionStackAsString(Exception ex) {
        List<?> exStack = (Arrays.stream(ex.getStackTrace()).collect(Collectors.toList()));
        return exStack.stream().map(x -> x.toString() + "\n").toString();
    }
}
