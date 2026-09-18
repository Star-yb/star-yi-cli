package com.star.admin.web;

import jakarta.servlet.http.HttpServletRequest;
import org.springframework.ui.Model;
import org.springframework.web.bind.annotation.ControllerAdvice;
import org.springframework.web.bind.annotation.ModelAttribute;

/**
 * 为所有超级管理员页面注入动态侧栏菜单数据。
 */
@ControllerAdvice
public class AdminMenuModelAdvice {

    private final AdminMenuRegistry adminMenuRegistry;

    public AdminMenuModelAdvice(AdminMenuRegistry adminMenuRegistry) {
        this.adminMenuRegistry = adminMenuRegistry;
    }

    @ModelAttribute
    public void enrichAdminMenus(HttpServletRequest request, Model model) {
        String uri = request.getRequestURI();
        if (uri == null || !uri.startsWith("/admin")) {
            return;
        }
        if (uri.equals("/admin/login") || uri.startsWith("/admin/login?")) {
            return;
        }
        model.addAttribute("adminMenuCategories", adminMenuRegistry.getMenuCategories());
    }
}
