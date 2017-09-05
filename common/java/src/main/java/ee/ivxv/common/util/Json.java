package ee.ivxv.common.util;

import com.fasterxml.jackson.core.JsonGenerator;
import com.fasterxml.jackson.core.JsonParser;
import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.DeserializationContext;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.SerializerProvider;
import com.fasterxml.jackson.databind.deser.std.StdDeserializer;
import com.fasterxml.jackson.databind.module.SimpleModule;
import com.fasterxml.jackson.databind.ser.std.StdSerializer;
import java.io.IOException;
import java.io.InputStream;
import java.io.OutputStream;
import java.nio.file.Files;
import java.nio.file.Path;
import java.time.Instant;

/**
 * Json is a class for reading and writing JSON format.
 */
public class Json {

    /**
     * Reads the specified path as JSON file and parses the result into new instance of the
     * specified type.
     * 
     * @param path Location of a JSON file.
     * @param type The target class to parse the result into.
     * @return Instance of the target type.
     * @throws Exception if an i/o or parsing error occurs.
     */
    public static <T> T read(Path path, Class<T> type) throws Exception {
        return read(Files.newInputStream(path), type);
    }

    public static <T> T read(InputStream in, Class<T> type) throws Exception {
        ObjectMapper mapper = getMapper();

        return mapper.readValue(in, type);
    }

    /**
     * Writes the specified object in JSON format at the specified path. All parent folders of the
     * specified path will be created.
     * 
     * @param o The object to write
     * @param path Location of the JSON file
     * @throws Exception if an i/o or parsing error occurs.
     */
    public static void write(Object o, Path path) throws Exception {
        if (path.getParent() != null) {
            Files.createDirectories(path.getParent());
        }
        write(o, Files.newOutputStream(path));
    }

    public static void write(Object o, OutputStream out) throws Exception {
        ObjectMapper mapper = getMapper();

        // TODO Using pretty print for debugging
        mapper.writerWithDefaultPrettyPrinter().writeValue(out, o);
    }

    /**
     * @return Creates and returns a mapper that is properly set up for serialization and
     *         deserialization as required in this project.
     */
    private static ObjectMapper getMapper() {
        ObjectMapper mapper = new ObjectMapper();
        SimpleModule module = new SimpleModule();

        module.addSerializer(Instant.class, new InstantSerializer());
        module.addDeserializer(Instant.class, new InstantDeserializer());

        mapper.registerModule(module);

        return mapper;
    }

    private static class InstantSerializer extends StdSerializer<Instant> {
        private static final long serialVersionUID = 4737351183174945801L;

        protected InstantSerializer() {
            super(Instant.class);
        }

        @Override
        public void serialize(Instant value, JsonGenerator gen, SerializerProvider provider)
                throws IOException {
            gen.writeString(value.toString());
        }
    }

    private static class InstantDeserializer extends StdDeserializer<Instant> {
        private static final long serialVersionUID = 4737351183174945801L;

        protected InstantDeserializer() {
            super(Instant.class);
        }

        @Override
        public Instant deserialize(JsonParser p, DeserializationContext ctxt)
                throws IOException, JsonProcessingException {
            return Instant.parse(p.getText());
        }
    }

}
