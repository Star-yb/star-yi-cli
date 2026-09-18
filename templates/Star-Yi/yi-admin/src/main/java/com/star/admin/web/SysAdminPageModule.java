package com.star.admin.web;

import org.springframework.stereotype.Component;

import java.util.List;

/**
 * yi-admin 内置系统管理菜单（用户、角色、权限等）。
 */
@Component
public class SysAdminPageModule implements AdminPageModule {

    @Override
    public String moduleId() {
        return "yi-admin-sys";
    }

    @Override
    public int moduleOrder() {
        return 0;
    }

    @Override
    public List<AdminMenuItem> menuItems() {
        return List.of(
                item("user", "系统管理", "用户管理", "/admin/user", "bx bx-user", 10),
                item("role", "系统管理", "角色管理", "/admin/role", "bx bx-shield-quarter", 20),
                item("permission", "系统管理", "权限管理", "/admin/permission", "bx bx-lock-alt", 30),
                item("login-log", "系统监控", "登录日志", "/admin/login-log", "bx bx-log-in-circle", 10),
                item("api-docs", "开发工具", "接口文档", "/admin/api-docs", "bx bx-book-open", 10)
        );
    }

    private static AdminMenuItem item(
            String menuKey, String category, String label, String path, String icon, int order) {
        return new AdminMenuItem(menuKey, category, label, path, icon, order);
    }
}
