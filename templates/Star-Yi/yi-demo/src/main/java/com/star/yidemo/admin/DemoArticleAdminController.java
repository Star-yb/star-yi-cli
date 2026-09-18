package com.star.yidemo.admin;

import com.star.admin.web.AdminPageAttributes;
import com.star.common.page.PageQuery;
import com.star.common.page.PageResult;
import com.star.yidemo.model.dto.DemoArticleSpecification;
import com.star.yidemo.model.entity.DemoArticle;
import com.star.yidemo.service.DemoArticleService;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.tags.Tag;
import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Controller;
import org.springframework.ui.Model;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestMapping;

/**
 * 演示文章超级管理员页面（模板位于本模块 {@code resources/admin/}）。
 */
@Tag(name = "超级管理员-演示文章", description = "DemoArticle Thymeleaf 管理页")
@Controller
@RequestMapping("/admin/demo/articles")
@RequiredArgsConstructor
public class DemoArticleAdminController {

    private final DemoArticleService demoArticleService;

    @Operation(summary = "演示文章管理页")
    @GetMapping
    public String index(PageQuery pageQuery,
                        DemoArticleSpecification specification,
                        Model model) {
        if (pageQuery.getPage() < 1) {
            pageQuery.setPage(1);
        }
        PageResult<DemoArticle> pageResult = demoArticleService.page(pageQuery, specification);

        AdminPageAttributes.enrich(model, "demo-article", "开发演示", "演示文章");
        model.addAttribute("articlesPageResult", pageResult);
        return "yi-demo/index";
    }
}
