package com.star.admin.support;

import com.star.admin.model.entity.LoginLogs;
import com.star.admin.model.entity.LoginLogsTable;
import com.star.common.page.PageQuery;
import com.star.common.page.PageResult;
import com.star.common.page.PageUtils;
import jakarta.annotation.Nullable;
import lombok.RequiredArgsConstructor;
import org.babyfish.jimmer.spring.repository.support.SpringPageFactory;
import org.babyfish.jimmer.sql.JSqlClient;
import org.babyfish.jimmer.sql.ast.Predicate;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.stereotype.Component;

import java.time.LocalDateTime;
import java.util.ArrayList;
import java.util.List;

/**
 * 登录日志查询与清理（管理端专用，不经 Dao/Service 分层）。
 */
@Component
@RequiredArgsConstructor
public class LoginLogSupport {

    private static final LoginLogsTable T = LoginLogsTable.$;

    private final JSqlClient sql;

    public PageResult<LoginLogs> list(
            PageQuery pageQuery,
            @Nullable String username,
            @Nullable Integer loginType,
            @Nullable Integer status
    ) {
        if (pageQuery.getSort() == null || pageQuery.getSort().isBlank()) {
            pageQuery.setSort("loginTime");
            pageQuery.setOrder("DESC");
        }
        Predicate predicate = buildPredicate(username, loginType, status);
        return PageUtils.paginate(
                pageQuery,
                () -> queryList(predicate),
                pageable -> queryPage(pageable, predicate)
        );
    }

    public long count() {
        return sql.createQuery(T).select(T).fetchUnlimitedCount();
    }

    public int deleteByIds(List<Long> ids) {
        if (ids == null || ids.isEmpty()) {
            return 0;
        }
        return sql.createDelete(T)
                .where(T.id().in(ids))
                .execute();
    }

    public int deleteOlderThanDays(int days) {
        if (days <= 0) {
            return 0;
        }
        LocalDateTime before = LocalDateTime.now().minusDays(days);
        return sql.createDelete(T)
                .where(T.loginTime().lt(before))
                .execute();
    }

    public int deleteKeepLatest(int keepCount) {
        if (keepCount <= 0) {
            return deleteAll();
        }
        List<Long> keepIds = sql.createQuery(T)
                .orderBy(T.loginTime().desc())
                .select(T.id())
                .limit(keepCount)
                .execute();
        if (keepIds.isEmpty()) {
            return 0;
        }
        long total = count();
        if (total <= keepCount) {
            return 0;
        }
        return sql.createDelete(T)
                .where(Predicate.not(T.id().in(keepIds)))
                .execute();
    }

    public int deleteAll() {
        return sql.createDelete(T)
                .where(Predicate.sql("1 = 1"))
                .execute();
    }

    private List<LoginLogs> queryList(@Nullable Predicate predicate) {
        var query = sql.createQuery(T);
        if (predicate != null) {
            query.where(predicate);
        }
        return query.orderBy(T.loginTime().desc())
                .select(T)
                .execute();
    }

    private Page<LoginLogs> queryPage(Pageable pageable, @Nullable Predicate predicate) {
        var query = sql.createQuery(T);
        if (predicate != null) {
            query.where(predicate);
        }
        return query.orderBy(T.loginTime().desc())
                .select(T)
                .fetchPage(pageable.getPageNumber(), pageable.getPageSize(), SpringPageFactory.getInstance());
    }

    @Nullable
    private Predicate buildPredicate(
            @Nullable String username,
            @Nullable Integer loginType,
            @Nullable Integer status
    ) {
        List<Predicate> predicates = new ArrayList<>();
        if (username != null && !username.isBlank()) {
            predicates.add(T.username().ilike("%" + username.trim() + "%"));
        }
        if (loginType != null) {
            predicates.add(T.loginType().eq(loginType));
        }
        if (status != null) {
            predicates.add(T.status().eq(status));
        }
        if (predicates.isEmpty()) {
            return null;
        }
        return Predicate.and(predicates.toArray(Predicate[]::new));
    }
}
