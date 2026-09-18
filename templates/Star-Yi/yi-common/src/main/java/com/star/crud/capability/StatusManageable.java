package com.star.crud.capability;

/**
 * 状态型资源通用能力。
 */
public interface StatusManageable<ID> {

    void enable(ID id);

    void disable(ID id);
}
