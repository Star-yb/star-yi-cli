package com.star.admin.service.impl;

import cn.dev33.satoken.SaManager;
import com.star.admin.dao.RolesDao;
import com.star.admin.model.entity.Roles;
import com.star.admin.model.dto.*;
import com.star.admin.model.entity.RolesTable;
import com.star.admin.service.RolesService;
import com.star.common.page.PageQuery;
import com.star.common.page.PageResult;
import com.star.common.page.PageUtils;
import lombok.RequiredArgsConstructor;
import org.babyfish.jimmer.sql.JSqlClient;
import org.springframework.stereotype.Service;

import java.util.List;

/**
 * 角色服务实现类
 */
@Service
@RequiredArgsConstructor
public class RolesServiceImpl implements RolesService {

    private final RolesDao rolesDao;
    private final JSqlClient sql;
    private static final RolesTable R = RolesTable.$;

    @Override
    public Roles createRole(RoleCreateInputView role) {
        // 新建角色后，角色与权限关系初始为空，无需清理缓存
        return rolesDao.add(role, RolesView.METADATA.getFetcher());
    }

    @Override
    public Roles updateRole(RoleUpdateInputView role) {
        Roles updated = rolesDao.update(role, RolesView.METADATA.getFetcher());
        // 如果修改了角色编码或状态，可能影响权限缓存，这里根据角色编码清理缓存
        String cacheKey = "satoken:role-find-permission:" + updated.roleCode();
        SaManager.getSaTokenDao().deleteObject(cacheKey);
        return updated;
    }

    @Override
    public void deleteRole(Long id) {
        Roles role = rolesDao.findById(id);
        if (role != null) {
            // 删除角色前先清理该角色对应的权限缓存
            String cacheKey = "satoken:role-find-permission:" + role.roleCode();
            SaManager.getSaTokenDao().deleteObject(cacheKey);
            rolesDao.delete(id);
        }
    }

    @Override
    public Roles obtainRole(Long id) {
        return rolesDao.findById(id);
    }

    @Override
    public List<Roles> listRoles(RoleSpecification roleSpecification) {
        return rolesDao.queryList(
                roleSpecification,
                f->f.fetch(
                        RolesView.METADATA.getFetcher()
                )
        );
    }

    @Override
    public PageResult<Roles> listRoles(PageQuery pageQuery, RoleSpecification roleSpecification) {
        return PageUtils.paginate(
                pageQuery,
                () -> rolesDao.queryList(
                        roleSpecification,
                        root -> root.fetch(
                                RolesView.METADATA.getFetcher()

                        )
                ),
                pageable -> rolesDao.queryPage(
                        pageable,
                        roleSpecification,
                        root -> root.fetch(
                                RolesView.METADATA.getFetcher()
                        )
                )
        );
    }

    @Override
    public void enableRole(Long id) {
        rolesDao.updateStatus(id, 0);
        // 状态改变后，清理关联缓存
        Roles role = rolesDao.findById(id);
        if (role != null) {
            String cacheKey = "satoken:role-find-permission:" + role.roleCode();
            SaManager.getSaTokenDao().deleteObject(cacheKey);
        }
    }

    @Override
    public void disableRole(Long id) {
        rolesDao.updateStatus(id, 1);
        Roles role = rolesDao.findById(id);
        if (role != null) {
            String cacheKey = "satoken:role-find-permission:" + role.roleCode();
            SaManager.getSaTokenDao().deleteObject(cacheKey);
        }
    }

    @Override
    public int updatePermission(RolePermissionUpdateInputView input) {
        int i = rolesDao.updatePermission(input);

        // 这里假设前端会通过角色编辑接口提交包含权限的 DTO，这里只需要清理缓存
        Roles role = rolesDao.findById(input.getId());
        if (role != null) {
            String cacheKey = "satoken:role-find-permission:" + role.roleCode();
            SaManager.getSaTokenDao().deleteObject(cacheKey);
        }
        return i;
    }


    @Override
    public List<Roles> getUserRoles(Long userId) {
        // 查询用户详情时，使用 UsersView 中的角色信息
        return rolesDao.queryList(
                RolesTable.$.users(u -> u.id().eq(userId)),
                f -> f.fetch(
                        RolesView.METADATA.getFetcher()
                )
        );
    }


    @Override
    public void adminDelete(Long id) {
        rolesDao.adminDelete(id);
    }


}
