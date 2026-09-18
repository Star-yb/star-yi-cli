package com.star.admin.dao;

import com.star.admin.model.entity.Roles;
import com.star.admin.model.dto.*;
import com.star.admin.model.entity.RolesTable;
import com.star.admin.model.entity.RolesProps;
import jakarta.annotation.Nullable;
import org.babyfish.jimmer.Specification;
import org.babyfish.jimmer.spring.repo.support.AbstractJavaRepository;
import org.babyfish.jimmer.spring.repository.SpringOrders;
import org.babyfish.jimmer.sql.JSqlClient;
import org.babyfish.jimmer.sql.ast.Predicate;
import org.babyfish.jimmer.sql.ast.mutation.DeleteMode;
import org.babyfish.jimmer.sql.ast.mutation.SaveMode;
import org.babyfish.jimmer.sql.fetcher.Fetcher;
import org.springframework.stereotype.Repository;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.babyfish.jimmer.spring.repository.support.SpringPageFactory;

import java.util.Collections;
import java.util.List;

@Repository
public class RolesDao extends AbstractJavaRepository<Roles, Long> {

    private static final RolesTable table = RolesTable.$;

    public RolesDao(JSqlClient sql) {
        super(sql);
    }

    /**
     * 创建角色
     */
    public Roles add(RoleCreateInputView role,@Nullable Fetcher<Roles> fetcher) {
        return sql.getEntities()
                .saveCommand(role)
                .setMode(SaveMode.INSERT_IF_ABSENT)
                .execute(fetcher)
                .getModifiedEntity();
    }


    /**
     * 更新角色
     */
    public Roles update(RoleUpdateInputView role,@Nullable Fetcher<Roles> fetcher) {
        return sql.getEntities()
                .saveCommand(role)
                .setMode(SaveMode.UPDATE_ONLY)
                .execute(fetcher)
                .getModifiedEntity();
    }

    /**
     * 删除角色
     */
    public void delete(Long id) {
        sql.getEntities()
                .deleteCommand(Roles.class, id)
                .setMode(DeleteMode.AUTO)
                .execute();
    }

    /**
     * 修改角色状态
     *
     * @param id     角色ID
     * @param status 状态值（0 正常，1 禁用）
     */
    public void updateStatus(Long id, Integer status) {
        sql.createUpdate(table)
                .set(table.status(), status)
                .where(table.id().eq(id))
                .execute();
    }
    
    
//    权限分配
    public void updatePermission(Long id, List<Long> permissionIds, Boolean isAdd) {
        if (isAdd == null) {
            isAdd = true;
        }
        if (isAdd) {
        sql.getAssociations(RolesProps.PERMISSIONS).saveAll(
                Collections.singleton(id),
                permissionIds
        );
        } else {
            sql.getAssociations(RolesProps.PERMISSIONS).deleteAll(
                    Collections.singleton(id),
                    permissionIds
            );
        }
    }



    public int updatePermission(RolePermissionUpdateInputView rolePermissionUpdate) {
        return sql.save(rolePermissionUpdate).getTotalAffectedRowCount();

    }

    // ==================== 新封装：列表条件查询 一致的通用查询能力 ====================

    /**
     * 【新】通用非分页查询：使用 Predicate 条件 + 自定义 Selection，返回任意类型
     */
    public <R> List<R> queryList(
            Predicate predicate,
            java.util.function.Function<RolesTable, org.babyfish.jimmer.sql.ast.Selection<R>> selectionFn
    ) {
        return sql.createQuery(table).where(predicate)
                .select(selectionFn.apply(table))
                .execute();
    }

    /**
     * 【新】通用非分页查询：使用 Specification 条件 + 自定义 Selection，返回任意类型
     */
    public <R> List<R> queryList(
            Specification<Roles> specification,
            java.util.function.Function<RolesTable, org.babyfish.jimmer.sql.ast.Selection<R>> selectionFn
    ) {
        return sql.createQuery(table).where(specification)
                .select(selectionFn.apply(table))
                .execute();
    }

    /**
     * 【新】通用分页查询：使用 Predicate 条件 + 自定义 Selection，返回任意类型
     */
    public <R> Page<R> queryPage(
            Pageable pageable,
            Predicate predicate,
            java.util.function.Function<RolesTable, org.babyfish.jimmer.sql.ast.Selection<R>> selectionFn
    ) {
        return sql.createQuery(table).where(predicate)
                .orderBy(SpringOrders.toOrders(table, pageable.getSort()))
                .select(selectionFn.apply(table))
                .fetchPage(
                        pageable.getPageNumber(),
                        pageable.getPageSize(),
                        SpringPageFactory.getInstance()
                );
    }

    /**
     * 【新】通用分页查询：使用 Specification 条件 + 自定义 Selection，返回任意类型
     */
    public <R> Page<R> queryPage(
            Pageable pageable,
            Specification<Roles> specification,
            java.util.function.Function<RolesTable, org.babyfish.jimmer.sql.ast.Selection<R>> selectionFn
    ) {
        return sql.createQuery(table).where(specification)
                .orderBy(SpringOrders.toOrders(table, pageable.getSort()))
                .select(selectionFn.apply(table))
                .fetchPage(
                        pageable.getPageNumber(),
                        pageable.getPageSize(),
                        SpringPageFactory.getInstance()
                );
    }



    //    超级管理员真删除方法
    public void adminDelete(Long id) {
        sql.createDelete(table)
                .where(table.id().eq(id))
                .setMode(DeleteMode.PHYSICAL)
                .execute();
    }
}
