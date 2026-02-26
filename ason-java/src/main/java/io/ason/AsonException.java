package io.ason;

public class AsonException extends RuntimeException {
    public AsonException(String message) { super(message); }
    public AsonException(String message, Throwable cause) { super(message, cause); }
}
