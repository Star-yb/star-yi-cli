package com.star.yidemo.dao;

import com.star.crud.dao.BaseJimmerDao;
import com.star.yidemo.model.entity.DemoArticle;
import com.star.yidemo.model.entity.DemoArticleTable;
import org.babyfish.jimmer.sql.JSqlClient;
import org.springframework.stereotype.Repository;

@Repository
public class DemoArticleDao extends BaseJimmerDao<DemoArticle, Long, DemoArticleTable> {

    public DemoArticleDao(JSqlClient sql) {
        super(sql, DemoArticle.class, DemoArticleTable.$);
    }
}
