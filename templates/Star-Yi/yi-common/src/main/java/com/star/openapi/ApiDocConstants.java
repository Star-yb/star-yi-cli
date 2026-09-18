package com.star.openapi;

/**
 * OpenAPI / Swagger 文档常量，与 {@code sa-token.token-name} 保持一致。
 */
public final class ApiDocConstants {

    private ApiDocConstants() {
    }

    /** Swagger Authorize 与请求头名称（satoken） */
    public static final String SECURITY_SCHEME_SATOKEN = "satoken";
}
