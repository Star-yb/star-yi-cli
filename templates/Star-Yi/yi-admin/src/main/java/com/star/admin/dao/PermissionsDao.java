package com.star.admin.dao;

import com.star.admin.model.dto.PermissionCreateInputView;
import com.star.admin.model.dto.PermissionUpdateInputView;
import com.star.admin.model.dto.PermissionsView;
import com.star.admin.model.entity.Permissions;
import com.star.admin.model.entity.PermissionsTable;
import jakarta.annotation.Nullable;
import org.babyfish.jimmer.Specification;
import org.babyfish.jimmer.spring.repo.support.AbstractJavaRepository;
import org.babyfish.jimmer.spring.repository.SpringOrders;
import org.babyfish.jimmer.spring.repository.support.SpringPageFactory;
import org.babyfish.jimmer.sql.JSqlClient;
import org.babyfish.jimmer.sql.ast.Predicate;
import org.babyfish.jimmer.sql.ast.mutation.DeleteMode;
import org.babyfish.jimmer.sql.ast.mutation.SaveMode;
import org.babyfish.jimmer.sql.fetcher.Fetcher;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.stereotype.Repository;

import java.util.List;

@Repository
public class PermissionsDao extends AbstractJavaRepository<Permissions, Long> {

    private static final PermissionsTable table = PermissionsTable.$;

    public PermissionsDao(JSqlClient sql) {
        super(sql);
    }

    /**
     * 创建权限
     */
    public Permissions add(PermissionCreateInputView permission, @Nullable Fetcher<Permissions> fetcher) {
        return sql.getEntities()
                .saveCommand(permission)
                .setMode(SaveMode.INSERT_IF_ABSENT)
                .execute(fetcher)
                .getModifiedEntity();
    }

    /**
     * 更新权限
     */
    public Permissions update(PermissionUpdateInputView permission, @Nullable Fetcher<Permissions> fetcher) {
        return sql.getEntities()
                .saveCommand(permission)
                .setMode(SaveMode.UPDATE_ONLY)
                .execute(fetcher)
                .getModifiedEntity();
    }

    /**
     * 删除权限
     */
    public void delete(Long id) {
        sql.getEntities()
                .deleteCommand(Permissions.class, id)
                .setMode(DeleteMode.AUTO)
                .execute();
    }

    /**
     * 修改权限状态
     *
     * @param id     权限ID
     * @param status 状态值（0 正常，1 禁用）
     */
    public void updateStatus(Long id, Integer status) {
        sql.createUpdate(table)
                .set(table.status(), status)
                .where(table.id().eq(id))
                .execute();
    }


    // ==================== 新封装：列表条件查询 一致的通用查询能力 ====================

    /**
     * 【新】通用非分页查询：使用 Predicate 条件 + 自定义 Selection，返回任意类型
     */
    public <R> List<R> queryList(
            Predicate predicate,
            java.util.function.Function<PermissionsTable, org.babyfish.jimmer.sql.ast.Selection<R>> selectionFn
    ) {
        return sql.createQuery(table).where(predicate)
                .select(selectionFn.apply(table))
                .execute();
    }

    /**
     * 【新】通用非分页查询：使用 Specification 条件 + 自定义 Selection，返回任意类型
     */
    public <R> List<R> queryList(
            Specification<Permissions> specification,
            java.util.function.Function<PermissionsTable, org.babyfish.jimmer.sql.ast.Selection<R>> selectionFn
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
            java.util.function.Function<PermissionsTable, org.babyfish.jimmer.sql.ast.Selection<R>> selectionFn
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
            Specification<Permissions> specification,
            java.util.function.Function<PermissionsTable, org.babyfish.jimmer.sql.ast.Selection<R>> selectionFn
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
