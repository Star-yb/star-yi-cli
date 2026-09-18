package com.star.common.page;


import lombok.Data;
import org.springframework.data.domain.PageRequest;
import org.springframework.data.domain.Pageable;
import org.springframework.data.domain.Sort;

import java.util.ArrayList;
import java.util.List;

@Data
public class PageQuery {
    /**
     * 前端页码（从 1 开始）。
     * <p>
     * 默认 0：用于区分前端是否传了该参数；当 page <= 0 时表示不分页。
     */
    private Integer page = 0;

    /**
     * 每页条数（默认 10）。
     */
    private Integer size = 10;

    /**
     * 排序字段（默认按 id）。
     */
    private String sort = "id";

    /**
     * 排序方向（ASC/DESC），默认 ASC。
     */
    private String order = "ASC";

    /**
     * 是否需要分页：page > 0 视为前端已正确传值。
     */
    public boolean isPaged() {
        return page != null && page > 0;
    }

    /**
     * 将本地页码（从 1 开始）转换为 Spring Data 的 Pageable（从 0 开始）。
     * <p>
     * 如果 page <= 0，则返回 Pageable.unpaged()（调用方应根据 isPaged() 选择使用分页/非分页接口）。
     */
    public Pageable toPageable() {
        if (!isPaged()) {
            return Pageable.unpaged();
        }

        int safeSize = (size == null || size <= 0) ? 10 : size;
        int pageIndex = page - 1; // 前端从 1 开始，Spring Data 从 0 开始

        Sort.Direction direction;
        try {
            direction = Sort.Direction.fromString(order);
        } catch (Exception e) {
            direction = Sort.Direction.ASC;
        }

        if (sort == null || sort.isBlank()) {
            return PageRequest.of(pageIndex, safeSize);
        }

        Sort sortObj = buildSort(direction);
        return PageRequest.of(pageIndex, safeSize, sortObj);
    }

    /**
     * 解析排序字段，支持以下格式：
     * <ul>
     *     <li>`sort=superAdmin,nickname`（多个字段，统一使用 order）</li>
     *     <li>`sort=superAdmin:DESC,nickname:ASC`（每个字段单独方向）</li>
     * </ul>
     */
    private Sort buildSort(Sort.Direction defaultDirection) {
        String[] rawFields = sort.split(",");
        List<Sort.Order> orders = new ArrayList<>();

        for (String raw : rawFields) {
            if (raw == null) {
                continue;
            }
            String token = raw.trim();
            if (token.isEmpty()) {
                continue;
            }

            // 支持 field:ASC / field:DESC
            if (token.contains(":")) {
                String[] parts = token.split(":", 2);
                String field = parts[0].trim();
                String dirText = parts[1].trim();
                if (!field.isEmpty()) {
                    try {
                        Sort.Direction dir = Sort.Direction.fromString(dirText);
                        orders.add(new Sort.Order(dir, field));
                        continue;
                    } catch (Exception ignore) {
                        // 方向非法时回退到默认方向
                    }
                }
                if (!field.isEmpty()) {
                    orders.add(new Sort.Order(defaultDirection, field));
                }
                continue;
            }

            // 仅字段名，使用默认方向
            orders.add(new Sort.Order(defaultDirection, token));
        }

        if (orders.isEmpty()) {
            return Sort.unsorted();
        }
        return Sort.by(orders);
    }
}
