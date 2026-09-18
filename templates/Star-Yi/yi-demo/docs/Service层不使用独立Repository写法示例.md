# Service 层不使用独立 Repository 类的写法示例

> **说明**：本文件仅提供**写法参考**，项目默认仍使用 `repository/DemoArticleRepository.java` + `DemoArticleServiceImpl` 注入该类的结构。  
> 若要在业务中采用「无独立 `XxxRepository` 文件」的方式，可将下面逻辑合并进你的 `XxxServiceImpl`，**不要**删除现有 `DemoArticleRepository`，除非你已确认团队规范允许。

---

## 思路

- `AbstractCrudService` 只要求 `repository()` 返回 **`CrudRepository` 的实现**。
- 该实现可以是**独立类**，也可以在 **`ServiceImpl` 构造方法里用匿名类**继承 `AbstractJimmerCrudRepository`，与独立 `DemoArticleRepository` **逻辑等价**。

---

## 示例代码（等价于当前 `DemoArticleRepository` + `DemoArticleServiceImpl`）

```java
@Service
public class DemoArticleServiceImpl extends AbstractCrudService<
        DemoArticle,
        Long,
        DemoArticleCreateInput,
        DemoArticleUpdateInput,
        DemoArticleSpecification> implements DemoArticleService {

    private final DemoArticleDao demoArticleDao;

    private final CrudRepository<
            DemoArticle,
            Long,
            DemoArticleCreateInput,
            DemoArticleUpdateInput,
            DemoArticleSpecification> crudRepository;

    public DemoArticleServiceImpl(DemoArticleDao demoArticleDao) {
        this.demoArticleDao = demoArticleDao;
        this.crudRepository = new AbstractJimmerCrudRepository<
                DemoArticle,
                Long,
                DemoArticleCreateInput,
                DemoArticleUpdateInput,
                DemoArticleSpecification,
                DemoArticleTable>() {
            @Override
            protected DemoArticleDao dao() {
                return DemoArticleServiceImpl.this.demoArticleDao;
            }

            @Override
            protected Fetcher<DemoArticle> fetcher() {
                return DemoArticleView.METADATA.getFetcher();
            }
        };
    }

    @Override
    protected CrudRepository<
            DemoArticle,
            Long,
            DemoArticleCreateInput,
            DemoArticleUpdateInput,
            DemoArticleSpecification> repository() {
        return crudRepository;
    }
}
```

---

## 与当前工程默认结构的对应关系

| 方式 | 说明 |
|------|------|
| **默认（推荐团队可读性）** | `DemoArticleRepository` + `DemoArticleServiceImpl` 注入 `DemoArticleRepository` |
| **本文示例** | 去掉 `DemoArticleRepository` 文件，把匿名类块挪进 `DemoArticleServiceImpl`（二选一，勿重复两套） |

更完整的架构说明见根目录 `docs/通用CRUD封装使用说明.md` §3.1。
