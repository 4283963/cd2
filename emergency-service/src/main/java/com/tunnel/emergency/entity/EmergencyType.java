package com.tunnel.emergency.entity;

public enum EmergencyType {
    NONE(0),
    ACCIDENT(1),
    FIRE(2),
    FAULT(3);

    private final int code;

    EmergencyType(int code) {
        this.code = code;
    }

    public int getCode() {
        return code;
    }

    public static EmergencyType fromCode(int code) {
        for (EmergencyType type : values()) {
            if (type.code == code) {
                return type;
            }
        }
        return NONE;
    }
}
