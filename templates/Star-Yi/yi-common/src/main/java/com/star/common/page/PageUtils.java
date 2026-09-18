package com.star.common.page;

import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;

import java.util.List;
import java.util.function.Function;
import java.util.function.Supplier;

/**
 * 分页通用工具：减少 Service 层重复的分页/非分页分支代码。
 */
public final class PageUtils {

    private PageUtils() {
    }

    /**
     * 根据 PageQuery 决定使用分页或非分页，并统一返回 PageResult。
     *
     * @param pageQuery            分页参数
     * @param unpagedSupplier     非分页查询：返回全部 records
     * @param pagedFunction       分页查询：入参 pageable，返回 Page 数据
     */
    public static <T> PageResult<T> paginate(
            PageQuery pageQuery,
            Supplier<List<T>> unpagedSupplier,
            Function<Pageable, Page<T>> pagedFunction
    ) {
        if (pageQuery == null || !pageQuery.isPaged()) {
            return PageResult.ofUnpaged(unpagedSupplier.get());
        }
        Pageable pageable = pageQuery.toPageable();
        return PageResult.ofPaged(pagedFunction.apply(pageable));
    }
}

