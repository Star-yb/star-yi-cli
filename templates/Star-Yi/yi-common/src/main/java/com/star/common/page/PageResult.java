package com.star.common.page;

import lombok.Data;
import org.springframework.data.domain.Page;

import java.util.List;

/**
 * 通用分页返回对象，统一分页与非分页场景的返回结构。
 */
@Data
public class PageResult<T> {

    /**
     * 是否分页查询
     */
    private boolean paged;

    /**
     * 当前页（前端语义：从 1 开始；非分页时为 0）
     */
    private int page;

    /**
     * 每页条数（非分页时为 records.size()）
     */
    private int size;

    /**
     * 是否是第一页
     */
    private boolean first;

    /**
     * 是否是最后一页
     */
    private boolean last;

    /**
     * 总记录数
     */
    private long count;

    /**
     * 实际数据
     */
    private List<T> results;

    public static <T> PageResult<T> ofUnpaged(List<T> records) {
        PageResult<T> result = new PageResult<>();
        result.setPaged(false);
        result.setPage(0);
        result.setSize(records == null ? 0 : records.size());
        // 非分页：只有一页语义
        result.setFirst(true);
        result.setLast(true);
        result.setCount(records == null ? 0 : records.size());
        result.setResults(records);
        return result;
    }

    /**
     * 总页数（前端分页展示用；无数据时为 0）。
     */
    public int getTotalPages() {
        if (count <= 0) {
            return 0;
        }
        if (!paged || size <= 0) {
            return 1;
        }
        return (int) ((count + size - 1) / size);
    }

    public static <T> PageResult<T> ofPaged(Page<T> pageData) {
        PageResult<T> result = new PageResult<>();
        result.setPaged(true);
        result.setPage(pageData.getNumber() + 1); // Spring Page 从 0 开始，转换为前端从 1 开始
        result.setSize(pageData.getSize());
        // 分页：直接使用 Page 的语义判断
        result.setFirst(pageData.isFirst());
        result.setLast(pageData.isLast());
        result.setCount(pageData.getTotalElements());
        result.setResults(pageData.getContent());
        return result;
    }
}

