package com.star.admin.common.satoken;

import cn.dev33.satoken.SaManager;
import cn.dev33.satoken.stp.StpInterface;
import cn.dev33.satoken.stp.StpUtil;
import com.star.admin.model.entity.PermissionsTable;
import com.star.admin.model.entity.RolesTable;
import com.star.admin.model.entity.UsersTable;    // 假设用户实体类为 UsersTable
import org.babyfish.jimmer.sql.JSqlClient;
import org.springframework.stereotype.Component;

import java.util.ArrayList;
import java.util.List;
import java.util.Objects;
import java.util.UUID;

@Component
public class StpInterfaceImpl implements StpInterface {

    private final JSqlClient sql;

    public StpInterfaceImpl(JSqlClient sql) {
        this.sql = sql;
    }

    /**
     * 返回一个账号所拥有的权限码集合
     */
    @Override
    @SuppressWarnings("unchecked")
    public List<String> getPermissionList(Object loginId, String loginType) {
        List<String> permissions = new ArrayList<>();

        // 1. 先拿到角色列表
        List<String> roleList = getRoleList(loginId, loginType);
        if (roleList.isEmpty()) {
            return permissions;
        }

        // 2. 如果是超级管理员，直接返回所有权限（通配符）
        if (roleList.contains("*")) {
            permissions.add("*");
            return permissions;
        }

        PermissionsTable p = PermissionsTable.$;

        // 3. 遍历角色，按角色维度做缓存
        for (String roleCode : roleList) {
            String cacheKey = "satoken:role-find-permission:" + roleCode;
            List<String> permissionList =
                    (List<String>) SaManager.getSaTokenDao().getObject(cacheKey);

            if (permissionList == null) {
                permissionList = sql.createQuery(p)
                        .where(p.roles(r -> r.roleCode().eq(roleCode)))
                        .select(p.permissionCode())
                        .distinct()
                        .execute();

                SaManager.getSaTokenDao().setObject(
                        cacheKey,
                        permissionList,
                        60L * 60 * 24 * 30
                );
            }

            if (permissionList != null && !permissionList.isEmpty()) {
                permissions.addAll(permissionList);
            }
        }

        return permissions;
    }

    /**
     * 返回一个账号所拥有的角色标识集合
     */
    @Override
    @SuppressWarnings("unchecked")
    public List<String> getRoleList(Object loginId, String loginType) {
//        UUID userId = UUID.fromString(Objects.toString(loginId));
        Long userId = Long.parseLong(loginId.toString()); // 关键修复点


        // 0. 先判断用户是否为超级管理员
        if (isSuperAdmin(userId)) {
            // 超级管理员固定返回 "*" 角色
            List<String> superRole = List.of("*");
            // 可选：将结果缓存，避免重复查询用户表
            String cacheKey = "satoken:loginId-find-role:" + loginId;
            SaManager.getSaTokenDao().setObject(cacheKey, superRole, 60L * 60 * 24 * 30);
            return superRole;
        }

        String cacheKey = "satoken:loginId-find-role:" + loginId;
        List<String> roleList = (List<String>) SaManager.getSaTokenDao().getObject(cacheKey);
        if (roleList != null) {
            return roleList;
        }

        RolesTable r = RolesTable.$;
        roleList = sql.createQuery(r)
                .where(r.users(u -> u.id().eq(userId)))
                .select(r.roleCode())
                .distinct()
                .execute();

        SaManager.getSaTokenDao().setObject(
                cacheKey,
                roleList,
                60L * 60 * 24 * 30
        );

        return roleList;
    }

    /**
     * 判断用户是否为超级管理员
     */
    private boolean isSuperAdmin(Long userId) {
        UsersTable u = UsersTable.$;
        // 查询 superAdmin 字段，假设 1 表示超级管理员，0 表示普通用户
        Integer superAdminValue = sql.createQuery(u)
                .where(u.id().eq(userId))
                .select(u.superAdmin())
                .fetchOptional()
                .orElse(0);
        return superAdminValue == 1;
    }
}