package com.star.admin.web;

/**
 * 超级管理员后台侧栏菜单项（由各业务模块通过 {@link AdminPageModule} 注册）。
 *
 * @param menuKey  菜单唯一键，与页面 {@code activeMenu} 对应以高亮当前项
 * @param category 侧栏分组标题（如「系统管理」「教育资源」）
 * @param label    菜单显示名称
 * @param path     页面路径（如 {@code /admin/ed/attachments}）
 * @param icon     BoxIcons 类名（如 {@code bx bx-folder}）
 * @param order    同分组内排序，越小越靠前
 */
public record AdminMenuItem(
        String menuKey,
        String category,
        String label,
        String path,
        String icon,
        int order
) {
}
