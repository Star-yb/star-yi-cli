package com.star.config.openapi;

import com.star.openapi.ApiDocConstants;
import io.swagger.v3.oas.models.Components;
import io.swagger.v3.oas.models.OpenAPI;
import io.swagger.v3.oas.models.info.Info;
import io.swagger.v3.oas.models.security.SecurityRequirement;
import io.swagger.v3.oas.models.security.SecurityScheme;
import org.springdoc.core.customizers.GlobalOpenApiCustomizer;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

/**
 * OpenAPI 文档信息与 satoken 鉴权方案。
 */
@Configuration
public class OpenApiConfigure {

    @Bean
    public OpenAPI openAPI() {
        return new OpenAPI()
                .info(new Info()
                        .title("{projectName} API")
                        .description("{projectName} 接口文档（OpenAPI 3 + Swagger UI）")
                        .version("0.0.1"))
                .components(new Components()
                        .addSecuritySchemes(ApiDocConstants.SECURITY_SCHEME_SATOKEN,
                                new SecurityScheme()
                                        .type(SecurityScheme.Type.APIKEY)
                                        .in(SecurityScheme.In.HEADER)
                                        .name(ApiDocConstants.SECURITY_SCHEME_SATOKEN)));
    }

    /**
     * 为需登录路径打上 satoken 安全要求；公开路径（如 /auth/**）不打标。
     */
    @Bean
    public GlobalOpenApiCustomizer satokenSecurityCustomizer() {
        return openApi -> {
            if (openApi.getPaths() == null) {
                return;
            }
            openApi.getPaths().forEach((path, pathItem) -> {
                if (isPublicPath(path)) {
                    return;
                }
                pathItem.readOperations().forEach(operation ->
                        operation.addSecurityItem(new SecurityRequirement()
                                .addList(ApiDocConstants.SECURITY_SCHEME_SATOKEN)));
            });
        };
    }

    /**
     * 与 {@link com.star.config.satoken.SaTokenConfigure#excludePaths} 中免登录前缀保持一致。
     */
    static boolean isPublicPath(String path) {
        return path.startsWith("/auth/")
                || path.startsWith("/test/")
                || path.startsWith("/swagger-ui")
                || path.startsWith("/v3/api-docs")
                || path.equals("/swagger-ui.html");
    }
}
