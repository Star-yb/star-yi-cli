package com.star.crud.service;

import com.star.common.page.PageQuery;
import com.star.common.page.PageResult;

import java.util.List;

/**
 * Service 模板所依赖的仓储抽象。
 * <p>
 * 这里不强绑定具体 DAO 实现，便于后续让 Jimmer DAO、聚合仓储或组合仓储共同复用。
 */
public interface CrudRepository<E, ID, C, U, S> {

    E create(C input);

    E update(U input);

    void delete(ID id);

    E obtain(ID id);

    List<E> list(S specification);

    PageResult<E> page(PageQuery pageQuery, S specification);
}
