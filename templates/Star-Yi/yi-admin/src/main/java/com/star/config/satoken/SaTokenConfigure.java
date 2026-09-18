package com.star.config.satoken;

import cn.dev33.satoken.SaManager;
import cn.dev33.satoken.context.SaHolder;
import cn.dev33.satoken.filter.SaServletFilter;
import cn.dev33.satoken.interceptor.SaInterceptor;
import cn.dev33.satoken.jwt.StpLogicJwtForSimple;
import cn.dev33.satoken.router.SaHttpMethod;
import cn.dev33.satoken.router.SaRouter;
import cn.dev33.satoken.stp.StpLogic;
import cn.dev33.satoken.stp.StpUtil;
import cn.dev33.satoken.thymeleaf.dialect.SaTokenDialect;
import cn.dev33.satoken.util.SaResult;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.web.servlet.config.annotation.InterceptorRegistry;
import org.springframework.web.servlet.config.annotation.WebMvcConfigurer;

import java.util.Arrays;
import java.util.List;


/**
 * [Sa-Token 权限认证] 配置类
 *
 * @author click33
 */
@Configuration
public class SaTokenConfigure implements WebMvcConfigurer {
    // Sa-Token 整合 jwt (Simple 简单模式)
    @Bean
    public StpLogic getStpLogicJwt() {
        return new StpLogicJwtForSimple();
    }

    // Sa-Token 标签方言 (Thymeleaf版)
    @Bean
    public SaTokenDialect getSaTokenDialect() {
        return new SaTokenDialect();
    }


    /**
     * 自定义 SaJwtUtil 生成 token 的算法
     */
//    @PostConstruct
//    public void setSaJwtTemplate() {
//        SaJwtUtil.setSaJwtTemplate(new SaJwtTemplate() {
//            @Override
//            public String generateToken(JWT jwt, String keyt) {
//                System.out.println("------ 自定义了 token 生成算法");
//                return super.generateToken(jwt, keyt);
//            }
//        });
//    }


    /**
     * 注册 Sa-Token 拦截器打开注解鉴权功能
     */
    @Override
    public void addInterceptors(InterceptorRegistry registry) {


        // 注册 Sa-Token 拦截器打开注解鉴权功能 开启全局登入
        registry.addInterceptor(new SaInterceptor(handler -> {
//            获取SaTokenContext 上下文

            System.out.println("---------- Sa-Token 全局过滤器");


            SaRouter
                    .match("/**")    // 拦截的 path 列表，可以写多个 */
                    .notMatch(excludePaths())        // 排除掉的 path 列表，可以写多个
                    .check(r -> StpUtil.checkLogin());        // 要执行的校验动作，可以写完整的 lambda 表达式

//            超级管理员页面
            SaRouter.match("/admin/**", r -> StpUtil.checkRole("*"));
        })
        ).addPathPatterns("/**").excludePathPatterns("/error");
    }

    public List<String> excludePaths() {
        // 此处仅为示例，实际项目你可以写任意代码来查询这些path
        return Arrays.asList(
                "/static/**", "/**.png",
                "**/**.html", "**/**.css", "**/**.js",
                "/auth/**", "/test/**",
                "/swagger-ui.html", "/swagger-ui/**",
                "/v3/api-docs/**"
        );

    }

    /**
     * 注册 [Sa-Token 全局过滤器]
     */
    @Bean
    public SaServletFilter getSaServletFilter() {
        return new SaServletFilter()

                // 指定 [拦截路由] 与 [放行路由]
                .addInclude("/**").addExclude("/favicon.ico")

                // 认证函数: 每次请求执行
                .setAuth(obj -> {
                    SaManager.getLog().debug("-----请求类型{} 请求path={}  提交token={}", SaHolder.getRequest().getMethod(), SaHolder.getRequest().getRequestPath(), StpUtil.getTokenValue());
                    // ...
                })

                // 异常处理函数：每次认证函数发生异常时执行此函数
                .setError(e -> {
                    System.out.println("---------- 异常处理");
                    return SaResult.error(e.getMessage());
                })

                // 前置函数：在每次认证函数之前执行
                .setBeforeAuth(obj -> {
                    SaHolder.getResponse()

                            // ---------- 设置跨域响应头 ----------
                            // 允许指定域访问跨域资源
                            .setHeader("Access-Control-Allow-Origin", "*")
                            // 允许所有请求方式
                            .setHeader("Access-Control-Allow-Methods", "*")
                            // 允许的header参数
                            .setHeader("Access-Control-Allow-Headers", "*")
                            // 有效时间
                            .setHeader("Access-Control-Max-Age", "3600")
                    ;

                    // 如果是预检请求，则立即返回到前端
                    SaRouter.match(SaHttpMethod.OPTIONS)
                            .free(r -> System.out.println("--------OPTIONS预检请求，不做处理"))
                            .back();
                })
                ;
    }


}
