package com.star.admin.support;

import cn.dev33.satoken.stp.StpUtil;
import cn.dev33.satoken.stp.parameter.SaLoginParameter;

/**
 * Sa-Token 登录会话：API 走请求头，Thymeleaf 后台走 Cookie，避免同浏览器 Cookie 互相覆盖。
 *
 * <p>要点：</p>
 * <ul>
 *   <li>{@code createLoginSession} 只创建服务端会话并返回 token，不会调用 {@code setTokenValue}，因此不会写 Cookie/响应头</li>
 *   <li>{@code StpUtil.login} = {@code createLoginSession} + {@code setTokenValue}，在全局 {@code is-read-cookie=true} 时会写 Cookie</li>
 *   <li>{@code setIsWriteHeader} 仅控制是否写入<strong>响应头</strong>，与 Cookie 无关</li>
 * </ul>
 *
 * @see <a href="https://sa-token.cc/doc.html#/up/not-cookie">Sa-Token 非 Cookie 模式</a>
 */
public final class SaAuthLoginSupport {

    /** 前后端分离 SPA / REST API */
    public static final String DEVICE_API = "api";

    /** Thymeleaf 超级管理员后台 */
    public static final String DEVICE_ADMIN_WEB = "admin-web";

    private SaAuthLoginSupport() {
    }

    /**
     * API 登录：仅创建会话，token 由接口 JSON 返回，前端放入 {@code satoken} 请求头。
     */
    public static String establishApiSession(long userId) {
        SaLoginParameter param = SaLoginParameter.create()
                .setDeviceType(DEVICE_API)
                // 覆盖全局 is-share:true，每次登录独立 token，避免多端共用一个 token 互相影响
                .setIsShare(false);
        // return StpUtil.getStpLogic().createLoginSession(userId, param);
        return StpUtil.createLoginSession(userId, param);
    }

    /**
     * Thymeleaf 后台登录：完整 login 流程，由框架将 token 写入 Cookie（依赖全局 is-read-cookie）。
     */
    public static void establishAdminWebSession(long userId) {
        StpUtil.login(userId, SaLoginParameter.create()
                .setDeviceType(DEVICE_ADMIN_WEB)
                .setIsLastingCookie(true)
                // 与 API 端隔离，避免 is-share:true 时复用/共享 token
                .setIsShare(false)
                // 后台走 Cookie，不需要把 token 再写到响应头
                .setIsWriteHeader(false));
    }
}
