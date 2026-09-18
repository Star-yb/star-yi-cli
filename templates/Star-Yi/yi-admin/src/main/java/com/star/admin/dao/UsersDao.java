package com.star.admin.dao;


import com.star.admin.model.dto.UserCreateInputView;
import com.star.admin.model.dto.UserRoleUpdateInputView;
import com.star.admin.model.dto.UserUpdateInputView;
import com.star.admin.model.entity.Users;
import com.star.admin.model.entity.UsersTable;
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

import java.time.LocalDateTime;
import java.util.List;


@Repository
public class UsersDao extends AbstractJavaRepository<Users, Long> {
    private static final UsersTable table = UsersTable.$;

    public UsersDao(JSqlClient sql) {
        super(sql);
    }

    public Users add(UserCreateInputView users,@Nullable Fetcher<Users> fetcher){
        return sql.getEntities()
                .saveCommand(users)
                .setMode(SaveMode.INSERT_IF_ABSENT)
                .execute(fetcher)
                .getModifiedEntity();
    }
    public Users update(UserUpdateInputView users,@Nullable Fetcher<Users> fetcher){
        return sql.getEntities()
                .saveCommand(users)
                .setMode(SaveMode.UPDATE_ONLY)
                .execute(fetcher)
                .getModifiedEntity();
    }


/*
     // 新增：创建用户，返回视图类型（例如 UsersView）
 public UsersView add(UserCreateInputView users) {
     return sql.getEntities()
             .saveCommand(users)
             .setMode(SaveMode.INSERT_IF_ABSENT)
             .execute(UsersView.class)
             .getModifiedView();
 }


     // 新增：更新用户，返回视图类型
 public UsersView update(UserUpdateInputView users) {
     return sql.getEntities()
             .saveCommand(users)
             .setMode(SaveMode.UPDATE_ONLY)
             .execute(UsersView.class)
             .getModifiedView();
 }
*/




    public void updateLastLoginTime(long id) {
        sql.createUpdate(table)
                .set(table.lastLoginTime(), LocalDateTime.now())
                .where(table.id().eq(id))
                .execute();
    }

    //    修改密码
    public void updatePassword(long id, String password){
        sql.createUpdate(table)
                .set(table.password(), password)
                .where(table.id().eq(id))
                .execute();
    }

    public void delete(long id){
        sql.getEntities()
                .deleteCommand(Users.class, id)
                .setMode(DeleteMode.AUTO)
                .execute();
    }

    //    修改用户状态
    public void updateStatus(long id, Integer status){
        sql.createUpdate(table)
                .set(table.status(), status)
                .where(table.id().eq(id))
                .execute();
    }

    public int updateRole(UserRoleUpdateInputView userRoleUpdateInputView){
        return sql.save(userRoleUpdateInputView).getTotalAffectedRowCount();

    }


    // ==================== 新封装：不改动旧方法的前提下，提供更通用的查询能力 ====================

    /**
     * 【新】通用非分页查询：使用 Predicate 条件 + 自定义 Selection，返回任意类型
     */
    public <R> List<R> queryList(
            Predicate predicate,
            java.util.function.Function<UsersTable, org.babyfish.jimmer.sql.ast.Selection<R>> selectionFn
    ) {
        return sql.createQuery(table).where(predicate)
                .select(selectionFn.apply(table))
                .execute();
    }

    /**
     * 【新】通用非分页查询：使用 Specification 条件 + 自定义 Selection，返回任意类型
     */
    public <R> List<R> queryList(
            Specification<Users> specification,
            java.util.function.Function<UsersTable, org.babyfish.jimmer.sql.ast.Selection<R>> selectionFn
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
            java.util.function.Function<UsersTable, org.babyfish.jimmer.sql.ast.Selection<R>> selectionFn
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
            Specification<Users> specification,
            java.util.function.Function<UsersTable, org.babyfish.jimmer.sql.ast.Selection<R>> selectionFn
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
    public void adminDelete(long id) {
        sql.createDelete(table)
                .where(table.id().eq(id))
                .setMode(DeleteMode.PHYSICAL)
                .execute();
    }

}
