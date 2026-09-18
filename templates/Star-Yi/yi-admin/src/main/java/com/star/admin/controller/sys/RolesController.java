package com.star.admin.controller.sys;

import cn.dev33.satoken.util.SaResult;
import com.star.admin.model.dto.RoleCreateInputView;
import com.star.admin.model.dto.RolePermissionUpdateInputView;
import com.star.admin.model.dto.RoleSpecification;
import com.star.admin.model.dto.RoleUpdateInputView;
import com.star.admin.service.RolesService;
import com.star.common.page.PageQuery;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.tags.Tag;
import lombok.RequiredArgsConstructor;
import org.springframework.web.bind.annotation.*;

import java.util.UUID;

/**
 * 角色控制器
 */
@Tag(name = "角色管理", description = "角色 CRUD、启停与权限绑定，路径 /roles")
@RestController
@RequestMapping("/roles")
@RequiredArgsConstructor
public class RolesController {

    private final RolesService rolesService;

    @Operation(summary = "分页查询角色列表")
    @GetMapping
    public SaResult listRoles(
            PageQuery pageQuery,
            RoleSpecification roleSpecification) {
        return SaResult.data(rolesService.listRoles(pageQuery,roleSpecification));
    }

    @Operation(summary = "创建角色")
    @PostMapping
    public SaResult createRole(@RequestBody RoleCreateInputView role) {
        return SaResult.data(rolesService.createRole(role));
    }

    @Operation(summary = "更新角色")
    @PutMapping("/{id}")
    public SaResult updateRole(@PathVariable Long id, @RequestBody RoleUpdateInputView role) {
        role.setId(id);
        return SaResult.data(rolesService.updateRole(role));
    }

    @Operation(summary = "删除角色", description = "逻辑删除")
    @DeleteMapping("/{id}")
    public SaResult deleteRole(@PathVariable Long id) {
        rolesService.deleteRole(id);
        return SaResult.ok("删除成功");
    }

    @Operation(summary = "角色详情")
    @GetMapping("/{id}")
    public SaResult getRoleById(@PathVariable Long id) {
        return SaResult.data(rolesService.obtainRole(id));
    }

    @Operation(summary = "启用角色")
    @PostMapping("/{id}/enable")
    public SaResult enableRole(@PathVariable Long id) {
        rolesService.enableRole(id);
        return SaResult.ok("启用成功");
    }

    @Operation(summary = "禁用角色")
    @PostMapping("/{id}/disable")
    public SaResult disableRole(@PathVariable Long id) {
        rolesService.disableRole(id);
        return SaResult.ok("禁用成功");
    }

    @Operation(summary = "更新角色权限绑定")
    @PutMapping("/{id}/permission")
    public SaResult updatePermission(@PathVariable Long id, @RequestBody RolePermissionUpdateInputView input) {
        return SaResult.data(rolesService.updatePermission(input));
    }
    @Operation(summary = "查询用户拥有的角色")
    @GetMapping("/user/{userId}")
    public SaResult getUserRoles(@PathVariable Long userId) {
        return SaResult.data(rolesService.getUserRoles(userId));
    }
}
