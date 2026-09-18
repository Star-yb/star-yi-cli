package com.star.utils;

import cn.dev33.satoken.secure.SaSecureUtil;

/**
 * 认证相关的统一加密工具。
 * 统一封装后，后续切换算法时只需改这一处。
 */
public final class AuthCryptoUtils {

    private AuthCryptoUtils() {
    }

    /**
     * 使用 Sa-Token 的 SHA-256 进行摘要。
     */
    public static String hash(String plainText) {
        return SaSecureUtil.sha256(plainText);
    }

    /**
     * 校验明文与摘要是否匹配。
     */
    public static boolean matches(String plainText, String hashedText) {
        if (plainText == null || hashedText == null) {
            return false;
        }
        return hash(plainText).equals(hashedText);
    }
}
