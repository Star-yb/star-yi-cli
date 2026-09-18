package com.star.admin.service.impl;

import com.star.admin.dao.UsersDao;
import com.star.admin.model.dto.*;
import com.star.admin.model.entity.Users;
import com.star.admin.model.entity.UsersFetcher;
import com.star.admin.model.entity.UsersTable;
import com.star.admin.service.UsersService;
import com.star.utils.AuthCryptoUtils;
import com.star.common.exception.BusinessException;
import com.star.common.page.PageQuery;
import com.star.common.page.PageResult;
import com.star.common.page.PageUtils;
import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Service;

import java.util.List;

/**
 * 用户服务实现类
 */
@Service
@RequiredArgsConstructor
public class UsersServiceImpl implements UsersService {
    private final UsersDao usersDao;
    private static final UsersTable T = UsersTable.$;


    @Override
    public Users createUser(UserCreateInputView user) {
        return usersDao.add(user, UsersView.METADATA.getFetcher());
    }

    @Override
    public Users updateUser(UserUpdateInputView user) {
        return usersDao.update(user, UsersView.METADATA.getFetcher());
    }

    @Override
    public void deleteUser(Long id) {
        usersDao.delete(id);
    }

    @Override
    public Users obtainUser(Long id) {
        return usersDao.findById(id, UsersView.METADATA.getFetcher());
    }


    @Override
    public List<Users> listUsers(UserSpecification usersSpecification) {
        return usersDao.queryList(
                usersSpecification,
                root -> root.fetch(
                        UsersView.METADATA.getFetcher()
                                .add("roleIds")
                )
        );
    }

    @Override
    public PageResult<Users> listUsers(PageQuery pageQuery, UserSpecification usersSpecification) {
/*
 {  boolean paged = pageQuery.isPaged();
        if (!paged) {
            // 前端未传入有效分页参数：不分页，返回统一结构
            List<Users> records = usersDao.queryList(
                    usersSpecification,
                    root -> root.fetch(UsersView.METADATA.getFetcher())
            );
            return PageResult.ofUnpaged(records);
        }
        Pageable pageable = pageQuery.toPageable();
        Page<Users> pageData = usersDao.queryPage(
                pageable,
                usersSpecification,
                root -> root.fetch(UsersView.METADATA.getFetcher())
        );
        return PageResult.ofPaged(pageData);
    }
    */
        return PageUtils.paginate(
                pageQuery,
                () -> usersDao.queryList(
                        usersSpecification,
                        root -> root.fetch(
                                UsersView.METADATA.getFetcher()
                                        .add("roleIds")
                        )
                ),
                pageable -> usersDao.queryPage(
                        pageable,
                        usersSpecification,
                        root -> root.fetch(
                                UsersView.METADATA.getFetcher()
                                        .add("roleIds")
                        )
                )
        );
    }

    @Override
    public void enableUser(Long id) {
        usersDao.updateStatus(id, 0);
    }

    @Override
    public void disableUser(Long id) {
        usersDao.updateStatus(id, 1);
    }


    @Override
    public void changePassword(Long userId, String oldPassword, String newPassword) {
        Users user = usersDao.findById(userId, UsersFetcher.$.password());
        if (user == null) {
            throw new BusinessException("用户不存在").code(404);
        }
        if (!AuthCryptoUtils.matches(oldPassword, user.password())) {
            throw new BusinessException("原密码不正确").code(400);
        }
//        新密码不能为空，长度至少为6
        if (newPassword.length() < 6) {
            throw new BusinessException("新密码不能为空，长度至少为6").code(400);
        }
        newPassword = AuthCryptoUtils.hash(newPassword);

        // 使用用户更新输入视图只更新密码字段
        usersDao.updatePassword(userId, newPassword);
    }

    @Override
    public List<Users> getRoleUsers(Long roleId) {
        // 通过角色反查用户列表
        return usersDao.queryList(
                UsersTable.$.roles(r -> r.id().eq(roleId)),
                f -> f.fetch(UsersView.METADATA.getFetcher())
        );
    }


    @Override
    public int updateRole(UserRoleUpdateInputView userRoleUpdateInputView) {
        return usersDao.updateRole(userRoleUpdateInputView);
    }

    @Override
    public void adminDelete(Long id) {
        usersDao.adminDelete(id);
    }

    @Override
    public void resetPassword(Long id) {
        String newPassword = AuthCryptoUtils.hash("123456");
        usersDao.updatePassword(id, newPassword);
    }
}
