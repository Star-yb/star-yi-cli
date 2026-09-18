package com.star.yidemo.service.impl;

import com.star.crud.service.AbstractCrudService;
import com.star.crud.service.CrudRepository;
import com.star.yidemo.model.dto.DemoArticleCreateInput;
import com.star.yidemo.model.dto.DemoArticleSpecification;
import com.star.yidemo.model.dto.DemoArticleUpdateInput;
import com.star.yidemo.model.entity.DemoArticle;
import com.star.yidemo.repository.DemoArticleRepository;
import com.star.yidemo.service.DemoArticleService;
import org.springframework.stereotype.Service;

@Service
public class DemoArticleServiceImpl extends AbstractCrudService<
        DemoArticle,
        Long,
        DemoArticleCreateInput,
        DemoArticleUpdateInput,
        DemoArticleSpecification> implements DemoArticleService {

    private final DemoArticleRepository demoArticleRepository;

    public DemoArticleServiceImpl(DemoArticleRepository demoArticleRepository) {
        this.demoArticleRepository = demoArticleRepository;
    }

    @Override
    protected CrudRepository<
            DemoArticle,
            Long,
            DemoArticleCreateInput,
            DemoArticleUpdateInput,
            DemoArticleSpecification> repository() {
        return demoArticleRepository;
    }
}
