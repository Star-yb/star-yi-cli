# 超级管理员页面扩展机制

Education-Yi 的超级管理员后台（Thymeleaf，`/admin/**`）由 `yi-admin` 统一启动与鉴权。业务模块（`yi-demo`、`apps/ed-*` 等）可在**各自 JAR 内**注册管理页与侧栏菜单，无需修改 `yi-admin` 的 `WebController` 或 `fragments.html`。

---

## 1. 设计目标

| 问题 | 方案 |
|------|------|
| 附件、课程等业务管理页不应全部堆在 `yi-admin/resources` | 模板放在各模块 `resources/admin/` |
| 每加一个业务就要改侧栏 HTML | `AdminPageModule` 插件式注册菜单 |
| 多 JAR 的 Thymeleaf 模板如何加载 | `AdminThymeleafConfigure` 增加 `classpath:/admin/` 解析器 |

采用 **SPI + Spring Bean 自动发现**：各模块实现 `AdminPageModule` 并标注 `@Component`，`yi-admin` 启动时聚合菜单并注入页面 Model。

---

## 2. 架构示意

```
┌─────────────────────────────────────────────────────────────┐
│ yi-admin（启动入口）                                          │
│  AdminMenuRegistry      ← 收集所有 AdminPageModule Bean      │
│  AdminMenuModelAdvice   ← /admin/** 注入 adminMenuCategories │
│  AdminThymeleafConfigure← 解析 classpath:/admin/ 模板         │
│  SysAdminPageModule     ← 内置系统管理菜单                    │
│  fragments.html         ← 侧栏遍历 adminMenuCategories      │
└─────────────────────────────────────────────────────────────┘
         ▲                    ▲                    ▲
         │                    │                    │
   ┌─────┴─────┐      ┌───────┴───────┐    ┌──────┴──────┐
   │  yi-demo  │      │ ed-attachment │    │  ed-course  │  …
   │ PageModule│      │ PageModule    │    │ PageModule  │
   │ Controller│      │ Controller    │    │ Controller  │
   │ admin/…   │      │ admin/…       │    │ admin/…     │
   └───────────┘      └───────────────┘    └─────────────┘
```

**契约层（yi-common）**

| 类 | 说明 |
|----|------|
| `AdminPageModule` | 模块扩展接口：`moduleId()`、`moduleOrder()`、`menuItems()` |
| `AdminMenuItem` | 单条菜单：menuKey、category、label、path、icon、order |
| `AdminMenuCategory` | 按 category 分组后的菜单 |
| `AdminPageAttributes` | 填充 `activeMenu`、`pageCategory`、`pageName` |

**聚合层（yi-admin）**

| 类 | 说明 |
|----|------|
| `AdminMenuRegistry` | 合并、排序、分组所有模块菜单 |
| `AdminMenuModelAdvice` | 为 `/admin/**`（除登录页）注入 `adminMenuCategories` |
| `AdminThymeleafConfigure` | `classpath:/admin/` 模板解析，与 `templates/` 并存 |
| `SysAdminPageModule` | 用户/角色/权限/登录日志/接口文档等内置菜单 |

---

## 3. 接入步骤（三步）

以 `yi-demo` 为例，完整可参考：

- `yi-demo/.../admin/DemoArticleAdminPageModule.java`
- `yi-demo/.../admin/DemoArticleAdminController.java`
- `yi-demo/src/main/resources/admin/yi-demo/index.html`

### 3.1 注册侧栏菜单

```java
@Component
public class DemoArticleAdminPageModule implements AdminPageModule {

    @Override
    public String moduleId() {
        return "yi-demo";
    }

    @Override
    public int moduleOrder() {
        return 90; // 越小越靠前
    }

    @Override
    public List<AdminMenuItem> menuItems() {
        return List.of(new AdminMenuItem(
                "demo-article",           // menuKey，与 activeMenu 一致
                "开发演示",                  // 侧栏分组标题
                "演示文章",                  // 菜单文字
                "/admin/demo/articles",     // 页面 URL
                "bx bx-news",               // BoxIcons 图标类
                10                          // 组内排序
        ));
    }
}
```

### 3.2 页面 Controller

- 使用 `@Controller`（非 `@RestController`）
- 路径以 `/admin/` 开头（走超级管理员角色 `*` 鉴权）
- 返回视图名对应 `admin/` 下模板路径

```java
@Controller
@RequestMapping("/admin/demo/articles")
@RequiredArgsConstructor
public class DemoArticleAdminController {

    private final DemoArticleService demoArticleService;

    @GetMapping
    public String index(PageQuery pageQuery,
                        DemoArticleSpecification specification,
                        Model model) {
        if (pageQuery.getPage() < 1) {
            pageQuery.setPage(1);
        }
        AdminPageAttributes.enrich(model, "demo-article", "开发演示", "演示文章");
        model.addAttribute("articlesPageResult",
                demoArticleService.page(pageQuery, specification));
        return "yi-demo/index";  // → admin/yi-demo/index.html
    }
}
```

### 3.3 Thymeleaf 模板

放在**本模块**：

```
src/main/resources/admin/<视图目录>/index.html
```

视图名 `yi-demo/index` 解析为 `classpath:/admin/yi-demo/index.html`。

模板须复用 `yi-admin` 公共片段，且 **HTML 骨架必须与 `admin.css` 一致**：

```html
<body>
<aside th:replace="~{common/fragments :: sidebar('demo-article')}"></aside>
<main class="main-wrapper">
    <header th:replace="~{common/fragments :: topbar('演示文章')}"></header>
    <div class="content-area">
        <!-- 页面内容 -->
    </div>
</main>
<div th:replace="~{common/fragments :: common_modals}"></div>
<div th:replace="~{common/fragments :: common_scripts}"></div>
</body>
```

> **注意**：不要使用 `admin-layout`、`main-content` 等自定义外层容器。`body` 为 `display: flex`，直接子元素应为 `aside` + `main.main-wrapper`。参考 `yi-admin/templates/sys/login-log.html`。

常用样式类：`search-card`、`table-card`、`page-header`、`page-title`、`page-subtitle`。

分页统一使用公共片段（含总页数、页码跳转，自动保留筛选参数）：

```html
<th:block th:replace="~{common/fragments :: adminPagination(${articlesPageResult}, '/admin/demo/articles')}"></th:block>
```

> 注意：第一个参数必须是 `${pageResult变量}`，不能写裸变量名，否则 Thymeleaf 会当作字符串字面量。

---

## 4. 约定一览

| 项 | 约定 |
|----|------|
| Maven 依赖 | 业务模块依赖 `yi-common`；`yi-admin` 的 `pom.xml` 依赖该业务模块 |
| 包扫描 | 启动类 `com.star.AppLication` 默认扫描 `com.star.**` |
| 菜单注册 | 实现 `AdminPageModule` + `@Component` |
| 页面路由 | `/admin/...` |
| 模板目录 | `src/main/resources/admin/` |
| 视图名 | 与相对路径一致，如 `ed-attachment/index` |
| 侧栏高亮 | `AdminPageAttributes.enrich(..., activeMenu, ...)` 的 `activeMenu` = `AdminMenuItem.menuKey()` |
| 面包屑分组 | `pageCategory` 显示在顶栏面包屑中间段 |
| 鉴权 | `/admin/**` 需登录且角色 `*`（超级管理员） |

---

## 5. 已接入示例

| 模块 | 菜单 | 路径 | 模板 |
|------|------|------|------|
| yi-admin | 系统管理 / 系统监控 / 开发工具 | `/admin/user` 等 | `templates/sys/*.html` |
| yi-demo | 开发演示 → 演示文章 | `/admin/demo/articles` | `admin/yi-demo/index.html` |
| ed-attachment | 教育资源 → 附件管理 | `/admin/ed/attachments` | `admin/ed-attachment/index.html` |

---

## 6. 新模块检查清单

- [ ] `yi-admin/pom.xml` 已添加业务模块依赖
- [ ] 实现 `AdminPageModule` 并注册 `@Component`
- [ ] `@Controller` 映射 `/admin/...`
- [ ] 模板在 `resources/admin/`，骨架符合 `body > aside + main.main-wrapper`
- [ ] `activeMenu` 与 `menuKey` 一致
- [ ] `mvn compile` 通过
- [ ] 超级管理员登录后侧栏出现新菜单，页面布局正常

---

## 7. 与 REST CRUD 的关系

| 类型 | 路径前缀 | 控制器基类 | 用途 |
|------|----------|------------|------|
| REST API | `/demo/**`、`/ed/**` | `BaseCrudController` | 前后端分离 SPA、接口调用 |
| 管理页面 | `/admin/**` | 手写 `@Controller` | Thymeleaf SSR 超级管理员后台 |

管理页可通过服务端 `Service` 直接查库（如本文示例），也可在页面内用 jQuery 调用 REST（参考 `templates/sys/user.html`）。

---

## 8. 常见问题

**Q：侧栏没有新菜单？**  
确认模块已被 `yi-admin` 依赖、`*AdminPageModule` 带 `@Component`、包名在 `com.star` 下。

**Q：页面布局错乱、内容挤在角落？**  
检查是否使用了错误的布局容器；对照 `login-log.html` 修正骨架。

**Q：模板找不到？**  
确认文件在 `resources/admin/`，视图名与路径一致；执行 `mvn compile` 后重启 `yi-admin`。

**Q：能否一个模块注册多个菜单？**  
可以，`menuItems()` 返回多条 `AdminMenuItem` 即可，对应多个 `@Controller`。
