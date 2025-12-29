package io.github.athxx.ason;

/**
 * Exception thrown when parsing ASON fails.
 */
public class ParseException extends Exception {
    
    public ParseException(String message) {
        super(message);
    }
    
    public ParseException(String message, Throwable cause) {
        super(message, cause);
    }
}

