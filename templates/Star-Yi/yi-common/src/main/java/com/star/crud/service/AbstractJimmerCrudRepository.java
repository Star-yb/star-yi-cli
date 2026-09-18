package com.star.crud.service;

import com.star.common.page.PageQuery;
import com.star.common.page.PageResult;
import com.star.common.page.PageUtils;
import com.star.crud.dao.BaseJimmerDao;
import jakarta.annotation.Nullable;
import org.babyfish.jimmer.Input;
import org.babyfish.jimmer.Specification;
import org.babyfish.jimmer.sql.ast.table.spi.AbstractTypedTable;
import org.babyfish.jimmer.sql.fetcher.Fetcher;

import java.util.List;

/**
 * 基于 {@link BaseJimmerDao} 的通用仓储桥接实现。
 * <p>
 * 适用于：
 * 1. Create / Update 输入都是 Jimmer Input
 * 2. 列表和分页查询都基于 Jimmer Specification
 * 3. 返回值统一使用同一套 Fetcher
 */
public abstract class AbstractJimmerCrudRepository<
        E,
        ID,
        C extends Input<E>,
        U extends Input<E>,
        S extends Specification<E>,
        T extends AbstractTypedTable<E>
        > implements CrudRepository<E, ID, C, U, S> {

    protected abstract BaseJimmerDao<E, ID, T> dao();

    @Nullable
    protected abstract Fetcher<E> fetcher();

    @Override
    public E create(C input) {
        return dao().add(input, fetcher());
    }

    @Override
    public E update(U input) {
        return dao().update(input, fetcher());
    }

    @Override
    public void delete(ID id) {
        dao().delete(id);
    }

    @Override
    public E obtain(ID id) {
        Fetcher<E> fetcher = fetcher();
        if (fetcher != null) {
            return dao().getById(id, fetcher);
        }
        return dao().getById(id);
    }

    @Override
    public List<E> list(S specification) {
        return dao().queryList(
                specification,
                root -> fetcher() == null ? root : root.fetch(fetcher())
        );
    }

    @Override
    public PageResult<E> page(PageQuery pageQuery, S specification) {
        return PageUtils.paginate(
                pageQuery,
                () -> dao().queryList(
                        specification,
                        root -> fetcher() == null ? root : root.fetch(fetcher())
                ),
                pageable -> dao().queryPage(
                        pageable,
                        specification,
                        root -> fetcher() == null ? root : root.fetch(fetcher())
                )
        );
    }
}
