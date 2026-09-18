# {projectName} 项目结构与架构评审（yi-admin / yi-common）

## 1. 当前项目结构（基于代码扫描）

### 1.1 Maven 多模块

- **父工程**：`{projectArtifactId}`（`packaging=pom`）
- **子模块**：
  - **`yi-admin`**：Spring Boot Web 应用（打包 `jar`，启动类 `com.star.AppLication`）
  - **`yi-common`**：公共代码（异常/监听/工具类等）

### 1.2 模块职责与依赖关系

- **`yi-admin`**
  - 负责 Web 入口、鉴权/拦截器配置（Sa-Token）、对外 Controller。
  - 依赖 `yi-common`。
- **`yi-common`**
  - 公共异常体系（`BusinessException`、`GlobalExceptionHandler`）
  - Sa-Token 监听器（`MySaTokenListener`）
  - 工具类（`MD5Utils`）

### 1.3 包结构（目前的落地情况）

- `yi-admin`
  - `com.star.AppLication`
  - `com.star.config.satoken.*`（Sa-Token 拦截器、Servlet Filter、Thymeleaf 变量注入）
  - `com.star.admin.controller.auth.AuthController`
  - `com.star.admin.controller.test.TestController`（空）
- `yi-common`
  - `com.star.common.exception.BusinessException`
  - `com.star.common.globalexception.GlobalExceptionHandler`
  - `com.star.common.satoken.MySaTokenListener`
  - `com.star.utils.MD5Utils`

### 1.4 配置与基础设施

- **配置文件**：`yi-admin/src/main/resources/application.yaml`
- **端口**：`8500`
- **模板/静态资源**：配置里声明了 Thymeleaf 与 `static-locations`，但 `resources/` 下当前未看到 `templates/` 与 `static/` 目录落地。
- **数据库**：SQLite（`jdbc:sqlite:D:/.../star.db`，绝对路径） + Druid 连接池
- **Redis**：配置存在（`password: 123456`）
- **ORM**：父工程引入 `jimmer-spring-boot-starter`，并在 yaml 中配置了 `SQLiteDialect`
- **鉴权**：Sa-Token + sa-token-jwt（Simple 模式 `StpLogicJwtForSimple`）

---

## 2. 架构理解（当前形态）

目前项目更接近“**脚手架/骨架阶段**”：

- Web 应用入口在 `yi-admin`，已接入 Sa-Token，并通过 `SaInterceptor` 做全局登录校验。
- 异常处理在 `yi-common`，通过 `@RestControllerAdvice` 统一返回 `SaResult`。
- 数据层/领域层尚未形成（未看到 entity/model、repository/dao、service 的落地）。

整体调用链（当前）：

- HTTP 请求 -> Sa-Token 全局 Filter（CORS + 日志 + 异常处理）
- -> Spring MVC -> SaInterceptor（按路由匹配做 `checkLogin`）
- -> Controller
- -> （未来应进入 Service/Domain/Repository，但目前缺失）

---

## 3. 结构与规范建议（针对现状的“可落地”建议）

### 3.1 目录/分层建议（建议采纳）

在 `yi-admin` 内建议尽快形成稳定分层，避免后期重构成本上升：

- `controller`：只做参数校验、返回封装、调用 service
- `service`：业务编排（事务、跨聚合调用）
- `domain`（可选）：领域对象/领域服务（如果你后续会做复杂业务，建议引入）
- `repository` / `dao`：数据库访问（Jimmer 的 `Repository` 或 `SqlClient` 相关封装）
- `model`：
  - `entity`（或 Jimmer entity 类型定义）
  - `dto`（请求/响应对象）
  - `vo`（视图对象）
- `config`：仅放配置类（Sa-Token、WebMvc、Jackson、Swagger/OpenAPI、Jimmer 等）
- `infrastructure`（可选）：缓存、外部系统、消息队列、对象存储等

`yi-common` 建议拆分为更明确的公共能力包：

- `common-core`（或继续 `yi-common`，但按 package 细分）：
  - `exception`
  - `web`（统一返回体/分页/请求追踪等）
  - `util`
  - `constant`

如果后续规模会增大（后台管理 + 业务模块多），可以演进为：

- `yi-admin`：仅 Web 入口
- `yi-system`：系统模块（用户/角色/权限/菜单/字典/配置）
- `yi-biz-xxx`：具体业务域
- `yi-common`：跨模块公共

### 3.2 Maven 依赖管理建议（建议采纳）

当前父工程 `pom.xml`（packaging=pom）里直接放了大量 `<dependencies>`，这会 **让所有子模块都继承同一套依赖**，短期方便但长期容易出现：

- 模块边界变模糊（common 也会被动带上 web/jdbc 等）
- 依赖膨胀、打包体积变大、冲突更难排查

建议调整为：

- 父工程仅做：
  - 版本管理（`dependencyManagement`）
  - 插件与编译配置
- 每个子模块自行声明自己需要的依赖（`yi-common` 不应默认继承 `spring-boot-starter-web/jdbc` 等）

### 3.3 Spring Boot 版本风险（重点建议）

父工程使用：`spring-boot-starter-parent 4.0.3`。

- Spring Boot 4.x 属于较新的大版本（生态适配、插件、第三方依赖兼容性风险更高）。
- 如果不是“明确要用 4.x 的特性”，建议优先选择 **Spring Boot 3.4.x**（更成熟、资料更多、兼容性更稳定）。

### 3.4 配置与安全建议（重点建议）

当前 `application.yaml` 中存在几个高风险点：

- **SQLite 数据库路径是绝对路径**：`jdbc:sqlite:D:/.../star.db`
  - 建议改为相对路径（如 `jdbc:sqlite:./data/star.db`）
  - 并把 DB 文件放到项目外部目录（不要放在 `resources`，否则打包/运行环境不可控）
- **Redis 密码明文**：`password: 123456`
- **JWT 密钥明文**：`jwt-secret-key: ...`

建议：

- 引入 `application-dev.yaml` / `application-prod.yaml`，敏感配置通过环境变量/启动参数注入
- 本地默认配置可留，但生产必须通过外部化配置覆盖

### 3.5 Sa-Token 配置建议（建议采纳）

`SaTokenConfigure` 同时注册了：

- `SaInterceptor`（WebMvc 拦截器）
- `SaServletFilter`（全局过滤器）

这两者的职责容易交叉，建议明确分工：

- Filter 只做：CORS、统一日志、统一异常兜底（可选）
- 鉴权建议统一在一种机制里完成（更推荐 `SaInterceptor` + 注解鉴权/路由鉴权），避免“同一请求被重复校验/重复处理”。

另外：

- 代码中存在 `System.out.println`：建议替换为统一日志框架（slf4j）
- 目前全局 CORS 允许 `*`：建议在生产环境限制 `Origin` 白名单

### 3.6 密码/加密工具建议（重点建议）

`yi-common` 提供了 `MD5Utils`。

- **MD5 不适合用于密码存储**（可被彩虹表/暴力破解）。
- 建议密码方案：
  - BCrypt（Spring Security 的 `BCryptPasswordEncoder`）
  - 或 Argon2（更强，但需要依赖支持）

MD5 可以用于非安全场景（例如：文件校验、非敏感签名等），但不要用于用户密码。

### 3.7 数据库演进建议（建议采纳）

目前未看到 SQL 迁移/初始化方案。建议尽早引入：

- Flyway 或 Liquibase

这样你后续的“用户/角色/权限”表结构变更会可追踪、可回滚、可在 CI/CD 中自动应用。

---

## 4. 是否需要改结构？结论（给你一个直接建议）

- **现在不需要大改模块数量**（当前功能还在骨架阶段）。
- **需要尽快把分层与依赖边界定下来**：
  - 父 POM 做版本管理
  - 子模块声明依赖
  - `yi-admin` 内形成 controller/service/repository/model 的基本骨架

这样后续你再加“系统模块（RBAC）/业务模块”时，不会出现所有代码都堆在 controller/config 里的情况。

---

## 5. 后续新增功能的建议规范（可直接作为团队约定）

- **新增一个业务能力**（例如“用户管理”）必须包含：
  - Controller（接口层）
  - DTO/VO（入参与出参）
  - Service（业务编排）
  - Repository/DAO（数据访问）
  - 迁移脚本（Flyway/Liquibase）
- **禁止**：Controller 直接写 SQL/直接操作 SqlClient
- **统一返回**：建议自定义统一返回体（不要全项目到处用 `SaResult`，可以在最外层适配 SaResult；或者统一封装一层 `Result<T>`）

