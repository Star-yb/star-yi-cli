package com.star.crud.service;

import com.star.common.page.PageQuery;
import com.star.common.page.PageResult;

import java.util.List;

/**
 * 通用 CRUD 操作约定。
 */
public interface BaseCrudOperations<E, ID, C, U, S> {

    E create(C input);

    E update(U input);

    void delete(ID id);

    E obtain(ID id);

    List<E> list(S specification);

    PageResult<E> page(PageQuery pageQuery, S specification);
}
