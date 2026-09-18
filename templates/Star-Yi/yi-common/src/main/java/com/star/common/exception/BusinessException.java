package com.star.common.exception;

import lombok.Getter;

/**
 * 业务异常类
 * 用于抛出带有特定错误码和消息的业务异常
 * 支持链式调用设置错误码，默认错误码为500
 */
@Getter
public class BusinessException extends RuntimeException {

    private int code = 500; // 默认错误码为500

    public BusinessException(String message) {
        super(message);
    }

    public BusinessException(String message, Throwable cause) {
        super(message, cause);
    }

    /**
     * 设置错误码
     *
     * @param code 错误码
     * @return 当前BusinessException实例，支持链式调用
     */
    public BusinessException code(int code) {
        this.code = code;
        return this;
    }

}
