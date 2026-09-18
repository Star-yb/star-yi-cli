package com.star.yidemo.controller.demo;

import com.star.crud.controller.BaseCrudController;
import com.star.crud.service.BaseCrudOperations;
import com.star.yidemo.model.dto.DemoArticleCreateInput;
import com.star.yidemo.model.dto.DemoArticleSpecification;
import com.star.yidemo.model.dto.DemoArticleUpdateInput;
import com.star.yidemo.model.entity.DemoArticle;
import com.star.yidemo.service.DemoArticleService;
import io.swagger.v3.oas.annotations.tags.Tag;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

@Tag(name = "演示-文章", description = "DemoArticle 标准 CRUD 示例，路径 /demo/articles")
@RestController
@RequestMapping("/demo/articles")
public class DemoArticleController extends BaseCrudController<
        DemoArticle,
        Long,
        DemoArticleCreateInput,
        DemoArticleUpdateInput,
        DemoArticleSpecification> {

    private final DemoArticleService demoArticleService;

    public DemoArticleController(DemoArticleService demoArticleService) {
        this.demoArticleService = demoArticleService;
    }

    @Override
    protected BaseCrudOperations<
            DemoArticle,
            Long,
            DemoArticleCreateInput,
            DemoArticleUpdateInput,
            DemoArticleSpecification> service() {
        return demoArticleService;
    }

    @Override
    protected void bindId(Long id, DemoArticleUpdateInput input) {
        input.setId(id);
    }
}
