package com.star.admin.web;

import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Component;

import java.util.ArrayList;
import java.util.Comparator;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;

/**
 * 聚合所有 {@link AdminPageModule} 注册的超级管理员侧栏菜单。
 */
@Slf4j
@Component
public class AdminMenuRegistry {

    private final List<AdminPageModule> modules;

    public AdminMenuRegistry(List<AdminPageModule> modules) {
        this.modules = modules == null ? List.of() : modules;
        this.modules.forEach(m ->
                log.info("已注册超级管理员页面模块: {} (order={})", m.moduleId(), m.moduleOrder()));
    }

    /**
     * 按分组返回排序后的侧栏菜单（供 Thymeleaf 侧栏渲染）。
     */
    public List<AdminMenuCategory> getMenuCategories() {
        List<AdminMenuItem> flat = modules.stream()
                .sorted(Comparator.comparingInt(AdminPageModule::moduleOrder)
                        .thenComparing(AdminPageModule::moduleId))
                .flatMap(m -> m.menuItems().stream())
                .sorted(Comparator.comparing(AdminMenuItem::order)
                        .thenComparing(AdminMenuItem::label))
                .toList();

        Map<String, List<AdminMenuItem>> grouped = new LinkedHashMap<>();
        for (AdminMenuItem item : flat) {
            grouped.computeIfAbsent(item.category(), k -> new ArrayList<>()).add(item);
        }

        return grouped.entrySet().stream()
                .map(e -> new AdminMenuCategory(e.getKey(), List.copyOf(e.getValue())))
                .toList();
    }
}
