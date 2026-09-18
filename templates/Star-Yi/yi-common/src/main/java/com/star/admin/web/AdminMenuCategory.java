package com.star.admin.web;

import java.util.List;

/**
 * 侧栏菜单分组（按 category 聚合后的展示结构）。
 */
public record AdminMenuCategory(String name, List<AdminMenuItem> items) {
}
