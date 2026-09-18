package com.star.admin.service;

import com.star.admin.model.dto.*;
import com.star.admin.model.entity.Roles;
import com.star.common.page.PageQuery;
import com.star.common.page.PageResult;
import org.springframework.stereotype.Service;

import java.util.List;

/**
 * 角色服务接口
 */
@Service
public interface RolesService {

    /** 创建角色 */
    Roles createRole(RoleCreateInputView role);

    /** 更新角色信息 */
    Roles updateRole(RoleUpdateInputView role);

    /** 删除角色 */
    void deleteRole(Long id);

    /** 根据ID查询角色 */
    Roles obtainRole(Long id);

    /** 分页查询角色列表 */
    List<Roles> listRoles(RoleSpecification roleSpecification);

    PageResult<Roles> listRoles(PageQuery pageQuery, RoleSpecification roleSpecification);

    // ==================== 角色状态管理 ====================
    /** 启用角色 */
    void enableRole(Long id);

    /** 禁用角色 */
    void disableRole(Long id);


    // ==================== 角色权限关联 ====================
//    /** 给角色分配权限 */
//    void assignPermissions(Long roleId, List<Long> permissionIds);
//
//    /** 移除角色权限 */
//    void removePermissions(Long roleId, List<Long> permissionIds);
    int updatePermission(RolePermissionUpdateInputView input);


    /** 获取用户的角色列表 */
    List<Roles> getUserRoles(Long userId);

    void adminDelete(Long id);


}
