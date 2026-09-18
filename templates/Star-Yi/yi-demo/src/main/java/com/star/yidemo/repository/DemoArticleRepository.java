package com.star.yidemo.repository;

import com.star.crud.service.AbstractJimmerCrudRepository;
import com.star.yidemo.dao.DemoArticleDao;
import com.star.yidemo.model.dto.DemoArticleCreateInput;
import com.star.yidemo.model.dto.DemoArticleSpecification;
import com.star.yidemo.model.dto.DemoArticleUpdateInput;
import com.star.yidemo.model.dto.DemoArticleView;
import com.star.yidemo.model.entity.DemoArticle;
import com.star.yidemo.model.entity.DemoArticleTable;
import org.babyfish.jimmer.sql.fetcher.Fetcher;
import org.springframework.stereotype.Repository;

@Repository
public class DemoArticleRepository extends AbstractJimmerCrudRepository<
        DemoArticle,
        Long,
        DemoArticleCreateInput,
        DemoArticleUpdateInput,
        DemoArticleSpecification,
        DemoArticleTable> {

    private final DemoArticleDao demoArticleDao;

    public DemoArticleRepository(DemoArticleDao demoArticleDao) {
        this.demoArticleDao = demoArticleDao;
    }

    @Override
    protected DemoArticleDao dao() {
        return demoArticleDao;
    }

    @Override
    protected Fetcher<DemoArticle> fetcher() {
        return DemoArticleView.METADATA.getFetcher();
    }
}
