package com.star.crud.dao;

import jakarta.annotation.Nullable;
import org.babyfish.jimmer.Input;
import org.babyfish.jimmer.Specification;
import org.babyfish.jimmer.spring.repo.support.AbstractJavaRepository;
import org.babyfish.jimmer.spring.repository.SpringOrders;
import org.babyfish.jimmer.spring.repository.support.SpringPageFactory;
import org.babyfish.jimmer.sql.JSqlClient;
import org.babyfish.jimmer.sql.ast.Predicate;
import org.babyfish.jimmer.sql.ast.Selection;
import org.babyfish.jimmer.sql.ast.mutation.DeleteMode;
import org.babyfish.jimmer.sql.ast.mutation.SaveMode;
import org.babyfish.jimmer.sql.ast.table.spi.AbstractTypedTable;
import org.babyfish.jimmer.sql.fetcher.Fetcher;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;

import java.util.List;
import java.util.function.Function;

/**
 * 通用 Jimmer DAO 基类。
 * <p>
 * 保留与当前项目 DAO 一致的风格：
 * 1. 继续继承 {@link AbstractJavaRepository}
 * 2. 提供新增、修改、删除
 * 3. 提供 Predicate / Specification 两套分页与非分页查询能力
 */
public abstract class BaseJimmerDao<E, ID, T extends AbstractTypedTable<E>>
        extends AbstractJavaRepository<E, ID> {

    protected final T table;
    private final Class<E> entityType;

    protected BaseJimmerDao(JSqlClient sql, Class<E> entityType, T table) {
        super(sql);
        this.entityType = entityType;
        this.table = table;
    }

    protected final Class<E> entityType() {
        return entityType;
    }

    protected final T table() {
        return table;
    }

    public <I extends Input<E>> E add(I input, @Nullable Fetcher<E> fetcher) {
        return sql.getEntities()
                .saveCommand(input)
                // .setMode(SaveMode.INSERT_IF_ABSENT) // INSERT_IF_ABSENT是如果实体存在则不插入，不存在则插入
                .setMode(SaveMode.INSERT_ONLY) // INSERT_ONLY是无条件插入数据
                .execute(fetcher)
                .getModifiedEntity();
    }

    public <I extends Input<E>> E update(I input, @Nullable Fetcher<E> fetcher) {
        return sql.getEntities()
                .saveCommand(input)
                .setMode(SaveMode.UPDATE_ONLY) // UPDATE_ONLY是无条件更新数据
                .execute(fetcher)
                .getModifiedEntity();
    }

    public void delete(ID id) {
        sql.getEntities()
                .deleteCommand(entityType, id)
                .setMode(DeleteMode.AUTO)
                .execute();
    }

    public E getById(ID id) {
        return findById(id);
    }

    public E getById(ID id, Fetcher<E> fetcher) {
        return findById(id, fetcher);
    }

    public <R> List<R> queryList(
            Predicate predicate,
            Function<T, Selection<R>> selectionFn
    ) {
        return sql.createQuery(table).where(predicate)
                .select(selectionFn.apply(table))
                .execute();
    }

    public <R> List<R> queryList(
            Specification<E> specification,
            Function<T, Selection<R>> selectionFn
    ) {
        return sql.createQuery(table).where(specification)
                .select(selectionFn.apply(table))
                .execute();
    }

    public <R> Page<R> queryPage(
            Pageable pageable,
            Predicate predicate,
            Function<T, Selection<R>> selectionFn
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

    public <R> Page<R> queryPage(
            Pageable pageable,
            Specification<E> specification,
            Function<T, Selection<R>> selectionFn
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
}
