package com.star.yidemo.admin;

import com.star.admin.web.AdminMenuItem;
import com.star.admin.web.AdminPageModule;
import org.springframework.stereotype.Component;

import java.util.List;

/**
 * yi-demo 向超级管理员后台注册演示菜单。
 */
@Component
public class DemoArticleAdminPageModule implements AdminPageModule {

    @Override
    public String moduleId() {
        return "yi-demo";
    }

    @Override
    public int moduleOrder() {
        return 90;
    }

    @Override
    public List<AdminMenuItem> menuItems() {
        return List.of(
                new AdminMenuItem(
                        "demo-article",
                        "开发演示",
                        "演示文章",
                        "/admin/demo/articles",
                        "bx bx-news",
                        10
                )
        );
    }
}
