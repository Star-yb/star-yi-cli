package com.star.admin.service;


import com.star.admin.model.entity.Users;
import jakarta.annotation.Nullable;
import org.springframework.stereotype.Service;

import java.util.List;

@Service
public interface AuthService {

    Users login(String username, String password);

    /** 登录会话建立成功后更新最后登录时间 */
    void onLoginSuccess(long userId);

    /** 用户是否处于可登录状态（{@code status == 0}） */
    boolean isUserActive(long userId);

    /**
     * 按登录账号（用户名或手机号）解析用户 ID，不存在则返回 null。
     */
    @Nullable
    Long resolveUserIdByLoginAccount(String account);

    /**
     * 解析登出日志等场景使用的登录账号展示名。
     */
    String resolveLoginUsername(long userId);

    /** 获取用户的所有权限编码列表 */
    List<String> getUserPermissionCodes(Long userId);

    /** 获取用户的所有角色编码列表 */
    List<String> getUserRoleCodes(Long userId);
}
