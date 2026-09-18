package com.star.admin.service.impl;

import cn.dev33.satoken.SaManager;
import com.star.admin.dao.PermissionsDao;
import com.star.admin.model.entity.Permissions;
import com.star.admin.model.dto.*;
import com.star.admin.model.entity.PermissionsFetcher;
import com.star.admin.model.entity.PermissionsTable;
import com.star.admin.service.PermissionsService;
import com.star.common.page.PageQuery;
import com.star.common.page.PageResult;
import com.star.common.page.PageUtils;
import lombok.RequiredArgsConstructor;
import org.babyfish.jimmer.sql.fetcher.Fetcher;
import org.springframework.stereotype.Service;

import java.util.List;

/**
 * 权限服务实现类
 */
@Service
@RequiredArgsConstructor
public class PermissionsServiceImpl implements PermissionsService {

    private final PermissionsDao permissionsDao;

    @Override
    public Permissions createPermission(PermissionCreateInputView permission) {
        return permissionsDao.add(permission, PermissionsView.METADATA.getFetcher());
    }

    @Override
    public Permissions updatePermission(PermissionUpdateInputView permission) {
        // 权限本身变更不会直接影响缓存 key，但可以在需要时扩展
        return permissionsDao.update(permission, PermissionsView.METADATA.getFetcher());
    }

    @Override
    public void deletePermission(Long id) {
        permissionsDao.delete(id);
    }

    @Override
    public Permissions obtainPermission(Long id) {
        return permissionsDao.findById(id);
    }

    @Override
    public List<Permissions> listPermissions(PermissionSpecification permissionSpecification) {
        return permissionsDao.queryList(
                permissionSpecification,
                f -> f.fetch(
                        PermissionsView.METADATA.getFetcher()
                )
        );
    }


    @Override
    public PageResult<Permissions> listPermissions(PageQuery pageQuery, PermissionSpecification permissionSpecification) {
        return PageUtils.paginate(
                pageQuery,
                () -> permissionsDao.queryList(
                        permissionSpecification,
                        root -> root.fetch(
                                PermissionsView.METADATA.getFetcher()
                        )
                ),
                pageable -> permissionsDao.queryPage(
                        pageable,
                        permissionSpecification,
                        root -> root.fetch(
                                PermissionsView.METADATA.getFetcher()
                        )
                )
        );
    }


    @Override
    public void enablePermission(Long id) {
        permissionsDao.updateStatus(id, 0);
        // 启用权限时，可能影响多个角色的权限集合，这里只清缓存，由下次鉴权自动刷新
        // searchData(prefix, keyword, start, size, sortType)，当前版本最后一个参数为 boolean
        SaManager.getSaTokenDao().searchData("satoken:role-find-permission:", null, 0, 10_000, false)
                .forEach(SaManager.getSaTokenDao()::deleteObject);
    }

    @Override
    public void disablePermission(Long id) {
        permissionsDao.updateStatus(id, 1);
        // 禁用权限同样清理与角色相关的权限缓存
        SaManager.getSaTokenDao().searchData("satoken:role-find-permission:", null, 0, 10_000, false)
                .forEach(SaManager.getSaTokenDao()::deleteObject);
    }

    @Override
    public List<Permissions> getPermissionRoles(Long roleId) {
        // 这里可以根据业务后续用 Table 查询补充实现，目前先返回空集合以避免 NPE
        // 通过角色反查用户列表
        return permissionsDao.queryList(
                PermissionsTable.$.roles(r -> r.id().eq(roleId)),
                f -> f.fetch(PermissionsView.METADATA.getFetcher())
        );
    }

    @Override
    public List<Permissions> getChildPermissions(Long parentId) {
        // 简单实现：所有权限中筛选 parentId 相同的
        return permissionsDao.queryList(
                PermissionsTable.$.parentId().eq(parentId),
                f -> f.fetch(
                        PermissionsFetcher.$.allScalarFields()
                                .recursiveParent()
                                .recursiveChildPermissions()
                )
        );
    }


    @Override
    public PageResult<Permissions> getPermissionsTree(PageQuery pageQuery, PermissionSpecification permissionSpecification) {
        Fetcher<Permissions> fetcher = PermissionsView.METADATA.getFetcher();

        Fetcher<Permissions> add = fetcher.add("parent").add("childPermissions",
                PermissionsFetcher.$.allScalarFields());


        return PageUtils.paginate(
                pageQuery,
                () -> permissionsDao.queryList(
                        permissionSpecification,
                        f -> f.fetch(
                                PermissionsTreeView.METADATA.getFetcher()
                                        .add("parentId")
                        )
                ),
                pageable -> permissionsDao.queryPage(
                        pageable,
                        permissionSpecification,
                        f -> f.fetch(
                                PermissionsTreeView.METADATA.getFetcher()
                                        .add("parentId")

                        )
                )
        );
    }


    @Override
    public void adminDelete(Long id) {
        permissionsDao.adminDelete(id);
    }
}
