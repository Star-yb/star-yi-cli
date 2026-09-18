package com.star.crud.controller;

import cn.dev33.satoken.util.SaResult;
import com.star.common.page.PageQuery;
import com.star.crud.service.BaseCrudOperations;
import io.swagger.v3.oas.annotations.Operation;
import org.springframework.web.bind.annotation.*;

/**
 * 通用 CRUD Controller 基类。
 * <p>
 * 是否使用该基类由具体模块自行决定，但公共层保留这一能力，供后续模块直接复用。
 */
public abstract class BaseCrudController<E, ID, C, U, S> {

    protected abstract BaseCrudOperations<E, ID, C, U, S> service();

    /**
     * 由子类决定如何把路径 ID 注入更新对象。
     */
    protected void bindId(ID id, U input) {
    }

    @Operation(summary = "创建")
    @PostMapping
    public SaResult create(@RequestBody C input) {
        return SaResult.data(service().create(input));
    }

    @Operation(summary = "更新")
    @PutMapping("/{id}")
    public SaResult update(@PathVariable ID id, @RequestBody U input) {
        bindId(id, input);
        return SaResult.data(service().update(input));
    }

    @Operation(summary = "删除")
    @DeleteMapping("/{id}")
    public SaResult delete(@PathVariable ID id) {
        service().delete(id);
        return SaResult.ok("删除成功");
    }

    @Operation(summary = "详情")
    @GetMapping("/{id}")
    public SaResult obtain(@PathVariable ID id) {
        return SaResult.data(service().obtain(id));
    }

    @Operation(summary = "分页列表")
    @GetMapping
    public SaResult list(PageQuery pageQuery, S specification) {
        return SaResult.data(service().page(pageQuery, specification));
    }
}
