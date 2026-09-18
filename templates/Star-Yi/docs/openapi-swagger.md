# OpenAPI / Swagger 接入说明（Education-Yi）

本文说明本项目如何接入 **springdoc-openapi**（OpenAPI 3 + Swagger UI），以及各文件的分工与使用方式。

**全文案例统一以 `yi-demo` 模块的 `DemoArticle`（演示文章 CRUD）为例**，该模块是项目内标准 CRUD 金标准，与 `apps/*` 业务模块写法一致。

---

## 0. Demo 案例一览（DemoArticle）

| 项 | 路径 / 值 |
|----|-----------|
| 实体 | `yi-demo/.../model/entity/DemoArticle.java` |
| DTO | `yi-demo/src/main/dto/DemoArticle.dto` |
| Controller | `yi-demo/.../controller/demo/DemoArticleController.java` |
| REST 前缀 | `/demo/articles` |
| Swagger 分组（建议） | `@Tag(name = "演示-文章")` 加在 Controller 类上 |
| 继承基类 | `BaseCrudController` → 自动带 5 个 `@Operation`（创建/更新/删除/详情/分页） |

**Demo 在文档中的 5 个接口（由 `BaseCrudController` 提供）：**

| 方法 | 路径 | 文档摘要 |
|------|------|----------|
| `POST` | `/demo/articles` | 创建 |
| `PUT` | `/demo/articles/{id}` | 更新 |
| `DELETE` | `/demo/articles/{id}` | 删除 |
| `GET` | `/demo/articles/{id}` | 详情 |
| `GET` | `/demo/articles` | 分页列表（`PageQuery` + `DemoArticleSpecification`） |

`yi-demo` 已被 `yi-admin` 依赖，启动 `yi-admin` 后扫描 `com.star` 时会自动包含上述接口，**无需**在 `yi-demo` 中再引入 springdoc starter。

---

## 1. 架构：为什么拆成两层？

| 层级 | 模块 | 依赖 / 文件 | 作用（以 Demo 为例） |
|------|------|-------------|----------------------|
| 注解层 | `yi-common` | `swagger-annotations-jakarta` | `DemoArticleController` 可写 `@Tag`；`BaseCrudController` 为 Demo 的 5 个接口提供 `@Operation` |
| 常量层 | `yi-common` | `ApiDocConstants.java` | Swagger Authorize 与登录 token 头均用 `satoken` |
| 运行时层 | `yi-admin` | `springdoc-openapi-starter-webmvc-ui` | 启动后生成 `/demo/articles` 等全部 `com.star` 下接口文档 |
| 运行时配置 | `yi-admin` | `OpenApiConfigure.java` | 文档标题、satoken 方案；`/demo/**` 需登录，`/auth/**` 免鉴权 |
| 运行时配置 | `yi-admin` | `application.yaml → springdoc.*` | `packages-to-scan: com.star` 扫到 `com.star.yidemo.*` |
| 鉴权放行 | `yi-admin` | `SaTokenConfigure#excludePaths` | 仅放行 Swagger 页面，**不放行** `/demo/articles` |

**为何不把完整 springdoc 放进 `yi-common`？**

- `yi-demo`、`apps/*` 都是 **jar 库**，不单独启动 Spring。
- 完整 starter 只需在 **`yi-admin` 装一份**，启动时扫描 `com.star.yidemo.controller.demo.DemoArticleController` 即可出现在 UI 中。
- 业务模块只通过 `yi-common` 拿 **注解** 即可（与 Demo 相同方式）。

---

## 2. 接入流程（按实施顺序，对照 Demo）

### 步骤 1：父 POM 统一管理版本

**文件：** 根目录 `pom.xml`

```xml
<springdoc.version>3.0.3</springdoc.version>
<swagger-annotations.version>2.2.30</swagger-annotations.version>
```

在 `dependencyManagement` 中声明：

- `springdoc-openapi-starter-webmvc-ui` — 兼容 **Spring Boot 4.x**（仅 `yi-admin` 引用）
- `swagger-annotations-jakarta` — 经 `yi-common` 传给 `yi-demo`、`apps/*`

### 步骤 2：`yi-common` 传递 Swagger 注解能力

**文件：** `yi-common/pom.xml`

```xml
<dependency>
    <groupId>io.swagger.core.v3</groupId>
    <artifactId>swagger-annotations-jakarta</artifactId>
</dependency>
```

`yi-demo` 已依赖 `yi-common`，因此 **Demo 模块不用再加** swagger 依赖。

**文件：** `yi-common/.../openapi/ApiDocConstants.java`

- `SECURITY_SCHEME_SATOKEN = "satoken"`
- 调试 Demo 接口时，Authorize 填写的头名称即此常量

**文件：** `yi-common/.../crud/controller/BaseCrudController.java`

Demo 的 `DemoArticleController` 继承本类后，文档中自动出现：

```text
创建 / 更新 / 删除 / 详情 / 分页列表
```

无需在 Demo Controller 里为每个方法再写 `@Operation`（若要覆盖摘要，可在子类方法上单独写）。

### 步骤 3：`yi-admin` 引入 springdoc 运行时

**文件：** `yi-admin/pom.xml`

```xml
<dependency>
    <groupId>org.springdoc</groupId>
    <artifactId>springdoc-openapi-starter-webmvc-ui</artifactId>
</dependency>
```

`yi-admin` 的 `pom.xml` 中已有对 `yi-demo` 的依赖 → 启动后 Demo 接口进入同一套文档。

### 步骤 4：YAML 配置 springdoc 行为

**文件：** `yi-admin/src/main/resources/application.yaml`（约 85–99 行）

| 配置项 | 含义 | 对 Demo 的影响 |
|--------|------|----------------|
| `springdoc.api-docs.enabled` | 是否输出 OpenAPI JSON | 含 `/demo/articles` 等路径定义 |
| `springdoc.api-docs.path` | JSON 地址 | `GET /v3/api-docs` 可搜 `demo` |
| `springdoc.swagger-ui.enabled` | 是否启用 UI | 左侧可见 Demo 分组 |
| `springdoc.swagger-ui.path` | UI 入口 | 浏览器打开调试 Demo |
| `springdoc.packages-to-scan` | 扫描包 | `com.star` 包含 `com.star.yidemo` |
| `springdoc.show-actuator` | Actuator 进文档 | 本项目 `false` |

### 步骤 5：Java 配置文档信息与 satoken

**文件：** `yi-admin/.../config/openapi/OpenApiConfigure.java`

1. **`openAPI()` Bean** — 文档标题、注册 `satoken` Header 方案。  
2. **`satokenSecurityCustomizer()` Bean** — 为需登录路径打标。  
   - **公开**：`/auth/**`（先调登录拿 token）  
   - **需 satoken**：`/demo/articles` 及全部业务 API  

### 步骤 6：Sa-Token 放行文档 URL

**文件：** `yi-admin/.../config/satoken/SaTokenConfigure.java` → `excludePaths()`

| 路径 | 是否放行 | 说明 |
|------|----------|------|
| `/swagger-ui.html`、`/swagger-ui/**` | ✅ | 打开文档页 |
| `/v3/api-docs/**` | ✅ | 拉取 OpenAPI JSON |
| `/demo/articles` | ❌ | Demo 业务接口，须 Authorize |
| `/test/**` | ✅ | 项目测试接口（若有） |

须与 `OpenApiConfigure#isPublicPath` 保持一致。

### 步骤 7：Demo Controller 加分组（推荐）

**文件：** `yi-demo/.../controller/demo/DemoArticleController.java`

在现有类上增加 `@Tag`（依赖已由 `yi-common` 传递）：

```java
import io.swagger.v3.oas.annotations.tags.Tag;

@Tag(name = "演示-文章", description = "DemoArticle 标准 CRUD 示例，路径 /demo/articles")
@RestController
@RequestMapping("/demo/articles")
public class DemoArticleController extends BaseCrudController<
        DemoArticle,
        Long,
        DemoArticleCreateInput,
        DemoArticleUpdateInput,
        DemoArticleSpecification> {
    // service()、bindId() 与现有一致，无需为 Swagger 改逻辑
}
```

**DTO 与文档参数（Demo 已有，供对照）：**

`yi-demo/src/main/dto/DemoArticle.dto`：

- `DemoArticleCreateInput` → `POST /demo/articles` 请求体  
- `DemoArticleUpdateInput` → `PUT /demo/articles/{id}` 请求体  
- `DemoArticleSpecification` → `GET /demo/articles` 查询参数（如 `articleCode`、`articleTitle`）  
- `DemoArticleView` → 响应中展示字段（隐藏 `deletedTime`）

复制新业务时：**照 Demo 的 dto + Controller + 五层结构**，再在 Controller 上加自己的 `@Tag` 即可。

---

## 3. 本地使用（用 Demo 走通一遍）

### 3.1 启动

```bash
mvn clean compile
mvn -pl yi-admin spring-boot:run
```

默认端口：**8500**。

### 3.2 打开文档

| 用途 | URL |
|------|-----|
| Swagger UI | http://localhost:8500/swagger-ui.html |
| OpenAPI JSON | http://localhost:8500/v3/api-docs |

在 UI 左侧找到 **「演示-文章」** 分组（若已加 `@Tag`），或按路径搜索 `demo/articles`。

### 3.3 用 Demo 调试需登录的接口

1. 展开 **auth** 相关接口，调用 `POST /auth/login`，从响应取得 token。  
2. 点击右上角 **Authorize**，填入 token（**只填值**，不要 `Bearer ` 前缀）。  
3. 在 **演示-文章** 分组中依次验证：  
   - `POST /demo/articles` — Body 示例：

     ```json
     {
       "articleCode": "DEMO-001",
       "articleTitle": "Swagger 测试文章",
       "content": "通过文档页创建"
     }
     ```

   - `GET /demo/articles` — 分页列表，可加查询参数 `articleCode`、`articleTitle`  
   - `GET /demo/articles/{id}` — 详情  
   - `PUT /demo/articles/{id}` — 更新标题或内容  
   - `DELETE /demo/articles/{id}` — 逻辑删除  

请求头名称：`satoken`（与 `ApiDocConstants.SECURITY_SCHEME_SATOKEN` 一致）。

### 3.4 Demo 与 `/test/**` 的区别

- `/test/**` 在 `excludePaths` 中 **免登录**，一般不进 Swagger 业务分组。  
- `/demo/articles` **需要登录**，专门用来验证「文档 + Sa-Token」是否配置正确。  
- 新业务（如 `apps/ed-school` 的 `/ed/schools`）与 Demo **同一套** Authorize 流程，只是路径和 `@Tag` 名称不同。

---

## 4. 从 Demo 复制到新业务模块

以 `DemoArticle` 为模板，新增 `apps/xxx` 时：

| Demo（yi-demo） | 新业务（apps/*） |
|-----------------|------------------|
| `DemoArticle.java` | 你的 `@Entity` |
| `DemoArticle.dto` | `YourEntity.dto` |
| `DemoArticleDao` → `Repository` → `Service` → `Controller` | 同结构五层 |
| `@RequestMapping("/demo/articles")` | 如 `/ed/schools` |
| `@Tag(name = "演示-文章")` | `@Tag(name = "学校")` 等 |

Checklist：

- [ ] 模块 `pom.xml` 依赖 `yi-common`（与 Demo 相同）  
- [ ] `yi-admin/pom.xml` 增加对新 `apps/xxx` 模块的依赖（Demo 已依赖 `yi-demo`）  
- [ ] Controller 继承 `BaseCrudController` 并实现 `bindId`（见 `DemoArticleController`）  
- [ ] Controller 上加 `@Tag`（可选，便于 Swagger 分组）  
- [ ] **不要**在 `apps/*` 或 `yi-demo` 中引入 `springdoc-openapi-starter-webmvc-ui`  
- [ ] 若新增免登录 API 前缀，同步改 `SaTokenConfigure#excludePaths` 与 `OpenApiConfigure#isPublicPath`  

---

## 5. 相关文件索引（Demo 标 ★）

| 文件 | 说明 |
|------|------|
| `pom.xml` | springdoc / swagger-annotations 版本管理 |
| `yi-common/pom.xml` | 注解依赖（Demo 间接使用） |
| `yi-common/.../ApiDocConstants.java` | token 常量 |
| `yi-common/.../BaseCrudController.java` | ★ Demo 五个接口的 `@Operation` |
| ★ `yi-demo/.../DemoArticleController.java` | 案例 Controller |
| ★ `yi-demo/src/main/dto/DemoArticle.dto` | 案例 DTO / Specification |
| ★ `yi-demo/.../DemoArticle.java` | 案例实体 |
| `yi-admin/pom.xml` | springdoc UI + 依赖 yi-demo |
| `yi-admin/.../OpenApiConfigure.java` | OpenAPI Bean 与 satoken |
| `yi-admin/.../application.yaml` | springdoc 配置 |
| `yi-admin/.../SaTokenConfigure.java` | 文档路径放行 |

---

## 6. 常见问题（结合 Demo）

**Q：Swagger UI 里看不到 `/demo/articles`？**  
A：确认 `yi-admin` 依赖了 `yi-demo`；`springdoc.packages-to-scan` 为 `com.star`；`DemoArticleController` 带 `@RestController`。

**Q：Demo 接口一直 401？**  
A：`/demo/articles` 未放行登录，须先 `/auth/login` 再 Authorize；头名必须是 `satoken`。

**Q：`yi-demo` 编译报找不到 `@Tag`？**  
A：确认 `yi-demo` 依赖 `yi-common`，并执行根目录 `mvn compile`。

**Q：Swagger UI 页面本身 401？**  
A：检查 `SaTokenConfigure#excludePaths` 是否包含 `/swagger-ui/**`、`/v3/api-docs/**`。

**Q：列表查询参数从哪来？**  
A：Demo 的 `DemoArticleSpecification` 定义了 `articleCode`、`articleTitle` 等，对应 `GET /demo/articles` 的 query 参数，springdoc 会自动展示。
