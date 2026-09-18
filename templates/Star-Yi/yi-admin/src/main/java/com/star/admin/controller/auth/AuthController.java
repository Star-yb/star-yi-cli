package com.star.admin.controller.auth;

import cn.dev33.satoken.stp.StpUtil;
import cn.dev33.satoken.util.SaResult;
import com.star.admin.model.dto.LoginInput;
import com.star.admin.model.dto.LoginView;
import com.star.admin.model.dto.VerifyInput;
import com.star.admin.model.dto.VerifyView;
import com.star.admin.model.entity.Users;
import com.star.admin.service.AuthService;
import com.star.admin.support.LoginLogRecorder;
import com.star.admin.support.SaAuthLoginSupport;
import com.star.common.exception.BusinessException;
import jakarta.servlet.http.HttpServletRequest;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.tags.Tag;
import lombok.RequiredArgsConstructor;
import org.springframework.web.bind.annotation.*;

@Tag(name = "认证", description = "登录、登出与 Token 校验，路径 /auth")
@RestController
@RequestMapping("/auth")
@RequiredArgsConstructor
public class AuthController {

    private final AuthService authService;
    private final LoginLogRecorder loginLogRecorder;
    private final HttpServletRequest request;

    @Operation(summary = "用户登录", description = "免鉴权；成功返回 satoken 及用户信息")
    @PostMapping("login")
    public SaResult login(@RequestBody LoginInput users) {
        String username = users.getUsername();
        try {
            Users login = authService.login(username, users.getPassword());
            String token = SaAuthLoginSupport.establishApiSession(login.id());
            authService.onLoginSuccess(login.id());
            loginLogRecorder.recordLoginSuccess(request, login.id(), username, token);

            LoginView loginView = new LoginView(login);
            loginView.setToken(token);
            loginView.setTokenTimeout(StpUtil.getTokenTimeout(token));
            return SaResult.data(loginView);
        } catch (BusinessException e) {
            Long userId = authService.resolveUserIdByLoginAccount(username);
            loginLogRecorder.recordLoginFailure(request, username, userId, e.getMessage());
            throw e;
        }
    }

    @Operation(
            summary = "用户登出",
            description = "需携带 satoken；仅注销当前请求的 token，不影响同账号其他端会话，也不影响 Thymeleaf 超管 Cookie"
    )
    @GetMapping("logout")
    public SaResult logout() {
        if (StpUtil.isLogin()) {
            long userId = StpUtil.getLoginIdAsLong();
            String username = authService.resolveLoginUsername(userId);
            String tokenValue = StpUtil.getTokenValue();
            loginLogRecorder.recordLogout(request, userId, username, tokenValue);
            // 按 token 注销：is-share=false 时同账号多端各持独立 token，一端退出不得连带其他端。
            // 切勿使用 StpUtil.logout(loginId, device)：那会踢掉该账号该 deviceType 下全部会话。
            if (tokenValue != null && !tokenValue.isBlank()) {
                StpUtil.logoutByTokenValue(tokenValue);
            }
        }
        return SaResult.ok();
    }

    @Operation(summary = "校验 Token", description = "免鉴权；根据 token 判断是否过期及关联用户")
    @PostMapping("verify")
    public SaResult validateToken(@RequestBody VerifyInput input) {
        String tokenValue = input.getToken() != null ? input.getToken() : "";

        VerifyView verifyView = new VerifyView();
        verifyView.setToken(tokenValue);

        if (tokenValue.isBlank()) {
            verifyView.setExpired(true);
            verifyView.setTokenTimeout(0L);
            verifyView.setUserId("");
            verifyView.setUserStatus(false);
            return SaResult.data(verifyView); 
        }

        Object loginId = StpUtil.getLoginIdByToken(tokenValue);
        boolean expired = loginId == null;

        verifyView.setTokenId(loginId);
        verifyView.setTokenTimeout(StpUtil.getTokenTimeout(tokenValue));
        verifyView.setExpired(expired);

        if (loginId != null) {
            long userId = Long.parseLong(loginId.toString());
            verifyView.setUserId(String.valueOf(userId));
            verifyView.setUserStatus(authService.isUserActive(userId));
        } else {
            verifyView.setUserId("");
            verifyView.setUserStatus(false);
        }

        return SaResult.data(verifyView);
    }
}
