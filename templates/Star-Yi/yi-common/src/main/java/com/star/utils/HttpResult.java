package com.star.utils;

import java.util.Objects;

/**
 * OkHttp 同步请求的简要结果。
 */
public final class HttpResult {

    private final int statusCode;
    private final String body;

    public HttpResult(int statusCode, String body) {
        this.statusCode = statusCode;
        this.body = body;
    }

    public int statusCode() {
        return statusCode;
    }

    public String body() {
        return body;
    }

    public boolean isSuccessful() {
        return statusCode >= 200 && statusCode < 300;
    }

    @Override
    public boolean equals(Object o) {
        if (this == o) {
            return true;
        }
        if (o == null || getClass() != o.getClass()) {
            return false;
        }
        HttpResult that = (HttpResult) o;
        return statusCode == that.statusCode && Objects.equals(body, that.body);
    }

    @Override
    public int hashCode() {
        return Objects.hash(statusCode, body);
    }

    @Override
    public String toString() {
        return "HttpResult{" + "statusCode=" + statusCode + ", body='" + body + '\'' + '}';
    }
}
