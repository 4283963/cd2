package com.tunnel.emergency.entity;

public enum EscapeLightStatus {
    OFF(0),
    ON(1),
    BLINK(2);

    private final int code;

    EscapeLightStatus(int code) {
        this.code = code;
    }

    public int getCode() {
        return code;
    }

    public static EscapeLightStatus fromCode(int code) {
        for (EscapeLightStatus status : values()) {
            if (status.code == code) {
                return status;
            }
        }
        return OFF;
    }
}
