plugins {
    // 编译 Kotlin 源码。apply false：只在根工程登记版本，不把根工程当成要编译的模块。
    kotlin("jvm") version "2.3.21" apply false
    // 让带 Spring 注解的 Kotlin 类默认 open，Spring 才能生成代理。
    kotlin("plugin.spring") version "2.3.21" apply false
    // Spring Boot 4.1.1。可执行插件只由 yi-admin 启用，根工程不启动应用。
    id("org.springframework.boot") version "4.1.1" apply false
    // 按 Spring Boot 的版本清单解析依赖版本，子模块里多数 Spring 依赖不用手写版本号。
    id("io.spring.dependency-management") version "1.1.7" apply false
    // Jimmer 用 KSP 生成实体的 Props、Fetcher、Draft，以及 dto 文件对应的类型。
    id("com.google.devtools.ksp") version "2.3.9" apply false
}

// 根工程描述，出现在 gradle projects 的输出里。
description = "Star-Yi Arc"


// 根工程和全部子模块共用的坐标与仓库。
allprojects {
    // Maven 坐标中的 groupId。
    group = "com.star"
    // Maven 坐标中的 version。
    version = "0.0.1-SNAPSHOT"

    repositories {
        // 依赖从 Maven Central 下载。
        mavenCentral()
    }
}

// 只作用于子模块。根工程没有源码，也不套用这些插件和依赖。
subprojects {
    // 真正启用上面登记过的 Kotlin / Spring 版本管理插件。
    apply(plugin = "org.jetbrains.kotlin.jvm")
    apply(plugin = "org.jetbrains.kotlin.plugin.spring")
    apply(plugin = "io.spring.dependency-management")
    // 每个子模块都可能放实体或 dto，生成配置统一放在这里。
    apply(plugin = "com.google.devtools.ksp")

    // IDE 同步会向每个子模块要这个任务；Kotlin 插件默认只在根工程注册它。
    tasks.maybeCreate("prepareKotlinBuildScriptModel")

    extensions.configure<JavaPluginExtension> {
        toolchain {
            // 编译和运行使用 JDK 21。
            languageVersion.set(JavaLanguageVersion.of(21))
        }
    }

    extensions.configure<org.jetbrains.kotlin.gradle.dsl.KotlinJvmProjectExtension> {
        compilerOptions {
            // 按 JSR-305 注解严格处理 Java 可空性；注解默认落在构造参数对应的属性上。
            freeCompilerArgs.addAll("-Xjsr305=strict", "-Xannotation-default-target=param-property")
        }
        sourceSets.named("main") {
            // 让编译和 IDE 都能看到 KSP 生成的源码。没有这个目录时，命令行仍能编译，IDE 会找不到生成类。
            kotlin.srcDir("build/generated/ksp/main/kotlin")
        }
    }

    extensions.configure<io.spring.gradle.dependencymanagement.dsl.DependencyManagementExtension> {
        imports {
            // 导入与 Spring Boot 插件相同版本的 BOM，统一 Spring 相关依赖版本。
            mavenBom(org.springframework.boot.gradle.plugin.SpringBootPlugin.BOM_COORDINATES)
        }
    }

    dependencies {
        // Kotlin 反射。Spring 创建 Bean、绑定配置时需要它。
        "implementation"("org.jetbrains.kotlin:kotlin-reflect")
        // Web 与 Spring MVC。各业务模块都要编译自己的 Controller，所以放在父工程。
        "implementation"("org.springframework.boot:spring-boot-starter-web")
        // 单元测试与 Spring 测试支持，只在测试源码中可用。
        "testImplementation"("org.springframework.boot:spring-boot-starter-test")
        // 让 Kotlin 测试代码使用 JUnit 5。
        "testImplementation"("org.jetbrains.kotlin:kotlin-test-junit5")
        // Gradle 运行测试时加载 JUnit 平台，只在测试运行期需要。
        "testRuntimeOnly"("org.junit.platform:junit-platform-launcher")
        
        // Sa-Token 权限认证，在线文档：https://sa-token.com
        "implementation"("cn.dev33:sa-token-spring-boot4-starter:1.46.0")
        "implementation"("org.babyfish.jimmer:jimmer-spring-boot-starter:0.12.2")
        // 扫描本模块的实体和 src/main/dto。
        "ksp"("org.babyfish.jimmer:jimmer-ksp:0.12.2")
    
    }

    tasks.withType<Test> {
        // 测试任务使用 JUnit Platform，才能跑 JUnit 5。
        useJUnitPlatform()
    }
}
