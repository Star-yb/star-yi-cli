package com.star.admin.service;

import com.star.admin.model.dto.*;
import com.star.admin.model.entity.Users;
import com.star.common.page.PageQuery;
import com.star.common.page.PageResult;
import org.springframework.stereotype.Service;

import java.util.List;

/**
 * 用户服务接口
 */
@Service
public interface UsersService {

    Users createUser(UserCreateInputView user);

    Users updateUser(UserUpdateInputView user);

    void deleteUser(Long id);

    Users obtainUser(Long id);

    List<Users> listUsers(UserSpecification usersSpecification);

    PageResult<Users> listUsers(PageQuery pageQuery, UserSpecification usersSpecification);

    // ==================== 用户状态管理 ====================


    /** 启用用户 */
    void enableUser(Long id);

    /** 禁用用户 */
    void disableUser(Long id);

    // ==================== 用户角色关联 ====================
//    /** 给用户分配角色 */
//    void assignRoles(Long userId, List<Long> roleIds);
//
//    /** 移除用户角色 */
//    void removeRoles(Long userId, List<Long> roleIds);

    int updateRole(UserRoleUpdateInputView userRoleUpdateInputView);




    /** 修改密码 */
    void changePassword(Long userId, String oldPassword, String newPassword);

    // ==================== 角色用户关联 ====================
    /** 获取角色下的用户列表 */
    List<Users> getRoleUsers(Long roleId);


//

    void adminDelete(Long id);


    void resetPassword(Long id);

}
