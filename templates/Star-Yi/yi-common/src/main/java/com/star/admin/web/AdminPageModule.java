package com.star.admin.web;

import java.util.List;

/**
 * 超级管理员后台页面模块扩展点（插件式注册）。
 *
 * <p>各业务应用（如 {@code ed-attachment}、{@code ed-course}）实现本接口并注册为 Spring Bean，
 * {@code yi-admin} 启动时自动收集菜单项并渲染到统一侧栏。</p>
 *
 * <p>约定：</p>
 * <ul>
 *   <li>页面 Controller 路径以 {@code /admin/} 为前缀</li>
 *   <li>Thymeleaf 模板放在依赖 JAR 的 {@code classpath:/admin/<模块名>/} 下</li>
 *   <li>视图名与模板路径一致，如 {@code ed-attachment/index} → {@code admin/ed-attachment/index.html}</li>
 *   <li>页面骨架须与 {@code admin.css} 一致：{@code body > aside + main.main-wrapper > header + div.content-area}，
 *       参考 {@code yi-admin/templates/sys/login-log.html}</li>
 * </ul>
 */
public interface AdminPageModule {

    /**
     * 模块标识，用于日志与排序（建议与 Maven artifact 一致，如 {@code ed-attachment}）。
     */
    String moduleId();

    /**
     * 模块整体排序，越小越靠前（影响其菜单项在全局合并时的相对顺序）。
     */
    default int moduleOrder() {
        return 100;
    }

    /**
     * 本模块贡献的超级管理员侧栏菜单。
     */
    List<AdminMenuItem> menuItems();
}
