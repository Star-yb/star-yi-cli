package com.star.config;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.boot.context.event.ApplicationReadyEvent;
import org.springframework.context.ApplicationListener;
import org.springframework.core.env.Environment;
import org.springframework.stereotype.Component;

/**
 * 应用完全就绪后输出启动摘要（在 Tomcat 启动、Bean 初始化完成之后）。
 */
@Component
public class ApplicationStartupListener implements ApplicationListener<ApplicationReadyEvent> {

    private static final Logger log = LoggerFactory.getLogger(ApplicationStartupListener.class);

    private final Environment environment;

    public ApplicationStartupListener(Environment environment) {
        this.environment = environment;
    }

    @Override
    public void onApplicationEvent(ApplicationReadyEvent event) {
        int port = resolvePort();
        String contextPath = resolveContextPath();
        String baseUrl = "http://localhost:" + port + contextPath;
        String appName = environment.getProperty("spring.application.name", "application");

        log.info("""
                
                ----------------------------------------------------------
                  启动完毕 · {}
                  API 根地址:     {}
                  Swagger 文档:   {}/swagger-ui.html
                  管理后台:       {}/admin/login
                ----------------------------------------------------------
                """, appName, baseUrl, baseUrl, baseUrl);
    }

    /**
     * 优先取 Spring Boot 启动后写入的运行时端口 local.server.port（与 Tomcat 日志一致），
     * 否则读取 application.yaml 中的 server.port。
     */
    private int resolvePort() {
        Integer port = environment.getProperty("local.server.port", Integer.class);
        if (port != null) {
            return port;
        }
        return environment.getRequiredProperty("server.port", Integer.class);
    }

    /** 从 application.yaml 的 server.servlet.context-path（或 contextPath）读取。 */
    private String resolveContextPath() {
        String contextPath = environment.getProperty("server.servlet.context-path", "");
        if (contextPath == null || contextPath.isBlank() || "/".equals(contextPath)) {
            return "";
        }
        return contextPath.endsWith("/") ? contextPath.substring(0, contextPath.length() - 1) : contextPath;
    }
}
