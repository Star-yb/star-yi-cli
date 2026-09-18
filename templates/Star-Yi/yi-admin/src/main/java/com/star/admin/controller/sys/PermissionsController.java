package com.star.admin.controller.sys;

import cn.dev33.satoken.util.SaResult;
import com.star.admin.model.dto.PermissionCreateInputView;
import com.star.admin.model.dto.PermissionSpecification;
import com.star.admin.model.dto.PermissionUpdateInputView;
import com.star.admin.service.PermissionsService;
import com.star.common.page.PageQuery;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.tags.Tag;
import lombok.RequiredArgsConstructor;
import org.springframework.web.bind.annotation.*;

/**
 * 权限控制器
 */
@Tag(name = "权限管理", description = "权限 CRUD、启停与树形查询，路径 /permissions")
@RestController
@RequestMapping("/permissions")
@RequiredArgsConstructor
public class PermissionsController {

    private final PermissionsService permissionsService;

    @Operation(summary = "分页查询权限列表")
    @GetMapping
    public SaResult listPermissions(PageQuery pageQuery, PermissionSpecification permissionSpecification) {
        return SaResult.data(permissionsService.listPermissions(pageQuery,permissionSpecification));
    }

    @Operation(summary = "创建权限")
    @PostMapping
    public SaResult createPermission(@RequestBody PermissionCreateInputView permission) {
        return SaResult.data(permissionsService.createPermission(permission));
    }

    @Operation(summary = "更新权限")
    @PutMapping("/{id}")
    public SaResult updatePermission(@PathVariable Long id, @RequestBody PermissionUpdateInputView permission) {
        permission.setId(id);
        return SaResult.data(permissionsService.updatePermission(permission));
    }

    @Operation(summary = "删除权限", description = "逻辑删除")
    @DeleteMapping("/{id}")
    public SaResult deletePermission(@PathVariable Long id) {
        permissionsService.deletePermission(id);
        return SaResult.ok("删除成功");
    }

    @Operation(summary = "权限详情")
    @GetMapping("/{id}")
    public SaResult getPermissionById(@PathVariable Long id) {
        return SaResult.data(permissionsService.obtainPermission(id));
    }

    @Operation(summary = "启用权限")
    @PostMapping("/{id}/enable")
    public SaResult enablePermission(@PathVariable Long id) {
        permissionsService.enablePermission(id);
        return SaResult.ok("启用成功");
    }

    @Operation(summary = "禁用权限")
    @PostMapping("/{id}/disable")
    public SaResult disablePermission(@PathVariable Long id) {
        permissionsService.disablePermission(id);
        return SaResult.ok("禁用成功");
    }

    @Operation(summary = "查询角色已绑定权限")
    @GetMapping("/role/{roleId}")
    public SaResult getPermissionsByRole(@PathVariable Long roleId) {
        return SaResult.data(permissionsService.getPermissionRoles(roleId));
    }

    @Operation(summary = "查询子权限列表", description = "按父级 ID 查询直接子节点")
    @GetMapping("/parent/{parentId}")
    public SaResult getPermissionsByParent(@PathVariable Long parentId) {
        return SaResult.data(permissionsService.getChildPermissions(parentId));
    }

    @Operation(summary = "权限树查询", description = "树形结构分页返回")
    @GetMapping("/tree")
    public SaResult getPermissionsTree(PageQuery pageQuery, PermissionSpecification permissionSpecification) {
        return SaResult.data(permissionsService.getPermissionsTree(pageQuery,permissionSpecification));
    }
}
