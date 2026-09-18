package com.star.admin.web;

import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.thymeleaf.spring6.templateresolver.SpringResourceTemplateResolver;
import org.thymeleaf.templatemode.TemplateMode;

import java.nio.charset.StandardCharsets;

/**
 * 支持从各业务 JAR 的 {@code classpath:/admin/} 加载 Thymeleaf 模板。
 *
 * <p>与 yi-admin 自带的 {@code classpath:/templates/} 并存：本解析器优先尝试
 * {@code admin/<模块>/<页面>.html}，不存在时回退到默认 templates 目录。</p>
 */
@Configuration
public class AdminThymeleafConfigure {

    @Bean
    public SpringResourceTemplateResolver adminModuleTemplateResolver() {
        SpringResourceTemplateResolver resolver = new SpringResourceTemplateResolver();
        resolver.setPrefix("classpath:/admin/");
        resolver.setSuffix(".html");
        resolver.setTemplateMode(TemplateMode.HTML);
        resolver.setCharacterEncoding(StandardCharsets.UTF_8.name());
        resolver.setCheckExistence(true);
        resolver.setOrder(1);
        resolver.setName("adminModuleTemplateResolver");
        return resolver;
    }
}
