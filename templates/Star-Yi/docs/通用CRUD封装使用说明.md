# 通用 CRUD 封装使用说明

本文档说明 `yi-common` 模块下 `com.star.crud` 通用封装的设计与用法，并结合 `yi-demo` 模块中的 `DemoArticle` 示例说明如何在新业务中落地一套接口。

---

## 1. 适用范围与前置条件

### 1.1 技术前提

- 实体使用 **Jimmer** 定义（`@Entity` 接口），并继承 `yi-common` 中的 `BaseEntity`（若需要统一审计字段）。
- 输入 / 视图 / 查询规格通过 **`src/main/dto/*.dto`** 定义，由 Jimmer APT 生成 `Input`、`View`、`Specification` 等 Java 类。
- 数据访问层 DAO 继承 **`BaseJimmerDao`**，底层仍继承 Jimmer 的 **`AbstractJavaRepository`**。

### 1.2 模块依赖

- 业务模块需在 `pom.xml` 中依赖 **`yi-common`**（示例：`yi-demo` 已依赖 `yi-common`）。
- 父工程需已配置 **Jimmer APT**（`jimmer-apt`）与 **Lombok** 注解处理器，以便编译期生成 DTO 与 Table 等类。

### 1.3 运行说明

`yi-demo` 为演示模块：若需**独立启动** Spring Boot，需自行补充：

- 带 `@SpringBootApplication` 的启动类（通常放在 `com.star.yidemo` 包下）；
- `application.yaml` / `application.yml`（数据源、Jimmer 方言等，可参考 `yi-admin`）。

仅作为**代码模板参考**时，可直接阅读本文与 `yi-demo` 源码结构。

---

## 2. 包结构总览（`com.star.crud`）

| 路径 | 说明 |
|------|------|
| `com.star.crud.dao.BaseJimmerDao` | DAO 基类：增删改、按 ID 查、Predicate/Specification 的列表与分页查询 |
| `com.star.crud.service.BaseCrudOperations` | 服务层对外操作契约：`create/update/delete/obtain/list/page` |
| `com.star.crud.service.CrudRepository` | 仓储契约，与 `BaseCrudOperations` 对齐，供 `AbstractCrudService` 委托 |
| `com.star.crud.service.BaseCrudService` | 空壳抽象服务，实现 `BaseCrudOperations` |
| `com.star.crud.service.AbstractCrudService` | 模板方法服务：统一调用 `repository()`，并提供 `beforeCreate/afterCreate/...` 等扩展点 |
| `com.star.crud.service.AbstractJimmerCrudRepository` | 基于 `BaseJimmerDao` 的通用仓储实现，用统一 `Fetcher` 完成 CRUD 与列表/分页 |
| `com.star.crud.controller.BaseCrudController` | 通用 REST：`SaResult` 包装，标准 CRUD 路由 |
| `com.star.crud.capability.StatusManageable` | **可选**能力接口（启用/禁用），**未**接入通用 Controller，按业务模块自行使用 |

---

## 3. 分层职责与调用关系

```
HTTP → BaseCrudController
         → DemoArticleService (extends AbstractCrudService)
              → DemoArticleRepository (extends AbstractJimmerCrudRepository)
                   → DemoArticleDao (extends BaseJimmerDao)
                        → JSqlClient / AbstractJavaRepository
```

- **Controller**：可选继承 `BaseCrudController`，只需实现 `service()` 与 `bindId`（更新时把路径 `id` 写入更新 DTO）。
- **Service**：继承 `AbstractCrudService`，实现 `repository()` 返回你的 `CrudRepository` 实现；业务差异通过重写 Hook 完成。
- **Repository**：继承 `AbstractJimmerCrudRepository`，实现 `dao()` 与 `fetcher()`（通常 `XXXView.METADATA.getFetcher()`）。
- **DAO**：继承 `BaseJimmerDao`，构造函数传入 `JSqlClient`、实体 `Class`、对应 `XxxTable.$`。

### 3.1 「Repository 层」是否必须单独存在？

**结论：不是必须。** 在本封装里，真正必须的是 **`CrudRepository` 这一契约**（供 `AbstractCrudService` 委托），而不是「必须新建一个名为 `XxxRepository` 的包和类」。

| 概念 | 说明 |
|------|------|
| `CrudRepository` 接口 | **需要有人实现**，`AbstractCrudService.repository()` 才能工作。 |
| 独立的 `repository` 包 / `XxxRepository.java` | **仅为组织代码的一种方式**，与认证模块里「只有 Service + DAO」的习惯可以兼容。 |

与 **`yi-admin` 认证**（如 `AuthServiceImpl`）的对比：

- 认证流程通常是 **登录校验、Token、会话**，不走「标准 CRUD + View Fetcher + Specification 列表」这一套；实现上往往是 **`AuthService` 直接注入 `UsersDao`**，调用 `queryList` / `save` 等即可。
- 因此**没有** `AuthRepository` **并不代表**后续 CRUD 模块也不能用 Repository；只是说明：**非通用 CRUD 场景不必强行多一层**。

后续业务模块可以按偏好选择：

1. **保留独立 `XxxRepository`（当前 `yi-demo` 做法）**  
   - 文件：`yi-demo/.../repository/DemoArticleRepository.java`  
   - 职责：继承 `AbstractJimmerCrudRepository`，只负责把 **`DemoArticleDao`** 与 **`DemoArticleView.METADATA.getFetcher()`** 绑在一起。  
   - 优点：Service 薄、职责清晰、和「DDD 仓储」命名一致。  
   - 缺点：多一个类文件。

2. **不用独立类，在 `ServiceImpl` 内用匿名类 / 内部类实现 `CrudRepository`**  
   - 仍使用 `AbstractCrudService` + `AbstractJimmerCrudRepository` 的逻辑，但 **不单独建** `DemoArticleRepository.java`。  
   - `repository()` 返回 `new AbstractJimmerCrudRepository<...>() { ... }`，在匿名类里实现 `dao()` 与 `fetcher()`（与 `DemoArticleRepository` 内容相同，只是挪到 `ServiceImpl` 构造或字段初始化里）。  
   - 优点：对外仍是 **Controller → Service → DAO**，目录里看不到 `repository` 包。  
   - 缺点：`ServiceImpl` 会变长，适合中小型资源。

3. **不用 `AbstractCrudService`，Service 只注入 `XxxDao`（与认证风格一致）**  
   - `XxxServiceImpl` **不继承** `AbstractCrudService`，直接注入 `DemoArticleDao`，在方法里调用 `dao.add` / `dao.queryPage` 等；若需要分页统一结构，可继续调用 `PageUtils.paginate`。  
   - 优点：层数最少，和「认证只认 DAO」一致。  
   - 缺点：失去 `AbstractCrudService` 的 Hook 模板与 `BaseCrudController` 与 `BaseCrudOperations` 的默认对齐，需自己实现接口方法或不用通用 Controller。

4. **`ServiceImpl` 实现 `CrudRepository`（`repository()` 返回 `this`）**  
   - 技术上可行：让 `XxxServiceImpl extends AbstractCrudService implements CrudRepository<...>`，并把 `AbstractJimmerCrudRepository` 里的委托逻辑**复制或抽取**到本类。  
   - 一般不推荐：易与 `AbstractJimmerCrudRepository` 重复；更推荐 **方案 1 或 2**。

**推荐选择：**

- 标准资源型 CRUD、希望长期维护：**独立 `XxxRepository`（现状）** 或 **内部类继承 `AbstractJimmerCrudRepository`**。  
- 极简模块、与认证一致、几乎无通用分页：**Service + DAO** 即可，**不必**为凑分层而加 Repository。

#### 与 `DemoArticleRepository` 等价的「无独立文件」示例思路（伪代码）

下面与 `DemoArticleRepository.java` 逻辑一致，仅将仓储实现收进 `ServiceImpl`（示意，非仓库内真实文件）：

```java
// DemoArticleServiceImpl 内字段或构造中：
private final CrudRepository<DemoArticle, Long, DemoArticleCreateInput,
        DemoArticleUpdateInput, DemoArticleSpecification> crudRepository =
        new AbstractJimmerCrudRepository<
                DemoArticle, Long, DemoArticleCreateInput, DemoArticleUpdateInput,
                DemoArticleSpecification, DemoArticleTable>() {
            @Override
            protected DemoArticleDao dao() {
                return demoArticleDao;
            }
            @Override
            protected Fetcher<DemoArticle> fetcher() {
                return DemoArticleView.METADATA.getFetcher();
            }
        };

@Override
protected CrudRepository<...> repository() {
    return crudRepository;
}
```

这样**没有** `repository/DemoArticleRepository.java`，仍完整使用通用 CRUD 封装。

---

## 4. `BaseJimmerDao` 能力说明

继承 `AbstractJavaRepository<E, ID>`，并持有本实体对应的 `Table` 常量。

### 4.1 写操作

| 方法 | 说明 |
|------|------|
| `add(I input, Fetcher<E> fetcher)` | `INSERT_IF_ABSENT` |
| `update(I input, Fetcher<E> fetcher)` | `UPDATE_ONLY` |
| `delete(ID id)` | 按主键删除（`DeleteMode.AUTO`） |

### 4.2 读操作

| 方法 | 说明 |
|------|------|
| `getById(ID id)` / `getById(ID id, Fetcher<E> fetcher)` | 委托 `findById` |

### 4.3 查询（与业务 DAO 中「通用查询」风格一致）

| 方法 | 说明 |
|------|------|
| `queryList(Predicate, Function<T, Selection<R>>)` | 非分页，条件为 Predicate |
| `queryList(Specification<E>, Function<T, Selection<R>>)` | 非分页，条件为 Specification |
| `queryPage(Pageable, Predicate, ...)` | 分页 + Predicate |
| `queryPage(Pageable, Specification<E>, ...)` | 分页 + Specification |

子类若有特殊 SQL（关联、批量、原生更新等），在继承 `BaseJimmerDao` 的 DAO 中**追加**方法即可。

---

## 5. `AbstractCrudService` 扩展点

子类实现 `protected abstract CrudRepository<...> repository()`。

可选重写（模板方法 Hook）：

- `beforeCreate` / `afterCreate`
- `beforeUpdate` / `afterUpdate`
- `beforeDelete` / `afterDelete`

用于：校验、清缓存、记日志等，**不包含**通用「状态字段」业务（状态由各实体自行设计）。

---

## 6. `AbstractJimmerCrudRepository` 行为约定

适用于：**Create/Update 均为 Jimmer `Input<E>`**，列表与分页使用 **`Specification<E>`**，且列表/详情/写回使用**同一套** `Fetcher`（来自 View 的 `METADATA`）。

- `create` / `update`：调用 `dao().add` / `dao().update`，并传入 `fetcher()`。
- `obtain`：按 `fetcher` 是否存在决定 `getById` 是否带 Fetcher。
- `list` / `page`：在 `queryList` / `queryPage` 的 selection 中使用 `root.fetch(fetcher())` 组装结果形状。

若某资源需要「列表不要关联、详情要关联」等差异，不要强行用这一套，应在具体 Repository 中拆分或覆写方法。

---

## 7. `BaseCrudController` 路由与约定

子类需标注 `@RestController` 与 `@RequestMapping("/你的前缀")`，并实现：

- `protected abstract BaseCrudOperations<...> service();`
- `protected void bindId(ID id, U input)`：PUT 更新时把路径上的 `id` 设进 body（如 `input.setId(id)`）。

### 7.1 默认映射（与 `DemoArticleController` 一致）

| HTTP | 路径 | 行为 |
|------|------|------|
| POST | `/{prefix}` | 创建 |
| PUT | `/{prefix}/{id}` | 更新（会先 `bindId`） |
| DELETE | `/{prefix}/{id}` | 删除 |
| GET | `/{prefix}/{id}` | 详情 `obtain` |
| GET | `/{prefix}` | 分页/非分页列表，使用 `PageQuery` + `Specification` 查询参数 |

返回统一为 **`SaResult`**（与现有 `yi-admin` 风格一致）。

### 7.2 列表与分页参数

列表入口调用的是 **`service().page(pageQuery, specification)`**，内部使用 `com.star.common.page` 下的：

- **`PageQuery`**：`page`（≤0 表示不分页）、`size`、`sort`、`order` 等；
- **`PageResult`**：统一分页与非分页结构。

前端传参方式与 `yi-admin` 中 `UsersController` 等一致（查询规格由 Jimmer 生成的 `Specification` 类绑定请求参数）。

---

## 8. `yi-demo` 示例：`DemoArticle` 端到端说明

### 8.1 文件一览

| 文件 | 作用 |
|------|------|
| `model/entity/DemoArticle.java` | Jimmer 实体；注释内含建表 SQL |
| `src/main/dto/DemoArticle.dto` | View / CreateInput / UpdateInput / Specification |
| `dao/DemoArticleDao.java` | 继承 `BaseJimmerDao` |
| `repository/DemoArticleRepository.java` | 继承 `AbstractJimmerCrudRepository`，指定 `DemoArticleView` 的 Fetcher（**可改为** §3.1 中的内联实现，非强制单独文件） |
| `service/DemoArticleService.java` | 继承 `BaseCrudOperations` |
| `service/impl/DemoArticleServiceImpl.java` | 继承 `AbstractCrudService` |
| `controller/demo/DemoArticleController.java` | 继承 `BaseCrudController`，前缀 `/demo/articles` |

### 8.2 建表

在 `DemoArticle.java` 的 JavaDoc 中已给出 **MySQL** 示例 DDL，核心字段包括：`article_code`、`article_title`、`content` 及 `BaseEntity` 要求的审计与逻辑删除字段。

### 8.3 DTO 命名与生成类

`DemoArticle.dto` 中导出包为 `com.star.yidemo.model.dto`，编译后应生成（名称以 APT 为准）例如：

- `DemoArticleView` + `METADATA.getFetcher()`
- `DemoArticleCreateInput`
- `DemoArticleUpdateInput`（`dynamic`，支持局部更新）
- `DemoArticleSpecification`

### 8.4 接口前缀

示例 Controller：`@RequestMapping("/demo/articles")`，即：

- 创建：`POST /demo/articles`
- 更新：`PUT /demo/articles/{id}`
- 删除：`DELETE /demo/articles/{id}`
- 详情：`GET /demo/articles/{id}`
- 列表：`GET /demo/articles`（带 `PageQuery` + `DemoArticleSpecification` 参数）

---

## 9. 新业务接入步骤（ checklist ）

1. 新建 Jimmer 实体接口（表名、字段、`@Id` 等）。
2. 在 `src/main/dto/YourEntity.dto` 中定义 `View`、`CreateInput`、`UpdateInput`、`Specification`。
3. 执行 Maven **compile**，确认 `target/generated-sources` 下生成 DTO 与 `XxxTable`。
4. 编写 `XxxDao extends BaseJimmerDao<E, ID, T>`。
5. **实现 `CrudRepository`（二选一或组合）**  
   - **方式 A**：独立类 `XxxRepository extends AbstractJimmerCrudRepository<...>`（与 `DemoArticleRepository` 相同）。  
   - **方式 B**：在 `XxxServiceImpl` 内用匿名类 / 内部类继承 `AbstractJimmerCrudRepository`，见 **§3.1**。  
   - **方式 C**：不使用 `AbstractCrudService` 时，可跳过 `CrudRepository`，`XxxServiceImpl` 直接注入 `XxxDao` 实现业务方法。
6. 编写 `XxxService` + `XxxServiceImpl`：若用通用服务模板，则 `extends AbstractCrudService` 且 `repository()` 返回上一步的 `CrudRepository` 实例。
7. 编写 `XxxController extends BaseCrudController`（可选），实现 `service()` 与 `bindId`。
8. （可选）在 `XxxServiceImpl` 中重写 Hook 处理业务副作用。

---

## 10. 常见问题

### 10.1 为什么 `BaseCrudOperations` 用 `obtain` 而不是 `getById`？

与仓储层命名对齐，避免与 Jimmer `Repository` 自带方法语义混淆；含义相同：按主键查询一条。

### 10.2 能否不用 `BaseCrudController`？

可以。通用层提供基类是**可选**的；业务模块可继续手写 `@RestController`，直接注入 Service 调用即可。

### 10.3 状态启用/禁用要不要走通用 CRUD？

**不要**强行塞进通用 Controller。`StatusManageable` 仅作能力接口参考；是否启用、字段类型、路由风格因业务而异，请在具体模块单独定义接口。

### 10.4 编译报错找不到 `DemoArticleTable` / `DemoArticleView`

请先执行 **`mvn compile`**（或 IDE 触发编译），确保 Jimmer APT 已生成代码；并确认 `*.dto` 中 `export` 的实体与 `package` 正确。

### 10.5 我自己写的认证模块没有 Repository，后面应用是否也必须加？

**不必须。** 详见 **§3.1**。

- **认证**：多为会话与鉴权，通常 **Service + DAO** 即可，不必套用 `CrudRepository` / `AbstractJimmerCrudRepository`。
- **标准 CRUD 模块**：若使用 `AbstractCrudService`，则必须有 **`CrudRepository` 的实现**（可以是独立 `XxxRepository` 类，也可以是 `ServiceImpl` 里的匿名内部类）；若不用该模板，则可与认证一样 **只保留 Service + DAO**。

---

## 11. 相关代码路径速查

- 通用封装：`yi-common/src/main/java/com/star/crud/`
- 示例模块：`yi-demo/src/main/java/com/star/yidemo/`
- 示例 DTO：`yi-demo/src/main/dto/DemoArticle.dto`
- **不写独立 Repository 的写法示例（仅文档，不替换默认代码）**：`yi-demo/docs/Service层不使用独立Repository写法示例.md`

---

*文档版本：与当前仓库 `com.star.crud` 与 `yi-demo` 示例代码同步描述；若后续调整基类方法签名，请同步更新本文。*
