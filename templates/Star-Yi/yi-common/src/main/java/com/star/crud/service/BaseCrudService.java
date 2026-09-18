package com.star.crud.service;

/**
 * 通用 CRUD 服务基类定义。
 * <p>
 * 用于表达统一的服务契约；真正带模板方法和扩展 Hook 的实现见 {@link AbstractCrudService}。
 */
public abstract class BaseCrudService<E, ID, C, U, S>
        implements BaseCrudOperations<E, ID, C, U, S> {
}
