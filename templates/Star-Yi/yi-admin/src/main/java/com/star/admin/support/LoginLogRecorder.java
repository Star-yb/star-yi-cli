package com.star.admin.support;

import com.star.admin.model.entity.LoginLogs;
import com.star.admin.model.entity.LoginLogsDraft;
import jakarta.annotation.Nullable;
import jakarta.servlet.http.HttpServletRequest;
import lombok.RequiredArgsConstructor;
import org.babyfish.jimmer.sql.JSqlClient;
import org.babyfish.jimmer.sql.ast.mutation.SaveMode;
import org.springframework.stereotype.Component;

import java.time.LocalDateTime;

/**
 * API 登录日志写入器：直接通过 Jimmer 落库，不经过 Dao/Service 分层。
 */
@Component
@RequiredArgsConstructor
public class LoginLogRecorder {

    public static final int TYPE_LOGIN = 1;
    public static final int TYPE_LOGOUT = 2;
    public static final int STATUS_SUCCESS = 0;
    public static final int STATUS_FAILURE = 1;

    private final JSqlClient sql;

    public void recordLoginSuccess(
            HttpServletRequest request,
            long userId,
            String username,
            @Nullable String token
    ) {
        save(request, userId, username, TYPE_LOGIN, STATUS_SUCCESS, null, token);
    }

    public void recordLoginFailure(
            HttpServletRequest request,
            String username,
            @Nullable Long userId,
            String failReason
    ) {
        save(request, userId, username, TYPE_LOGIN, STATUS_FAILURE, failReason, null);
    }

    public void recordLogout(
            HttpServletRequest request,
            long userId,
            String username,
            @Nullable String token
    ) {
        save(request, userId, username, TYPE_LOGOUT, STATUS_SUCCESS, null, token);
    }

    private void save(
            HttpServletRequest request,
            @Nullable Long userId,
            String username,
            int loginType,
            int status,
            @Nullable String failReason,
            @Nullable String token
    ) {
        ClientInfo client = ClientInfo.from(request);
        LoginLogs log = LoginLogsDraft.$.produce(draft -> {
            if (userId != null) {
                draft.setUserId(userId);
            }
            draft.setUsername(username);
            draft.setLoginType(loginType);
            draft.setStatus(status);
            draft.setFailReason(failReason);
            draft.setIpAddress(client.ip());
            draft.setBrowser(client.browser());
            draft.setOs(client.os());
            draft.setUserAgent(client.userAgent());
            draft.setTokenValue(maskToken(token));
            draft.setLoginTime(LocalDateTime.now());
        });
        sql.getEntities()
                .saveCommand(log)
                .setMode(SaveMode.INSERT_ONLY)
                .execute();
    }

    @Nullable
    private static String maskToken(@Nullable String token) {
        if (token == null || token.isBlank()) {
            return null;
        }
        if (token.length() <= 12) {
            return token;
        }
        return token.substring(0, 8) + "****";
    }

    private record ClientInfo(
            @Nullable String ip,
            @Nullable String browser,
            @Nullable String os,
            @Nullable String userAgent
    ) {
        static ClientInfo from(HttpServletRequest request) {
            if (request == null) {
                return new ClientInfo(null, null, null, null);
            }
            String ua = request.getHeader("User-Agent");
            return new ClientInfo(resolveIp(request), parseBrowser(ua), parseOs(ua), ua);
        }

        private static String resolveIp(HttpServletRequest request) {
            String ip = request.getHeader("X-Forwarded-For");
            if (ip != null && !ip.isBlank()) {
                int comma = ip.indexOf(',');
                return comma > 0 ? ip.substring(0, comma).trim() : ip.trim();
            }
            ip = request.getHeader("X-Real-IP");
            if (ip != null && !ip.isBlank()) {
                return ip.trim();
            }
            return request.getRemoteAddr();
        }

        @Nullable
        private static String parseBrowser(@Nullable String ua) {
            if (ua == null) {
                return null;
            }
            if (ua.contains("Edg/")) {
                return "Edge";
            }
            if (ua.contains("Chrome/") && !ua.contains("Edg/")) {
                return "Chrome";
            }
            if (ua.contains("Firefox/")) {
                return "Firefox";
            }
            if (ua.contains("Safari/") && !ua.contains("Chrome/")) {
                return "Safari";
            }
            return "Other";
        }

        @Nullable
        private static String parseOs(@Nullable String ua) {
            if (ua == null) {
                return null;
            }
            if (ua.contains("Windows")) {
                return "Windows";
            }
            if (ua.contains("Mac OS")) {
                return "macOS";
            }
            if (ua.contains("Android")) {
                return "Android";
            }
            if (ua.contains("iPhone") || ua.contains("iPad")) {
                return "iOS";
            }
            if (ua.contains("Linux")) {
                return "Linux";
            }
            return "Other";
        }
    }
}
