package com.star.admin.service;

import com.star.admin.model.dto.PermissionCreateInputView;
import com.star.admin.model.dto.PermissionSpecification;
import com.star.admin.model.dto.PermissionUpdateInputView;
import com.star.admin.model.dto.PermissionsView;
import com.star.admin.model.entity.Permissions;
import com.star.common.page.PageQuery;
import com.star.common.page.PageResult;
import org.springframework.stereotype.Service;

import java.util.List;

/**
 * 权限服务接口
 */
@Service
public interface PermissionsService {

    /** 创建权限 */
    Permissions createPermission(PermissionCreateInputView permission);

    /** 更新权限信息 */
    Permissions updatePermission(PermissionUpdateInputView permission);

    /** 删除权限 */
    void deletePermission(Long id);

    /** 根据ID查询权限 */
    Permissions obtainPermission(Long id);

    /** 查询所有权限列表 */
    List<Permissions> listPermissions(PermissionSpecification permissionSpecification);
    PageResult<Permissions> listPermissions(PageQuery pageQuery, PermissionSpecification permissionSpecification);

    // ==================== 权限树结构 ====================
    /** 获取权限树（用于菜单展示） */
//    List<PermissionTreeDTO> getPermissionTree();

    /** 获取父权限下的子权限 */
    List<Permissions> getChildPermissions(Long parentId);
    PageResult<Permissions> getPermissionsTree(PageQuery pageQuery, PermissionSpecification permissionSpecification);

    // ==================== 权限状态管理 ====================
    /** 启用权限 */
    void enablePermission(Long id);

    /** 禁用权限 */
    void disablePermission(Long id);


    // ==================== 权限角色关联 ====================
    /** 获取拥有该权限的角色列表 */
    List<Permissions> getPermissionRoles(Long roleId);

    void adminDelete(Long id);
    
}
