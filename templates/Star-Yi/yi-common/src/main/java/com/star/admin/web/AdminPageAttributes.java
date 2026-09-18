package com.star.admin.web;

import org.springframework.ui.Model;

/**
 * 超级管理员页面 Model 公共属性填充工具。
 */
public final class AdminPageAttributes {

    private AdminPageAttributes() {
    }

    /**
     * 填充页面标题、面包屑与侧栏高亮键。
     */
    public static void enrich(Model model, String activeMenu, String pageCategory, String pageName) {
        model.addAttribute("activeMenu", activeMenu);
        model.addAttribute("pageCategory", pageCategory);
        model.addAttribute("pageName", pageName);
    }
}
