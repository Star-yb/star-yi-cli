package com.star.yidemo.service;

import com.star.crud.service.BaseCrudOperations;
import com.star.yidemo.model.dto.DemoArticleCreateInput;
import com.star.yidemo.model.dto.DemoArticleSpecification;
import com.star.yidemo.model.dto.DemoArticleUpdateInput;
import com.star.yidemo.model.entity.DemoArticle;

public interface DemoArticleService extends BaseCrudOperations<
        DemoArticle,
        Long,
        DemoArticleCreateInput,
        DemoArticleUpdateInput,
        DemoArticleSpecification> {
}
