plugins {
    id("org.springframework.boot")
}

description = "yi-admin"

tasks.bootJar {
    archiveFileName.set("star-yi.jar")
}

dependencies {
    "implementation"(project(":yi-common"))
    "implementation"(project(":yi-demo"))
    "implementation"("org.springframework.boot:spring-boot-starter-actuator")
    "implementation"("org.springframework.boot:spring-boot-starter-jdbc")
    "implementation"("org.springframework.boot:spring-boot-starter-thymeleaf")
    "implementation"("org.apache.commons:commons-pool2")
    "implementation"("com.alibaba:druid-spring-boot-starter:1.2.16")
    "runtimeOnly"("org.postgresql:postgresql")
    "implementation"("cn.dev33:sa-token-jwt:1.46.0")
    "implementation"("cn.dev33:sa-token-thymeleaf:1.46.0")
    "implementation"("cn.dev33:sa-token-redis-template:1.46.0")
    "implementation"("org.springdoc:springdoc-openapi-starter-webmvc-ui:3.0.3")
    // 让 Jackson 能按构造参数反序列化 Kotlin data class。版本跟随 Spring Boot 的 Jackson 3。
    "implementation"("tools.jackson.module:jackson-module-kotlin")
}
