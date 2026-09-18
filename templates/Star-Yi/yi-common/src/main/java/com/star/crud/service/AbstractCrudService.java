package com.star.crud.service;

import com.star.common.page.PageQuery;
import com.star.common.page.PageResult;

import java.util.List;

/**
 * 通用 CRUD 服务模板。
 * <p>
 * 将固定流程下沉到基类，把业务差异保留给 Hook 扩展点。
 */
public abstract class AbstractCrudService<E, ID, C, U, S>
        extends BaseCrudService<E, ID, C, U, S> {

    protected abstract CrudRepository<E, ID, C, U, S> repository();

    protected void beforeCreate(C input) {
    }

    protected void afterCreate(E entity) {
    }

    protected void beforeUpdate(U input) {
    }

    protected void afterUpdate(E entity) {
    }

    protected void beforeDelete(ID id) {
    }

    protected void afterDelete(ID id) {
    }

    @Override
    public E create(C input) {
        beforeCreate(input);
        E entity = repository().create(input);
        afterCreate(entity);
        return entity;
    }

    @Override
    public E update(U input) {
        beforeUpdate(input);
        E entity = repository().update(input);
        afterUpdate(entity);
        return entity;
    }

    @Override
    public void delete(ID id) {
        beforeDelete(id);
        repository().delete(id);
        afterDelete(id);
    }

    @Override
    public E obtain(ID id) {
        return repository().obtain(id);
    }

    @Override
    public List<E> list(S specification) {
        return repository().list(specification);
    }

    @Override
    public PageResult<E> page(PageQuery pageQuery, S specification) {
        return repository().page(pageQuery, specification);
    }
}
