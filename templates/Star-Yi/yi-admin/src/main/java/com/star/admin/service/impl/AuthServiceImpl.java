package com.star.admin.service.impl;

import com.star.admin.dao.PermissionsDao;
import com.star.admin.dao.RolesDao;
import com.star.admin.dao.UsersDao;
import com.star.admin.model.dto.LoginInput;
import com.star.admin.model.entity.*;
import com.star.admin.service.AuthService;
import com.star.utils.AuthCryptoUtils;
import com.star.common.exception.BusinessException;
import lombok.RequiredArgsConstructor;
import org.babyfish.jimmer.sql.ast.Predicate;
import org.springframework.stereotype.Service;

import jakarta.annotation.Nullable;

import java.util.List;


@Service
@RequiredArgsConstructor
public class AuthServiceImpl implements AuthService {
    private final UsersDao usersDao;
    private final RolesDao rolesDao;
    private final PermissionsDao permissionsDao;
    private static final UsersTable T = UsersTable.$;


    @Override
    public Users login(String username, String password) {

        List<Users> usersList = usersDao.queryList(
                Predicate.or(
                        T.username().eq(username),
                        T.phone().eq(username)
                ),
                f -> f.fetch(LoginInput.METADATA.getFetcher())
        );
        if (usersList.isEmpty()) {
            throw new BusinessException("账号不存在").code(401);
        }
//        对比密码
        if (!AuthCryptoUtils.matches(password, usersList.getFirst().password())) {
            throw new BusinessException("账号或密码错误").code(401);
        }
        Users user = usersList.getFirst();

        if (user.status() != 0) {
            throw new BusinessException("账号已禁用").code(401);
        }
        return user;
    }

    @Override
    public void onLoginSuccess(long userId) {
        usersDao.updateLastLoginTime(userId);
    }

    @Override
    public boolean isUserActive(long userId) {
        Users user = usersDao.findById(userId);
        return user != null && user.status() == 0;
    }

    @Override
    @Nullable
    public Long resolveUserIdByLoginAccount(String account) {
        return usersDao.queryList(
                Predicate.or(T.username().eq(account), T.phone().eq(account)),
                UsersTable::id
        ).stream().findFirst().orElse(null);
    }

    @Override
    public String resolveLoginUsername(long userId) {
        Users user = usersDao.findById(userId);
        return user != null ? user.username() : String.valueOf(userId);
    }

    @Override
    public List<String> getUserPermissionCodes(Long userId) {
        return permissionsDao.queryList(
                PermissionsTable.$.roles(r -> r.users(
                        u -> u.id().eq(userId)
                )),
                PermissionsTable::permissionCode
        );


        // 直接复用 Sa-Token 的权限获取逻辑（内部会调用自定义的 StpInterfaceImpl）
//        return StpUtil.getPermissionList(userId);
    }

    @Override
    public List<String> getUserRoleCodes(Long userId) {

        // 查询用户详情时，使用 UsersView 中的角色信息
        return rolesDao.queryList(
                RolesTable.$.users(u -> u.id().eq(userId)),
                RolesTable::roleCode
        );


        // 直接复用 Sa-Token 的角色获取逻辑（内部会调用自定义的 StpInterfaceImpl）
//        return StpUtil.getRoleList(userId);
    }
}
